import { unpack } from 'msgpackr';

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

export async function getMsgpack(url: string): Promise<any> {
    const resp = await fetch(url);
    return unpack(await resp.arrayBuffer())
}

export type SnoId = number;
export type SnoName = number;
export type SnoGroupId = number;
export type SnoGroupName = string;
export type SnoGroups = {
    [id: SnoGroupId]: SnoGroupName
};
export type SnoNames = {
    [groupId: SnoGroupId]: {
        [snoId: SnoId]: SnoName,
    }
}

export const groups = await getMsgpack('/groups.mpk') as SnoGroups;

export const names = await getMsgpack('/names.mpk') as SnoNames;

export function snoGroupName(id: SnoGroupId, groupsOverride: SnoGroups | undefined = undefined): SnoGroupName {
    if (id === 255) {
        return "Unknown";
    }
    return (groupsOverride ?? groups)[id] ?? `Group_${id}`;
}

export function lookupSnoGroup(id: SnoGroupId) {
    for (let [groupId, m] of Object.entries(names)) {
        if (m.hasOwnProperty(id)) {
            return Number(groupId)
        }
    }
    return -1;
}


export function snoName(group: SnoGroupId, id: SnoId) {
    if (group > 250 || !names.hasOwnProperty(group)) {
        return `[Unknown] ${id === -1 ? 'Unknown' : id}`;
    }

    const groupName = snoGroupName(group);
    const groupNames = names[group];

    if (!groupNames || !groupNames.hasOwnProperty(id)) {
        return `[${groupName}] ${id}`
    }

    return `[${groupName}] ${groupNames[id]}`
}
