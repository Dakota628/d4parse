// From ZoneMapParams in WorldDefinition
const zoneArtScale = 1.066667; // tZoneMapParams.fZoneArtScale
const zoneArtCenter = [1449.000000, 2909.000000]; // tZoneMapParams.vecZoneArtCenter
const zoneMapParamsScale = 5; // Scale of texture relative to zone map params

// Pixels <-> Leaflet map units
const mapUnitPerTile = 64;
const mapSize = 40;
const tileSize = 512;
const pxPerMapUnit = tileSize / mapUnitPerTile;

function pxToMapUnit(px) {
    return [px[0] / pxPerMapUnit, px[1] / pxPerMapUnit];
}

function subPx(a, b) {
    return [a[0] - b[0], a[1] - b[1]];
}

// Calculated constants
const min = [0, 0];
const max = [tileSize * mapSize, tileSize * mapSize];
const origin = [zoneArtCenter[0] * zoneMapParamsScale, zoneArtCenter[1] * zoneMapParamsScale];
const ptScale = 1 + ((1 - zoneArtScale) * zoneMapParamsScale);

const originMapUnits = pxToMapUnit(origin);
const minMapUnits = subPx(pxToMapUnit(min), originMapUnits);
const maxMapUnits = subPx(pxToMapUnit(max), originMapUnits);

// D4 CRS
const D4Projection = L.extend({}, L.Projection.LonLat, {
    project: function (latlng) {
        let point = L.Projection.LonLat.project(latlng);
        return rotate(point, -45);
    },
    unproject: function (point) {
        point = rotate(point, 45);
        return L.Projection.LonLat.unproject(point);
    },
});

const D4CRS = L.extend({}, L.CRS.Simple, {
    projection: D4Projection,
    transformation: new L.Transformation(ptScale, originMapUnits[0], ptScale, originMapUnits[1]),
});

// Markers
const markerColors = {
    'Actor': 'green',
    'AmbientSound': 'lightpurple',
    'Encounter': 'red',
    'EffectGroup': 'blue',
    'FogVolume': 'mintcream',
    'Light': 'yellow',
    'MarkerSet': 'lightgray',
    'Material': 'orange',
    'Particle': 'brown',
    'Quest': 'lightblue',
    'Sound': 'purple',
    'Unknown': 'darkgray',
    'Weather': 'lightskyblue',
};

