package main

import (
	"errors"
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/slog"
	"os"
	"path/filepath"
	"strings"
)

const (
	sanctuaryEasternContinentSnoId = 69068
	// TODO: gbid spawns
)

var (
	outputFile = filepath.Join("map", "markers.mpk")
	sceneTypes = map[string]string{
		"(Game)":         "Game",
		"(Lighting)":     "Lighting",
		"(Merged)":       "Merged",
		"(Cameras)":      "Cameras",
		"(Props)":        "Props",
		"(Merged_Props)": "Merged Props",
		"(Events)":       "Events",
		"(Clickies)":     "Clickies",
		"(Population)":   "Population",
		"(Ambient)":      "Ambient",
	}
)

type MarkerData struct {
	RefSno    int32   `msgpack:"r"`
	SourceSno int32   `msgpack:"s"`
	Name      string  `msgpack:"n"`
	Desc      string  `msgpack:"d"`
	X         float32 `msgpack:"x"`
	Y         float32 `msgpack:"y"`
	Z         float32 `msgpack:"z"`
}

type Polygon [][2]float32

func getSceneGroup(prefix string, snoName string) string {
	for suffix, group := range sceneTypes {
		if strings.HasSuffix(snoName, suffix) {
			return fmt.Sprintf("%s - %s", prefix, group)
		}
	}
	return prefix
}

func mergeMarkerGroups(a, b map[string][]MarkerData) {
	for key, v := range b {
		a[key] = append(a[key], v...)
	}
}

func loadGlobalMarkers(baseMetaPath string, toc d4.Toc, worldSnoId int32) ([]MarkerData, error) {
	meta, err := d4.ReadSnoMetaFile(filepath.Join(baseMetaPath, "Global", "global_markers.glo"))
	if err != nil {
		return nil, err
	}

	gd, ok := meta.Meta.(*d4.GlobalDefinition)
	if !ok {
		return nil, errors.New("not global definition")
	}

	var md []MarkerData

	gd.Walk(func(k string, v d4.Object, next d4.WalkNext) {
		// Get global marker actor
		gma, ok := v.(*d4.GlobalMarkerActor)
		if !ok || gma == nil {
			next()
			return
		}

		// If it's not in SEC, skip
		if gma.SnoWorld.Id != worldSnoId {
			return
		}

		// Get referenced sno name
		refSnoGroup, refSnoName := toc.Entries.GetName(gma.SnoActor.Id)

		// Add marker data
		md = append(md, MarkerData{
			RefSno:    gma.SnoActor.Id,
			SourceSno: meta.Id.Value,
			Name:      fmt.Sprintf("[%s] %s", refSnoGroup, refSnoName),     // Lookup szMarkerName?
			Desc:      fmt.Sprintf("Gizmo Type: %d", gma.EGizmoType.Value), // TODO
			X:         gma.TWorldTransform.Wp.X,
			Y:         gma.TWorldTransform.Wp.Y,
			Z:         gma.TWorldTransform.Wp.Z,
		})
	})

	return md, nil
}

func loadWorldScene(baseMetaPath string, toc d4.Toc, worldId int32, sceneChunk *d4.SceneChunk) (map[string][]MarkerData, error) {
	sceneSnoGroup, sceneSnoName := toc.Entries.GetName(sceneChunk.Snoname.Id)
	worldSnoPath := filepath.Join(baseMetaPath, "Scene", sceneSnoName+sceneSnoGroup.Ext())
	sceneSnoMeta, err := d4.ReadSnoMetaFile(worldSnoPath)
	if err != nil {
		return nil, err
	}

	sd, ok := sceneSnoMeta.Meta.(*d4.SceneDefinition)
	if !ok {
		return nil, errors.New("not scene definition")
	}

	md := make(map[string][]MarkerData)

	for _, layer := range sd.ArLayers.Value {
		markerSetSnoGroup, markerSetSnoName := toc.Entries.GetName(layer.Id)
		markerSetSnoPath := filepath.Join(baseMetaPath, "MarkerSet", markerSetSnoName+markerSetSnoGroup.Ext())
		markerSetSnoMeta, err := d4.ReadSnoMetaFile(markerSetSnoPath)
		if err != nil {
			return nil, err
		}

		msd, ok := markerSetSnoMeta.Meta.(*d4.MarkerSetDefinition)
		if !ok {
			return nil, errors.New("not marker set definition")
		}

		msd.Walk(func(k string, v d4.Object, next d4.WalkNext) {
			marker, ok := v.(*d4.Marker)
			if !ok {
				next()
				return
			}

			refSnoGroup, refSnoName := toc.Entries.GetName(marker.Snoname.Id)

			sceneGroup := getSceneGroup("World", markerSetSnoName)
			md[sceneGroup] = append(md[sceneGroup], MarkerData{
				RefSno:    marker.Snoname.Id,
				SourceSno: markerSetSnoMeta.Id.Value,
				Name:      fmt.Sprintf("[%s] %s", refSnoGroup, refSnoName),
				Desc:      fmt.Sprintf("Type: %d", marker.EType.Value),
				X:         sceneChunk.Transform.Wp.X + marker.Transform.Wp.X,
				Y:         sceneChunk.Transform.Wp.Y + marker.Transform.Wp.Y,
				Z:         sceneChunk.Transform.Wp.Z + marker.Transform.Wp.Z,
				// TODO: could also add scale to add a larger radius on hover
			})
		})
	}

	return md, nil
}

