package ml

import (
	"errors"
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/Dakota628/d4parse/pkg/d4/util"
	"path/filepath"
)

type ValueCountMap map[string]map[string]uint

func (f ValueCountMap) Add(path string, value string) {
	if sub, ok := f[path]; ok {
		sub[value] += 1
	} else {
		f[path] = map[string]uint{
			value: 1,
		}
	}
}

func (f ValueCountMap) ValueCount(path string, value string) uint {
	if sub, ok := f[path]; ok {
		return sub[value]
	}
	return 0
}

func (f ValueCountMap) GetValueProbability(metaObj d4.Walkable, count uint) map[string]float32 {
	if count == 0 {
		panic("count must be greater than 0")
	}
	flCount := float32(count)

	probs := make(map[string]float32, len(f))
	WalkSerPathValues(metaObj, func(path string, value string) {
		probs[path] = float32(f.ValueCount(path, value)) / flCount
	})

	return probs
}

func WalkSerPathValues(metaObj d4.Walkable, cb func(string, string)) {
	metaObj.Walk(func(k string, v d4.Object, next d4.WalkNext, d ...any) {
		path := fmt.Sprintf("%s.%s", d[0].(string), k)

		switch x := v.(type) {
		case *d4.DT_NULL:
			cb(path, "null")
		case *d4.DT_BYTE:
			cb(path, fmt.Sprintf("%d", x.Value))
		case *d4.DT_WORD:
			cb(path, fmt.Sprintf("%d", x.Value))
		case *d4.DT_ENUM:
			cb(path, fmt.Sprintf("%d", x.Value))
		case *d4.DT_INT:
			cb(path, fmt.Sprintf("%d", x.Value))
		case *d4.DT_FLOAT:
			cb(path, fmt.Sprintf("%f", x.Value))
		case *d4.DT_SNO:
			cb(path, fmt.Sprintf("%d", x.Id))
		case *d4.DT_SNO_NAME:
			cb(path, fmt.Sprintf("%d", x.Id))
		case *d4.DT_GBID:
			cb(path, fmt.Sprintf("%d/%d", x.Group, x.Value))
		case *d4.DT_STARTLOC_NAME:
			cb(path, fmt.Sprintf("%d", x.Value))
		case *d4.DT_UINT:
			cb(path, fmt.Sprintf("%d", x.Value))
		case *d4.DT_ACD_NETWORK_NAME:
			cb(path, fmt.Sprintf("%d", x.Value))
		case *d4.DT_SHARED_SERVER_DATA_ID:
			cb(path, fmt.Sprintf("%d", x.Value))
		case *d4.DT_INT64:
			cb(path, fmt.Sprintf("%d", x.Value))
		case *d4.DT_STRING_FORMULA:
			cb(path, x.Value)
		case *d4.DT_CSTRING[*d4.DT_BYTE]:
			cb(path, x.Value)
		case *d4.DT_CHARARRAY:
			cb(path, string(x.Value))
		case *d4.DT_RGBACOLOR:
			cb(path, fmt.Sprintf("%d,%d,%d,%d", x.R, x.G, x.B, x.A))
		case *d4.DT_RGBACOLORVALUE:
			cb(path, fmt.Sprintf("%f,%f,%f,%f", x.R, x.G, x.B, x.A))
		case *d4.DT_BCVEC2I:
			cb(path, fmt.Sprintf("%f,%f", x.X, x.Y))
		case *d4.DT_VECTOR2D:
			cb(path, fmt.Sprintf("%f,%f", x.X, x.Y))
		case *d4.DT_VECTOR3D:
			cb(path, fmt.Sprintf("%f,%f,%f", x.X, x.Y, x.Z))
		case *d4.DT_VECTOR4D:
			cb(path, fmt.Sprintf("%f,%f,%f,%f", x.X, x.Y, x.Z, x.W))
		}

		next(path)
	}, "")
}

func ExtractGroupValueCounts(dataPath string, toc *d4.Toc, group d4.SnoGroup) (ValueCountMap, uint, error) {
	var eachErr error
	var count uint
	f := make(ValueCountMap)

	if err := util.EachSnoMeta(dataPath, toc, group, func(meta d4.SnoMeta) bool {
		// Get walkable meta
		metaObj, ok := meta.Meta.(d4.Walkable)
		if !ok {
			eachErr = errors.New("meta object not walkable")
			return false
		}

		WalkSerPathValues(metaObj, f.Add)
		count += 1
		return true
	}); err != nil {
		return nil, 0, err
	}

	if eachErr != nil {
		return nil, 0, eachErr
	}

	return f, count, nil
}

func DetermineSnoValueFreqs(dataPath string, group d4.SnoGroup, snoId int32) (map[string]float32, error) {
	// Read toc
	toc, err := d4.ReadTocFile(filepath.Join(dataPath, "base", "CoreTOC.dat"))
	if err != nil {
		return nil, err
	}

	// Get counts for group
	counts, l, err := ExtractGroupValueCounts(dataPath, toc, group)
	if err != nil {
		return nil, err
	}

	// Get sno meta by id
	metaPath := util.MetaPathById(dataPath, toc, snoId)
	metaObj, err := d4.ReadSnoMetaFile(metaPath, toc)
	if err != nil {
		return nil, err
	}

	return counts.GetValueProbability(metaObj.Meta.(d4.Walkable), l), nil
}
