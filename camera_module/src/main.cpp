#include <Arduino.h>
#include <WiFi.h>
#include <WebServer.h>
#include <HTTPClient.h>
#include <soc/rtc_cntl_reg.h>
#include <esp_camera.h>

#include <camera_effects.h>
#include <camera_wb_modes.h>
#include <camera_gain_ceilings.h>
#include <settings.h>

extern "C" uint8_t temprature_sens_read();

WiFiClient web_client;

void connect_wifi();
void initialize_camera();
void update_camera_settings();
void upload_snapshot();
camera_fb_t *get_camera_snapshot();

void setup()
{
  // Disable brownout
  WRITE_PERI_REG(RTC_CNTL_BROWN_OUT_REG, 0);

  // LED_BUILTIN (GPIO33) has inverted logic false => LED on
  pinMode(LED_BUILTIN, OUTPUT);
  pinMode(LED_FLASH, OUTPUT);

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

  connect_wifi();
  initialize_camera();
}

unsigned long lastFrame = millis();
void loop()
{
  if (millis() - lastFrame >= 5 * 1000)
  {
    upload_snapshot();
    lastFrame = millis();
  }
}

String read_http_response(WiFiClient *client)
{
  unsigned long startTime = millis();
  while (!client->available())
  {
    if (millis() - startTime > BACKEND_SERVER_TIMEOUT_MS)
    {
      log_e("[HTTP] Response timed out");
      client->stop();

      return "";
    }
  }
  
  String response;
  while (client->available())
  {
    response += client->readStringUntil('\r');
  }
  
  client->stop();
  return response;
}

void upload_snapshot()
{
  camera_fb_t *snapshot = get_camera_snapshot();
  if (snapshot == nullptr)
  {
    return;
  }

  log_v("Attempting to connect to backend server at %s:%d", BACKEND_SERVER_HOSTNAME, BACKEND_SERVER_PORT);

  if (web_client.connect(BACKEND_SERVER_HOSTNAME, BACKEND_SERVER_PORT))
  {
    log_i("Uploading snapshot to backend server at %s", IMAGE_CLASSIFICATION_ENDPOINT);

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

    web_client.println("POST " IMAGE_CLASSIFICATION_ENDPOINT " HTTP/1.1");
    web_client.println("Host: " BACKEND_SERVER_HOSTNAME);
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

    String response = read_http_response(&web_client);
    Serial.println(response.c_str());
    log_i("Response from server: %s", response.c_str());
  }
  else
  {
    log_e("Failed to connect to backend server at %s:%s", BACKEND_SERVER_HOSTNAME, BACKEND_SERVER_PORT);
    esp_camera_fb_return(snapshot);
  }
}

void connect_wifi()
{
  WiFi.mode(WIFI_STA);
  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);
  
  log_v("Attempting to connect to Wi-Fi network: %s", WIFI_SSID);
  
  int tries = 0;
  
  while (WiFi.status() != WL_CONNECTED)
  {
    Serial.print(".");
    tries++;
    
    if (tries > WIFI_RETRIES)
    {
      Serial.println();

      log_e("Failed to connect to Wi-Fi network. Restarting..");
      delay(1000);
      ESP.restart();
      
      return;
    }

    delay(500);
  }
  
  log_i("Connected to Wi-Fi network: %s", WiFi.SSID());
  log_i("Local IP address: %s", WiFi.localIP().toString().c_str());
}

void initialize_camera()
{
  log_v("initialize_camera");
  esp_err_t err = esp_camera_init(&camera_settings);

  if (err != ESP_OK)
  {
    log_e("Camera initialization failed with error code 0x%0x", err);
    delay(1000);
    ESP.restart();
  }
  
  update_camera_settings();
}

void update_camera_settings()
{
  auto camera = esp_camera_sensor_get();
  if (camera == nullptr)
  {
    log_e("Unable to get camera sensor");
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
}

camera_fb_t *get_camera_snapshot()
{
  log_v("get_camera_snapshot");

  camera_fb_t *fb = nullptr;

  // Remove old images stored in the frame buffer
  // int frame_buffers = camera_settings.fb_count;
  // while (frame_buffers--) {
  //   fb = esp_camera_fb_get();
  //   if (fb == nullptr) break;
  //
  //   esp_camera_fb_return(fb);
  //   fb = nullptr;
  // }
  
  fb = esp_camera_fb_get();
  if (fb == nullptr)
  {
    log_e("Unable to obtain frame buffer from the camera");
    delay(1000);
    ESP.restart();
  }

  return fb;
}
