#pragma once

#define WIFI_SSID "jsph273"
#define WIFI_PASSWORD "halohihalohello"
#define WIFI_RETRIES 20

#define BACKEND_SERVER_HOSTNAME "172.20.10.3"
#define BACKEND_SERVER_PORT 3123
#define BACKEND_SERVER_TIMEOUT_MS 5 * 1000

#define IMAGE_CLASSIFICATION_ENDPOINT "/image-classification"

#define IMAGE_UPLOAD_CHUNK_SIZE 1024

#define DEFAULT_BRIGHTNESS  0
#define DEFAULT_CONTRAST  0
#define DEFAULT_SATURATION  0
#define DEFAULT_EFFECT  "Normal"
#define DEFAULT_WHITE_BALANCE true
#define DEFAULT_WHITE_BALANCE_GAIN true
#define DEFAULT_WHITE_BALANCE_MODE "Auto"
#define DEFAULT_EXPOSURE_CONTROL true
#define DEFAULT_AEC2 true
#define DEFAULT_AE_LEVEL 0
#define DEFAULT_AEC_VALUE 300
#define DEFAULT_GAIN_CONTROL true
#define DEFAULT_AGC_GAIN 0
#define DEFAULT_GAIN_CEILING "2X"
#define DEFAULT_BPC false
#define DEFAULT_WPC true
#define DEFAULT_RAW_GAMMA true
#define DEFAULT_LENC true
#define DEFAULT_HORIZONTAL_MIRROR false
#define DEFAULT_VERTICAL_MIRROR false
#define DEFAULT_DCW true
#define DEFAULT_COLORBAR false

#define DEFAULT_LED_INTENSITY 0

constexpr camera_config_t camera_settings = {
  .pin_pwdn = 32,
  .pin_reset = -1,
  .pin_xclk = 0,
  .pin_sscb_sda = 26,
  .pin_sscb_scl = 27,
  .pin_d7 = 35,
  .pin_d6 = 34,
  .pin_d5 = 39,
  .pin_d4 = 36,
  .pin_d3 = 21,
  .pin_d2 = 19,
  .pin_d1 = 18,
  .pin_d0 = 5,
  .pin_vsync = 25,
  .pin_href = 23,
  .pin_pclk = 22,
  .xclk_freq_hz = 20000000,
  .ledc_timer = LEDC_TIMER_1,
  .ledc_channel = LEDC_CHANNEL_1,
  .pixel_format = PIXFORMAT_JPEG,
  .frame_size = FRAMESIZE_SVGA,
  .jpeg_quality = 12,
  .fb_count = 2,
  .fb_location = CAMERA_FB_IN_PSRAM,
  .grab_mode = CAMERA_GRAB_LATEST,
};
