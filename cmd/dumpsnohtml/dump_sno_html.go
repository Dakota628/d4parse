package main

import (
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/Dakota628/d4parse/pkg/d4/html"
	"golang.org/x/exp/slog"
	"os"
	"path/filepath"
	"sync"
)

func main() {
	if len(os.Args) < 3 {
		slog.Error("usage: dumpsnohtml coreTocFile snoMetaFile")
		os.Exit(1)
	}

	tocFilePath := os.Args[1]
	snoMetaPath := os.Args[2]

	toc, err := d4.ReadTocFile(tocFilePath)
	if err != nil {
		slog.Error("failed to read toc file", slog.Any("error", err))
		os.Exit(1)
	}

	snoMeta, err := d4.ReadSnoMetaFile(snoMetaPath)
	if err != nil {
		slog.Error("failed to read sno meta file", slog.Any("error", err))
		os.Exit(1)
	}
	outputPath := filepath.Join("docs", "sno", fmt.Sprintf("%d.html", snoMeta.Id.Value))

	htmlGen := html.NewGenerator(toc, &sync.Map{})
	htmlGen.Add(&snoMeta)
	if err = os.WriteFile(outputPath, []byte(htmlGen.String()), 0666); err != nil {
		slog.Error("failed write html to output file", slog.Any("error", err))
		os.Exit(1)
	}
}
