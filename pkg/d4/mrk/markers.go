package mrk

import (
	"errors"
	"fmt"
	"github.com/Dakota628/d4parse/pkg/d4"
	"github.com/Dakota628/d4parse/pkg/d4/util"
	"github.com/Dakota628/d4parse/pkg/pb"
	"github.com/bmatcuk/doublestar/v4"
	mapset "github.com/deckarep/golang-set/v2"
	"google.golang.org/protobuf/proto"
	"os"
	"path/filepath"
	"sync"
)

const (
	adHocWorkers = 100
)

// UniqueMarkerData ...
type UniqueMarkerData struct {
	mu      sync.Mutex
	seen    mapset.Set[string]
	Markers []*pb.MarkerData
}

func NewUniqueMarkerData() *UniqueMarkerData {
	return &UniqueMarkerData{
		seen: mapset.NewSet[string](),
	}
}

func (u *UniqueMarkerData) Add(data ...*pb.MarkerData) {
	for _, d := range data {
		pbb, err := proto.Marshal(d)
		if err != nil {
			panic(err)
		}
		if u.seen.Add(string(pbb)) {
			u.mu.Lock()
			u.Markers = append(u.Markers, d)
			u.mu.Unlock()
		}
	}
}

// MarkerExtractor ...
type MarkerExtractor struct {
	dumpPath     string
	baseMetaPath string
	outputPath   string
	toc          d4.Toc

	data          *pb.MapData
	uniqueMarkers *UniqueMarkerData
	seenSnos      mapset.Set[int32]

	snoId int32
}

func NewMarkerExtractor(dumpPath string, outputPath string, toc d4.Toc) *MarkerExtractor {
	return &MarkerExtractor{
		dumpPath:     dumpPath,
		baseMetaPath: filepath.Join(dumpPath, "base", "meta"),
		outputPath:   outputPath,
		toc:          toc,

		data:          &pb.MapData{},
		uniqueMarkers: NewUniqueMarkerData(),
		seenSnos:      mapset.NewSet[int32](),
	}
}

func (e *MarkerExtractor) getMaxNativeZoom(tiles int32) (zoom uint32) {
	for ; tiles%2 == 0 && tiles > 0; tiles /= 2 {
		zoom++
	}
	return
}

func (e *MarkerExtractor) getBoundingBox(snoMeta *d4.SnoMeta, transform *Transform) (*pb.AABB, error) {
	if snoMeta == nil {
		return nil, nil
	}

	switch x := snoMeta.Meta.(type) {
	case *d4.ActorDefinition:
		off := transform.GetRelWorldPos(&x.AabbBounds.Wp)
		ext := transform.GetRelWorldPos(&x.AabbBounds.WvExt)
		return &pb.AABB{
			Offset: pb.Vec3ToPoint3D(off),
			Ext:    pb.Vec3ToPoint3D(ext),
		}, nil
	}

	return nil, nil
}

func (e *MarkerExtractor) getDataSnos(m *d4.Marker) (snos []int32) {
	m.PtBase.Walk(func(k string, v d4.Object, next d4.WalkNext, d ...any) {
		switch x := v.(type) {
		case *d4.DT_SNO:
			if x.Id > 0 {
				snos = append(snos, x.Id)
			}
		case *d4.DT_SNO_NAME:
			if x.Id > 0 {
				snos = append(snos, x.Id)
			}
		}
		next()
	})
	return
}

