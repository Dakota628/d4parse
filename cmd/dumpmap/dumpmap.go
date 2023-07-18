package main

import (
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/bmatcuk/doublestar/v4"
	"golang.org/x/exp/slog"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		slog.Error("usage: dumpmap d4DataPath outputPath")
		os.Exit(1)
	}

	d4DataPath := os.Args[1]
	outputPath := os.Args[2]

	texMetaGlobPath := filepath.Join(
		d4DataPath,
		"base", "meta", "Texture", "zmap_Sanctuary_Eastern_Continent_[0-9][0-9]_[0-9][0-9].tex",
	)
	payloadBasePath := filepath.Join(d4DataPath, "base", "payload", "Texture")
	paylowBasePath := filepath.Join(d4DataPath, "base", "paylow", "Texture")

	texMetaFile, err := doublestar.FilepathGlob(texMetaGlobPath)
	if err != nil {
		slog.Error("Failed to glob map texture meta files", slog.Any("error", err))
		os.Exit(1)
	}

	for _, texMetaFilePath := range texMetaFile {
		baseFileName := filepath.Base(texMetaFilePath)
		payloadFilePath := filepath.Join(payloadBasePath, baseFileName)
		paylowFilePath := filepath.Join(paylowBasePath, baseFileName)

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
				slog.Any("payloadFilePay", payloadFilePath),
				slog.Any("paylowFilePath", paylowFilePath),
			)
			os.Exit(1)
		}

		// Load texture
		mipMaps, err := d4.LoadTexture(texDef, payloadFilePath, paylowFilePath)
		if err != nil {
			slog.Error(
				"Failed to load texture",
				slog.Any("error", err),
				slog.String("in", baseFileName),
			)
			os.Exit(1)
		}

		maxLevel := 7
		for level, img := range mipMaps {
			// Construct output file path
			l := len(baseFileName)
			parts := strings.Split(baseFileName[l-9:l-4], "_")
			x, _ := strconv.Atoi(parts[0])
			y, _ := strconv.Atoi(parts[1])
			z := maxLevel - level

			tileFileName := fmt.Sprintf("%d_%d_%d.png", x, y, z)
			outputTilePath := filepath.Join(outputPath, tileFileName)

			// Write texture
			f, err := os.Create(outputTilePath)
			if err != nil {
				slog.Error(
					"Failed to create output file",
					slog.Any("error", err),
					slog.String("in", baseFileName),
					slog.String("out", outputTilePath),
				)
				os.Exit(1)
			}

			if err = png.Encode(f, img); err != nil {
				slog.Error(
					"Failed to encode output PNG",
					slog.Any("error", err),
					slog.String("in", baseFileName),
					slog.String("out", outputTilePath),
				)
				os.Exit(1)
			}

			slog.Info("Wrote tile", slog.String("in", baseFileName), slog.String("out", outputTilePath))
		}
	}
}
