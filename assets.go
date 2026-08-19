package main

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

//go:embed assets/lofi-girl-squared.jpg
var sidebarImageBytes []byte

func decodeSidebarImage() (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(sidebarImageBytes))
	return img, err
}