func (e *MarkerExtractor) loadGlobalMarkers(worldSnoId int32, transform *Transform) error {
	metaPath := util.BaseFilePath(e.dumpPath, util.FileTypeMeta, d4.SnoGroupGlobal, "global_markers")
	meta, err := d4.ReadSnoMetaFile(metaPath)
	if err != nil {
		return err
	}

	gd, ok := meta.Meta.(*d4.GlobalDefinition)
	if !ok {
		return errors.New("not global definition")
	}

	errs := util.NewErrors()

	for _, content := range gd.PtContent.Value {
		gwd, ok := content.(*d4.GlobalWaypointData)
		if !ok {
			continue
		}

		util.DoWorkSlice(adHocWorkers, gwd.ArGlobalMarkerActors.Value, func(gma *d4.GlobalMarkerActor) {
			// If it's not in the world, skip
			if gma.SnoWorld.Id != worldSnoId {
				return
			}

			// Update transform
			currTransform := transform.AddPR(&gma.TWorldTransform)
			vec := currTransform.GetWorldPosition()

			// Add referenced actor
			if gma.SnoActor.Id > 0 {
				refSnoGroup, _ := e.toc.Entries.GetName(gma.SnoActor.Id)
				e.uniqueMarkers.Add(&pb.MarkerData{
					RefSnoGroup: int32(refSnoGroup),
					RefSno:      gma.SnoActor.Id,
					SourceSno:   meta.Id.Value,
					Position:    pb.Vec3ToPoint3D(vec),
					Extra: &pb.ExtraMarkerData{
						GizmoType: &gma.EGizmoType.Value,
					},
				})
			}

			// Add referenced marker set
			if _, markerSetName := e.toc.Entries.GetName(gma.SnoMarkerSet.Id, d4.SnoGroupMarkerSet); markerSetName != "" {
				markerSetMetaPath := util.BaseFilePath(e.dumpPath, util.FileTypeMeta, d4.SnoGroupMarkerSet, markerSetName)
				markerSetMeta, err := d4.ReadSnoMetaFile(markerSetMetaPath)
				if err != nil {
					errs.Add(err)
					return
				}

				if err := e.addMarkerSet(markerSetMeta, transform); err != nil {
					errs.Add(err)
					return
				}
			}
		})
	}

	return errs.Err()
}

func (e *MarkerExtractor) addMarkerSno(snoMeta *d4.SnoMeta, transform *Transform) error {
	if snoMeta == nil {
		return nil
	}

	switch x := snoMeta.Meta.(type) {
	case *d4.ActorDefinition:
		snoGroup, snoName := e.toc.Entries.GetName(x.SnoPrefabAttachment.Id)
		if snoGroup <= 0 {
			return nil
		}

		metaPath := util.BaseFilePath(e.dumpPath, util.FileTypeMeta, snoGroup, snoName)
		meta, err := d4.ReadSnoMetaFile(metaPath)
		if err != nil {
			return err
		}

		if err := e.addMarkerSet(meta, transform); err != nil {
			return err
		}
	}

	return nil
}

func (e *MarkerExtractor) getMarkerGroupHashes(marker *d4.Marker) []uint32 {
	return nil // Note: DWGroupHash was removed
	//hashes := mapset.NewThreadUnsafeSet[uint32]()
	//for _, markerLink := range marker.PtMarkerLinks.Value {
	//	hashes.Add(markerLink.DwGroupHash.Value)
	//}
	//return hashes.ToSlice()
}

func (e *MarkerExtractor) addRawMarker(marker *d4.Marker, sourceSno int32, transform *Transform) error {
	var snoMeta *d4.SnoMeta
	if snoGroup, snoName := e.toc.Entries.GetName(marker.Snoname.Id); snoGroup > 0 {
		metaPath := util.BaseFilePath(e.dumpPath, util.FileTypeMeta, snoGroup, snoName)
		if meta, err := d4.ReadSnoMetaFile(metaPath); err == nil {
			snoMeta = &meta
		}
	}

	bounds, err := e.getBoundingBox(snoMeta, transform)
	if err != nil {
		return err
	}

	e.uniqueMarkers.Add(&pb.MarkerData{
		RefSnoGroup: marker.Snoname.Group,
		RefSno:      marker.Snoname.Id,
		SourceSno:   sourceSno,
		DataSnos:    e.getDataSnos(marker),
		Position:    pb.Vec3ToPoint3D(transform.GetWorldPosition()),
		Extra: &pb.ExtraMarkerData{
			MarkerType: &marker.EType.Value,
			Bounds:     bounds,
		},
		MarkerHash:        &marker.DwHash.Value,
		MarkerGroupHashes: e.getMarkerGroupHashes(marker),
	})

	if err = e.addMarkerSno(snoMeta, transform); err != nil {
		return err
	}

	return nil
}

