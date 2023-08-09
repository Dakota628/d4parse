package tex

import (
	"errors"
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/Dakota628/d4parse/pkg/d4/util"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/nfnt/resize"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const mapTileSize = 512 // TODO: currently all map tiles are 512 @ level 0

type TileCoord struct {
	Row uint
	Col uint
}

type MapTiles struct {
	Rows                    uint
	Cols                    uint
	TexturePaths            map[TileCoord]string
	ConditionalTexturePaths map[TileCoord]string
}

func NewMapTiles(rows uint, cols uint) *MapTiles {
	return &MapTiles{
		Rows:                    rows,
		Cols:                    cols,
		TexturePaths:            make(map[TileCoord]string),
		ConditionalTexturePaths: make(map[TileCoord]string),
	}
}

func (t *MapTiles) validateIndex(row uint, col uint) {
	switch {
	case row >= t.Rows:
		panic("row too large")
	case col >= t.Cols:
		panic("col too large")
	case row < 0:
		panic("row too small")
	case col < 0:
		panic("col too small")
	}
}

func (t *MapTiles) GetPath(row uint, col uint) (path string, ok bool) {
	t.validateIndex(row, col)
	path, ok = t.TexturePaths[TileCoord{row, col}]
	return
}

func (t *MapTiles) SetPath(row uint, col uint, path string) {
	t.validateIndex(row, col)
	t.TexturePaths[TileCoord{row, col}] = path
}

func (t *MapTiles) Each(cb func(coord TileCoord, path string)) {
	for coord, path := range t.TexturePaths {
		cb(coord, path)
	}
}

func coordsFromName(name string) (TileCoord, error) {
	texSuffix := d4.SnoGroupTexture.Ext()
	if !strings.HasSuffix(name, texSuffix) {
		return TileCoord{}, errors.New("not a texture file name")
	}
	suffixLen := len(texSuffix)

	l := len(name)
	colStr := name[l-suffixLen-2 : l-suffixLen]
	rowStr := name[l-suffixLen-5 : l-suffixLen-3]

	if rowStr[0] == '0' {
		rowStr = rowStr[1:]
	}
	if colStr[0] == '0' {
		colStr = colStr[1:]
	}

	row, err := strconv.Atoi(rowStr)
	if err != nil {
		return TileCoord{}, err
	}
	col, err := strconv.Atoi(colStr)
	if err != nil {
		return TileCoord{}, err
	}

	return TileCoord{
		Row: uint(row),
		Col: uint(col),
	}, nil
}

func FindMapTextures(dataPath string, worldName string) (*MapTiles, int32, error) {
	worldMetaPath := util.MetaPathByName(dataPath, d4.SnoGroupWorld, worldName)
	worldMeta, err := d4.ReadSnoMetaFile(worldMetaPath)
	if err != nil {
		return nil, -1, err
	}
	worldDef, err := d4.GetDefinition[*d4.WorldDefinition](worldMeta)
	if err != nil {
		return nil, -1, err
	}

	// TODO: check hasZoneMap
	mapTiles := NewMapTiles(
		uint(worldDef.TZoneMapParams.Unk_3620f37.Value),
		uint(worldDef.TZoneMapParams.Unk_c60b9b0.Value),
	)

	texMetaGlob := util.BaseFilePattern(
		dataPath,
		util.FileTypeMeta,
		d4.SnoGroupTexture,
		worldName,
		"zmap_",
		"_[0-9][0-9]_[0-9][0-9]",
	)
	texMetaPaths, err := doublestar.FilepathGlob(texMetaGlob, doublestar.WithFilesOnly())
	for _, texMetaPath := range texMetaPaths {
		coords, err := coordsFromName(texMetaPath)
		if err != nil {
			return nil, -1, err
		}
		mapTiles.SetPath(coords.Row, coords.Col, texMetaPath)
	}

	return mapTiles, worldMeta.Id.Value, nil
}

func mapImgRect(coord TileCoord, tileSize int) image.Rectangle {
	x := int(coord.Row) * tileSize
	y := int(coord.Col) * tileSize
	return image.Rect(x, y, x+tileSize, y+tileSize)
}

func zoomLevels(maxTiles int) []int {
	levels := []int{maxTiles}
	for maxTiles%2 == 0 && maxTiles > 0 {
		maxTiles /= 2
		levels = slices.Insert(levels, 0, maxTiles)
	}
	return levels
}

func tileSubImg(img *image.NRGBA, size int, x int, y int) image.Image {
	x *= size
	y *= size
	return img.SubImage(image.Rect(x, y, x+size, y+size))
}

func generateTiles(mapImg *image.NRGBA, outputDir string, level int, tiles int, tileSize int) error {
	subImgSize := mapImg.Bounds().Dx() / tiles

	baseDir := filepath.Join(outputDir, fmt.Sprintf("%d", level))
	err := os.MkdirAll(baseDir, 0655)
	if err != nil {
		return err
	}

	// Create channel of coords
	coords := make(chan [2]int, tiles*tiles)
	for x := 0; x < tiles; x++ {
		for y := 0; y < tiles; y++ {
			coords <- [2]int{x, y}
		}
	}
	close(coords)

	// Process each coord
	errs := util.NewErrors()
	util.DoWorkChan(100, coords, func(coord [2]int) {
		subImg := tileSubImg(mapImg, subImgSize, coord[0], coord[1])
		tile := resize.Resize(uint(tileSize), uint(tileSize), subImg, resize.Lanczos3)

		p := filepath.Join(baseDir, fmt.Sprintf("%d_%d.png", coord[0], coord[1]))
		f, err := os.Create(p)
		if err != nil {
			errs.Add(err)
			return
		}

		err = png.Encode(f, tile)
		if err != nil {
			errs.Add(err)
			return
		}
	})

	return errs.Err()
}

func WriteMapTiles(mapTiles *MapTiles, outputDir string) (err error) {
	mapImg := image.NewNRGBA(image.Rect(0, 0, int(mapTiles.Rows*mapTileSize), int(mapTiles.Cols*mapTileSize)))

	// Load each tile onto one large map image
	mapTiles.Each(func(coord TileCoord, texMetaPath string) {
		var texMeta d4.SnoMeta
		texMeta, err = d4.ReadSnoMetaFile(texMetaPath)
		if err != nil {
			return
		}
		var texDef *d4.TextureDefinition
		texDef, err = d4.GetDefinition[*d4.TextureDefinition](texMeta)
		if err != nil {
			return
		}

		texPayloadPath := util.ChangePathType(texMetaPath, util.FileTypePayload)
		texPaylowPath := util.ChangePathType(texMetaPath, util.FileTypePaylow)

		var texs map[int]image.Image
		texs, err = LoadTexture(texDef, texPayloadPath, texPaylowPath, 0)
		if err != nil {
			return
		}

		img, ok := texs[0] // Only need largest mipmap. We'll do the downscaling.
		if !ok {
			return
		}

		if img.Bounds().Dx() != mapTileSize || img.Bounds().Dy() != mapTileSize {
			slog.Warn("Resized map tile", slog.String("texMetaPath", texMetaPath))
			img = resize.Resize(mapTileSize, mapTileSize, img, resize.Lanczos3)
		}

		draw.Draw(
			mapImg,
			mapImgRect(coord, mapTileSize),
			img,
			image.Pt(0, 0),
			draw.Src,
		)
	})
	if err != nil {
		return
	}

	// Split into tiles for each level
	if mapTiles.Rows != mapTiles.Cols {
		return errors.New("map is not a square") // TODO: we can support rectangle if there's a use case
	}

	for level, tiles := range zoomLevels(int(mapTiles.Rows)) {
		if err = generateTiles(mapImg, outputDir, level, tiles, mapTileSize); err != nil {
			return err
		}
	}

	return
}
