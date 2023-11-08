package main

import (
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/Dakota628/d4parse/pkg/d4/diff"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/slog"
	"os"
	"strings"
)

func groupName(group d4.SnoGroup) string {
	if n := group.String(); n != "Unknown" {
		return n
	}
	return fmt.Sprintf("unk_%d", group)
}

func changedString(c diff.Change) string {
	var ss []string

	if c.NameChanged {
		ss = append(ss, "name")
	}

	if c.MetaChanged {
		ss = append(ss, "meta")
	}

	if c.PayloadChanged {
		ss = append(ss, "payload")
	}

	if c.XMLChanged {
		ss = append(ss, "server")
	}

	return strings.Join(ss, ",")
}

func main() {
	if len(os.Args) < 3 {
		slog.Error("usage: diff oldManifest newManifest")
		os.Exit(1)
	}

	oldPath := os.Args[1]
	newPath := os.Args[2]

	// Load the manifests
	var oldManifest diff.Manifest
	var newManifest diff.Manifest

	oldData, err := os.ReadFile(oldPath)
	if err != nil {
		panic(err)
	}
	if err := msgpack.Unmarshal(oldData, &oldManifest); err != nil {
		panic(err)
	}

	newData, err := os.ReadFile(newPath)
	if err != nil {
		panic(err)
	}
	if err := msgpack.Unmarshal(newData, &newManifest); err != nil {
		panic(err)
	}

	// Do the diff
	d := oldManifest.Diff(&newManifest)

	// Open files
	fAdded, err := os.Create("added.txt")
	if err != nil {
		panic(err)
	}

	fRemoved, err := os.Create("removed.txt")
	if err != nil {
		panic(err)
	}

	fChanged, err := os.Create("changed.txt")
	if err != nil {
		panic(err)
	}

	// Write added
	for _, a := range d.Added {
		if _, err := fmt.Fprintf(fAdded, "[%s] %s\n", groupName(a.Group), a.Name); err != nil {
			panic(err)
		}
	}

	// Write removed
	for _, r := range d.Removed {
		if _, err := fmt.Fprintf(fRemoved, "[%s] %s\n", groupName(r.Group), r.Name); err != nil {
			panic(err)
		}
	}

	// Write changed
	for _, c := range d.Changes {
		if _, err := fmt.Fprintf(
			fChanged,
			"[%s] %s (%s)\n",
			groupName(c.New.Group),
			c.New.Name,
			changedString(c),
		); err != nil {
			panic(err)
		}
	}
}
