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

type MarkerExtractor struct {
	dumpPath     string
	baseMetaPath string
	outputPath   string
	toc          d4.Toc

	data          *pb.MapData
	uniqueMarkers *UniqueMarkerData
	snoId         int32
}

func NewMarkerExtractor(dumpPath string, outputPath string, toc d4.Toc) *MarkerExtractor {
	return &MarkerExtractor{
		dumpPath:     dumpPath,
		baseMetaPath: filepath.Join(dumpPath, "base", "meta"),
		outputPath:   outputPath,
		toc:          toc,

		data:          &pb.MapData{},
		uniqueMarkers: NewUniqueMarkerData(),
	}
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

func (e *MarkerExtractor) loadGlobalMarkers(worldSnoId int32) error {
	metaPath := util.BaseFilePath(e.dumpPath, util.FileTypeMeta, d4.SnoGroupGlobal, "global_markers")
	meta, err := d4.ReadSnoMetaFile(metaPath)
	if err != nil {
		return err
	}

	gd, ok := meta.Meta.(*d4.GlobalDefinition)
	if !ok {
		return errors.New("not global definition")
	}

	gd.Walk(func(k string, v d4.Object, next d4.WalkNext, d ...any) {
		// Get global marker actor
		gma, ok := v.(*d4.GlobalMarkerActor)
		if !ok || gma == nil {
			next()
			return
		}

		// If it's not in the world, skip
		if gma.SnoWorld.Id != worldSnoId {
			return
		}

		// Get referenced sno name
		refSnoGroup, _ := e.toc.Entries.GetName(gma.SnoActor.Id)

		// Add marker data
		e.uniqueMarkers.Add(&pb.MarkerData{
			RefSnoGroup: int32(refSnoGroup),
			RefSno:      gma.SnoActor.Id,
			SourceSno:   meta.Id.Value,
			Position: &pb.Point3D{
				X: gma.TWorldTransform.Wp.X,
				Y: gma.TWorldTransform.Wp.Y,
				Z: gma.TWorldTransform.Wp.Z,
			},
			Extra: &pb.ExtraMarkerData{
				GizmoType: &gma.EGizmoType.Value,
			},
		})
	})

	return nil
}

func (e *MarkerExtractor) maxNativeZoom(tiles int32) (zoom uint32) {
	for ; tiles%2 == 0 && tiles > 0; tiles /= 2 {
		zoom++
	}
	return
}

func (e *MarkerExtractor) addMarker(
	marker *d4.Marker,
	sourceSno int32,
	baseX float32,
	baseY float32,
	baseZ float32,
) error {
	x := baseX + marker.Transform.Wp.X
	y := baseY + marker.Transform.Wp.Y
	z := baseZ + marker.Transform.Wp.Z

	// Add the marker
	e.uniqueMarkers.Add(&pb.MarkerData{
		RefSnoGroup: marker.Snoname.Group,
		RefSno:      marker.Snoname.Id,
		SourceSno:   sourceSno,
		DataSnos:    e.getDataSnos(marker),
		Position: &pb.Point3D{
			X: x,
			Y: y,
			Z: z,
		},
		Extra: &pb.ExtraMarkerData{
			MarkerType: &marker.EType.Value,
			Scale: &pb.Point3D{
				X: marker.VecScale.X,
				Y: marker.VecScale.Y,
				Z: marker.VecScale.Z,
			},
		},
	})

	// If it's a nested MarkerSet, unwrap it
	if d4.SnoGroup(marker.Snoname.Group) == d4.SnoGroupMarkerSet {
		if _, markerSetName := e.toc.Entries.GetName(marker.Snoname.Id, d4.SnoGroupMarkerSet); markerSetName != "" {
			metaPath := util.BaseFilePath(e.dumpPath, util.FileTypeMeta, d4.SnoGroupMarkerSet, markerSetName)
			meta, err := d4.ReadSnoMetaFile(metaPath)
			if err != nil {
				return err
			}

			if err := e.addMarkerSet(meta, x, y, z); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *MarkerExtractor) addMarkerSet(
	markerSetSnoMeta d4.SnoMeta,
	baseX float32,
	baseY float32,
	baseZ float32,
) error {
	msd, ok := markerSetSnoMeta.Meta.(*d4.MarkerSetDefinition)
	if !ok {
		panic("not marker set definition")
	}

	errs := util.NewErrors()
	msd.Walk(func(k string, v d4.Object, next d4.WalkNext, d ...any) {
		defer next()

		marker, ok := v.(*d4.Marker)
		if !ok {
			return
		}

		if err := e.addMarker(marker, markerSetSnoMeta.Id.Value, baseX, baseY, baseZ); err != nil {
			errs.Add(err)
			return
		}
	})

	return errs.Err()
}

func (e *MarkerExtractor) loadWorldRelatedMarkers(worldName string) error {
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

		if err := e.addMarkerSet(markerSetSnoMeta, 0, 0, 0); err != nil {
			errs.Add(err)
			return
		}
	})

	return errs.Err()
}

func (e *MarkerExtractor) loadSubzoneMarkers(subZoneId int32) error {
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

	// Add related marker sets
	relatedMarkerSetPattern := util.BaseFilePattern(e.dumpPath, util.FileTypeMeta, d4.SnoGroupMarkerSet, subZoneName, "", " (*)")
	relatedMarkerSetPaths, err := doublestar.FilepathGlob(relatedMarkerSetPattern)
	if err != nil {
		return err
	}

	seenMarkerSet := mapset.NewSet[string]()
	errs := util.NewErrors()

	util.DoWorkSlice(adHocWorkers, relatedMarkerSetPaths, func(markerSetSnoPath string) {
		markerSetSnoMeta, err := d4.ReadSnoMetaFile(markerSetSnoPath)
		if err != nil {
			panic(err)
		}

		if !seenMarkerSet.Add(markerSetSnoPath) {
			return
		}

		msd, ok := markerSetSnoMeta.Meta.(*d4.MarkerSetDefinition)
		if !ok {
			panic("not marker set definition")
		}

		msd.Walk(func(k string, v d4.Object, next d4.WalkNext, d ...any) {
			defer next()

			marker, ok := v.(*d4.Marker)
			if !ok {
				return
			}

			if err := e.addMarker(marker, markerSetSnoMeta.Id.Value, 0, 0, 0); err != nil {
				errs.Add(err)
				return
			}
		})
	})

	// Add subzone marker sets
	sd.Walk(func(k string, v d4.Object, next d4.WalkNext, d ...any) {
		szMsEntry, ok := v.(*d4.SubzoneWorldMarkerSetEntry)
		if !ok {
			next()
			return
		}

		markerSetSnoGroup, markerSetSnoName := e.toc.Entries.GetName(szMsEntry.SnoMarkerSet.Id, d4.SnoGroupMarkerSet)
		markerSetSnoPath := util.BaseFilePath(e.dumpPath, util.FileTypeMeta, markerSetSnoGroup, markerSetSnoName)
		markerSetSnoMeta, err := d4.ReadSnoMetaFile(markerSetSnoPath)
		if err != nil {
			errs.Add(err)
			return
		}

		if !seenMarkerSet.Add(markerSetSnoPath) {
			next()
			return
		}

		if err := e.addMarkerSet(markerSetSnoMeta, 0, 0, 0); err != nil {
			errs.Add(err)
			return
		}
	})

	return errs.Err()
}

func (e *MarkerExtractor) loadSceneMarkers(sceneId int32, baseX float32, baseY float32, baseZ float32) error {
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
			if err := e.addMarker(marker, markerSetSnoMeta.Id.Value, baseX, baseY, baseZ); err != nil {
				errs.Add(err)
				return
			}
		}
	})

	return nil
}

