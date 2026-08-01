package service

import (
	"bytes"
	"image"
	"image/png"
)

func splitImageVerticalHalves(src image.Image) (left image.Image, right image.Image, splitX int) {
	bounds := src.Bounds()
	splitX = bounds.Min.X + bounds.Dx()/2
	leftRect := image.Rect(bounds.Min.X, bounds.Min.Y, splitX, bounds.Max.Y)
	rightRect := image.Rect(splitX, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
	switch typed := src.(type) {
	case interface {
		SubImage(r image.Rectangle) image.Image
	}:
		return typed.SubImage(leftRect), typed.SubImage(rightRect), splitX
	default:
		leftImg := image.NewRGBA(leftRect)
		rightImg := image.NewRGBA(rightRect)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < splitX; x++ {
				leftImg.Set(x, y, src.At(x, y))
			}
			for x := splitX; x < bounds.Max.X; x++ {
				rightImg.Set(x, y, src.At(x, y))
			}
		}
		return leftImg, rightImg, splitX
	}
}

func encodeImagePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}
