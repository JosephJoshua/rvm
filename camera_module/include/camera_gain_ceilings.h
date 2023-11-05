#pragma once

#include <esp_camera.h>

typedef struct {
  gainceiling_t code;
  const char *name;
} camera_gain_ceiling;

const camera_gain_ceiling camera_gain_ceilings[] {
  { GAINCEILING_2X, "2X" },
  { GAINCEILING_4X, "4X" },
  { GAINCEILING_8X, "8X" },
  { GAINCEILING_16X, "16X" },
  { GAINCEILING_32X, "32X" },
  { GAINCEILING_64X, "64X" },
  { GAINCEILING_128X, "128X" },
};

gainceiling_t get_camera_gain_ceiling_code(const char *name)
{
  for (auto ceil : camera_gain_ceilings)
  {
    if (ceil.name == name)
    {
      return ceil.code;
    }
  }
  
  return GAINCEILING_2X;
}
