#include <Arduino.h>
#include <WiFi.h>
#include <IotWebConf.h>
#include <IotWebConfTParameter.h>
#include <soc/rtc_cntl_reg.h>
#include <esp_camera.h>

#include <http_utils.h>
#include <camera_effects.h>
#include <camera_wb_modes.h>
#include <camera_gain_ceilings.h>
#include <settings.h>

extern "C" uint8_t temprature_sens_read();

DNSServer dns_server;
WebServer web_server(80);
WiFiClient web_client;

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

// -- IotWebConf method declarations
void setup_iot_web_conf();
bool validate_config_form(iotwebconf::WebRequestWrapper *wrapper);
void handle_root();

void initialize_camera();
void update_camera_settings();
void upload_snapshot();
camera_fb_t *get_camera_snapshot();

void setup()
{
    // Disable brownout
    WRITE_PERI_REG(RTC_CNTL_BROWN_OUT_REG, 0);

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

    setup_iot_web_conf();
    initialize_camera();
}

unsigned long lastFrame = millis();
void loop()
{
    iot_web_conf.doLoop();

    if (iot_web_conf.getState() != iotwebconf::NetworkState::OnLine)
    {
        return;
    }

    if (millis() - lastFrame >= 5 * 1000)
    {
        upload_snapshot();
        lastFrame = millis();
    }
}

void upload_snapshot()
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

    if (web_client.connect(backend_server_hostname_param.value(), backend_server_port_param.value()))
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

        web_client.printf("POST %s HTTP/1.1\r\n", image_classification_endpoint_param.value());
        web_client.printf("Host: %s\r\n", backend_server_hostname_param.value());
        web_client.println("Content-Length: " + String(totalLength));
        web_client.println("Content-Type: multipart/form-data; boundary=camera_module");
        web_client.println("Accept: text/plain");
        web_client.println("Accept-Charset: ISO-8859-1,utf-8;q=0.7,*;q=0.7");
        web_client.println("User-Agent: ESP32CAM/Camera-Module");
        web_client.println();

        web_client.print(head);

        uint8_t *buffer = snapshot->buf;
        for (size_t i = 0; i < imageLength; i += IMAGE_UPLOAD_CHUNK_SIZE)
        {
            if (i + IMAGE_UPLOAD_CHUNK_SIZE < imageLength)
            {
                web_client.write(buffer, IMAGE_UPLOAD_CHUNK_SIZE);
                buffer += IMAGE_UPLOAD_CHUNK_SIZE;
            }
            else
            {
                web_client.write(buffer, imageLength - i);
            }
        }

        web_client.print(tail);
        esp_camera_fb_return(snapshot);

        String response = read_http_response(&web_client, backend_server_timeout_param.value());
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

void setup_iot_web_conf()
{
    log_v("[IotWebConf] Initializing..");

    backend_server_param_group.addItem(&backend_server_hostname_param);
    backend_server_param_group.addItem(&backend_server_port_param);
    backend_server_param_group.addItem(&backend_server_timeout_param);
    backend_server_param_group.addItem(&image_classification_endpoint_param);

    iot_web_conf.setStatusPin(WIFI_STATUS_PIN);
    iot_web_conf.addParameterGroup(&backend_server_param_group);
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
        "</ul>"
        "Go <a href='/config'>here</a> to change values."
        "</body>"
        "</html>"
        "\n";

    web_server.send(200, "text/html", s.c_str());
}
