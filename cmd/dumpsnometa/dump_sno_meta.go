package main

import (
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/davecgh/go-spew/spew"
	"golang.org/x/exp/slog"
	"os"
	"strings"
)

var spewConfig = spew.ConfigState{
	Indent:                  "  ",
	MaxDepth:                0,
	DisableMethods:          true,
	DisablePointerMethods:   true,
	DisablePointerAddresses: true,
	DisableCapacities:       true,
	ContinueOnMethod:        true,
	SortKeys:                false,
	SpewKeys:                false,
}

func main() {
	if len(os.Args) < 2 {
		slog.Error("usage: dumpsnometa snoMetaFile")
		os.Exit(1)
	}

	snoMetaPath := os.Args[1]

	snoMeta, err := d4.ReadSnoMetaFile(snoMetaPath)
	if err != nil {
		slog.Error("failed to read sno meta file", slog.Any("error", err))
		os.Exit(1)
	}

	dump := spewConfig.Sdump(snoMeta)
	dump = strings.Replace(dump, "github.com/Dakota628/d4parse/pkg/", "", -1)
	dump = strings.Replace(dump, "[*d4.", "[", -1)
	dump = strings.Replace(dump, "(*d4.", "(", -1)
	dump = strings.Replace(dump, "[d4.", "[", -1)
	dump = strings.Replace(dump, "(d4.", "(", -1)
	dump = strings.Replace(dump, "[]*d4.", "[]", -1)
	print(dump)
}
