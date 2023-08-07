import {LRUCache} from 'typescript-lru-cache';
import {unpack} from "msgpackr";
import {error} from "jquery";

export const docsBaseUrl = "https://docs.diablo.farm"

// Sno types
export namespace Sno {
    export type Id = number;
    export type Name = string;
    export type GroupId = number;
    export type GroupName = string;
    export type Groups = {
        [id: GroupId]: GroupName
    };
    export type Names = {
        [groupId: GroupId]: {
            [snoId: Id]: Name,
        }
    }
    export type DisplayInfo = {
        title: string,
        id: Id,
    }
}


// Utility
export async function getMsgpack(url: string): Promise<any> {
    const resp = await fetch(url);
    if (!resp.ok) {
        throw error("Msgpack request failed");
    }
    return unpack(await resp.arrayBuffer());
}

// World data
const worldDataCache = new LRUCache<number, any>({
    maxSize: 3,
});

export async function getWorldData(baseUrl: string, worldId: number): Promise<any> {
    if (worldDataCache.has(worldId)) {
        return worldDataCache.get(worldId);
    }

    const result = await getMsgpack(`${baseUrl}/data/${worldId}.mpk`);
    worldDataCache.set(worldId, result);
    return result;
}

// Groups and names
let groupsCache: Sno.Groups | undefined;
let namesCache: Sno.Names | undefined;

export async function groups(): Promise<Sno.Groups> {
    if (!groupsCache) {
        groupsCache = await getMsgpack(`${docsBaseUrl}/groups.mpk`);
    }
    return groupsCache!;
}

export async function names(): Promise<Sno.Names> {
    if (!namesCache) {
        namesCache = await getMsgpack(`${docsBaseUrl}/names.mpk`);
    }
    return namesCache!;
}

export async function snoGroupName(id: Sno.GroupId, gs?: Sno.Groups): Promise<Sno.GroupName> {
    if (id === 255) {
        return "Unknown";
    }
    gs ??= await groups();
    return gs[id] ?? `Group_${id}`;
}

export async function lookupSnoGroup(id: Sno.Id, ns?: Sno.Names): Promise<Sno.GroupId> {
    ns ??= await names();
    for (let [groupId, m] of Object.entries(ns)) {
        if (m.hasOwnProperty(id)) {
            return Number(groupId)
        }
    }
    return -1;
}

export async function snoName(group: Sno.GroupId, id: Sno.Id, ns?: Sno.Names): Promise<string | undefined> {
    ns ??= await names();
    return (ns[group] ?? {})[id] ?? undefined;
}

export async function snoTitle(group: Sno.GroupId, id: Sno.Id, gs?: Sno.Groups, ns?: Sno.Names): Promise<string> {
    ns ??= await names();

    if (group > 250 || !ns.hasOwnProperty(group)) {
        return `[Unknown] ${id === -1 ? 'Unknown' : id}`;
    }

    const groupName = await snoGroupName(group, gs);
    const name = await snoName(group, id);
    return `[${groupName}] ${name ?? id}`
}

export async function getDisplayInfo(id: Sno.Id, group?: Sno.GroupId, gs?: Sno.Groups, ns?: Sno.Names): Promise<Sno.DisplayInfo> {
    group ??= await lookupSnoGroup(id, ns);
    return {
        title: await snoTitle(group, id, gs, ns),
        id,
    } as Sno.DisplayInfo;
}

// Marker data
export const defaultMarkerColor = 0x000000;
export const markerColors = new Map<string, number>([
    ['Actor', 0x00ff00],
    ['AmbientSound', 0x9775fa],
    ['Encounter', 0xff0000],
    ['EffectGroup', 0x1864ab],
    ['FogVolume', 0x38d9a9],
    ['Light', 0xfcc419],
    ['MarkerSet', 0x0000ff],
    ['Material', 0xe8590c],
    ['Particle', 0x7b3f00],
    ['Quest', 0x74c0fc],
    ['Sound', 0x862e9c],
    ['Unknown', defaultMarkerColor],
    ['Weather', 0xc5f6fa],
]);

export const markerMetaNames = new Map<string, string>([
    ['mt', 'Marker Type'],
    ['gt', 'Gizmo Type'],
])