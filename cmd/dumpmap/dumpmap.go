package main

import (
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/bmatcuk/doublestar/v4"
	"golang.org/x/exp/slog"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 3 {
		slog.Error("usage: dumpmap d4DataPath outputPath")
		os.Exit(1)
	}

	d4DataPath := os.Args[1]
	outputPath := os.Args[2]

	texMetaGlobPath := filepath.Join(d4DataPath, "base", "meta", "Texture", "zmap_Sanctuary_Eastern_Continent_*_*.tex")
	texPayBasePath := filepath.Join(d4DataPath, "base", "payload", "Texture")

	texMetaFile, err := doublestar.FilepathGlob(texMetaGlobPath)
	if err != nil {
		slog.Error("Failed to glob map texture meta files", slog.Any("error", err))
		os.Exit(1)
	}

	for _, texMetaFilePath := range texMetaFile {
		baseFileName := filepath.Base(texMetaFilePath)
		texPayFilePath := filepath.Join(texPayBasePath, baseFileName)

		// Read texture definition
		snoMeta, err := d4.ReadSnoMetaFile(texMetaFilePath)
		if err != nil {
			slog.Error("Failed to read tex def sno meta file", slog.Any("error", err))
			os.Exit(1)
		}

		texDef, ok := snoMeta.Meta.(*d4.TextureDefinition)
		if !ok {
			slog.Error(
				"Failed to load texture definition",
				slog.Any("error", err),
				slog.Any("texMetaFilePath", texMetaFilePath),
				slog.Any("texPayFilePath", texPayFilePath),
			)
			os.Exit(1)
		}

		// Load texture
		_, err = d4.LoadTexture(texDef, texPayFilePath, func(img image.Image) {
			// Construct output file path
			l := len(baseFileName)
			outputTilePath := filepath.Join(outputPath, baseFileName[l-9:l-4]+".png")

			// Write texture
			f, err := os.Create(outputTilePath)
			if err != nil {
				slog.Error("Failed to create output file", slog.Any("error", err))
				os.Exit(1)
			}

			if err = png.Encode(f, img); err != nil {
				slog.Error("Failed to encode output PNG", slog.Any("error", err))
				os.Exit(1)
			}

			slog.Info("Wrote tile", slog.String("int", baseFileName), slog.String("out", outputTilePath))
		})
		if err != nil {
			slog.Error("Failed to load texture", slog.Any("error", err))
			os.Exit(1)
		}
	}
}
