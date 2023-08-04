package main

import (
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/Dakota628/d4parse/pkg/d4/util"
	mapset "github.com/deckarep/golang-set/v2"
	"golang.org/x/exp/slog"
	"os"
	"path/filepath"
)

var (
	coreTocPath = filepath.Join("base", "CoreTOC.dat")
)

type Sno struct {
	Group d4.SnoGroup
	Id    int32
	Name  string
}

func groupName(group d4.SnoGroup) string {
	if n := group.String(); n != "Unknown" {
		return n
	}
	return fmt.Sprintf("unk_%d", group)
}

func main() {
	if len(os.Args) < 3 {
		slog.Error("usage: diff oldDump newDump")
		os.Exit(1)
	}

	oldPath := os.Args[1]
	newPath := os.Args[2]

	oldTocPath := filepath.Join(oldPath, coreTocPath)
	newTocPath := filepath.Join(newPath, coreTocPath)

	oldToc, err := d4.ReadTocFile(oldTocPath)
	if err != nil {
		slog.Error("failed to read oldToc toc file", slog.Any("error", err))
		os.Exit(1)
	}

	newToc, err := d4.ReadTocFile(newTocPath)
	if err != nil {
		slog.Error("failed to read new toc file", slog.Any("error", err))
		os.Exit(1)
	}

	// Get old entries
	oldSet := mapset.NewThreadUnsafeSet[Sno]()
	for group, m := range oldToc.Entries {
		for id, name := range m {
			oldSet.Add(Sno{
				Group: group,
				Id:    id,
				Name:  name,
			})
		}
	}

	// Get new entries
	newSet := mapset.NewThreadUnsafeSet[Sno]()
	for group, m := range newToc.Entries {
		for id, name := range m {
			newSet.Add(Sno{
				Group: group,
				Id:    id,
				Name:  name,
			})
		}
	}

	// Write changes
	fAdded, err := os.Create("samples/added.txt")
	if err != nil {
		panic(err)
	}

	fRemoved, err := os.Create("samples/removed.txt")
	if err != nil {
		panic(err)
	}

	fChanged, err := os.Create("samples/changed.txt")
	if err != nil {
		panic(err)
	}

	added := newSet.Difference(oldSet)
	removed := oldSet.Difference(newSet)
	common := newSet.Intersect(oldSet)

	// Write added
	added.Each(func(a Sno) bool {
		if _, err := fmt.Fprintf(fAdded, "[%s] %s\n", groupName(a.Group), a.Name); err != nil {
			panic(err)
		}
		return false
	})

	// Write removed
	removed.Each(func(r Sno) bool {
		if _, err := fmt.Fprintf(fRemoved, "[%s] %s\n", groupName(r.Group), r.Name); err != nil {
			panic(err)
		}
		return false
	})

	// Write changed
	common.Each(func(sno Sno) bool {
		oldMetaPath := util.MetaPathByName(oldPath, sno.Group, sno.Name)
		newMetaPath := util.MetaPathByName(newPath, sno.Group, sno.Name)

		oldMeta, err := d4.ReadSnoMetaHeader(oldMetaPath)
		if err != nil {
			if _, err := fmt.Fprintf(fRemoved, "[%s] %s (compare failed)\n", groupName(sno.Group), sno.Name); err != nil {
				panic(err)
			}
			return false
		}
		newMeta, err := d4.ReadSnoMetaHeader(newMetaPath)
		if err != nil {
			if _, err := fmt.Fprintf(fRemoved, "[%s] %s (compare failed)\n", groupName(sno.Group), sno.Name); err != nil {
				panic(err)
			}
			return false
		}

		// Check XML hash
		if oldMeta.DwXMLHash != newMeta.DwXMLHash {
			if _, err := fmt.Fprintf(fChanged, "[%s] %s (meta changed)\n", groupName(sno.Group), sno.Name); err != nil {
				panic(err)
			}
			return false
		}

		//// Check payloads
		//oldPayloadPath := util.ChangePathType(oldMetaPath, util.FileTypePayload)
		//newPayLoadPath := util.ChangePathType(newMetaPath, util.FileTypePayload)
		//
		//oldPayloadExists := true
		//newPayloadExists := true
		//if _, err := os.Stat(oldPayloadPath); err != nil {
		//	oldPayloadExists = false
		//}
		//if _, err := os.Stat(newPayLoadPath); err != nil {
		//	newPayloadExists = false
		//}
		//
		//if oldPayloadExists != newPayloadExists {
		//	var status string
		//	if newPayloadExists {
		//		status = "payload added"
		//	} else {
		//		status = "payload removed"
		//	}
		//	if _, err := fmt.Fprintf(fChanged, "[%s] %s (%s)\n", groupName(sno.Group), sno.Name, status); err != nil {
		//		panic(err)
		//	}
		//	return true
		//}
		//
		//if oldPayloadExists && newPayloadExists {
		//	oldPayloadData, err := os.ReadFile(oldPayloadPath)
		//	if err != nil {
		//		panic(err)
		//	}
		//	newPayloadData, err := os.ReadFile(newPayLoadPath)
		//	if err != nil {
		//		panic(err)
		//	}
		//
		//	if !bytes.Equal(oldPayloadData, newPayloadData) {
		//		if _, err := fmt.Fprintf(fChanged, "[%s] %s (payload changed)\n", groupName(sno.Group), sno.Name); err != nil {
		//			panic(err)
		//		}
		//	}
		//	return true
		//}

		return false
	})
}