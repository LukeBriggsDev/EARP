package main

import (
	"image"
	"image/color"
	"math"
)

func pixelDifference(c1 color.Color, c2 color.Color) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	diffR := math.Abs(float64(int64(r1) - int64(r2)))
	diffG := math.Abs(float64(int64(g1) - int64(g2)))
	diffB := math.Abs(float64(int64(b1) - int64(b2)))
	return (diffR * 299 / 1000) + (diffG * 587 / 1000) + (diffB * 114 / 1000)

}

func ImageDifference(im1 image.Image, im2 image.Image) float64 {
	total := 0.0
	grey := image.NewGray(image.Rect(0, 0, 200, 200))
	for y := im1.Bounds().Min.Y; y < im1.Bounds().Max.Y; y++ {
		for x := im1.Bounds().Min.X; x < im1.Bounds().Max.X; x++ {
			pxlDiff := pixelDifference(im1.At(x, y), im2.At(x, y))
			grey.SetGray(x, y, color.Gray{Y: uint8((pxlDiff / 65535) * 255)})
			total += pxlDiff
		}
	}
	return total
}
