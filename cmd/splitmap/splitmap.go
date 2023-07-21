package main

import (
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

type SplittableImage interface {
	image.Image
	SubImage(r image.Rectangle) image.Image
}

const (
	tileSize = 512
)

var (
	mapDir    = filepath.Join("docs", "map")
	outputDir = filepath.Join(mapDir, "maptiles")
	levels    = []int{
		5,
		10,
		20,
		40,
	}
)

func tileSubImg(combined SplittableImage, size int, x int, y int) image.Image {
	x *= size
	y *= size
	return combined.SubImage(image.Rect(
		x, y, x+size, y+size,
	))
}

func generateTiles(combined SplittableImage, level int, tiles int, tileSize int) {
	subImgSize := combined.Bounds().Dx() / tiles

	baseDir := filepath.Join(outputDir, fmt.Sprintf("%d", level))
	err := os.MkdirAll(baseDir, 0655)
	if err != nil {
		panic(err)
	}

	for x := 0; x < tiles; x++ {
		for y := 0; y < tiles; y++ {
			subImg := tileSubImg(combined, subImgSize, x, y)
			tile := resize.Resize(uint(tileSize), uint(tileSize), subImg, resize.Lanczos3)

			p := filepath.Join(baseDir, fmt.Sprintf("%d_%d.png", x, y))
			f, err := os.Create(p)
			if err != nil {
				panic(err)
			}

			err = png.Encode(f, tile)
			if err != nil {
				panic(err)
			}
		}
	}
}

func main() {
	f, err := os.Open(filepath.Join(mapDir, "combined.png"))
	if err != nil {
		panic(err)
	}
	img, err := png.Decode(f)
	if err != nil {
		panic(err)
	}
	combined := img.(SplittableImage)

	for level, tiles := range levels {
		generateTiles(combined, level, tiles, tileSize)
	}
}
