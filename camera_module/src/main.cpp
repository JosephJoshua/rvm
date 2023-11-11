#include <Arduino.h>
#include <MQTT.h>
#include <WiFi.h>
#include <WiFiClientSecure.h>
#include <IotWebConf.h>
#include <IotWebConfTParameter.h>
#include <soc/rtc_cntl_reg.h>
#include <esp_camera.h>

#include <http_utils.h>
#include <camera_effects.h>
#include <camera_wb_modes.h>
#include <camera_gain_ceilings.h>
#include <settings.h>
#include <action.h>

extern "C" uint8_t temprature_sens_read();

Action action_needed = action::None{};
bool is_builtin_led_on = false;

DNSServer dns_server;
WebServer web_server(80);
WiFiClient http_client;

MQTTClient mqtt_client(25000);
WiFiClientSecure mqtt_web_client;

IotWebConf iot_web_conf(THING_NAME, &dns_server, &web_server, INITIAL_AP_PASSWORD, CONFIG_VERSION);

iotwebconf::ParameterGroup backend_server_param_group = iotwebconf::ParameterGroup("backend_server_group", "Backend server");

iotwebconf::TextTParameter<CONFIG_STRING_MAX_LENGTH> backend_server_hostname_param =
    iotwebconf::Builder<iotwebconf::TextTParameter<CONFIG_STRING_MAX_LENGTH>>("backend_server_hostname").
    label("Backend server hostname").
    defaultValue("").
    build();

iotwebconf::IntTParameter<uint16_t> backend_server_port_param =
    iotwebconf::Builder<iotwebconf::IntTParameter<uint16_t>>("backend_server_port").
    label("Backend server port").
    defaultValue(80).
    min(0).
    max(65535).
    build();

iotwebconf::IntTParameter<uint16_t> backend_server_timeout_param =
    iotwebconf::Builder<iotwebconf::IntTParameter<uint16_t>>("backend_server_timeout").
    label("Backend server timeout (ms)").
    defaultValue(5 * 1000).
    min(0).
    build();

iotwebconf::TextTParameter<CONFIG_STRING_MAX_LENGTH> image_classification_endpoint_param =
    iotwebconf::Builder<iotwebconf::TextTParameter<CONFIG_STRING_MAX_LENGTH>>("image_classification_endpoint").
    label("Image classification endpoint").
    defaultValue("/image-classification").
    build();

iotwebconf::ParameterGroup mqtt_param_group = iotwebconf::ParameterGroup("mqtt_group", "MQTT");

iotwebconf::TextTParameter<CONFIG_STRING_MAX_LENGTH> mqtt_host_param =
    iotwebconf::Builder<iotwebconf::TextTParameter<CONFIG_STRING_MAX_LENGTH>>("mqtt_host").
    label("MQTT broker host").
    defaultValue("").
    build();

iotwebconf::IntTParameter<uint16_t> mqtt_port_param =
    iotwebconf::Builder<iotwebconf::IntTParameter<uint16_t>>("mqtt_port").
    label("MQTT broker port").
    defaultValue(8883).
    min(0).
    max(65535).
    build();

iotwebconf::TextTParameter<CONFIG_STRING_MAX_LENGTH> mqtt_username_param =
    iotwebconf::Builder<iotwebconf::TextTParameter<CONFIG_STRING_MAX_LENGTH>>("mqtt_username").
    label("MQTT username").
    defaultValue("").
    build();

iotwebconf::TextTParameter<CONFIG_STRING_MAX_LENGTH> mqtt_password_param =
    iotwebconf::Builder<iotwebconf::TextTParameter<CONFIG_STRING_MAX_LENGTH>>("mqtt_password").
    label("MQTT password").
    defaultValue("").
    build();

// -- IotWebConf method declarations
void initialize_iot_web_conf();
bool validate_config_form(iotwebconf::WebRequestWrapper *wrapper);
void handle_root();

