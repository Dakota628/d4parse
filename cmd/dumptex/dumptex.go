package main

import (
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/Dakota628/d4parse/pkg/d4/tex"
	"golang.org/x/exp/slog"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 4 {
		slog.Error("usage: dumptex dataPath texName outputPrefix")
		os.Exit(1)
	}

	dataPath := os.Args[1]
	texName := os.Args[2]
	outputPrefix := os.Args[3]

	texName += ".tex"
	texDefPath := filepath.Join(dataPath, "base", "meta", "Texture", texName)
	texPayloadPath := filepath.Join(dataPath, "base", "payload", "Texture", texName)
	texPaylowPath := filepath.Join(dataPath, "base", "paylow", "Texture", texName)

	toc, err := d4.ReadTocFile(filepath.Join(dataPath, "base", "CoreTOC.dat"))
	if err != nil {
		slog.Error("Failed to read toc file", slog.Any("error", err))
		os.Exit(1)
	}

	snoMeta, err := d4.ReadSnoMetaFile(texDefPath, toc)
	if err != nil {
		slog.Error("Failed to read tex def sno meta file", slog.Any("error", err))
		os.Exit(1)
	}

	texDef, ok := snoMeta.Meta.(*d4.TextureDefinition)
	if !ok {
		slog.Error("Provided texDefPath was not a texture definition sno meta", slog.Any("error", err))
		os.Exit(1)
	}

	mipMaps, err := tex.LoadTexture(texDef, texPayloadPath, texPaylowPath)
	if err != nil {
		slog.Error("Failed to load texture", slog.Any("error", err))
		os.Exit(1)
	}

	for level, img := range mipMaps {
		outputFilePath := fmt.Sprintf("%s-%d.png", outputPrefix, level)

		f, err := os.Create(outputFilePath)
		if err != nil {
			slog.Error("Failed to create output file", slog.Any("error", err))
			os.Exit(1)
		}

		if err = png.Encode(f, img); err != nil {
			slog.Error("Failed to encode output PNG", slog.Any("error", err))
			os.Exit(1)
		}
	}
}
