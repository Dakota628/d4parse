package main

import (
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/Dakota628/d4parse/pkg/d4/mrk"
	"github.com/Dakota628/d4parse/pkg/d4/util"
	"golang.org/x/exp/slog"
	"os"
	"path/filepath"
)

const (
	workers = 100
)

var (
	outputBasePath = filepath.Join("data", "mapdata")
)

func main() {
	if len(os.Args) < 2 {
		slog.Error("usage: dumpmapdata d4DataPath")
		os.Exit(1)
	}

	d4DataPath := os.Args[1]
	tocFilePath := filepath.Join(d4DataPath, "base", "CoreTOC.dat")

	// Read TOC
	slog.Info("Reading TOC file...")
	toc, err := d4.ReadTocFile(tocFilePath)
	if err != nil {
		slog.Error("Failed to read toc file", slog.Any("error", err))
		os.Exit(1)
	}

	// Generate scene data
	util.DoWorkMap(workers, toc.Entries[d4.SnoGroupScene], func(sceneSnoId int32, sceneSnoName string) {
		slog.Info("Generating scene data...", slog.Int("id", int(sceneSnoId)), slog.String("name", sceneSnoName))
		me := mrk.NewMarkerExtractor(d4DataPath, outputBasePath, toc)

		if err := me.AddScene(sceneSnoId); err != nil {
			slog.Error("Failed generate scene data", slog.Any("error", err), slog.Int("id", int(sceneSnoId)), slog.String("name", sceneSnoName))
			os.Exit(1)
		}

		if err := me.Write(); err != nil {
			slog.Error("Failed write scene data", slog.Any("error", err), slog.Int("id", int(sceneSnoId)), slog.String("name", sceneSnoName))
			os.Exit(1)
		}
	})

	// Generate world data
	util.DoWorkMap(workers, toc.Entries[d4.SnoGroupWorld], func(worldSnoId int32, worldSnoName string) {
		slog.Info("Generating world data...", slog.Int("id", int(worldSnoId)), slog.String("name", worldSnoName))
		me := mrk.NewMarkerExtractor(d4DataPath, outputBasePath, toc)

		if err := me.AddWorld(worldSnoId); err != nil {
			slog.Error("Failed generate world data", slog.Any("error", err), slog.Int("id", int(worldSnoId)), slog.String("name", worldSnoName))
			os.Exit(1)
		}

		if err := me.Write(); err != nil {
			slog.Error("Failed write world data", slog.Any("error", err), slog.Int("id", int(worldSnoId)), slog.String("name", worldSnoName))
			os.Exit(1)
		}
	})
}