void initialize_camera();
void update_camera_settings();
void classify_image();
camera_fb_t *get_camera_snapshot();

void initialize_mqtt();
bool connect_mqtt();
void handle_mqtt_message(String &topic, String &payload);
bool publish_image_to_mqtt_broker(const String &requestId, int tries);

void setup()
{
    // Disable brownout.
    WRITE_PERI_REG(RTC_CNTL_BROWN_OUT_REG, 0);

    // Built-in LED works in reverse logic (HIGH = off, LOW = on).
    pinMode(LED_BUILTIN, OUTPUT);

#ifdef CORE_DEBUG_LEVEL
    Serial.begin(115200);
    Serial.setDebugOutput(true);
#endif

    log_i("CPU Freq: %d Mhz, %d core(s)", getCpuFrequencyMhz(), ESP.getChipCores());
    log_i("Free heap: %d bytes", ESP.getFreeHeap());
    log_i("SDK version: %s", ESP.getSdkVersion());

    if (psramFound())
    {
        psramInit();
        log_v("PSRAM found and initialized");
    }

    digitalWrite(LED_BUILTIN, LOW);
    is_builtin_led_on = true;

    initialize_iot_web_conf();
    initialize_camera();

#ifdef DATA_CAPTURE_MODE
    initialize_mqtt();
#endif
}

void loop()
{
    iot_web_conf.doLoop();
    mqtt_client.loop();

    if (iot_web_conf.getState() != iotwebconf::OnLine)
    {
        return;
    }

#ifdef DATA_CAPTURE_MODE
    if (!mqtt_client.connected())
    {
        connect_mqtt();
        return;
    }
#endif

    if (is_builtin_led_on)
    {
        digitalWrite(LED_BUILTIN, HIGH);
        is_builtin_led_on = false;
    }

    std::visit(overloaded{
        [](const action::None &) { },
        [](const action::ClassifyImage &)
        {
            log_i("[Scheduler] Classify image action triggered");
            classify_image();

            action_needed = action::None{};
        },
        [](const action::PublishImageToMQTTBroker &action)
        {
            log_i("[Scheduler] Publish image to MQTT broker action triggered");

            if (publish_image_to_mqtt_broker(action.request_id, action.tries))
            {
                action_needed = action::PublishImageToMQTTBroker{ action.request_id, action.tries + 1 };
            }
            else
            {
                action_needed = action::None{};
            }
        },
    }, action_needed);
}

void classify_image()
{
    if (iot_web_conf.getState() != iotwebconf::NetworkState::OnLine)
    {
        return;
    }

    camera_fb_t *snapshot = get_camera_snapshot();
    if (snapshot == nullptr)
    {
        return;
    }

    log_v("[Image Classification] Attempting to connect to backend server at %s:%d", backend_server_hostname_param.value(), backend_server_port_param.value());

    if (http_client.connect(backend_server_hostname_param.value(), backend_server_port_param.value()))
    {
        log_i("[Image Classification] Uploading snapshot to backend server at %s", image_classification_endpoint_param.value());

        const char *head =
            "--camera_module"
            "\r\n"
            "Content-Disposition: form-data; name=\"image\"; filename=\"esp32-camera.jpg\""
            "\r\n"
            "Content-Type: image/jpeg"
            "\r\n"
            "\r\n";

        const char *tail =
            "\r\n"
            "--camera_module--"
            "\r\n";

        size_t imageLength = snapshot->len;
        size_t extraLength = strlen(head) + strlen(tail);
        size_t totalLength = imageLength + extraLength;

        http_client.printf("POST %s HTTP/1.1\r\n", image_classification_endpoint_param.value());
        http_client.printf("Host: %s\r\n", backend_server_hostname_param.value());
        http_client.println("Content-Length: " + String(totalLength));
        http_client.println("Content-Type: multipart/form-data; boundary=camera_module");
        http_client.println("Accept: text/plain");
        http_client.println("Accept-Charset: ISO-8859-1,utf-8;q=0.7,*;q=0.7");
        http_client.println("User-Agent: ESP32CAM/Camera-Module");
        http_client.println();

        http_client.print(head);

        uint8_t *buffer = snapshot->buf;
        for (size_t i = 0; i < imageLength; i += IMAGE_UPLOAD_CHUNK_SIZE)
        {
            if (i + IMAGE_UPLOAD_CHUNK_SIZE < imageLength)
            {
                http_client.write(buffer, IMAGE_UPLOAD_CHUNK_SIZE);
                buffer += IMAGE_UPLOAD_CHUNK_SIZE;
            }
            else
            {
                http_client.write(buffer, imageLength - i);
            }
        }

        http_client.print(tail);
        esp_camera_fb_return(snapshot);

        String response = read_http_response(&http_client, backend_server_timeout_param.value());
        Serial.println(response.c_str());
        log_i("[Image Classification] Response from server: %s", response.c_str());
    }
    else
    {
        log_e("[Image Classification] Failed to connect to backend server at %s:%s", backend_server_hostname_param.value(), backend_server_port_param.value());
        esp_camera_fb_return(snapshot);
    }
}

