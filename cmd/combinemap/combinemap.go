package main

import (
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/color"
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

func diffMap() {
	oldF, err := os.Open("samples/old.png")
	if err != nil {
		panic(err)
	}
	old, err := png.Decode(oldF)
	if err != nil {
		panic(err)
	}

	newF, err := os.Open("samples/new.png")
	if err != nil {
		panic(err)
	}
	new_, err := png.Decode(newF)
	if err != nil {
		panic(err)
	}

	diff := image.NewRGBA(old.Bounds())
	red := color.RGBA{
		R: 255,
		G: 0,
		B: 0,
		A: 255,
	}
	for x := 0; x < old.Bounds().Dx(); x++ {
		for y := 0; y < old.Bounds().Dy(); y++ {
			or, og, ob, oa := old.At(x, y).RGBA()
			nr, ng, nb, na := new_.At(x, y).RGBA()

			if or != nr || og != ng || ob != nb || oa != na {
				diff.Set(x, y, red)
			}
		}
	}

	diffF, err := os.Create("samples/diff.png")
	if err != nil {
		panic(err)
	}
	err = png.Encode(diffF, diff)
	if err != nil {
		panic(err)
	}
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
