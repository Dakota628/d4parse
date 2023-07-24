package ml

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/minio/highwayhash"
	"golang.org/x/exp/slices"
	"hash"
)

func hashKey(s string) []byte {
	h := sha256.New()
	h.Write([]byte(s))
	return h.Sum(nil)
}

func BytesToFeature(h hash.Hash64, b []byte) float64 {
	defer h.Reset()
	_, err := h.Write(b)
	if err != nil {
		panic(err) // should never happen
	}
	return float64(h.Sum64())
}

func ExtractGroupFeatures(dataPath string, group d4.SnoGroup) (map[int]int32, [][]float64, error) {
	var eachErr error

	featureKeys := mapset.NewThreadUnsafeSet[string]()
	featureMaps := make([]map[string]float64, 0)
	snoIds := make(map[int]int32)

	h, err := highwayhash.New64(hashKey(group.String()))
	if err != nil {
		return nil, nil, err
	}

	if err := d4.EachSnoMeta(dataPath, group, func(meta d4.SnoMeta) bool {
		// Get walkable meta
		metaObj, ok := meta.Meta.(d4.Walkable)
		if !ok {
			eachErr = errors.New("meta object not walkable")
			return false
		}

		featureMap := make(map[string]float64)

		// Walk each meta while maintaining a path
		metaObj.Walk(func(k string, v d4.Object, next d4.WalkNext, d ...any) {
			path := fmt.Sprintf("%s.%s", d[0].(string), k)

			// Check if it's a leaf type
			leaf := true
			switch x := v.(type) {
			case *d4.DT_NULL:
				featureMap[path] = 0
			case *d4.DT_BYTE:
				featureMap[path] = float64(x.Value)
			case *d4.DT_WORD:
				featureMap[path] = float64(x.Value)
			case *d4.DT_ENUM:
				featureMap[path] = float64(x.Value)
			case *d4.DT_INT:
				featureMap[path] = float64(x.Value)
			case *d4.DT_FLOAT:
				featureMap[path] = float64(x.Value)
			case *d4.DT_SNO:
				featureMap[path] = float64(x.Id)
			case *d4.DT_SNO_NAME:
				featureMap[path] = float64(x.Id)
			case *d4.DT_GBID:
				featureMap[path] = float64((uint64(x.Group) << 32) | uint64(x.Value))
			case *d4.DT_STARTLOC_NAME:
				featureMap[path] = float64(x.Value)
			case *d4.DT_UINT:
				featureMap[path] = float64(x.Value)
			case *d4.DT_ACD_NETWORK_NAME:
				featureMap[path] = float64(x.Value)
			case *d4.DT_SHARED_SERVER_DATA_ID:
				featureMap[path] = float64(x.Value)
			case *d4.DT_INT64:
				featureMap[path] = float64(x.Value)
			case *d4.DT_STRING_FORMULA:
				featureMap[path] = BytesToFeature(h, []byte(x.Value))
			case *d4.DT_CSTRING[*d4.DT_BYTE]:
				featureMap[path] = BytesToFeature(h, []byte(x.Value))
			case *d4.DT_CHARARRAY:
				featureMap[path] = BytesToFeature(h, []byte(string(x.Value)))
			case *d4.DT_RGBACOLOR:
				featureMap[path] = float64(
					(uint64(x.R) << 32) | (uint64(x.G) << 16) | (uint64(x.B) << 8) | uint64(x.A),
				)
			case *d4.DT_RGBACOLORVALUE:
				// Note: could probably truncate to float16 and combine to float64 instead
				featureMap[path] = BytesToFeature(
					h,
					[]byte(fmt.Sprintf("%f.%f.%f.%f", x.R, x.G, x.B, x.A)),
				)
			case *d4.DT_BCVEC2I:
				featureMap[path] = float64(
					(uint64(x.X) << 8) | uint64(x.Y),
				)
			case *d4.DT_VECTOR2D:
				featureMap[path] = float64(
					(uint64(x.X) << 8) | uint64(x.Y),
				)
			case *d4.DT_VECTOR3D:
				featureMap[path] = float64(
					(uint64(x.X) << 16) | (uint64(x.Y) << 8) | uint64(x.Z),
				)
			case *d4.DT_VECTOR4D:
				featureMap[path] = float64(
					(uint64(x.X) << 32) | (uint64(x.Y) << 16) | (uint64(x.Z) << 8) | uint64(x.W),
				)
			default:
				leaf = false
			}

			if leaf {
				featureKeys.Add(path)
			}

			next(path)
		}, "")

		snoIds[len(featureMaps)] = meta.Id.Value
		featureMaps = append(featureMaps, featureMap)

		return true
	}); err != nil {
		return nil, nil, err
	}

	if eachErr != nil {
		return nil, nil, eachErr
	}

	return snoIds, keysToFeatures(featureKeys, featureMaps), nil
}

func keyMeans(featureMaps []map[string]float64) map[string]float64 {
	means := make(map[string]float64)
	for _, fm := range featureMaps {
		for key, value := range fm {
			if currMean, ok := means[key]; ok {
				means[key] = (currMean + value) / 2
			} else {
				means[key] = value
			}
		}
	}
	return means
}

func keysToFeatures(featureKeys mapset.Set[string], featureMaps []map[string]float64) (features [][]float64) {
	means := keyMeans(featureMaps)
	keys := featureKeys.ToSlice()
	slices.Sort(keys)

	for _, fm := range featureMaps {
		featureVec := make([]float64, len(keys))
		for i, key := range keys {
			if v, ok := fm[key]; ok {
				featureVec[i] = v
			} else {
				featureVec[i] = means[key]
			}
		}
		features = append(features, featureVec)
	}

	return
}