func (e *MarkerExtractor) addNestedMarkerSet(marker *d4.Marker, transform *Transform) error {
	if d4.SnoGroup(marker.Snoname.Group) != d4.SnoGroupMarkerSet {
		return nil
	}

	_, markerSetName := e.toc.Entries.GetName(marker.Snoname.Id, d4.SnoGroupMarkerSet)
	if markerSetName == "" {
		return nil
	}

	metaPath := util.BaseFilePath(e.dumpPath, util.FileTypeMeta, d4.SnoGroupMarkerSet, markerSetName)
	meta, err := d4.ReadSnoMetaFile(metaPath)
	if err != nil {
		return err
	}

	if err = e.addMarkerSet(meta, transform); err != nil {
		return err
	}

	return err
}

func (e *MarkerExtractor) addMarker(marker *d4.Marker, sourceSno int32, transform *Transform, isInstance bool) error {
	markerTransform := transform.AddPRAndScale(&marker.Transform, &marker.VecScale)

	if err := e.addRawMarker(marker, sourceSno, markerTransform); err != nil {
		return err
	}

	if err := e.addNestedMarkerSet(marker, markerTransform); err != nil {
		return err
	}

	return nil
}

func (e *MarkerExtractor) addMarkerSet(markerSetSnoMeta d4.SnoMeta, transform *Transform) error {
	e.seenSnos.Add(markerSetSnoMeta.Id.Value)

	msd, ok := markerSetSnoMeta.Meta.(*d4.MarkerSetDefinition)
	if !ok {
		return nil
	}

	errs := util.NewErrors()

	util.DoWorkSlice(adHocWorkers, msd.TMarkerSet.Value, func(marker *d4.Marker) {
		if err := e.addMarker(marker, markerSetSnoMeta.Id.Value, transform, false); err != nil {
			errs.Add(err)
			return
		}
	})

	return errs.Err()
}

func (e *MarkerExtractor) loadWorldRelatedMarkers(worldName string, transform *Transform) error {
	markerSetPattern := util.BaseFilePattern(
		e.dumpPath,
		util.FileTypeMeta,
		d4.SnoGroupMarkerSet,
		fmt.Sprintf("%s*_Content", worldName),
		"",
		"*",
	)
	files, err := doublestar.FilepathGlob(markerSetPattern)
	if err != nil {
		return err
	}

	errs := util.NewErrors()

	util.DoWorkSlice(adHocWorkers, files, func(file string) {
		markerSetSnoMeta, err := d4.ReadSnoMetaFile(file)
		if err != nil {
			errs.Add(err)
			return
		}

		if e.seenSnos.Contains(markerSetSnoMeta.Id.Value) {
			return
		}

		if err := e.addMarkerSet(markerSetSnoMeta, transform); err != nil {
			errs.Add(err)
			return
		}
	})

	return errs.Err()
}

