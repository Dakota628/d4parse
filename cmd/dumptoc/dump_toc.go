package main

import (
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/davecgh/go-spew/spew"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
	"os"
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
		slog.Error("usage: dumptoc tocFile")
		os.Exit(1)
	}

	tocFilePath := os.Args[1]

	toc, err := d4.ReadTocFile(tocFilePath)
	if err != nil {
		slog.Error("failed to read toc file", slog.Any("error", err))
		os.Exit(1)
	}

	yamlBytes, err := yaml.Marshal(toc.Entries)
	if err != nil {
		slog.Error("failed to marshal toc entries as yaml", slog.Any("error", err))
		os.Exit(1)
	}
	os.Stdout.Write(yamlBytes)
}
