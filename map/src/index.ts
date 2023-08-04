import {Application, BaseTexture, ENV, MIPMAP_MODES, SCALE_MODES, settings} from "pixi.js";
import {WorldMap} from "./world-map";
import {Stats} from "stats.ts";
import {Vec2} from "./util";
import {createWorldWorker, loadWorld} from "./workers/util";
import $ from "jquery";
import '@selectize/selectize';
import {names} from "./workers/data";

// Setup constants
const worldSnoGroup = 48;
const sceneSnoGroup = 33;
let currentWorldId = 69068;


//
// Setup pixijs
//
settings.PREFER_ENV = ENV.WEBGL2;
settings.FAIL_IF_MAJOR_PERFORMANCE_CAVEAT = true;

BaseTexture.defaultOptions.mipmap = MIPMAP_MODES.ON;
BaseTexture.defaultOptions.scaleMode = SCALE_MODES.NEAREST;

//
// Create canvas
//
const view = document.createElement("canvas");
view.setAttribute('id', 'map');
document.body.appendChild(view);

//
// Create stats
//
const stats = new Stats();
stats.showPanel(0);
stats.dom.id = 'stats'
document.body.appendChild(stats.dom);

//
// Create pixi app
//
export const app = new Application({
    view,
    width: window.innerWidth,
    height: window.innerHeight,
    antialias: false,
    autoDensity: true,
    backgroundColor: 0x0,
    resolution: window.devicePixelRatio,
    powerPreference: 'high-performance',
});

app.renderer.resize(window.innerWidth, window.innerHeight);

//
// Create world map
//
const map = new WorldMap(app, {
    stats,
    tileSize: new Vec2(512, 512),
    bounds: new Vec2(0, 0),
    minNativeZoom: 0,
    maxNativeZoom: 0,
    getTileUrl: () => '',
    onMarkerClick: (marker, global, _) => {
        // Show tooltip
        const tooltip = $("#tooltip");
        tooltip.css('left', global.x - 2);
        tooltip.css('top', global.y - 2);
        tooltip.show();

        // Update tooltip title
        $("#tooltip-title").html(`<a class="snoRef" href="../sno/${marker.ref.id}.html">${marker.ref.title}</a>`);

        // Update tooltip body
        const body = $("#tooltip-body");
        body.empty();

        const dl = $("<dl></dl>");

        // -- Source
        dl.append(`<dt>Source</dt><dd><a class="snoRef"  href="../sno/${marker.source.id}.html">${marker.source.title}</a></dd>`);

        // -- Data SNOs
        if (marker.data.length > 0) {
            dl.append('<dt>Data</dt>');
            const dd = $('<dd></dd>');
            for (let data of marker.data) {
                dd.append(`<a class="snoRef" href="../sno/${data.id}.html">${data.title}</a>`);
            }
            dl.append(dd);
        }

        // -- Metadata
        for (const [k, v] of marker.meta) {
            dl.append(`<dt>${k}</dt><dd>${v}</dd>`);
        }

        body.append(dl);

        // -- Coordinates
        body.append(`<div class="coords">${marker.x.toFixed(6)}, ${marker.y.toFixed(6)}, ${marker.z.toFixed(6)}</div>`);
    },
    crs: {
        rotation: (Math.PI / 180) * 45,
        offset: new Vec2(0, 0),
        gridSize: new Vec2(0, 0),
        scale: new Vec2(0, 0),
    }
});

//
// Start app
//
app.ticker.start();

window.addEventListener("resize", () => {
    map.resize(window.innerWidth, window.innerHeight);
});

//
// Load world
//
const worker = createWorldWorker(map);
loadWorld(map, worker, currentWorldId);

//
// Tooltip Handlers
//
const hideTooltip = () => $("#tooltip").hide();
$("#tooltip-close").on('click', hideTooltip);
map.tileContainer.on('mousedown', hideTooltip);
map.tileContainer.on('wheel', hideTooltip);

//
// Search handlers
//
(<any>window).onSearch = (e: any) => {
    let query: string | undefined = $(e).val()?.toString().toLowerCase();
    query = query == '' ? undefined : query;

    map.clearMarkers();
    loadWorld(map, worker, currentWorldId, {markers: true}, query);
};

//
// World switch handlers
//
// Create options
names('').then((names) => {
    const options = new Array<any>();

    for (const [snoId, snoName] of Object.entries(names[worldSnoGroup])) {
        options.push({
            group: 'World',
            label: snoName,
            value: snoId,
        })
    }

    for (const [snoId, snoName] of Object.entries(names[sceneSnoGroup])) {
        options.push({
            group: 'Scene',
            label: snoName,
            value: snoId,
        })
    }

    // Create selectize
    const $worldSelect = $('#world-select').selectize({
        create: false,
        persist: false,
        sortField: [
            { field: 'group', direction: 'desc' },
            { field: 'label', direction: 'asc' },
        ],
        valueField: 'value',
        labelField: 'label',
        searchField: ['group', 'label'],
        optgroupField: 'group',
        optgroupLabelField: 'group',
        optgroupValueField: 'group',
        optgroupOrder: ['World', 'Scene'],
        lockOptgroupOrder: true,
        optgroups: [
            { group: 'World' },
            { group: 'Scene' },
        ],
        options,
        onChange: (worldId) => {
            if (worldId) {
                currentWorldId = worldId;
                map.clear();
                loadWorld(map, worker, worldId);
            }
        }
    });

    // Set default value
    $worldSelect[0].selectize.setValue(currentWorldId, true);
});