func (e *MarkerExtractor) loadSubzoneMarkers(subZoneId int32, transform *Transform) error {
	errs := util.NewErrors()

	// Load sub zone
	subZoneGroup, subZoneName := e.toc.Entries.GetName(subZoneId, d4.SnoGroupSubzone)
	subZonePath := util.BaseFilePath(e.dumpPath, util.FileTypeMeta, subZoneGroup, subZoneName)
	subZoneMeta, err := d4.ReadSnoMetaFile(subZonePath)
	if err != nil {
		return err
	}

	sd, ok := subZoneMeta.Meta.(*d4.SubzoneDefinition)
	if !ok {
		return errors.New("not sub zone definition")
	}

	// Add subzone marker sets
	util.DoWorkSlice(adHocWorkers, sd.ArWorldMarkerSets.Value, func(subZoneMarkerSetEntry *d4.SubzoneWorldMarkerSetEntry) {
		markerSetSnoGroup, markerSetSnoName := e.toc.Entries.GetName(subZoneMarkerSetEntry.SnoMarkerSet.Id, d4.SnoGroupMarkerSet)
		markerSetSnoPath := util.BaseFilePath(e.dumpPath, util.FileTypeMeta, markerSetSnoGroup, markerSetSnoName)
		markerSetSnoMeta, err := d4.ReadSnoMetaFile(markerSetSnoPath)
		if err != nil {
			errs.Add(err)
			return
		}

		if err := e.addMarkerSet(markerSetSnoMeta, transform); err != nil {
			errs.Add(err)
			return
		}
	})

	// Add related marker sets
	relatedMarkerSetPattern := util.BaseFilePattern(e.dumpPath, util.FileTypeMeta, d4.SnoGroupMarkerSet, subZoneName, "", " (*)")
	relatedMarkerSetPaths, err := doublestar.FilepathGlob(relatedMarkerSetPattern)
	if err != nil {
		return err
	}

	util.DoWorkSlice(adHocWorkers, relatedMarkerSetPaths, func(markerSetSnoPath string) {
		markerSetSnoMeta, err := d4.ReadSnoMetaFile(markerSetSnoPath)
		if err != nil {
			panic(err)
		}

		if err := e.addMarkerSet(markerSetSnoMeta, transform); err != nil {
			errs.Add(err)
			return
		}
	})

	return errs.Err()
}

func (e *MarkerExtractor) loadSceneMarkers(sceneId int32, transform *Transform) error {
	sceneSnoGroup, sceneSnoName := e.toc.Entries.GetName(sceneId, d4.SnoGroupScene)
	sceneSnoPath := util.BaseFilePath(e.dumpPath, util.FileTypeMeta, sceneSnoGroup, sceneSnoName)
	sceneSnoMeta, err := d4.ReadSnoMetaFile(sceneSnoPath)
	if err != nil {
		return err
	}

	sd, ok := sceneSnoMeta.Meta.(*d4.SceneDefinition)
	if !ok {
		return errors.New("not scene definition")
	}

	errs := util.NewErrors()
	util.DoWorkSlice(adHocWorkers, sd.ArLayers.Value, func(layer *d4.DT_SNO) {
		markerSetSnoGroup, markerSetSnoName := e.toc.Entries.GetName(layer.Id, d4.SnoGroupMarkerSet)
		markerSetSnoPath := util.BaseFilePath(e.dumpPath, util.FileTypeMeta, markerSetSnoGroup, markerSetSnoName)
		markerSetSnoMeta, err := d4.ReadSnoMetaFile(markerSetSnoPath)
		if err != nil {
			errs.Add(err)
			return
		}

		msd, ok := markerSetSnoMeta.Meta.(*d4.MarkerSetDefinition)
		if !ok {
			errs.Add(errors.New("not marker set definition"))
			return
		}

		for _, marker := range msd.TMarkerSet.Value {
			if err := e.addMarker(marker, markerSetSnoMeta.Id.Value, transform, false); err != nil {
				errs.Add(err)
				return
			}
		}
	})

	return nil
}

func (e *MarkerExtractor) loadSceneChunkMarkers(sceneChunk *d4.SceneChunk, transform *Transform) error {
	// TODO: are there any cases where we need to apply the quaternion?
	return e.loadSceneMarkers(sceneChunk.Snoname.Id, transform.AddPR(&sceneChunk.Transform))
}

