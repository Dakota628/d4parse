function pxToMapUnit(px) {
    return [px[0] / pxPerMapUnit, px[1] / pxPerMapUnit];
}

function subPx(a, b) {
    return [a[0] - b[0], a[1] - b[1]];
}

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

const worldSnoGroup = 48;
const sceneSnoGroup = 33;
const defaultSnoId =  69068;

// Load groups and names data
loadData((groups, names) => {
    $(() => {
        // Add world selector
        const worldSelect = $("#worldSelect");
        worldSelect.select2({
            theme: "classic"
        });

        let worldSnos = Object.entries(names[worldSnoGroup]);
        worldSnos.sort((a,b) => a[1].localeCompare(b[1]))
        for (const [snoId, snoName] of worldSnos) {
            worldSelect.append(`<option value="${snoId}">[World] ${snoName}</option>`);
        }

        let sceneSnos = Object.entries(names[sceneSnoGroup]);
        sceneSnos.sort((a,b) => a[1].localeCompare(b[1]))
        for (const [snoId, snoName] of sceneSnos) {
            worldSelect.append(`<option value="${snoId}">[Scene] ${snoName}</option>`);
        }

        // Add world select
        worldSelect.change(function() {
            loadWorld(
                groups,
                names,
                $(this).val(),
                $(this).find('option:selected').text(),
            );
        });

        // Load base world
        window.flyTo = getFlyTo();
        worldSelect.val(flyTo.w ?? defaultSnoId).trigger('change');

        // Add search event
        $("#search").data('val', '').on('input', function(){
            const v = this.value;
            if (v !== $(this).data('val')) {
                clearTimeout(this.delay);
                this.delay = setTimeout(function () {
                    console.log("Searching for:", v);
                    $(this).data('val', v);
                    drawMarkers(groups, names, v);
                }.bind(this), 300);
            }
        });
    });
});

// TODO: add overlays for quest conditioned map updates
// TODO: add radius (on hover) around markers with a radius
// TODO: add rotated and non-rotated grid
// TODO: filter by gizmo type
// TODO: regex search
// TODO: assure cluster groups are removed during search
// TODO: styling consistent with docs eventually

function rotate(p, angle) {
    const rads = (Math.PI / 180) * angle;
    const cos = Math.cos(rads);
    const sin = Math.sin(rads);
    return L.point(
        (cos * p.x) + (sin * p.y),
        (cos * p.y) - (sin * p.x)
    );
}

const markerMetaNames = {
    'mt': 'Marker Type',
    'gt': 'Gizmo Type',
}

function markerPopup(groups, names, marker, title) {
    const popup = $('<div></div>'); // Container doesn't matter, we just want the inner html
    popup.append(`<a class="snoRef popupTitle" href="../sno/${marker.r}.html">${title}</a>`);

    // Add attributes
    const dl = $('<dl class="markerAttrs"></dl>');

    // -- Source
    dl.append(`<dt>Source</dt><dd><a class="snoRef"  href="../sno/${marker.s}.html">${snoName(groups, names, lookupSnoGroup(names, marker.s), marker.s)}</a></dd>`);

    // -- Data SNOs
    const dataSnos = marker.d ?? [];
    if (dataSnos.length > 0) {
        dl.append('<dt>Data</dt>');
        const dd = $('<dd></dd>');
        for (const dataSno of dataSnos) {
            const title = snoName(groups, names, lookupSnoGroup(names, dataSno), dataSno);
            dd.append(`<a class="snoRef" href="../sno/${marker.s}.html">${title}</a>`)
        }
        dl.append(dd);
    }

    // -- Metadata
    for (const [key, val] of Object.entries(marker.m ?? {})) {
        dl.append(`<dt>${markerMetaNames[key] ?? key}</dt><dd>${val}</dd>`);
    }

    popup.append(dl);

    // Add coordinates
    popup.append(`<i>${marker.x.toFixed(6)}, ${marker.y.toFixed(6)}, ${marker.z.toFixed(6)}</i>`)

    return popup.html();
}

function snoGroupName(groups, id) {
    if (id === 255) {
        return "Unknown";
    }
    return groups[id] ?? `Group_${id}`;
}

