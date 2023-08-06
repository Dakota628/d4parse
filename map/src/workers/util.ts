import {WorldMap} from "../world-map";
import {WorldReq, WorldReqRetrieve, WorldResp} from "./events";
import {Vec2} from "../util";
import $ from "jquery";

export function createWorldWorker(map: WorldMap, doneCb?: () => void): Worker {
    const worker = new Worker(
        new URL('./world.ts', import.meta.url),
        {
            type: "module",
        },
    )

    let gotMapData = false;
    worker.onmessage = (e: MessageEvent<WorldResp>) => {
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
            gotMapData = true;
        } else {
            // Done
            map.redraw(gotMapData);
            $("#loading").hide();
            $("#map").trigger('focus');
            gotMapData = false;

            if (doneCb) {
                doneCb();
            }
        }
    };

    return worker
}

export function loadWorld(
    map: WorldMap,
    worldWorker: Worker,
    worldId: number,
    retrieve: WorldReqRetrieve = {
        mapData: true,
        polygons: true,
        markers: true,
    },
    query?: string,
) {
    $("#loading").show();

    let baseUrl = new URL('.', window.location.toString()).toString();
    if (baseUrl.charAt(baseUrl.length - 1) == '/') {
        baseUrl = baseUrl.slice(0, -1);
    }

    map.config.getTileUrl = (tileCoord: Vec2, zoom: number): string => {
        return `${baseUrl}/${worldId}/${zoom}/${tileCoord.x}_${tileCoord.y}.png`
    };

    worldWorker.postMessage({
        baseUrl,
        worldId,
        retrieve,
        query,
    } as WorldReq)
}