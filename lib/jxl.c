#include <stdlib.h>
#include <string.h>

#include "jxl/decode.h"

void *allocate(size_t size);
void deallocate(void *ptr);

int decode(uint8_t *jxl_in, int jxl_in_size, int config_only, int decode_all, uint32_t *width, uint32_t *height,
        uint32_t *depth, uint32_t *count, uint8_t *delay, uint8_t *rgb_out);

__attribute__((export_name("allocate")))
void *allocate(size_t size) {
    return malloc(size);
}

__attribute__((export_name("deallocate")))
void deallocate(void *ptr) {
    free(ptr);
}

__attribute__((export_name("decode")))
int decode(uint8_t *jxl_in, int jxl_in_size, int config_only, int decode_all, uint32_t *width, uint32_t *height,
        uint32_t *depth, uint32_t *count, uint8_t *delay, uint8_t *rgb_out) {
    JxlDecoder* decoder = JxlDecoderCreate(NULL);

    if(JXL_DEC_SUCCESS != JxlDecoderSubscribeEvents(decoder, JXL_DEC_BASIC_INFO | JXL_DEC_FRAME | JXL_DEC_FULL_IMAGE)) {
        JxlDecoderDestroy(decoder);
        return 0;
    }

    JxlBasicInfo info;
    JxlFrameHeader header;
    JxlPixelFormat format = {4, JXL_TYPE_UINT8, JXL_NATIVE_ENDIAN, 0};

    JxlDecoderSetInput(decoder, jxl_in, jxl_in_size);
    JxlDecoderCloseInput(decoder);

    uint32_t n = 0;

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
            *depth = (uint32_t)info.bits_per_sample;

            if(info.bits_per_sample == 16) {
                format.data_type = JXL_TYPE_UINT16;
                format.endianness = JXL_BIG_ENDIAN;
            }

            if(config_only && !info.have_animation) {
                *count = 1;

                JxlDecoderDestroy(decoder);
                return 1;
            }
        } else if (status == JXL_DEC_FRAME) {
            if(JXL_DEC_SUCCESS != JxlDecoderGetFrameHeader(decoder, &header)) {
                JxlDecoderDestroy(decoder);
                return 0;
            };

            memcpy(delay + sizeof(uint32_t)*n, &header.duration, sizeof(uint32_t));
        } else if (status == JXL_DEC_NEED_IMAGE_OUT_BUFFER) {
            if(config_only) {
                n++; *count = n;
                JxlDecoderSkipCurrentFrame(decoder);
                continue;
            }

            size_t buf_size;
            if(JXL_DEC_SUCCESS != JxlDecoderImageOutBufferSize(decoder, &format, &buf_size)) {
                JxlDecoderDestroy(decoder);
                return 0;
            }

            if(JXL_DEC_SUCCESS != JxlDecoderSetImageOutBuffer(decoder, &format, rgb_out + buf_size*n, buf_size)) {
                JxlDecoderDestroy(decoder);
                return 0;
            }

            n++; *count = n;
        } else if (status == JXL_DEC_FULL_IMAGE) {
            if(!decode_all || (info.have_animation && header.is_last)) {
                JxlDecoderDestroy(decoder);
                return 1;
            }
        } else if (status == JXL_DEC_SUCCESS) {
            JxlDecoderDestroy(decoder);
            return 1;
        }
    }

    JxlDecoderDestroy(decoder);
    return 0;
}
