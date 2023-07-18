package main

import (
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"golang.org/x/exp/slog"
	"image/png"
	"os"
)

func main() {
	if len(os.Args) < 4 {
		slog.Error("usage: dumptex texDefPath texPayloadPath texPaylowPath outputPrefix")
		os.Exit(1)
	}

	texDefPath := os.Args[1]
	texPayloadPath := os.Args[2]
	texPaylowPath := os.Args[3]
	outputPrefix := os.Args[4]

	snoMeta, err := d4.ReadSnoMetaFile(texDefPath)
	if err != nil {
		slog.Error("Failed to read tex def sno meta file", slog.Any("error", err))
		os.Exit(1)
	}

	texDef, ok := snoMeta.Meta.(*d4.TextureDefinition)
	if !ok {
		slog.Error("Provided texDefPath was not a texture definition sno meta", slog.Any("error", err))
		os.Exit(1)
	}

	mipMaps, err := d4.LoadTexture(texDef, texPayloadPath, texPaylowPath)
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
