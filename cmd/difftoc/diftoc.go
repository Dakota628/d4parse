package main

import (
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	mapset "github.com/deckarep/golang-set/v2"
	"golang.org/x/exp/slices"
	"golang.org/x/exp/slog"
	"os"
)

func groupName(group d4.SnoGroup) string {
	if n := group.String(); n != "Unknown" {
		return n
	}
	return fmt.Sprintf("unk_%d", group)
}

func main() {
	if len(os.Args) < 3 {
		slog.Error("usage: difftoc old new")
		os.Exit(1)
	}

	oldPath := os.Args[1]
	newPath := os.Args[2]

	old, err := d4.ReadTocFile(oldPath)
	if err != nil {
		slog.Error("failed to read old toc file", slog.Any("error", err))
		os.Exit(1)
	}

	new_, err := d4.ReadTocFile(newPath)
	if err != nil {
		slog.Error("failed to read new toc file", slog.Any("error", err))
		os.Exit(1)
	}

	oldSet := mapset.NewSet[string]()
	newSet := mapset.NewSet[string]()

	for group, m := range old.Entries {
		for id, name := range m {
			oldSet.Add(fmt.Sprintf("%d %s/%s", id, groupName(group), name))
		}
	}

	for group, m := range new_.Entries {
		for id, name := range m {
			newSet.Add(fmt.Sprintf("%d %s/%s", id, groupName(group), name))
		}
	}

	fAdded, err := os.Create("samples/added.txt")
	if err != nil {
		panic(err)
	}

	fRemoved, err := os.Create("samples/removed.txt")
	if err != nil {
		panic(err)
	}

	added := newSet.Difference(oldSet)
	removed := oldSet.Difference(newSet)
	addedS := added.ToSlice()
	removedS := removed.ToSlice()
	slices.Sort(addedS)
	slices.Sort(removedS)

	fmt.Fprintf(fAdded, "ADDED:\n")
	for _, a := range addedS {
		if _, err := fmt.Fprintf(fAdded, "%s\n", a); err != nil {
			panic(err)
		}
	}

	fmt.Fprintf(fRemoved, "REMOVED:\n")
	for _, r := range removedS {
		if _, err := fmt.Fprintf(fRemoved, "%s\n", r); err != nil {
			panic(err)
		}
	}
}
