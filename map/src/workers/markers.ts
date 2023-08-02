import {MapData, Marker, MarkersReq, MarkersResp} from "./events";
import {unpack} from "msgpackr";
import {Point} from "pixi.js";
// import {snoGroupName} from "../data";

self.onmessage = (e: MessageEvent<MarkersReq>) => {
    // Get the data
    fetch(`${e.data.baseUrl}/data/${e.data.worldId}.mpk`).then((resp) => {
        resp.arrayBuffer().then((packed) => {
            const data = unpack(packed);

            // Send map data
            self.postMessage(new MarkersResp({
                mapData: data as MapData,
            }));

            // Add polygons
            for (let p of data.p ?? []) {
                const polygon = new Array<Point>();
                for (const wp of p) {
                    polygon.push(new Point(wp[1], wp[0]));
                }

                self.postMessage(new MarkersResp({
                    polygon: polygon,
                }));
            }

            // Add markers
            for (let m of data.m ?? []) {
                // const groupName = snoGroupName(m.g, e.data.groups);
                // const color = e.data.markerColors.get(groupName) ?? e.data.defaultMarkerColor;
                const groupName = e.data.groups[m.g] ?? 'Unknown';
                const color = e.data.markerColors.get(groupName) ?? e.data.defaultMarkerColor;

                self.postMessage(new MarkersResp({
                    marker: new Marker(
                        color,
                        m.y,
                        m.x,
                        m.z,
                        1,
                        1,
                        m.g,
                        m.r,
                        m.s,
                        m.d,
                        m.m,
                    )
                }));
            }

            // Signal done
            self.postMessage(new MarkersResp({}));
        });
    });
};

export {};