void initialize_camera()
{
    log_v("[Camera] Initializing..");
    esp_err_t err = esp_camera_init(&camera_settings);

    if (err != ESP_OK)
    {
        log_e("[Camera] Initialization failed with error code 0x%0x", err);
        delay(1000);
        ESP.restart();
    }

    update_camera_settings();

    log_v("[Camera] Ready.");
}

void update_camera_settings()
{
    log_v("[Camera] Updating settings..");

    auto camera = esp_camera_sensor_get();
    if (camera == nullptr)
    {
        log_e("[Camera] Failed to get camera sensor instance");
        return;
    }

    camera->set_brightness(camera, DEFAULT_BRIGHTNESS);
    camera->set_contrast(camera, DEFAULT_CONTRAST);
    camera->set_saturation(camera, DEFAULT_SATURATION);
    camera->set_special_effect(camera, get_camera_effect_code(DEFAULT_EFFECT));
    camera->set_whitebal(camera, DEFAULT_WHITE_BALANCE);
    camera->set_awb_gain(camera, DEFAULT_WHITE_BALANCE_GAIN);
    camera->set_wb_mode(camera, get_camera_wb_mode_code(DEFAULT_WHITE_BALANCE_MODE));
    camera->set_exposure_ctrl(camera, DEFAULT_EXPOSURE_CONTROL);
    camera->set_aec2(camera, DEFAULT_AEC2);
    camera->set_ae_level(camera, DEFAULT_AE_LEVEL);
    camera->set_aec_value(camera, DEFAULT_AEC_VALUE);
    camera->set_gain_ctrl(camera, DEFAULT_GAIN_CONTROL);
    camera->set_agc_gain(camera, DEFAULT_AGC_GAIN);
    camera->set_gainceiling(camera, get_camera_gain_ceiling_code(DEFAULT_GAIN_CEILING));
    camera->set_bpc(camera, DEFAULT_BPC);
    camera->set_wpc(camera, DEFAULT_WPC);
    camera->set_raw_gma(camera, DEFAULT_RAW_GAMMA);
    camera->set_lenc(camera, DEFAULT_LENC);
    camera->set_hmirror(camera, DEFAULT_HORIZONTAL_MIRROR);
    camera->set_vflip(camera, DEFAULT_VERTICAL_MIRROR);
    camera->set_dcw(camera, DEFAULT_DCW);
    camera->set_colorbar(camera, DEFAULT_COLORBAR);

    log_i("[Camera] Settings updated.");
}

camera_fb_t *get_camera_snapshot()
{
    log_v("[Camera] Getting snapshot..");

    camera_fb_t *fb = nullptr;

    fb = esp_camera_fb_get();
    if (fb == nullptr)
    {
        log_e("[Camera] Failed to obtain frame buffer from the camera");
        delay(1000);
        ESP.restart();
    }

    log_i("[Camera] Snapshot obtained. Size: %zu bytes", fb->len);
    return fb;
}

