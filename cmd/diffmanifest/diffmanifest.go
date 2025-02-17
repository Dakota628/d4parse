package main

import (
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/Dakota628/d4parse/pkg/d4/diff"
	"github.com/Dakota628/d4parse/pkg/d4/util"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/slog"
	"hash/crc32"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

const (
	workers = 2000
)

var (
	coreTocPath = filepath.Join("base", "CoreTOC.dat")
)

func main() {
	if len(os.Args) < 2 {
		slog.Error("usage: diff dump")
		os.Exit(1)
	}

	path := os.Args[1]
	tocPath := filepath.Join(path, coreTocPath)
	toc, err := d4.ReadTocFile(tocPath)
	if err != nil {
		slog.Error("failed to read oldToc toc file", slog.Any("error", err))
		os.Exit(1)
	}

	// Get entries
	var snos []diff.Sno

	for group, m := range toc.Entries {
		for id, name := range m {
			snos = append(snos, diff.Sno{
				Group: group,
				Id:    id,
				Name:  name,
			})
		}
	}

	// Create manifest
	var progress atomic.Uint64
	dm := diff.NewManifest()

	util.DoWorkSlice(workers, snos, func(sno diff.Sno) {
		defer func() {
			if i := progress.Add(1); i%1000 == 0 {
				slog.Info("Generating manifest...", slog.Uint64("progress", i))
			}
		}()

		if strings.TrimSpace(sno.Name) == "" {
			return
		}

		metaPath := util.FindLocalizedFile(path, util.FileTypeMeta, sno.Group, sno.Name)

		meta, err := d4.ReadSnoMetaFile(metaPath, toc)
		if err != nil {
			slog.Error("Error reading meta", slog.String("path", metaPath), slog.String("err", err.Error()))
			return
		}

		// Add meta hash
		h := crc32.NewIEEE()
		if err := meta.Hash(h); err != nil {
			panic(err)
		}
		sno.MetaHash = h.Sum32()

		// Add XML hash
		sno.XMLHash = meta.Header.DwXMLHash.Value

		// Add payload hash
		payloadPath := util.ChangePathType(metaPath, util.FileTypePayload)

		if _, err := os.Stat(payloadPath); err == nil {
			payloadData, err := os.ReadFile(payloadPath)
			if err != nil {
				panic(err)
			}
			sno.MetaHash = crc32.ChecksumIEEE(payloadData)
		}

		// Add to manifest
		dm.Add(sno)
	})

	b, err := msgpack.Marshal(dm)
	if err != nil {
		slog.Error("failed to marshal manifest file", slog.Any("error", err))
		os.Exit(1)
	}

	f, err := os.Create("diff-manifest.mpk")
	if err != nil {
		slog.Error("failed to create manifest file", slog.Any("error", err))
		os.Exit(1)
	}

	_, err = f.Write(b)
	if err != nil {
		slog.Error("failed to write manifest file", slog.Any("error", err))
		os.Exit(1)
	}
}
