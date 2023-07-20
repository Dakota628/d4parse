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

// Markers
const markerColors = {
    'Global': 'green',
    'World': 'blue',
    'World - Game': 'lightorange',
    'World - Lighting': 'yellow',
    'World - Merged': 'lightblue',
    'World - Cameras': 'gray',
    'World - Props': 'lightbrown',
    'World - Merged Props': 'brown',
    'World - Bounties': 'red',
    'World - Events': 'orange',
    'World - Clickies': 'pink',
    'World - Population': 'purple',
    'World - Ambient': 'lightpurple',
    'Subzone': 'blue',
    'Subzone - Game': 'lightorange',
    'Subzone - Lighting': 'yellow',
    'Subzone - Merged': 'lightblue',
    'Subzone - Cameras': 'gray',
    'Subzone - Props': 'lightbrown',
    'Subzone - Merged Props': 'brown',
    'Subzone - Bounties': 'red',
    'Subzone - Events': 'orange',
    'Subzone - Clickies': 'pink',
    'Subzone - Population': 'purple',
    'Subzone - Ambient': 'lightpurple',
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
        maxZoom: 5,
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

    const dataLayers = L.control.layers([], []).addTo(map);
    const searchMarkers = []

    loadMarkers((markers) => {
        // console.log(markers);

        for (const poly of markers["Polygons"]) {
            L.polygon(poly, {
                weight: 3,
                color: '#ffffff',
                fill: false,
                opacity: 0.1,
                interactive: false,
            }).addTo(map)
        }

        for (const [markerGroup, d] of Object.entries(markers["Markers"])) {
            const groupMarkers = []

            for (const marker of d) {
                if (marker.d !== "") {
                    marker.d += "<br/>"
                }

                const circle = L.circleMarker([marker.x, marker.y], {
                    radius: 5,
                    stroke: false,
                    fill: true,
                    fillOpacity: 0.75,
                    fillColor: markerColors[markerGroup],
                    title: marker.n,
                });
                circle.bindTooltip(
                    `<b> <a href="../sno/${marker.r}">${marker.n}</a></b><p>${marker.d}Source: <a href="../sno/${marker.s}">${marker.s}</a><br/>(${marker.x}, ${marker.y}, ${marker.z})</p>`,
                    { direction: 'center' },
                );

                groupMarkers.push(circle);
                // searchMarkers.push(circle);
            }

            const groupLayer = L.layerGroup(groupMarkers);
            dataLayers.addOverlay(groupLayer, markerGroup);
        }

        // const searchLayer = L.layerGroup(searchMarkers);
        // new L.control.search({
        //     layer: searchLayer,
        //     initial: false,
        //     propertyName: 'title',
        //     hideMarkerOnCollapse: true,
        //     zoom: 5,
        // }).addTo(map);
        // map.removeLayer(searchLayer); // Search layer is fake... just to facilitate the search
    });


    // TODO: add overlays for quest conditioned map updates
    // TODO: add radius (on hover) around markers with a radius
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

function loadMarkers(cb) {
    const req = new XMLHttpRequest();
    req.open("GET", "markers.mpk", true);
    req.responseType = "arraybuffer";
    req.onload = function () {
        cb(msgpack.deserialize(req.response))
    };
    req.send();
}
