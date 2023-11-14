import {WorldMapData, Marker, WorldReq, WorldResp} from "./events";
import {Point} from "pixi.js";
import {
    defaultMarkerColor,
    getDisplayInfo,
    getWorldData, lookupSnoGroup,
    markerColors,
    Sno,
    snoGroupName, snoName
} from "./data";
import * as liqe from "liqe";

console.log("new world worker!");

self.onmessage = async (e: MessageEvent<WorldReq>) => {
    const data = await getWorldData(e.data.baseUrl, e.data.worldId);

    // Send map data
    if (e.data.retrieve.mapData) {
        self.postMessage({
            mapData: {
                artCenterX: data.zoneArtCenter.x,
                artCenterY: data.zoneArtCenter.y,
                zoneArtScale: data.zoneArtScale,
                gridSize: data.gridSize,
                maxNativeZoom: data.maxNativeZoom,
                boundsX: data.bounds.x,
                boundsY: data.bounds.y,
            } as WorldMapData,
        } as WorldResp);
    }

    // Add polygons
    if (e.data.retrieve.polygons) {
        for (let p of data.polygons) {
            const polygon = new Array<Point>();
            for (const wp of p.vertices) {
                polygon.push(new Point(wp.y, wp.x));
            }

            self.postMessage({
                polygon: polygon,
            } as WorldResp);
        }
    }

    // Add markers
    if (e.data.retrieve.markers) {
        let query: liqe.LiqeQuery | undefined;
        if (e.data.query) {
            try {
                query = liqe.parse(e.data.query);
            } catch (e) {
                console.log("Error parsing search query:", e);
            }
        }

        for (let m of data.markers) {
            const refGroup = await snoGroupName(m.refSnoGroup);

            if (query) {
                const refName = await snoName(m.refSnoGroup, m.refSno);
                const srcGroupId = await lookupSnoGroup(m.sourceSno);
                const srcGroup = await snoGroupName(srcGroupId);
                const srcName = await snoName(srcGroupId, m.sourceSno);

                const searchObj: any = {
                    id: String(m.refSno),
                    group: refGroup,
                    name: refName,
                    source_id: String(m.sourceSno),
                    source_group: srcGroup,
                    source: srcName,
                    marker_hash: String(m.markerHash),
                    marker_group_hash: String(m.markerGroupHashes),
                    marker_type: m.extra.has_markerType ? m.extra.markerType : '',
                };

                if (!liqe.test(query, searchObj)) {
                    continue;
                }
            }

            const color = markerColors.get(refGroup) ?? defaultMarkerColor;

            // noinspection JSSuspiciousNameCombination
            self.postMessage({
                marker: {
                    color,
                    x: m.position.y, // Note: x and y are purposely swapped
                    y: m.position.x, // Note: x and y are purposely swapped
                    z: m.position.z,
                    w: 0.5,
                    h: 0.5,
                    boundX: m.has_extra && m.extra.has_bounds ? m.extra.bounds.offset.x : 0,
                    boundY: m.has_extra && m.extra.has_bounds ? m.extra.bounds.offset.y : 0,
                    boundW: m.has_extra && m.extra.has_bounds ? m.extra.bounds.ext.x : 0,
                    boundH: m.has_extra && m.extra.has_bounds ? m.extra.bounds.ext.y : 0,
                    ref: await getDisplayInfo(m.refSno, m.refSnoGroup),
                    source: await getDisplayInfo(m.sourceSno),
                    data: await Promise.all((m.dataSnos).map(
                        async (id: Sno.Id) => await getDisplayInfo(id),
                    )),
                    meta: [
                        ... m.extra.has_gizmoType ? [['Gizmo Type', m.extra.gizmoType]] : [],
                        ... m.extra.has_markerType ? [['Marker Type', m.extra.markerType]] : [],
                    ],
                } as Marker,
            } as WorldResp);
        }
    }

    // Signal done
    self.postMessage({});
};

export {};