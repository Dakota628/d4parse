package diff

import (
	"github.com/Dakota628/d4parse/pkg/d4"
	mapset "github.com/deckarep/golang-set/v2"
	"sync"
)

type Change struct {
	Old            Sno
	New            Sno
	MetaChanged    bool
	PayloadChanged bool
	XMLChanged     bool
	NameChanged    bool
}

func (c Change) HasChanged() bool {
	return c.MetaChanged || c.PayloadChanged || c.XMLChanged || c.NameChanged
}

type Diff struct {
	Added   []Sno
	Removed []Sno
	Changes []Change
}

type Sno struct {
	Group       d4.SnoGroup
	Id          int32
	Name        string `msgpack:"n"`
	MetaHash    uint32 `msgpack:"mh"`
	PayloadHash uint32 `msgpack:"ph"`
	XMLHash     uint32 `msgpack:"xh"`
}

type Manifest struct {
	Snos map[d4.SnoGroup]map[int32]Sno `msgpack:"snos"`
	mu   sync.Mutex
}

func NewManifest() *Manifest {
	return &Manifest{
		Snos: make(map[d4.SnoGroup]map[int32]Sno),
	}
}

func (m *Manifest) Add(s Sno) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sub, ok := m.Snos[s.Group]; ok {
		sub[s.Id] = s
		return
	}
	m.Snos[s.Group] = map[int32]Sno{
		s.Id: s,
	}
}

func (m *Manifest) Set() mapset.Set[[2]int32] {
	s := mapset.NewThreadUnsafeSet[[2]int32]()
	for group, sub := range m.Snos {
		for id, _ := range sub {
			s.Add(
				[2]int32{int32(group), id},
			)
		}
	}
	return s
}

