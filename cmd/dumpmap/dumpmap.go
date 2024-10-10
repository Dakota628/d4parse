package main

import (
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/Dakota628/d4parse/pkg/d4/tex"
	"github.com/Dakota628/d4parse/pkg/d4/util"
	"golang.org/x/exp/slog"
	"os"
	"path/filepath"
	"strconv"
)

const (
	workers = 1000
)

var (
	outputBasePath = filepath.Join("data", "maptiles")
)

func main() {
	if len(os.Args) < 2 {
		slog.Error("usage: dumpmap d4DataPath")
		os.Exit(1)
	}

	dataPath := os.Args[1]

	slog.Info("Reading TOC file...")
	toc, err := d4.ReadTocFile(filepath.Join(dataPath, "base", "CoreTOC.dat"))
	if err != nil {
		slog.Error("Failed to read toc file", slog.Any("error", err))
		os.Exit(1)
	}

	util.EachSnoMetaAsync(workers, dataPath, toc, d4.SnoGroupWorld, func(meta d4.SnoMeta) {
		slog.Info("Checking SNO...", slog.Int("id", int(meta.Id.Value)))

		if meta.Meta.(*d4.WorldDefinition).FHasZoneMap.Value != 1 {
			return
		}

		// Get world name
		_, worldName := toc.Entries.GetName(meta.Id.Value, d4.SnoGroupWorld)
		slog.Info("Dumping map...", slog.String("world", worldName))

		// Find the textures
		mapTiles, worldSnoId, err := tex.FindMapTextures(dataPath, worldName)
		if err != nil {
			slog.Error("Failed to find map textures", slog.Any("error", err))
			os.Exit(1)
		}

		if mapTiles.Rows == 0 || mapTiles.Cols == 0 || len(mapTiles.TexturePaths) == 0 {
			slog.Info("No map textures for world")
			os.Exit(0)
		}

		// Construct output base path
		tileOutputPath := filepath.Join(outputBasePath, strconv.Itoa(int(worldSnoId)))

		// Write the tiles
		if err = tex.WriteMapTiles(mapTiles, tileOutputPath); err != nil {
			slog.Error("Failed to write map tiles", slog.Any("error", err))
			os.Exit(1)
		}
	}, func(err error) {
		slog.Error("Failed finding world meta files", slog.Any("error", err))
		os.Exit(1)
	})
}
