#include <stdlib.h>

#include "jxl/decode.h"

void *allocate(size_t size);
void deallocate(void *ptr);

int decode(uint8_t *jxl_in, int jxl_in_size, int config_only, uint32_t *width, uint32_t *height, uint8_t *rgb_out);

__attribute__((export_name("allocate")))
void *allocate(size_t size) {
    return malloc(size);
}

__attribute__((export_name("deallocate")))
void deallocate(void *ptr) {
    free(ptr);
}

__attribute__((export_name("decode")))
int decode(uint8_t *jxl_in, int jxl_in_size, int config_only, uint32_t *width, uint32_t *height, uint8_t *rgb_out) {
    JxlDecoder* decoder = JxlDecoderCreate(NULL);

    if(JXL_DEC_SUCCESS != JxlDecoderSubscribeEvents(decoder, JXL_DEC_BASIC_INFO | JXL_DEC_FULL_IMAGE)) {
        JxlDecoderDestroy(decoder);
        return 0;
    }

    JxlBasicInfo info;
    JxlPixelFormat format = {4, JXL_TYPE_UINT8, JXL_LITTLE_ENDIAN, 0};

    JxlDecoderSetInput(decoder, jxl_in, jxl_in_size);
    JxlDecoderCloseInput(decoder);

    for(;;) {
        JxlDecoderStatus status = JxlDecoderProcessInput(decoder);

        if(status == JXL_DEC_ERROR) {
            JxlDecoderDestroy(decoder);
            return 0;
        } else if (status == JXL_DEC_NEED_MORE_INPUT) {
            JxlDecoderDestroy(decoder);
            return 0;
        } else if (status == JXL_DEC_BASIC_INFO) {
            if(JXL_DEC_SUCCESS != JxlDecoderGetBasicInfo(decoder, &info)) {
                JxlDecoderDestroy(decoder);
                return 0;
            }

            *width = (uint32_t)info.xsize;
            *height = (uint32_t)info.ysize;

            if(config_only) {
                JxlDecoderDestroy(decoder);
                return 1;
            }
        } else if (status == JXL_DEC_NEED_IMAGE_OUT_BUFFER) {
            size_t buf_size;
            if(JXL_DEC_SUCCESS != JxlDecoderImageOutBufferSize(decoder, &format, &buf_size)) {
                JxlDecoderDestroy(decoder);
                return 0;
            }

            if(JXL_DEC_SUCCESS != JxlDecoderSetImageOutBuffer(decoder, &format, rgb_out, buf_size)) {
                JxlDecoderDestroy(decoder);
                return 0;
            }
        } else if (status == JXL_DEC_FULL_IMAGE) {
            // If the image is an animation, more full frames may be decoded.
        } else if (status == JXL_DEC_SUCCESS) {
            JxlDecoderDestroy(decoder);
            return 1;
        }
    }

    JxlDecoderDestroy(decoder);
    return 0;
}
