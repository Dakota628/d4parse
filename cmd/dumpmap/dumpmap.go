package main

import (
	"github.com/Dakota628/d4parse/pkg/d4/util"
	"golang.org/x/exp/slog"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	if len(os.Args) < 4 {
		slog.Error("usage: dumpmap d4DataPath worldName outputPath")
		os.Exit(1)
	}

	dataPath := os.Args[1]
	worldName := os.Args[2]
	outputPath := os.Args[3]

	// Find the textures
	mapTiles, worldSnoId, err := util.FindMapTextures(dataPath, worldName)
	if err != nil {
		slog.Error("Failed to find map textures", slog.Any("error", err))
		os.Exit(1)
	}

	if mapTiles.Rows == 0 || mapTiles.Cols == 0 || len(mapTiles.TexturePaths) == 0 {
		slog.Info("No map textures for world")
		os.Exit(0)
	}

	// Construct output base path
	tileOutputPath := filepath.Join(outputPath, strconv.Itoa(int(worldSnoId)))

	// Write the tiles
	if err = util.WriteMapTiles(mapTiles, tileOutputPath); err != nil {
		slog.Error("Failed to write map tiles", slog.Any("error", err))
		os.Exit(1)
	}
}
