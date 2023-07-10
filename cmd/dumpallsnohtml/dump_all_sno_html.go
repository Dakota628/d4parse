package main

import (
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/Dakota628/d4parse/pkg/d4/html"
	"github.com/alphadose/haxmap"
	"github.com/bmatcuk/doublestar/v4"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	workers = 100
)

var (
	basePrefix = "base"
	metaPrefix = filepath.Join(basePrefix, "meta")
)

func generateHtmlWorker(c chan string, wg *sync.WaitGroup, toc d4.Toc, gbData d4.GbData, outputPath string) {
	defer wg.Done()

	snoPath := filepath.Join(outputPath, "sno")

	if err := os.MkdirAll(snoPath, 0700); err != nil {
		slog.Error("Error creating output dir", slog.Any("error", err), slog.String("snoPath", snoPath))
		return
	}

	for {
		snoMetaFilePath, ok := <-c
		if !ok {
			return
		}

		// Parse sno file
		snoMeta, err := d4.ReadSnoMetaFile(snoMetaFilePath)
		if err != nil {
			slog.Error("Error reading sno meta file", slog.Any("error", err), slog.String("snoMetaFilePath", snoMetaFilePath))
			continue
		}

		// Check GameBalance
		if gbDef, ok := snoMeta.Meta.(*d4.GameBalanceDefinition); ok {
			for _, gbHeader := range d4.GetGbHeader(gbDef) {
				gbData.Set(gbHeader.Gbid.Value, d4.GbInfo{
					SnoId: snoMeta.Id.Value,
					Name:  string(gbHeader.SzName.Value),
				})
			}
		}

		// Generate html
		htmlGen := html.NewGenerator(toc, gbData)
		htmlGen.Add(&snoMeta)

		// Write sno file
		snoHtmlPath := filepath.Join(snoPath, fmt.Sprintf("%d.html", snoMeta.Id.Value))
		snoHtml, err := os.Create(snoHtmlPath)
		if err != nil {
			slog.Error("Error creating html file", err, slog.String("snoHtmlPath", snoHtmlPath))
			continue
		}

		if _, err = snoHtml.WriteString(htmlGen.String()); err != nil {
			slog.Error("Error writing html file", err, slog.String("snoHtmlPath", snoHtmlPath))
			continue
		}
	}
}

func generateHtmlForFiles(toc d4.Toc, gbData d4.GbData, files []string, outputPath string) error {
	// Files arr to channel
	c := make(chan string, len(files))
	for _, file := range files {
		c <- file
	}
	close(c)

	// Start wokers
	wg := &sync.WaitGroup{}

	for i := uint(0); i < workers; i++ {
		wg.Add(1)
		go generateHtmlWorker(c, wg, toc, gbData, outputPath)
	}

	wg.Wait()
	return nil
}

func generateAllHtml(toc d4.Toc, gameDataPath string, outputPath string) error {
	// Make paths
	metaPath := filepath.Join(gameDataPath, metaPrefix)
	metaGlobPath := filepath.Join(metaPath, "**", "*.*")
	gameBalancePath := filepath.Join(metaPath, "GameBalance")

	// Get all data file names
	filesArr, err := doublestar.FilepathGlob(metaGlobPath)
	if err != nil {
		return err
	}

	slices.SortStableFunc(filesArr, func(a, b string) bool {
		return strings.HasPrefix(a, gameBalancePath)
	})

	// Split GameBalance
	var gameBalanceFiles []string

	for i := 0; i < len(filesArr); i++ {
		if !strings.HasPrefix(filesArr[i], gameBalancePath) {
			gameBalanceFiles = filesArr[:i]
			filesArr = filesArr[i:]
			break
		}
	}

	// Parse game balance files first
	gbData := haxmap.New[d4.GbId, d4.GbInfo]()
	if err = generateHtmlForFiles(toc, gbData, gameBalanceFiles, outputPath); err != nil {
		return err
	}
	if err = generateHtmlForFiles(toc, gbData, filesArr, outputPath); err != nil {
		return err
	}

	return nil
}

func main() {
	if len(os.Args) < 3 {
		slog.Error("usage: dumpallsnohtml d4DataPath outputPath")
		os.Exit(1)
	}

	d4DataPath := os.Args[1]
	outputPath := os.Args[2]
	tocFilePath := filepath.Join(d4DataPath, basePrefix, "CoreTOC.dat")

	toc, err := d4.ReadTocFile(tocFilePath)
	if err != nil {
		slog.Error("failed to read toc file", slog.Any("error", err))
		os.Exit(1)
	}

	if err = generateAllHtml(toc, d4DataPath, outputPath); err != nil {
		slog.Error("failed to generate html files", slog.Any("error", err))
		os.Exit(1)
	}
}