/**
 * -- IotWebConf method definitions
 */

void initialize_iot_web_conf()
{
    log_v("[IotWebConf] Initializing..");

    backend_server_param_group.addItem(&backend_server_hostname_param);
    backend_server_param_group.addItem(&backend_server_port_param);
    backend_server_param_group.addItem(&backend_server_timeout_param);
    backend_server_param_group.addItem(&image_classification_endpoint_param);

    mqtt_param_group.addItem(&mqtt_host_param);
    mqtt_param_group.addItem(&mqtt_port_param);
    mqtt_param_group.addItem(&mqtt_username_param);
    mqtt_param_group.addItem(&mqtt_password_param);

    iot_web_conf.addParameterGroup(&backend_server_param_group);
    iot_web_conf.addParameterGroup(&mqtt_param_group);

    iot_web_conf.setFormValidator(&validate_config_form);
    iot_web_conf.getApTimeoutParameter()->visible = true;

    iot_web_conf.init();

    log_v("[IotWebConf] Setting up web server..");

    web_server.on("/", handle_root);
    web_server.on("/config", []{ iot_web_conf.handleConfig(); });
    web_server.onNotFound([]{ iot_web_conf.handleNotFound(); });

    log_i("[IotWebConf] Ready. Wi-Fi SSID: %s", iot_web_conf.getThingName());
}

bool validate_config_form(iotwebconf::WebRequestWrapper *wrapper)
{
    log_v("[IotWebConf] Validating config form..");

    bool valid = true;

    String backend_server_hostname = wrapper->arg(backend_server_hostname_param.getId());
    String backend_server_port = wrapper->arg(backend_server_port_param.getId());
    String backend_server_timeout = wrapper->arg(backend_server_timeout_param.getId());
    String image_classification_endpoint = wrapper->arg(image_classification_endpoint_param.getId());

    backend_server_hostname.trim();
    image_classification_endpoint.trim();

    if (backend_server_hostname.length() == 0)
    {
        backend_server_hostname_param.errorMessage = "Backend server hostname is required";
        valid = false;
    }

    if (backend_server_port.length() == 0)
    {
        backend_server_port_param.errorMessage = "Backend server port is required";
        valid = false;
    }

    if (backend_server_timeout.length() == 0)
    {
        backend_server_timeout_param.errorMessage = "Backend server timeout is required";
        valid = false;
    }

    if (image_classification_endpoint.length() == 0)
    {
        image_classification_endpoint_param.errorMessage = "Image classification endpoint is required";
        valid = false;
    }

    log_i("[IotWebConf] Config form validation %s.", valid ? "passed" : "failed");
    return valid;
}

void handle_root()
{
    log_i("[WebServer] Handle /");

    if (iot_web_conf.handleCaptivePortal())
    {
        return;
    }

    String s = "<!DOCTYPE html><html lang=\"en\"><head><meta name=\"viewport\" content=\"width=device-width, initial-scale=1, user-scalable=no\"/>";

    s +=
        "<title>ESP32CAM/Camera Module Configuration</title></head>"
        "<body>"
        "<ul>"
        "<li>Backend server hostname: ";

    s += backend_server_hostname_param.value();

    s +=
        "</li>"
        "<li>Backend server port: ";

    s += String(backend_server_port_param.value());

    s +=
        "</li>"
        "<li>Backend server timeout (ms): ";

    s += String(backend_server_timeout_param.value());

    s +=
        "</li>"
        "<li>Image classification endpoint: ";

    s += image_classification_endpoint_param.value();

    s +=
        "</li>"
        "<li>MQTT broker host: ";

    s += mqtt_host_param.value();

    s +=
        "</li>"
        "<li>MQTT broker port: ";

    s += mqtt_port_param.value();

    s +=
        "</li>"
        "<li>MQTT username port: ";

    s += mqtt_username_param.value();

    s +=
        "</li>"
        "</ul>"
        "Go <a href='/config'>here</a> to change values."
        "</body>"
        "</html>"
        "\n";

    web_server.send(200, "text/html", s.c_str());
}

