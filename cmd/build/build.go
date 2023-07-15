package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/Dakota628/d4parse/pkg/d4/html"
	"github.com/bmatcuk/doublestar/v4"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	workers = 2000
)

var (
	basePrefix        = "base"
	metaPrefix        = filepath.Join(basePrefix, "meta")
	stringListsPrefix = filepath.Join("enUS_Text", "meta", "StringList")
)

func generateHtmlWorker(c chan string, wg *sync.WaitGroup, toc d4.Toc, gbData *d4.GbData, refs mapset.Set[[2]int32], outputPath string) {
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
			for _, gbHeader := range d4.GetGbidHeaders(gbDef) {
				gbName := d4.TrimNullTerminated(gbHeader.SzName.Value)
				gbData.Store(
					d4.DT_GBID{
						Group: gbDef.EGameBalanceType.Value,
						Value: d4.GbidHash(gbName),
					},
					d4.GbInfo{
						SnoId: snoMeta.Id.Value,
						Name:  gbName,
					},
				)
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

		// Send references to ref file writer
		for _, ref := range snoMeta.GetReferences(gbData) {
			refs.Add([2]int32{
				snoMeta.Id.Value,
				ref,
			})
		}
	}
}

func generateHtmlForFiles(toc d4.Toc, gbData *d4.GbData, refs mapset.Set[[2]int32], files []string, outputPath string) error {
	// Files arr to channel
	c := make(chan string, len(files))
	for _, file := range files {
		c <- file
	}
	close(c)

	// Start workers
	wg := &sync.WaitGroup{}
	for i := uint(0); i < workers; i++ {
		wg.Add(1)
		go generateHtmlWorker(c, wg, toc, gbData, refs, outputPath)
	}

	wg.Wait()
	return nil
}

func generateAllHtml(toc d4.Toc, refs mapset.Set[[2]int32], gameDataPath string, outputPath string) error {
	// Make paths
	metaPath := filepath.Join(gameDataPath, metaPrefix)
	baseMetaGlobPath := filepath.Join(metaPath, "**", "*.*")
	stringListsPath := filepath.Join(gameDataPath, stringListsPrefix)
	stringsMetaGlobPath := filepath.Join(stringListsPath, "**", "*.*")
	gameBalancePath := filepath.Join(metaPath, "GameBalance")

	// Get all data file names
	baseMetaFiles, err := doublestar.FilepathGlob(baseMetaGlobPath)
	if err != nil {
		return err
	}
	stringsMetaFiles, err := doublestar.FilepathGlob(stringsMetaGlobPath)
	if err != nil {
		return err
	}

	slices.SortStableFunc(baseMetaFiles, func(a, b string) bool {
		return strings.HasPrefix(a, gameBalancePath)
	})

	// Split GameBalance
	var gameBalanceFiles []string

	for i := 0; i < len(baseMetaFiles); i++ {
		if !strings.HasPrefix(baseMetaFiles[i], gameBalancePath) {
			gameBalanceFiles = baseMetaFiles[:i]
			baseMetaFiles = baseMetaFiles[i:]
			break
		}
	}

	// Parse game balance files first
	gbData := &sync.Map{}
	if err = generateHtmlForFiles(toc, gbData, refs, gameBalanceFiles, outputPath); err != nil {
		return err
	}
	if err = generateHtmlForFiles(toc, gbData, refs, append(stringsMetaFiles, baseMetaFiles...), outputPath); err != nil {
		return err
	}

	return nil
}

func generateGroupsFile(outputPath string) error {
	namesFilePath := filepath.Join(outputPath, "groups.mpk")

	groupMap := make(map[int32]string, d4.MaxSnoGroups-1+3)
	for i := d4.SnoGroup(-2); i < d4.MaxSnoGroups-1; i++ {
		groupMap[int32(i)] = i.String()
	}

	b, err := msgpack.Marshal(groupMap)
	if err != nil {
		return err
	}

	namesFile, err := os.Create(namesFilePath)
	if err != nil {
		return err
	}

	_, err = namesFile.Write(b)
	return err
}

func generateNamesFile(toc d4.Toc, outputPath string) error {
	namesFilePath := filepath.Join(outputPath, "names.mpk")

	b, err := msgpack.Marshal(toc.Entries)
	if err != nil {
		return err
	}

	namesFile, err := os.Create(namesFilePath)
	if err != nil {
		return err
	}

	_, err = namesFile.Write(b)
	return err
}

func generateRefsBin(refs mapset.Set[[2]int32], outputPath string) error {
	refsFilePath := filepath.Join(outputPath, "refs.bin")

	f, err := os.Create(refsFilePath)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)

	sortedRefs := refs.ToSlice()
	slices.SortStableFunc(sortedRefs, func(a, b [2]int32) bool {
		return a[1] < b[1]
	})

	// Write length
	if err = binary.Write(w, binary.LittleEndian, uint32(len(sortedRefs))); err != nil {
		return err
	}

	// Write refs
	for _, ref := range sortedRefs {
		if err = binary.Write(w, binary.LittleEndian, ref[0]); err != nil {
			return err
		}
		if err = binary.Write(w, binary.LittleEndian, ref[1]); err != nil {
			return err
		}
	}

	if err = w.Flush(); err != nil {
		return err
	}

	return nil
}

func main() {
	if len(os.Args) < 3 {
		slog.Error("usage: build d4DataPath outputPath")
		os.Exit(1)
	}

	d4DataPath := os.Args[1]
	outputPath := os.Args[2]
	tocFilePath := filepath.Join(d4DataPath, basePrefix, "CoreTOC.dat")

	if err := generateGroupsFile(outputPath); err != nil {
		slog.Error("Failed to generate groups files", slog.Any("error", err))
		os.Exit(1)
	}

	toc, err := d4.ReadTocFile(tocFilePath)
	if err != nil {
		slog.Error("Failed to read toc file", slog.Any("error", err))
		os.Exit(1)
	}

	if err = generateNamesFile(toc, outputPath); err != nil {
		slog.Error("Failed to generate names files", slog.Any("error", err))
		os.Exit(1)
	}

	refs := mapset.NewSet[[2]int32]()
	if err = generateAllHtml(toc, refs, d4DataPath, outputPath); err != nil {
		slog.Error("Failed to generate html files", slog.Any("error", err))
		os.Exit(1)
	}

	if err = generateRefsBin(refs, outputPath); err != nil {
		slog.Error("Failed to generate refs bin", slog.Any("error", err))
		os.Exit(1)
	}
}
