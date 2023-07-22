package main

import (
	"errors"
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/exp/slog"
	"os"
	"path/filepath"
)

const (
	sanctuaryEasternContinentSnoId = 69068
)

var (
	outputBasePath = filepath.Join("docs", "map")
)

type MarkerData struct {
	RefSnoGroup uint8          `msgpack:"g"`
	RefSno      int32          `msgpack:"r"`
	SourceSno   int32          `msgpack:"s"`
	X           float32        `msgpack:"x"`
	Y           float32        `msgpack:"y"`
	Z           float32        `msgpack:"z"`
	Metadata    map[string]any `msgpack:"m"`
}

type Polygon [][2]float32

type MapData struct {
	Markers  []MarkerData `msgpack:"m"`
	Polygons []Polygon    `msgpack:"p"`
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
		refSnoGroup, _ := toc.Entries.GetName(gma.SnoActor.Id)

		// Add marker data
		md = append(md, MarkerData{
			RefSnoGroup: uint8(refSnoGroup),
			RefSno:      gma.SnoActor.Id,
			SourceSno:   meta.Id.Value,
			X:           gma.TWorldTransform.Wp.X,
			Y:           gma.TWorldTransform.Wp.Y,
			Z:           gma.TWorldTransform.Wp.Z,
			Metadata: map[string]any{
				"gt": gma.EGizmoType.Value,
			},
		})
	})

	return md, nil
}

func loadWorldScene(baseMetaPath string, toc d4.Toc, worldId int32, sceneChunk *d4.SceneChunk) ([]MarkerData, error) {
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

	var md []MarkerData

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

			md = append(md, MarkerData{
				RefSnoGroup: uint8(marker.Snoname.Group),
				RefSno:      marker.Snoname.Id,
				SourceSno:   markerSetSnoMeta.Id.Value,
				X:           sceneChunk.Transform.Wp.X + marker.Transform.Wp.X,
				Y:           sceneChunk.Transform.Wp.Y + marker.Transform.Wp.Y,
				Z:           sceneChunk.Transform.Wp.Z + marker.Transform.Wp.Z,
				Metadata: map[string]any{
					"mt": marker.EType.Value,
				}, // TODO: could also add scale to add a larger radius on hover
			})
		})
	}

	return md, nil
}

func loadSubZone(baseMetaPath string, toc d4.Toc, worldId int32, subZoneId int32) ([]MarkerData, error) {
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

	var md []MarkerData

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
			md = append(md, MarkerData{
				RefSnoGroup: uint8(marker.Snoname.Group),
				RefSno:      marker.Snoname.Id,
				SourceSno:   markerSetSnoMeta.Id.Value,
				X:           marker.Transform.Wp.X,
				Y:           marker.Transform.Wp.Y,
				Z:           marker.Transform.Wp.Z,
				Metadata: map[string]any{
					"mt": marker.EType.Value,
				},
				// TODO: could also add scale to add a larger radius on hover
			})
		})
	})

	return md, nil
}

func loadWorldMarkers(baseMetaPath string, toc d4.Toc, worldId int32) ([]MarkerData, []Polygon, error) {
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
	var md []MarkerData

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
			md = append(md, sceneMarkers...)
		case *d4.SceneSpecification:
			for _, subZone := range x.ArSubzones.Value {
				subZoneMarkers, err := loadSubZone(baseMetaPath, toc, worldId, subZone.SnoSubzone.Id)
				if err != nil {
					panic(err) // Not a great solution but whatever
				}
				md = append(md, subZoneMarkers...)
			}
		}

		next()
	})

	return md, polygons, nil
}

func generateForWorld(baseMetaPath string, toc d4.Toc, worldSnoId int32) error {
	// Load markers
	var md MapData

	slog.Info("Loading global markers...")
	globalMarkers, err := loadGlobalMarkers(baseMetaPath, toc, worldSnoId)
	if err != nil {
		panic(err)
	}
	md.Markers = append(md.Markers, globalMarkers...)

	slog.Info("Loading Sanctuary_Eastern_Continent markers...")
	var worldMarkers []MarkerData
	worldMarkers, md.Polygons, err = loadWorldMarkers(baseMetaPath, toc, worldSnoId)
	if err != nil {
		return err
	}
	md.Markers = append(md.Markers, worldMarkers...)

	// Write marker data
	slog.Info("Generating map data file...")
	packed, err := msgpack.Marshal(md)
	if err != nil {
		return err
	}

	outputPath := filepath.Join(outputBasePath, "data", fmt.Sprintf("%d.mpk", worldSnoId))
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	if _, err = f.Write(packed); err != nil {
		return err
	}

	// TODO: also load quest markers

	return nil
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

	for worldSnoId, worldSnoName := range toc.Entries[d4.SnoGroupWorld] {
		if err := generateForWorld(baseMetaPath, toc, worldSnoId); err != nil {
			slog.Error("Failed generate world data", slog.Any("error", err), slog.Any("snoName", worldSnoName))
			os.Exit(1)
		}
	}
}
