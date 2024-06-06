#include <stdlib.h>
#include <webp/encode.h>

// Function to encode raw RGBA pixels to a WebP format
int encodeWebP(uint8_t* data, int width, int height, float quality, uint8_t** output, size_t* output_size) {
    *output_size = WebPEncodeRGBA(data, width, height, width * 4, quality, output);
    return *output != NULL;
}