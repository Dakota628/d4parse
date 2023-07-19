package main

import (
	"github.com/Dakota628/d4parse/pkg/bin"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/dave/jennifer/jen"
	"golang.org/x/exp/slog"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		slog.Error("usage: dumpformathash d4DataPath")
		os.Exit(1)
	}

	d4DataPath := os.Args[1]
	metaPath := filepath.Join(d4DataPath, "base", "meta")

	entries, err := os.ReadDir(metaPath)
	if err != nil {
		panic(err)
	}

	f := jen.NewFile("d4")
	var cases []jen.Code

	for _, e := range entries {
		// Read sno file header
		snoGroup := e.Name()
		snoGroupPath := filepath.Join(metaPath, snoGroup)
		entries, err := os.ReadDir(snoGroupPath)
		if err != nil {
			panic(err)
		}
		if len(entries) == 0 {
			continue
		}
		snoFilePath := filepath.Join(snoGroupPath, entries[0].Name())
		f, err := os.Open(snoFilePath)
		if err != nil {
			panic(err)
		}
		var snoFileHeader d4.SNOFileHeader
		if err := snoFileHeader.UnmarshalD4(bin.NewBinaryReader(f), nil); err != nil {
			panic(err)
		}

		if snoGroup == "137" {
			snoGroup = "Season"
		}

		// Generate case statement
		cases = append(
			cases,
			jen.Case(jen.Lit(int(snoFileHeader.DwFormatHash.Value))).Block(
				jen.Return(jen.Op("&").Id(snoGroup+"Definition").Block()),
			),
		)
	}

	cases = append(
		cases,
		jen.Default().Block(
			jen.Return(jen.Nil()),
		),
	)

	f.Func().Id("NewByFormatHash").Params(
		jen.Id("h").Int(),
	).Id("Object").Block(
		jen.Switch(jen.Id("h")).Block(cases...),
	).Line()

	if err := f.Save("samples/formathash.go"); err != nil {
		panic(err)
	}
}