func (e *MarkerExtractor) loadWorldMarkers(worldId int32, transform *Transform) error {
	// Get world definition
	_, worldSnoName := e.toc.Entries.GetName(worldId, d4.SnoGroupWorld)
	worldSnoPath := util.BaseFilePath(e.dumpPath, util.FileTypeMeta, d4.SnoGroupWorld, worldSnoName)
	worldSnoMeta, err := d4.ReadSnoMetaFile(worldSnoPath)
	if err != nil {
		return err
	}

	wd, ok := worldSnoMeta.Meta.(*d4.WorldDefinition)
	if !ok {
		return errors.New("not world definition")
	}

	errs := util.NewErrors()

	// Process default subzone
	if wd.SnoSubzoneDefault.Id > 0 {
		if err := e.loadSubzoneMarkers(wd.SnoSubzoneDefault.Id, transform); err != nil {
			return err
		}
	}

	// Process scene chunks and scene specs
	for _, serverData := range wd.PtServerData.Value {
		util.DoWorkSlice(adHocWorkers, serverData.PtSceneChunks.Value, func(sceneChunk *d4.SceneChunk) {
			if err := e.loadSceneChunkMarkers(sceneChunk, transform); err != nil {
				errs.Add(err)
				return
			}

			util.DoWorkSlice(adHocWorkers, sceneChunk.TSceneSpec.ArSubzones.Value, func(subZone *d4.SubzoneRelation) {
				if subZone.SnoSubzone.Id == wd.SnoSubzoneDefault.Id {
					return
				}

				if err := e.loadSubzoneMarkers(subZone.SnoSubzone.Id, transform); err != nil {
					errs.Add(err)
					return
				}
			})
		})
	}

	// Process screen static camps
	for _, regionBoundary := range wd.ArRegionBoundaries.Value {
		poly := &pb.Polygon{}
		for _, vec := range regionBoundary.ArPoints.Value {
			poly.Vertices = append(poly.Vertices, &pb.Point2D{
				X: vec.X,
				Y: vec.Y,
			})
		}
		e.data.Polygons = append(e.data.Polygons, poly)
	}

	// Load related markers (do this last since it's a hacky way -- we don't know how to get VFX, Audio or Road markers yet)
	if err = e.loadWorldRelatedMarkers(worldSnoName, transform); err != nil {
		return err
	}

	// TODO: conditional map overlays (Type_bef5a4a)

	// Update world data
	e.data.GridSize = wd.FlGridSize.Value
	e.data.ZoneArtScale = wd.TZoneMapParams.FlZoneArtScale.Value
	e.data.ZoneArtCenter = &pb.Point2D{
		X: wd.TZoneMapParams.VecZoneArtCenter.X,
		Y: wd.TZoneMapParams.VecZoneArtCenter.Y,
	}
	e.data.Bounds = &pb.Bounds{
		X: wd.TZoneMapParams.Unk_3620f37.Value,
		Y: wd.TZoneMapParams.Unk_c60b9b0.Value,
	}
	e.data.MaxNativeZoom = e.getMaxNativeZoom(wd.TZoneMapParams.Unk_3620f37.Value)

	return errs.Err()
}

func (e *MarkerExtractor) AddWorld(worldSnoId int32) error {
	e.snoId = worldSnoId

	// Load global markers
	if err := e.loadGlobalMarkers(worldSnoId, NewRootTransform()); err != nil {
		return err
	}

	// Load world markers
	if err := e.loadWorldMarkers(worldSnoId, NewRootTransform()); err != nil {
		return err
	}

	// TODO: gbid SpawnLocTypes

	return nil
}

func (e *MarkerExtractor) AddScene(sceneSnoId int32) error {
	e.snoId = sceneSnoId

	// Set map data
	e.data.GridSize = 96
	e.data.Bounds = &pb.Bounds{
		X: 0,
		Y: 0,
	}
	e.data.ZoneArtScale = 1
	e.data.ZoneArtCenter = &pb.Point2D{
		X: 0,
		Y: 0,
	}
	e.data.MaxNativeZoom = 0

	// Load scene
	if err := e.loadSceneMarkers(sceneSnoId, NewRootTransform()); err != nil {
		return err
	}

	return nil
}

func (e *MarkerExtractor) Get() *pb.MapData {
	e.data.Markers = e.uniqueMarkers.Markers
	return e.data
}

func (e *MarkerExtractor) Write() error {
	outputPath := filepath.Join(e.outputPath, fmt.Sprintf("%d.binpb", e.snoId))
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()

	pbb, err := proto.Marshal(e.Get())
	if err != nil {
		return err
	}

	if _, err = f.Write(pbb); err != nil {
		return err
	}

	return nil
}
