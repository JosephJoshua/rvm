#pragma once

typedef struct {
    int code;
    const char *name;
} camera_wb_mode;

const camera_wb_mode camera_wb_modes[] {
    { 0, "Auto" },
    { 1, "Sunny" },
    { 2, "Cloudy" },
    { 3, "Office" },
    { 4, "Home" },
};

int get_camera_wb_mode_code(const char *name)
{
    for (auto mode : camera_wb_modes)
    {
        if (mode.name == name)
        {
            return mode.code;
        }
    }

    return 0;
}
