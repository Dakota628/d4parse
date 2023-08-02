import {Application, BaseTexture, ENV, MIPMAP_MODES, SCALE_MODES, settings} from "pixi.js";
import {WorldMap} from "./world-map";
import {Stats} from "stats.ts";
import {ClosestPoint, Vec2} from "./util";
import {getWorker, loadWorld} from "./workers/util";
import $ from "jquery";
import {groups, lookupSnoGroup, markerMetaNames, snoName} from "./data";

// Setup pixijs
settings.PREFER_ENV = ENV.WEBGL2;
settings.FAIL_IF_MAJOR_PERFORMANCE_CAVEAT = true;

BaseTexture.defaultOptions.mipmap = MIPMAP_MODES.ON;
BaseTexture.defaultOptions.scaleMode = SCALE_MODES.NEAREST;

// Create canvas
const view = document.createElement("canvas");
document.body.appendChild(view);

// Create stats
const stats = new Stats();
stats.showPanel(0);
stats.dom.id = 'stats'
document.body.appendChild(stats.dom);

// Create pixi app
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

// Create world map
const map = new WorldMap(app, {
    stats,
    tileSize: new Vec2(512, 512),
    bounds: new Vec2(40, 40),
    minNativeZoom: 0,
    maxNativeZoom: 0,
    getTileUrl: (): string => {
        return ''
    },
    onMarkerClick: (markers, global, local) => {
        const marker = ClosestPoint(local, markers);
        if (!marker) {
            return
        }

        // Show tooltip
        const tooltip = $("#tooltip");
        tooltip.css('left', global.x - 2);
        tooltip.css('top', global.y - 2);
        tooltip.show();

        // Get sno info
        // const groupName = snoGroupName(marker.refSnoGroup);
        const title = snoName(marker.refSnoGroup, marker.refSnoId);

        // Update tooltip title
        $("#tooltip-title").html(`<a class="snoRef popupTitle" href="../sno/${marker.refSnoId}.html">${title}</a>`);

        // Update tooltip body
        const body = $("#tooltip-body");
        body.empty();

        const dl = $("<dl></dl>");

        // -- Source
        dl.append(`<dt>Source</dt><dd><a class="snoRef"  href="../sno/${marker.sourceSno}.html">${snoName(lookupSnoGroup(marker.sourceSno), marker.sourceSno)}</a></dd>`);

        // -- Data SNOs
        const dataSnos = marker.dataSnos ?? [];
        if (dataSnos.length > 0) {
            dl.append('<dt>Data</dt>');
            const dd = $('<dd></dd>');
            for (const dataSno of dataSnos) {
                const title = snoName(lookupSnoGroup(dataSno), dataSno);
                dd.append(`<a class="snoRef" href="../sno/${dataSno}.html">${title}</a>`);
            }
            dl.append(dd);
        }

        // -- Metadata
        for (const [key, val] of Object.entries(marker.meta ?? {})) {
            dl.append(`<dt>${markerMetaNames.get(key) ?? key}</dt><dd>${val}</dd>`);
        }

        body.append(dl);

        // -- Coordinates
        body.append(`<div class="coords">${marker.x.toFixed(6)}, ${marker.y.toFixed(6)}, ${marker.z.toFixed(6)}</div>`);
    },
    crs: {
        rotation: 0,
        offset: new Vec2(0, 0),
        gridSize: new Vec2(0, 0),
        scale: new Vec2(0, 0),
    }
});

// Start app
app.ticker.start();

window.addEventListener("resize", () => {
    map.resize(window.innerWidth, window.innerHeight);
});

// Event handlers
$("#tooltip-close").on('click', () => {
    $("#tooltip").hide();
});

$("#tooltip").on('mouseleave', () => {
    $("#tooltip").hide();
})

// Load world
const worker = getWorker(map);
loadWorld(map, worker, groups, 69068);