function lookupSnoGroup(names, id) {
    for (const [groupId, m] of Object.entries(names)) {
        if (m.hasOwnProperty(id)) {
            return groupId
        }
    }
    return -1;
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

function drawMarkers(groups, names, search) {
    search = search.toLowerCase();
    console.log("Drawing markers:", search);

    if (!window.m) {
        return;
    }

    if (window.dataLayers) {
        window.dataLayers.remove();
    }

    const m = mapData.m ?? [];
    const markers = {};

    let len = m.length;
    while (len--) {
        const marker = m[len];
        const groupName = snoGroupName(groups, marker.g);
        const title = snoName(groups, names, marker.g, marker.r);

        if (search && search.length > 0) {
            if (!title.toLowerCase().includes(search)) {
                continue;
            }
        }

        const circle = L.circleMarker([marker.x, marker.y], {
            radius: 5,
            stroke: false,
            fill: true,
            fillOpacity: 0.75,
            fillColor: markerColors[groupName],
        }).bindPopup(
            markerPopup(groups, names, marker, title),
            {direction: 'center'},
        );

        if (!markers.hasOwnProperty(groupName)) {
            markers[groupName] = L.markerClusterGroup({
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
        markers[groupName].addLayer(circle);
    }

    window.dataLayers = L.control.layers({}, markers).addTo(window.m);
    console.log("Done drawing markers:", search);
}

window.mapUnitPerTile = 64;
window.tileSize = 512;
window.pxPerMapUnit = tileSize / mapUnitPerTile;

function loadWorld(groups, names, worldSnoId, worldSnoName, cb) {
    console.log("Loading world:", worldSnoId, worldSnoName);

    // Show loading screen
    $("#loading").show();

    binaryRequest('GET', `data/${worldSnoId}.mpk`).then((data) => {
        window.mapData = msgpackr.unpack(data);

        if (!mapData.p && !mapData.m) {
            $("#loading").hide();
            alert("No data for Scene/World");
            return
        }

        if (window.m && window.m.remove) {
            window.m.remove();
        }
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
        const centerPxX = ((mapData.artCenterX * (tileSize / mapData.gridSize))) / mapData.zoneArtScale;
        const centerPxY = ((mapData.artCenterY * (tileSize / mapData.gridSize))) / mapData.zoneArtScale;
        const scale = (mapUnitPerTile * mapData.zoneArtScale) / (mapData.gridSize * mapData.zoneArtScale);

        const D4CRS = L.extend({}, L.CRS.Simple, {
            projection: D4Projection,
            transformation: new L.Transformation(
                scale,
                centerPxX / pxPerMapUnit,
                scale,
                centerPxY / pxPerMapUnit,
            ),
        });

        // Setup renderer
        const canvas = L.canvas();

        // Setup map
        window.m = L.map('map', {
            attributionControl: false,
            crs: D4CRS,
            renderer: canvas,
            contextmenu: true,
            contextmenuItems: [
                {
                    text: 'Show Coordinates',
                    callback: function (e) {
                        L.popup().setLatLng(e.latlng)
                            .setContent(`${e.latlng.lat.toFixed(6)}, ${e.latlng.lng.toFixed(6)}`)
                            .openOn(window.m);
                    }
                },
                {
                    text: 'Copy Link to Location',
                    callback: function(e) {
                        navigator.clipboard.writeText(
                            `${location.protocol}//${location.host}${location.pathname}?x=${e.latlng.lat}&y=${e.latlng.lng}&w=${worldSnoId}`
                        );
                    }
                }
            ]
        }).setView([flyTo.x ?? 0, flyTo.y ?? 0], flyTo.z ?? 0);
        window.flyTo = {};


        worldTileLayer(window.m, worldSnoId, worldSnoName, mapData);

        // Add markers
        L.circleMarker([0, 0], {
            radius: 5,
            stroke: false,
            fill: true,
            fillOpacity: 0.75,
            fillColor: "black",
        }).bindTooltip("This is the center of the world!").addTo(window.m);

        L.circleMarker([100, 100], {
            radius: 5,
            stroke: false,
            fill: true,
            fillOpacity: 0.75,
            fillColor: "black",
        }).bindTooltip("100, 100").addTo(window.m);

        L.circleMarker([-100, -100], {
            radius: 5,
            stroke: false,
            fill: true,
            fillOpacity: 0.75,
            fillColor: "black",
        }).bindTooltip("-100, -100").addTo(window.m);

        L.circleMarker([-100, 100], {
            radius: 5,
            stroke: false,
            fill: true,
            fillOpacity: 0.75,
            fillColor: "black",
        }).bindTooltip("-100, 100").addTo(window.m);

        L.circleMarker([100, -100], {
            radius: 5,
            stroke: false,
            fill: true,
            fillOpacity: 0.75,
            fillColor: "black",
        }).bindTooltip("100, -100").addTo(window.m);

        // Load world
        const p = mapData.p ?? [];

        // Polygons
        let len = p.length;
        while (len--) { // Using while has a measurable performance improvement... bc Javascript.
            L.polygon(p[len], {
                weight: 3,
                color: '#ACA491',
                fill: false,
                opacity: 0.5,
                interactive: false,
            }).addTo(window.m)
        }

        // Markers
        $("#search").val("")
        drawMarkers(groups, names, "");

        // Remove loading screen
        $("#loading").hide();
    }, console.error);
}

function worldTileLayer(map, worldSnoId, worldSnoName, mapData) {
    // Setup tiles
    return L.tileLayer(`maptiles/${worldSnoId}/{z}/{x}_{y}.png`, {
        tileSize: tileSize,
        maxZoom: 15,
        minZoom: -1,
        minNativeZoom: 0,
        maxNativeZoom: mapData.maxNativeZoom ?? 3,
        noWrap: true,
        tms: false,
    }).addTo(map);
}

function loadData(cb) {
    Promise.all([
        binaryRequest('GET', '../groups.mpk'),
        binaryRequest('GET', '../names.mpk'),
    ]).then((values) => {
        cb(
            msgpackr.unpack(values[0]),
            msgpackr.unpack(values[1]),
        );
    }, console.error);
}

function binaryRequest(method, url) {
    var xhr = new XMLHttpRequest();
    xhr.responseType = 'arraybuffer';
    return $.ajax({
        method,
        url,
        xhr: function() {
            return xhr;
        }
    })
}

function getFlyTo() {
    // TODO: also need to include world sno id
    const urlParams = new URLSearchParams(window.location.search);
    if (!urlParams.has('x') || !urlParams.has('y')) {
        return { x: 0, y: 0, z: 0, w: defaultSnoId };
    }
    const out = {
        x: Number(urlParams.get('x')),
        y: Number(urlParams.get('y')),
        z: Number(urlParams.get('z') ?? 6),
        w: Number(urlParams.get('w') ?? defaultSnoId)
    };
    urlParams.delete('x');
    urlParams.delete('y');
    urlParams.delete('w');
    return out;
}
