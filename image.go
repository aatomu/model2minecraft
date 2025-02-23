package main

import (
	"image"
	"io"
)

type pixel struct {
	c    Color
	x, y float64
}

func parseImage(f io.Reader) (p []pixel) {
	img, _, _ := image.Decode(f)
	bounds := img.Bounds()

	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			r, g, b, _ := img.At(x, y).RGBA()
			p = append(p, pixel{
				c: Color{
					r: uint8(r >> 8),
					g: uint8(g >> 8),
					b: uint8(b >> 8),
				},
				x: float64(x),
				y: float64(bounds.Max.Y - y),
			})
		}
	}

	return
}