func (e *MarkerExtractor) loadSceneChunkMarkers(sceneChunk *d4.SceneChunk) error {
	// TODO: are there any cases where we need to apply the quaternion?
	return e.loadSceneMarkers(
		sceneChunk.Snoname.Id,
		sceneChunk.Transform.Wp.X,
		sceneChunk.Transform.Wp.Y,
		sceneChunk.Transform.Wp.Z,
	)
}

func (e *MarkerExtractor) loadWorldMarkers(worldId int32) error {
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

	// Load related markers
	if err := e.loadWorldRelatedMarkers(worldSnoName); err != nil {
		return err
	}

	// Load subzone markers
	if wd.SnoSubzoneDefault.Id > 0 {
		if err := e.loadSubzoneMarkers(wd.SnoSubzoneDefault.Id); err != nil {
			return err
		}
	}

	// Load additional data from walking world definition
	var sceneChunks []*d4.SceneChunk
	var subZoneIds []int32

	wd.Walk(func(k string, v d4.Object, next d4.WalkNext, d ...any) {
		switch x := v.(type) {
		case *d4.ScreenStaticCamps:
			// Add polygons
			poly := &pb.Polygon{}
			for _, vec := range x.ArPoints.Value {
				poly.Vertices = append(poly.Vertices, &pb.Point2D{
					X: vec.X,
					Y: vec.Y,
				})
			}
			e.data.Polygons = append(e.data.Polygons, poly)
		case *d4.Type_bef5a4a:
			// TODO: conditional map overlays
		case *d4.SceneChunk:
			// Add scene chunk for async processing
			sceneChunks = append(sceneChunks, x)
		case *d4.SceneSpecification:
			// Add subzone id for async processing
			for _, subZone := range x.ArSubzones.Value {
				subZoneIds = append(subZoneIds, subZone.SnoSubzone.Id)
			}
		}

		next()
	})

	// Load some data async
	errs := util.NewErrors()

	util.DoWorkSlice(adHocWorkers, sceneChunks, func(sceneChunk *d4.SceneChunk) {
		if err := e.loadSceneChunkMarkers(sceneChunk); err != nil {
			errs.Add(err)
			return
		}
	})

	util.DoWorkSlice(adHocWorkers, subZoneIds, func(subZoneId int32) {
		if err := e.loadSubzoneMarkers(subZoneId); err != nil {
			errs.Add(err)
			return
		}
	})

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
	e.data.MaxNativeZoom = e.maxNativeZoom(wd.TZoneMapParams.Unk_3620f37.Value)

	return errs.Err()
}

func (e *MarkerExtractor) AddWorld(worldSnoId int32) error {
	e.snoId = worldSnoId

	// Load global markers
	if err := e.loadGlobalMarkers(worldSnoId); err != nil {
		return err
	}

	// Load world markers
	if err := e.loadWorldMarkers(worldSnoId); err != nil {
		return err
	}

	// TODO: also load quest markers?

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
	if err := e.loadSceneMarkers(sceneSnoId, 0, 0, 0); err != nil {
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
