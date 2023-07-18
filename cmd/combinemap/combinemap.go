package main

import (
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

func loadPng(x int, y int) image.Image {
	fn := filepath.Join("map", "tiles", fmt.Sprintf("%d_%d_7.png", x, y))
	f, err := os.Open(fn)
	if err != nil {
		panic(err)
	}

	img, err := png.Decode(f)
	if err != nil {
		panic(err)
	}

	return img
}

func rect(x int, y int, tileSize int) image.Rectangle {
	x *= tileSize
	y *= tileSize
	return image.Rect(x, y, x+tileSize, y+tileSize)
}

func main() {
	tileSize := 512
	dst := image.NewNRGBA(image.Rect(0, 0, tileSize*40, tileSize*40))

	for x := 0; x < 40; x++ {
		for y := 0; y < 40; y++ {
			src := loadPng(x, y)
			if src.Bounds().Dx() != tileSize || src.Bounds().Dy() != tileSize {
				src = resize.Resize(uint(tileSize), uint(tileSize), src, resize.Lanczos3)
			}

			draw.Draw(
				dst,
				rect(x, y, tileSize),
				src,
				image.Pt(0, 0),
				draw.Src,
			)
		}
	}

	fn := filepath.Join("map", "combined.png")
	f, err := os.Create(fn)
	if err != nil {
		panic(err)
	}

	err = png.Encode(f, dst)
	if err != nil {
		panic(err)
	}
}