func (m *Manifest) Diff(new *Manifest) Diff {
	old := m
	d := Diff{}

	oldSet := old.Set()
	newSet := new.Set()

	added := newSet.Difference(oldSet)
	removed := oldSet.Difference(newSet)
	common := newSet.Intersect(oldSet)

	// Add 'added' to changeset
	d.Added = make([]Sno, added.Cardinality())
	for a := range added.Iter() {
		d.Added = append(d.Added, new.Snos[d4.SnoGroup(a[0])][a[1]])
	}

	// Add 'removed' to changeset
	d.Removed = make([]Sno, removed.Cardinality())
	for r := range removed.Iter() {
		d.Removed = append(d.Removed, old.Snos[d4.SnoGroup(r[0])][r[1]])
	}

	// Add changed to changeset
	for c := range common.Iter() {
		snoGroup := d4.SnoGroup(c[0])
		snoId := c[1]

		change := Change{
			Old: old.Snos[snoGroup][snoId],
			New: new.Snos[snoGroup][snoId],
		}

		change.Old.Group = snoGroup

		//oldTocPath := filepath.Join(oldPath, coreTocPath)
		//newTocPath := filepath.Join(newPath, coreTocPath)
		//
		//oldToc, err := d4.ReadTocFile(oldTocPath)
		//if err != nil {
		//	slog.Error("failed to read oldToc toc file", slog.Any("error", err))
		//	os.Exit(1)
		//}
		//
		//newToc, err := d4.ReadTocFile(newTocPath)
		//if err != nil {
		//	slog.Error("failed to read new toc file", slog.Any("error", err))
		//	os.Exit(1)
		//}
		//
		//// Get old entries
		//oldSet := mapset.NewThreadUnsafeSet[Sno]()
		//for group, m := range oldToc.Entries {
		//	for id, name := range m {
		//		oldSet.Add(Sno{
		//			Group: group,
		//			Id:    id,
		//			Name:  name,
		//		})
		//	}
		//}
		//
		//// Get new entries
		//newSet := mapset.NewThreadUnsafeSet[Sno]()
		//for group, m := range newToc.Entries {
		//	for id, name := range m {
		//		newSet.Add(Sno{
		//			Group: group,
		//			Id:    id,
		//			Name:  name,
		//		})
		//	}
		//}
		//
		//// Write changes
		//fAdded, err := os.Create("samples/added.txt")
		//if err != nil {
		//	panic(err)
		//}
		//
		//fRemoved, err := os.Create("samples/removed.txt")
		//if err != nil {
		//	panic(err)
		//}
		//
		//fChanged, err := os.Create("samples/changed.txt")
		//if err != nil {
		//	panic(err)
		//}
		//
		//added := newSet.Difference(oldSet)
		//removed := oldSet.Difference(newSet)
		//common := newSet.Intersect(oldSet)
		//
		//// Write added
		//added.Each(func(a Sno) bool {
		//	if _, err := fmt.Fprintf(fAdded, "[%s] %s\n", groupName(a.Group), a.Name); err != nil {
		//		panic(err)
		//	}
		//	return false
		//})
		//
		//// Write removed
		//removed.Each(func(r Sno) bool {
		//	if _, err := fmt.Fprintf(fRemoved, "[%s] %s\n", groupName(r.Group), r.Name); err != nil {
		//		panic(err)
		//	}
		//	return false
		//})
		//
		//// Write changed
		//var progress atomic.Uint64
		//
		//util.DoWorkSlice(workers, common.ToSlice(), func(sno Sno) {
		//	defer func() {
		//		if i := progress.Add(1); i%1000 == 0 {
		//			slog.Info("Comparing snos...", slog.Uint64("progress", i))
		//		}
		//	}()
		//
		//	if strings.TrimSpace(sno.Name) == "" {
		//		return
		//	}
		//
		//	oldMetaPath := util.FindLocalizedFile(oldPath, util.FileTypeMeta, sno.Group, sno.Name)
		//	newMetaPath := util.FindLocalizedFile(newPath, util.FileTypeMeta, sno.Group, sno.Name)
		//
		//	oldMeta, err := d4.ReadSnoMetaFile(oldMetaPath)
		//	if err != nil {
		//		slog.Error("Error reading old meta", slog.String("path", oldMetaPath), slog.String("err", err.Error()))
		//		if _, err := fmt.Fprintf(fChanged, "[%s] %s (compare failed)\n", groupName(sno.Group), sno.Name); err != nil {
		//			panic(err)
		//		}
		//		return
		//	}
		//	newMeta, err := d4.ReadSnoMetaFile(newMetaPath)
		//	if err != nil {
		//		slog.Error("Error reading new meta", slog.String("path", newMetaPath), slog.String("err", err.Error()))
		//		if _, err := fmt.Fprintf(fChanged, "[%s] %s (compare failed)\n", groupName(sno.Group), sno.Name); err != nil {
		//			panic(err)
		//		}
		//		return
		//	}
		//
		//	// Log reasons
		//	var reasons []string
		//
		//	// Check data
		//	oldSer, err := msgpack.Marshal(oldMeta.Meta)
		//	if err != nil {
		//		panic(err)
		//	}
		//	newSer, err := msgpack.Marshal(newMeta.Meta)
		//	if err != nil {
		//		panic(err)
		//	}
		//
		//	if !bytes.Equal(oldSer, newSer) {
		//		reasons = append(reasons, "meta changed")
		//	}
		//
		//	// Check XML hash
		//	if oldMeta.Header.DwXMLHash != newMeta.Header.DwXMLHash {
		//		reasons = append(reasons, "possible server-side change")
		//	}
		//
		//	// Check payloads
		//	oldPayloadPath := util.ChangePathType(oldMetaPath, util.FileTypePayload)
		//	newPayLoadPath := util.ChangePathType(newMetaPath, util.FileTypePayload)
		//
		//	oldPayloadExists := true
		//	newPayloadExists := true
		//	if _, err := os.Stat(oldPayloadPath); err != nil {
		//		oldPayloadExists = false
		//	}
		//	if _, err := os.Stat(newPayLoadPath); err != nil {
		//		newPayloadExists = false
		//	}
		//
		//	if oldPayloadExists != newPayloadExists {
		//		if newPayloadExists {
		//			reasons = append(reasons, "payload added")
		//		} else {
		//			reasons = append(reasons, "payload removed")
		//		}
		//	}
		//
		//	if oldPayloadExists && newPayloadExists {
		//		oldPayloadData, err := os.ReadFile(oldPayloadPath)
		//		if err != nil {
		//			panic(err)
		//		}
		//		newPayloadData, err := os.ReadFile(newPayLoadPath)
		//		if err != nil {
		//			panic(err)
		//		}
		//
		//		if !bytes.Equal(oldPayloadData, newPayloadData) {
		//			reasons = append(reasons, "payload changed")
		//		}
		//	}
		//
		//	if len(reasons) > 0 {
		//		if _, err := fmt.Fprintf(fChanged, "[%s] %s (%s)\n", groupName(sno.Group), sno.Name, strings.Join(reasons, ",")); err != nil {
		//			panic(err)
		//		}
		//	}
		//})
		change.Old.Id = snoId
		change.New.Group = snoGroup
		change.New.Id = snoId

		change.PayloadChanged = change.Old.PayloadHash != change.New.PayloadHash
		change.XMLChanged = change.Old.XMLHash != change.New.XMLHash
		change.MetaChanged = change.Old.MetaHash != change.New.MetaHash
		change.NameChanged = change.Old.Name != change.New.Name

		if change.HasChanged() {
			d.Changes = append(d.Changes, change)
		}
	}

	return d
}
