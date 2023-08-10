import {Point} from "pixi.js";
import {Sno} from "./data";

export type WorldReqRetrieve = {
    readonly mapData?: boolean,
    readonly polygons?: boolean,
    readonly markers?: boolean,
}

export type WorldReq = {
    readonly baseUrl: string,
    readonly worldId: Sno.Id,
    readonly retrieve: WorldReqRetrieve,
    readonly query?: string,
}

export type WorldResp = {
    readonly marker: Marker | undefined,
    readonly polygon: Array<Point> | undefined;
    readonly mapData: WorldMapData | undefined;
}

export type Marker = {
    readonly color: number,
    readonly x: number,
    readonly y: number,
    readonly z: number,
    readonly w: number,
    readonly h: number,
    readonly ref: Sno.DisplayInfo,
    readonly source: Sno.DisplayInfo
    readonly data: Sno.DisplayInfo[],
    readonly meta: string[][],
}

export type WorldMapData = {
    readonly artCenterX: number,
    readonly artCenterY: number,
    readonly zoneArtScale: number,
    readonly gridSize: number,
    readonly maxNativeZoom: number,
    readonly boundsX: number,
    readonly boundsY: number,
}
