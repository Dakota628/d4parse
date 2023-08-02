import {WorldMap} from "../world-map";
import {MarkersReq, MarkersResp} from "./events";
import {Vec2} from "../util";
import {defaultMarkerColor, markerColors, SnoGroups} from "../data";

export function getWorker(map: WorldMap): Worker {
    const worker = new Worker(
        new URL('./markers.ts', import.meta.url),
        {
            type: "module",
        },
    )

    worker.onmessage = (e: MessageEvent<MarkersResp>) => {
        if (e.data.marker) {
            map.addMarker(e.data.marker);
        } else if (e.data.polygon) {
            map.addPolygon(e.data.polygon);
        } else if (e.data.mapData) {
            map.config.maxNativeZoom = e.data.mapData.maxNativeZoom;
            map.config.bounds = new Vec2(e.data.mapData.boundsX, e.data.mapData.boundsY);
            map.config.crs.offset = new Vec2(e.data.mapData.artCenterX, e.data.mapData.artCenterY);
            map.config.crs.gridSize = new Vec2(e.data.mapData.gridSize, e.data.mapData.gridSize);
            map.config.crs.scale = new Vec2(e.data.mapData.zoneArtScale, e.data.mapData.zoneArtScale);
        } else {
            // Done
            map.draw();
            map.drawMarkers();
        }
    };

    return worker
}

export function loadWorld(map: WorldMap, worker: Worker, groups: SnoGroups, worldId: number) {
    let baseUrl = new URL('.', window.location.toString()).toString();
    if (baseUrl.charAt(baseUrl.length - 1) == '/') {
        baseUrl = baseUrl.slice(0, -1);
    }

    map.config.getTileUrl = (tileCoord: Vec2, zoom: number): string => {
        return `${baseUrl}/${worldId}/${zoom}/${tileCoord.x}_${tileCoord.y}.png`
    };

    worker.postMessage(new MarkersReq(
        worldId,
        baseUrl,
        defaultMarkerColor,
        markerColors,
        groups,
    ));
}