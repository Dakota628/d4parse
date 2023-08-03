import {MapData, Marker, WorldReq, WorldResp} from "./events";
import {Point} from "pixi.js";
import {
    defaultMarkerColor,
    getDisplayInfo,
    getWorldData,
    markerColors,
    markerMetaNames,
    Sno,
    snoGroupName
} from "./data";

console.log("new world worker!");

self.onmessage = async (e: MessageEvent<WorldReq>) => {
    const data = await getWorldData(e.data.baseUrl, e.data.worldId);

    // Send map data
    if (e.data.retrieve.mapData) {
        self.postMessage({
            mapData: data as MapData,
        } as WorldResp);
    }

    // Add polygons
    if (e.data.retrieve.polygons) {
        for (let p of data.p ?? []) {
            const polygon = new Array<Point>();
            for (const wp of p) {
                polygon.push(new Point(wp[1], wp[0]));
            }

            self.postMessage({
                polygon: polygon,
            } as WorldResp);
        }
    }

    // Add markers
    if (e.data.retrieve.markers) {
        let size = 0.5;
        if (e.data.query) {
            size = 10;
        }

        for (let m of data.m ?? []) {
            const groupName = await snoGroupName(e.data.baseUrl, m.g);
            const color = markerColors.get(groupName) ?? defaultMarkerColor;

            const ref = await getDisplayInfo(e.data.baseUrl, m.r, m.g);
            if (e.data.query && !ref.title.toLowerCase().includes(e.data.query)) {
                continue;
            }

            // noinspection JSSuspiciousNameCombination
            self.postMessage({
                marker: {
                    color,
                    x: m.y, // Note: x and y are purposely swapped
                    y: m.x, // Note: x and y are purposely swapped
                    z: m.z,
                    w: size,
                    h: size,
                    ref,
                    source: await getDisplayInfo(e.data.baseUrl, m.s),
                    data: await Promise.all((m.d ?? []).map(
                        async (id: Sno.Id) => await getDisplayInfo(e.data.baseUrl, id),
                    )),
                    meta: Object.entries(m.m ?? {}).map(
                        ([k, v]) => [markerMetaNames.get(k) ?? k, v]
                    ),
                } as Marker,
            } as WorldResp);
        }
    }

    // Signal done
    self.postMessage({});
};

export {};