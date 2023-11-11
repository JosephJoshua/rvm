#pragma once

#define DATA_CAPTURE_MODE

/**
 * The side of the object this camera module is capturing.
 * Possible values: "front", "top"
 */
#define CAMERA_SIDE "top"

#define THING_NAME "ESP32CAM-Camera_Module-Top"
#define INITIAL_AP_PASSWORD "123456789"

/**
 * Configuration version number for this device.
 * Note: should be modified when the config structure changes.
 */
#define CONFIG_VERSION "0.2"
#define CONFIG_STRING_MAX_LENGTH 256

#define WIFI_STATUS_PIN LED_BUILTIN

#define IMAGE_UPLOAD_CHUNK_SIZE 1024

#define DATA_CAPTURE_MAX_ATTEMPTS 5
#define MQTT_CONNECTION_ATTEMPT_INTERVAL_MS 1000
#define MQTT_CAPTURE_REQUEST_TOPIC "capture/request"
#define MQTT_CAPTURE_COMPLETE_TOPIC "capture/complete"

#define DEFAULT_BRIGHTNESS    0
#define DEFAULT_CONTRAST      0
#define DEFAULT_SATURATION    0
#define DEFAULT_EFFECT    "Normal"
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
    .frame_size = FRAMESIZE_VGA,
    .jpeg_quality = 12,
    .fb_count = 2,
    .fb_location = CAMERA_FB_IN_PSRAM,
    .grab_mode = CAMERA_GRAB_LATEST,
};
