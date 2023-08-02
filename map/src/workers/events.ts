
import {Point} from "pixi.js";
import {SnoGroups} from "../data";

export class MarkersReq {
    constructor(
        readonly worldId: number,
        readonly baseUrl: string,
        readonly defaultMarkerColor: number,
        readonly markerColors: Map<String, number>,
        readonly groups: SnoGroups,
    ) {}
}

export class Marker {
    constructor(
        readonly color: number,
        readonly x: number,
        readonly y: number,
        readonly z: number,
        readonly width: number,
        readonly height: number,
        readonly refSnoGroup: number,
        readonly refSnoId: number,
        readonly sourceSno: number,
        readonly dataSnos: number[],
        readonly meta: object,
    ) {}
}

export type MapData = {
    artCenterX: number,
    artCenterY: number,
    zoneArtScale: number,
    gridSize: number,
    maxNativeZoom: number,
    boundsX: number,
    boundsY: number,
}

export class MarkersResp {
    readonly marker: Marker | undefined;
    readonly polygon: Array<Point> | undefined;
    readonly mapData: MapData | undefined;

    constructor(config: {
        marker?: Marker,
        polygon?: Array<Point>
        mapData?: MapData
    }) {
        ({
            marker: this.marker,
            polygon: this.polygon,
            mapData: this.mapData
        } = config);
    }
}