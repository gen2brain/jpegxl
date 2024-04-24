#include <stdlib.h>
#include <string.h>

#include "jxl/encode.h"

uint8_t* encode(uint8_t *rgb_in, int width, int height, size_t *size, int quality, int effort);

uint8_t* encode(uint8_t *rgb_in, int width, int height, size_t *size, int quality, int effort) {
    JxlEncoder* encoder = JxlEncoderCreate(NULL);

    JxlEncoderStatus status;
    JxlPixelFormat format = {4, JXL_TYPE_UINT8, JXL_NATIVE_ENDIAN, 0};

    JxlBasicInfo info;
    JxlEncoderInitBasicInfo(&info);
    info.xsize = width;
    info.ysize = height;
    info.bits_per_sample = 8;
    info.alpha_bits = 8;
    info.num_extra_channels = 1;

    if(quality == 100) {
        info.uses_original_profile = JXL_TRUE;
    }

    status = JxlEncoderSetBasicInfo(encoder, &info);
    if(status != JXL_ENC_SUCCESS) {
        JxlEncoderDestroy(encoder);
        return NULL;
    }

    JxlColorEncoding encoding = {};
    JxlColorEncodingSetToSRGB(&encoding, JXL_FALSE);

    status = JxlEncoderSetColorEncoding(encoder, &encoding);
    if(status != JXL_ENC_SUCCESS) {
        JxlEncoderDestroy(encoder);
        return NULL;
    }

    JxlEncoderFrameSettings* settings = JxlEncoderFrameSettingsCreate(encoder, NULL);
    JxlEncoderSetFrameDistance(settings, JxlEncoderDistanceFromQuality(quality));
    JxlEncoderFrameSettingsSetOption(settings, JXL_ENC_FRAME_SETTING_EFFORT, effort);
    if(quality == 100) {
        JxlEncoderSetFrameLossless(settings, JXL_TRUE);
    }

    status = JxlEncoderAddImageFrame(settings, &format, rgb_in, width * height * 4);
    if(status != JXL_ENC_SUCCESS) {
        JxlEncoderDestroy(encoder);
        return NULL;
    }

    JxlEncoderCloseInput(encoder);

    uint8_t* out;
    size_t offset = 0;
    uint8_t* next_out;
    size_t avail_out = 0;

    size_t count = 4096;
    out = (uint8_t*)malloc(4096);

    do {
        next_out = out + offset;
        avail_out = count - offset;

        status = JxlEncoderProcessOutput(encoder, &next_out, &avail_out);
        if(status == JXL_ENC_NEED_MORE_OUTPUT) {
            offset = next_out - out;
            count *= 2;
            out = (uint8_t*)realloc(out, count);
        } else if(status == JXL_ENC_ERROR) {
            JxlEncoderDestroy(encoder);
            return NULL;
        }
    } while(status != JXL_ENC_SUCCESS);

    *size = next_out - out;
    out = (uint8_t*)realloc(out, *size);

    JxlEncoderDestroy(encoder);
    return out;
}
