// From ZoneMapParams in WorldDefinition
const zoneArtScale = 1.066667; // tZoneMapParams.fZoneArtScale
const zoneArtCenter = [1449.000000, 2909.000000]; // tZoneMapParams.vecZoneArtCenter
const zoneMapParamsScale = 5; // Scale of texture relative to zone map params

// Pixels <-> Leaflet map units
const mapUnitPerTile = 64;
const mapSize = 40;
const tileSize = 512;
const pxPerMapUnit =  tileSize / mapUnitPerTile;

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
    project: function(latlng) {
        const point = L.Projection.LonLat.project(latlng);
        return rotate(point, -45);
    },
    unproject: function (point) {
        point = rotate(point, 45);
        return L.Projection.LonLat.unproject(point);
    },
})

const D4CRS = L.extend({}, L.CRS.Simple, {
    projection: D4Projection,
    transformation: new L.Transformation(ptScale, originMapUnits[0], ptScale, originMapUnits[1]),
});

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
        maxZoom: 4,
        minZoom: -1,
        minNativeZoom: 0,
        maxNativeZoom: 3,
        noWrap: true,
        tms: false,
    }).addTo(map);

    // Add map events
    map.on('click', function(e) {
        L.popup()
            .setLatLng(e.latlng)
            .setContent(`(${e.latlng.lat}, ${e.latlng.lng})`)
            .openOn(map);
        console.log(e);
    });

    // Add markers
    L.circle([0, 0], {
        radius: 1.5,
        stroke: false,
        fill: true,
        fillOpacity: 1.0,
        fillColor: "red",
    }).bindTooltip("This is the center of the world!").addTo(map);
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
