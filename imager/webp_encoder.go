package imager

/*
#cgo LDFLAGS: -lwebp
#include <stdlib.h>
#include "./webp_encoder.h"
*/
import "C"
import (
	"errors"
	"image"
	"image/draw"
	"unsafe"
)

// EncodeWebP encodes an image.Image to WebP format.
func EncodeWebP(img image.Image, quality float32) ([]byte, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var data []uint8

	// Image should already be in RGBA
	if rgbaImg, ok := img.(*image.RGBA); ok {
		data = rgbaImg.Pix
	} else {
		// Create an RGBA image with the same bounds as the input image
		rgba := image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)
		data = rgba.Pix
	}

	var output *C.uint8_t
	var outputSize C.size_t

	success := C.encodeWebP(
		(*C.uint8_t)(unsafe.Pointer(&data[0])),
		C.int(width),
		C.int(height),
		C.float(quality),
		&output,
		&outputSize,
	)

	if success == 0 {
		return nil, errors.New("WebP encoding of thumbnail failed")
	}

	defer C.free(unsafe.Pointer(output))

	encodedData := C.GoBytes(unsafe.Pointer(output), C.int(outputSize))
	return encodedData, nil
}