func loadSubZone(baseMetaPath string, toc d4.Toc, worldId int32, subZoneId int32) (map[string][]MarkerData, error) {
	subZoneGroup, subZoneName := toc.Entries.GetName(subZoneId)
	subZonePath := filepath.Join(baseMetaPath, "Subzone", subZoneName+subZoneGroup.Ext())
	subZoneMeta, err := d4.ReadSnoMetaFile(subZonePath)
	if err != nil {
		return nil, err
	}

	sd, ok := subZoneMeta.Meta.(*d4.SubzoneDefinition)
	if !ok {
		return nil, errors.New("not sub zone definition")
	}

	md := make(map[string][]MarkerData)

	sd.Walk(func(k string, v d4.Object, next d4.WalkNext) {
		szMsEntry, ok := v.(*d4.SubzoneWorldMarkerSetEntry)
		if !ok {
			next()
			return
		}

		markerSetSnoGroup, markerSetSnoName := toc.Entries.GetName(szMsEntry.SnoMarkerSet.Id)
		markerSetSnoPath := filepath.Join(baseMetaPath, "MarkerSet", markerSetSnoName+markerSetSnoGroup.Ext())
		markerSetSnoMeta, err := d4.ReadSnoMetaFile(markerSetSnoPath)
		if err != nil {
			panic(err)
		}

		msd, ok := markerSetSnoMeta.Meta.(*d4.MarkerSetDefinition)
		if !ok {
			panic("not marker set definition")
		}

		msd.Walk(func(k string, v d4.Object, next d4.WalkNext) {
			marker, ok := v.(*d4.Marker)
			if !ok {
				next()
				return
			}

			refSnoGroup, refSnoName := toc.Entries.GetName(marker.Snoname.Id)

			sceneGroup := getSceneGroup("Subzone", markerSetSnoName)
			md[sceneGroup] = append(md[sceneGroup], MarkerData{
				RefSno:    marker.Snoname.Id,
				SourceSno: markerSetSnoMeta.Id.Value,
				Name:      fmt.Sprintf("[%s] %s", refSnoGroup, refSnoName),
				Desc:      fmt.Sprintf("Type: %d", marker.EType.Value),
				X:         marker.Transform.Wp.X,
				Y:         marker.Transform.Wp.Y,
				Z:         marker.Transform.Wp.Z,
				// TODO: could also add scale to add a larger radius on hover
			})
		})
	})

	return md, nil
}

func loadWorldMarkers(baseMetaPath string, toc d4.Toc, worldId int32) (map[string][]MarkerData, []Polygon, error) {
	// Get sno file
	worldSnoGroup, worldSnoName := toc.Entries.GetName(worldId)
	worldSnoPath := filepath.Join(baseMetaPath, "World", worldSnoName+worldSnoGroup.Ext())
	worldSnoMeta, err := d4.ReadSnoMetaFile(worldSnoPath)
	if err != nil {
		return nil, nil, err
	}

	wd, ok := worldSnoMeta.Meta.(*d4.WorldDefinition)
	if !ok {
		return nil, nil, errors.New("not world definition")
	}

	// Load data from the world
	var polygons []Polygon
	md := make(map[string][]MarkerData)

	wd.Walk(func(k string, v d4.Object, next d4.WalkNext) {
		switch x := v.(type) {
		case *d4.ScreenStaticCamps:
			var poly Polygon
			for _, vec := range x.ArPoints.Value {
				poly = append(poly, [2]float32{
					vec.X,
					vec.Y,
				})
			}
			polygons = append(polygons, poly)
		case *d4.Type_bef5a4a:
			// TODO: conditional map overlays
		case *d4.SceneChunk:
			sceneMarkers, err := loadWorldScene(baseMetaPath, toc, worldId, x)
			if err != nil {
				panic(err) // Not a great solution but whatever
			}
			mergeMarkerGroups(md, sceneMarkers)
		case *d4.SceneSpecification:
			for _, subZone := range x.ArSubzones.Value {
				subZoneMarkers, err := loadSubZone(baseMetaPath, toc, worldId, subZone.SnoSubzone.Id)
				if err != nil {
					panic(err) // Not a great solution but whatever
				}
				mergeMarkerGroups(md, subZoneMarkers)
			}
		}

		next()
	})

	return md, polygons, nil
}

func main() {
	if len(os.Args) < 2 {
		slog.Error("usage: dumpmapdata d4DataPath")
		os.Exit(1)
	}

	d4DataPath := os.Args[1]
	baseMetaPath := filepath.Join(d4DataPath, "base", "meta")
	tocFilePath := filepath.Join(d4DataPath, "base", "CoreTOC.dat")

	// Read TOC
	slog.Info("Reading TOC file...")
	toc, err := d4.ReadTocFile(tocFilePath)
	if err != nil {
		slog.Error("Failed to read toc file", slog.Any("error", err))
		os.Exit(1)
	}

	// Load markers
	var polygons []Polygon
	markers := make(map[string][]MarkerData)

	slog.Info("Loading global markers...")
	markers["Global"], err = loadGlobalMarkers(baseMetaPath, toc, sanctuaryEasternContinentSnoId)
	if err != nil {
		panic(err)
	}

	slog.Info("Loading Sanctuary_Eastern_Continent markers...")
	var worldMarkers map[string][]MarkerData
	worldMarkers, polygons, err = loadWorldMarkers(baseMetaPath, toc, sanctuaryEasternContinentSnoId)
	if err != nil {
		panic(err)
	}
	mergeMarkerGroups(markers, worldMarkers)

	// Write marker data
	slog.Info("Generating markers file...")
	packed, err := msgpack.Marshal(map[string]any{
		"Markers":  markers,
		"Polygons": polygons,
	})
	if err != nil {
		panic(err)
	}

	f, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}

	if _, err = f.Write(packed); err != nil {
		panic(err)
	}
}
