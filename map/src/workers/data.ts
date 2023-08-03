import {LRUCache} from 'typescript-lru-cache';
import {unpack} from "msgpackr";
import {error} from "jquery";

// Sno types
export namespace Sno {
    export type Id = number;
    export type Name = number;
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
        console.log("msgpack request failed", resp);
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

export async function groups(baseUrl: string): Promise<Sno.Groups> {
    if (!groupsCache) {
        console.log("fetching groups");
        groupsCache = await getMsgpack(`${baseUrl}/groups.mpk`);
    }
    return groupsCache!;
}

export async function names(baseUrl: string): Promise<Sno.Names> {
    if (!namesCache) {
        console.log("fetching names");
        namesCache = await getMsgpack(`${baseUrl}/names.mpk`);
    }
    return namesCache!;
}

export async function snoGroupName(baseUrl: string, id: Sno.GroupId, gs?: Sno.Groups): Promise<Sno.GroupName> {
    if (id === 255) {
        return "Unknown";
    }
    gs ??= await groups(baseUrl);
    return gs[id] ?? `Group_${id}`;
}

export async function lookupSnoGroup(baseUrl: string, id: Sno.Id, ns?: Sno.Names): Promise<Sno.GroupId> {
    ns ??= await names(baseUrl);
    for (let [groupId, m] of Object.entries(ns)) {
        if (m.hasOwnProperty(id)) {
            return Number(groupId)
        }
    }
    return -1;
}

export async function snoTitle(baseUrl: string, group: Sno.GroupId, id: Sno.Id, gs?: Sno.Groups, ns?: Sno.Names): Promise<string> {
    ns ??= await names(baseUrl);

    if (group > 250 || !ns.hasOwnProperty(group)) {
        return `[Unknown] ${id === -1 ? 'Unknown' : id}`;
    }

    const groupName = await snoGroupName(baseUrl, group, gs);
    const groupNames = ns[group];

    if (!groupNames || !groupNames.hasOwnProperty(id)) {
        return `[${groupName}] ${id}`
    }

    return `[${groupName}] ${groupNames[id]}`
}

export async function getDisplayInfo(baseUrl: string, id: Sno.Id, group?: Sno.GroupId, gs?: Sno.Groups, ns?: Sno.Names): Promise<Sno.DisplayInfo> {
    group ??= await lookupSnoGroup(baseUrl, id, ns);
    return {
        title: await snoTitle(baseUrl, group, id, gs, ns),
        id,
    } as Sno.DisplayInfo;
}

// Marker data
export const defaultMarkerColor = 0x495057;
export const markerColors = new Map<string, number>([
    ['Actor', 0x2b8a3e],
    ['AmbientSound', 0x9775fa],
    ['Encounter', 0xc92a2a],
    ['EffectGroup', 0x1864ab],
    ['FogVolume', 0x38d9a9],
    ['Light', 0xfcc419],
    ['MarkerSet', 0xdee2e6],
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