// Main
window.addEventListener("DOMContentLoaded", () => {
    // Setup renderer
    window.canvas = L.canvas();

    // Setup map
    window.map = L.map('map', {
        attributionControl: false,
        crs: D4CRS,
        renderer: canvas,
        maxBounds: L.latLngBounds( // Basically magic at this point
            L.latLng(-970, -2890),
            L.latLng(-970, 2545),
        )
    }).setView([0, 0], 0);

    // Setup tiles
    window.tiles = L.tileLayer('maptiles/{z}/{x}_{y}.png', {
        tileSize: tileSize,
        maxZoom: 6,
        minZoom: -1,
        minNativeZoom: 0,
        maxNativeZoom: 3,
        noWrap: true,
        tms: false,
    }).addTo(map);

    // Add map events
    map.on('click', function (e) {
        L.popup()
            .setLatLng(e.latlng)
            .setContent(`${e.latlng.lat}, ${e.latlng.lng}`)
            .openOn(map);
    });

    // Add markers
    L.circleMarker([0, 0], {
        radius: 5,
        stroke: false,
        fill: true,
        fillOpacity: 0.75,
        fillColor: "black",
    }).bindTooltip("This is the center of the world!").addTo(map);

    window.dataLayers = L.control.layers([], []).addTo(map);

    loadData((data) => {
        const p = data.mapData.p;
        const m = data.mapData.m;

        // Polygons
        let len = p.length;
        while (len--) { // Using while has a measurable performance improvement... bc Javascript.
            L.polygon(p[len], {
                weight: 3,
                color: '#ffffff',
                fill: false,
                opacity: 0.1,
                interactive: false,
            }).addTo(map)
        }
        // Markers
        const groups = {};

        len = m.length;
        while (len--) {
            const marker = m[len];
            const groupName = snoGroupName(data.groups, marker.g);
            const title = snoName(data.groups, data.names, marker.g, marker.r);
            const color = markerColors[groupName];

            const circle = L.circleMarker([marker.x, marker.y], {
                radius: 5,
                stroke: false,
                fill: true,
                fillOpacity: 0.75,
                fillColor: markerColors[groupName],
            }).bindPopup(
                markerPopup(marker, title),
                {direction: 'center'},
            );

            if (!groups.hasOwnProperty(groupName)) {
                groups[groupName] = L.markerClusterGroup({
                    spiderfyOnMaxZoom: false,
                    removeOutsideVisibleBounds: true,
                    disableClusteringAtZoom: 3,
                    iconCreateFunction: function (cluster) {
                        var childCount = cluster.getChildCount();
                        return new L.DivIcon({
                            html: `<div>${childCount}</div>`,
                            className: `cluster group_${groupName}`,
                            iconSize: new L.Point(40, 40),
                        });
                    }
                });
            }
            groups[groupName].addLayer(circle);
        }

        for (const group of Object.keys(groups).sort()) {
            dataLayers.addOverlay(groups[group], group);
        }

        // Remove loading screen
        document.getElementById("loading").style.display = 'none';
    });

    // TODO: add overlays for quest conditioned map updates
    // TODO: add radius (on hover) around markers with a radius
    // TODO: add rotated and non-rotated grid
    // TODO: filter by gizmo type
    // TODO: expand marker sets on click
    // TODO: custom search with Fuse?
});

function rotate(p, angle) {
    const rads = (Math.PI / 180) * angle;
    const cos = Math.cos(rads);
    const sin = Math.sin(rads);
    return L.point(
        (cos * p.x) + (sin * p.y),
        (cos * p.y) - (sin * p.x)
    );
}

function markerPopup(marker, title) {
    let extra = '';
    const meta = marker.m ?? {};
    if (meta.hasOwnProperty('mt')) {
        extra += `Marker Type: ${meta.mt}<br/>`
    }
    if (meta.hasOwnProperty('gt')) {
        extra += `Gizmo Type: ${meta.gt}<br/>`
    }

    return `<b><a href="../sno/${marker.r}.html">${title}</a></b>
    <p>
    Source: <a href="../sno/${marker.s}.html">${marker.s}</a>
    <br/>
    ${extra}
    <br/>
    <i>${marker.x}, ${marker.y}, ${marker.z}</i>
    </p>`;
}

function snoGroupName(groups, id) {
    if (id === 255) {
        return "Unknown";
    }
    return groups[id] ?? `Group_${id}`;
}

function snoName(groups, names, group, id) {
    if (group > 250 || !names.hasOwnProperty(group)) {
        return `[Unknown] ${id === -1 ? 'Unknown' : id}`;
    }

    const groupName = snoGroupName(groups, group);
    names = names[group];

    if (!names.hasOwnProperty(id)) {
        return `[${groupName}] ${id}`
    }

    return `[${groupName}] ${names[id]}`
}

function loadData(cb) {
    Promise.all([
        binaryRequest('GET', 'map.mpk'),
        binaryRequest('GET', '../groups.mpk'),
        binaryRequest('GET', '../names.mpk'),
    ]).then((values) => {
        const mapData = msgpackr.unpack(values[0].currentTarget.response);
        const groups = msgpackr.unpack(values[1].currentTarget.response);
        const names = msgpackr.unpack(values[2].currentTarget.response);
        cb({mapData, groups, names});
    });
}

function binaryRequest(method, url) {
    return new Promise(function (resolve, reject) {
        const xhr = new XMLHttpRequest();
        xhr.open(method, url);
        xhr.responseType = 'arraybuffer';
        xhr.onload = resolve;
        xhr.onerror = reject;
        xhr.send();
    });
}
