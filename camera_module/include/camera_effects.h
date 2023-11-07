#pragma once

typedef struct {
    int code;
    const char *name;
} camera_effect;

const camera_effect camera_effects[] {
    { 0, "Normal" },
    { 1, "Negative " },
    { 2, "Grayscale" },
    { 3, "Red Tint" },
    { 4, "Green Tint" },
    { 5, "Blue Tint" },
    { 6, "Sepia" },
};

int get_camera_effect_code(const char *name)
{
    for (auto effect : camera_effects)
    {
        if (effect.name == name)
        {
            return effect.code;
        }
    }

    return 0;
}