/**
 * -- Data capture mode only.
 */

const char *hivemq_root_ca = \
    "-----BEGIN CERTIFICATE-----\n" \
    "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\n" \
    "-----END CERTIFICATE-----\n";

void initialize_mqtt()
{
    log_v("[MQTT] Initializing..");

    mqtt_web_client.setCACert(hivemq_root_ca);

    mqtt_client.begin(mqtt_host_param.value(), mqtt_port_param.value(), mqtt_web_client);
    mqtt_client.onMessage(handle_mqtt_message);
    mqtt_client.setKeepAlive(300);

    log_i("[MQTT] Initialized client");
}

unsigned long last_mqqtt_connection_attempt = 0;

bool connect_mqtt()
{
    if (millis() - last_mqqtt_connection_attempt < MQTT_CONNECTION_ATTEMPT_INTERVAL_MS)
    {
        return false;
    }

    log_v("[MQTT] Attempting to connect to MQTT broker at %s:%d", mqtt_host_param.value(), mqtt_port_param.value());

    if (!mqtt_client.connect(
        THING_NAME,
        mqtt_username_param.value(),
        mqtt_password_param.value()
    ))
    {
        log_i("[MQTT] Failed to connect to MQTT broker at %s:%d. Retrying in %d ms", mqtt_host_param.value(), mqtt_port_param.value(), MQTT_CONNECTION_ATTEMPT_INTERVAL_MS);

        last_mqqtt_connection_attempt = millis();
        return false;
    }

    log_i("[MQTT] Connected to broker at %s:%d", mqtt_host_param.value(), mqtt_port_param.value());

    mqtt_client.subscribe(MQTT_CAPTURE_REQUEST_TOPIC);
    log_i("[MQTT] Subscribed to " MQTT_CAPTURE_REQUEST_TOPIC);

    return true;
}

void handle_mqtt_message(String &topic, String &payload)
{
    log_v("[MQTT] Received message with topic %s and payload %s", topic, payload);

    if (topic == MQTT_CAPTURE_REQUEST_TOPIC)
    {
        log_i("[MQTT] Received capture request");
        action_needed = action::PublishImageToMQTTBroker{ payload, 0 };
    }
}

bool publish_image_to_mqtt_broker(const String &request_id, int tries)
{
    camera_fb_t *snapshot = get_camera_snapshot();
    if (snapshot == nullptr)
    {
        if (tries >= DATA_CAPTURE_MAX_ATTEMPTS)
        {
            log_e("[MQTT] Failed to obtain snapshot. Max attempts reached");
            return false;
        }
        else
        {
            log_e(
                "[MQTT] Failed to obtain snapshot. Retrying (%d/%d)",
                tries + 1,
                DATA_CAPTURE_MAX_ATTEMPTS
            );

            return true;
        }
    }

    String topic = (MQTT_CAPTURE_COMPLETE_TOPIC "/" CAMERA_SIDE "/") + request_id;

    if (!mqtt_client.publish(
        topic.c_str(),
        (char *)snapshot->buf,
        snapshot->len
    ))
    {
        esp_camera_fb_return(snapshot);

        if (tries >= DATA_CAPTURE_MAX_ATTEMPTS)
        {
            log_e("[MQTT] Failed to publish image to broker. Max attempts reached");
            return false;
        }
        else
        {
            log_e(
                "[MQTT] Failed to publish image to broker. Retrying (%d/%d)",
                tries + 1,
                DATA_CAPTURE_MAX_ATTEMPTS
            );

            return true;
        }
    }

    esp_camera_fb_return(snapshot);
    log_i("[MQTT] Image published to broker");

    return false;
}
