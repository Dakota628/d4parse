requirejs.config({
    "baseUrl": "js/lib",
    "paths": {
        "jquery": "//code.jquery.com/jquery-3.7.0.min",
        "dagre": "//unpkg.com/dagre@0.7.4/dist/dagre",
        "cytoscape": "//unpkg.com/cytoscape@3.25.0/dist/cytoscape.min",
        "cytoscape-dagre": "//unpkg.com/cytoscape-dagre@2.5.0/cytoscape-dagre",
    }
});

define(['jquery', 'cytoscape', 'cytoscape-dagre'], ($, cytoscape, cydagre) => {
    // Setup cytoscape
    cydagre(cytoscape);

    //  Load refs map
    loadRefs($);

    $(() => {
        // Read SNO info from HTML
        window.sno = {
            group: getSnoInfo("Group"),
            id: Number(getSnoInfo("ID")),
            name: getSnoInfo("Name"),
            file: getSnoInfo("File"),
        }

        // Set page title
        $(document).prop("title", `[${sno.group}] ${sno.name}`)

        // Add isInViewport func
        $.fn.isInViewport = function () {
            var elementTop = $(this).offset().top;
            var elementBottom = elementTop + $(this).outerHeight();
            var viewportTop = $(window).scrollTop();
            var viewportBottom = viewportTop + $(window).height();
            return elementBottom > viewportTop && elementTop < viewportBottom;
        };

        // Generate quest graph
        generateQuestGraph($, cytoscape);

        // Collapsable types
        $(".tn").on("click", function () {
            $(this).siblings().toggle();
        });

        // Field hover
        window.pathHint = $('<div class="pathHint"></div>').hide();
        $('body').append(pathHint);

        $(".fk").hover(
            function () {
                reversePath($(this), pathHint);
                pathHint.show();
            }, function () {
                pathHint.hide().empty();
            }
        );
    })
});

function getSnoInfo(key) {
    return $('.snoMeta .fn:contains("' + key + '")').closest('.f').find('.fv').text()
}

function reversePath(elem, pathHint) {
    let path = "";
    elem = elem.parents('.f, li').eq(0)
    while (elem.length) {
        if (elem.is('li')) {
            path = "[" + elem.index() + "]" + path;
        } else {
            path = "." + elem.find(".fk > .fn").first().text() + path;
        }
        elem = elem.parents('.f, li').eq(0);
    }
    pathHint.text("$" + path);
}

function refOffset(i) {
    return (8 * i) + 4
}

function searchRef(dv, pred) {
    let l = -1;
    let h = dv.getUint32(0, true);
    while (1 + l < h) {
        const m = l + ((h - l) >> 1);
        if (pred(dv.getInt32(refOffset(m) + 4, true))) {
            h = m;
        } else {
            l = m;
        }
    }
    return h;
}

function eachReferencedBy(dv, targetTo, cb) {
    const len = dv.getUint32(0, true);
    let i = searchRef(dv, x => targetTo <= x)
    if (i < 0) {
        return
    }
    for (; i < len; i++) {
        const off = refOffset(i);
        if (dv.getInt32(off + 4, true) !== targetTo) {
            break;
        }
        cb(dv.getInt32(off, true));
    }
}

function loadRefs($) {
    const metaEntry = $('<div class="f"><div class="fk"><div class="fn">Referenced By</div></div><div class="fv refs"></div></div>');
    const valNode = metaEntry.find('.fv');

    const req = new XMLHttpRequest();
    req.open("GET", "../refs.bin", true);
    req.responseType = "arraybuffer";
    req.onload = function (e) {
        const dv = new DataView(req.response);
        eachReferencedBy(dv, sno.id, function (from) {
            const link = $("<a></a>").attr("href", `${from}.html`).text(from);
            valNode.append(link);
        })
        $(() => $(".snoMeta").eq(0).append(metaEntry));
    };
    req.send();
}

function findType(elem, t) {
    return elem.find('.t:has(.tn:contains("' + t + '"))').filter(function () {
        return $(this).children('.tn').text() === t
    });
}

function getFieldValue(elem, f) {
    return elem.children('.f:has(> .fk > .fn:contains("' + f + '"))').children('.fv').filter(function () {
        return $(this).closest('.f').find('.fn').text() === f
    }).eq(0);
}

const questType = {
    definition: 'QuestDefinition',
    phase: 'QuestPhase',
    objectiveSet: 'QuestObjectiveSet',
    objectiveSetLink: 'QuestObjectiveSetLink',
    callback: 'QuestCallback',
};

function generateQuestGraph($, cytoscape) {
    // Determine graph nodes and edges
    const qd = findType($('body'), questType.definition);
    if (qd.length === 0) {
        return
    }

    const uidKey = 'dwUID';
    const destPhaseUidKey = 'dwDestinationPhaseUID'

    let nodes = [];
    let edges = [];

    findType(qd, questType.phase).each(function () {
        const phase = $(this);
        const phaseUidElem = getFieldValue(phase, uidKey);
        const phaseUid = phaseUidElem.text();

        nodes.push({
            group: 'nodes',
            data: {
                id: phaseUid,
                type: questType.phase,
                name: phaseUid,
                e: phase,
                eVal: phaseUidElem,
            },
        });

        findType(phase, questType.objectiveSet).each(function () {
            const objectiveSet = $(this);
            const objectiveSetUidElem = getFieldValue(objectiveSet, uidKey);
            const objectiveSetUid = objectiveSetUidElem.text();
            nodes.push({
                group: 'nodes',
                data: {
                    id: objectiveSetUid,
                    type: questType.objectiveSet,
                    name: objectiveSetUid,
                    e: objectiveSet,
                    eVal: objectiveSetUidElem,
                },
            });
            edges.push({
                group: 'edges',
                data: {
                    id: `${phaseUid}:${objectiveSetUid}`,
                    source: phaseUid,
                    target: objectiveSetUid,
                    color: "#f8f9fa",
                    shape: "triangle",
                },
            });

            findType(objectiveSet, questType.objectiveSetLink).each(function () {
                const link = $(this);
                const linkDestinationPhaseUidElem = getFieldValue(link, destPhaseUidKey);
                const linkDestinationPhaseUid = linkDestinationPhaseUidElem.text();
                edges.push({
                    group: 'edges',
                    data: {
                        id: `${objectiveSetUid}:${linkDestinationPhaseUid}`,
                        source: objectiveSetUid,
                        target: linkDestinationPhaseUid,
                        color: "#4c6ef5",
                        shape: "triangle",
                        e: link,
                        eVal: linkDestinationPhaseUidElem,
                    },
                });
            });

            findType(objectiveSet, questType.callback).each(function () {
                const callback = $(this);
                const callbackUidElem = getFieldValue(callback, uidKey);
                const callbackUid = callbackUidElem.text();
                nodes.push({
                    group: 'nodes',
                    data: {
                        id: callbackUid,
                        type: questType.callback,
                        name: callbackUid,
                        e: callback,
                        eVal: callbackUidElem,
                    },
                });
                edges.push({
                    group: 'edges',
                    data: {
                        id: `${objectiveSetUid}:${callbackUid}`,
                        source: objectiveSetUid,
                        target: callbackUid,
                        color: "#f8f9fa",
                        shape: "circle-triangle",
                    },
                });
            });
        });
    });

    // Construct the graph
    const metaEntry = $('<div class="extra"><div class="tn">Quest Graph</div><div id="questGraph"></div></div>')
    const cyDiv = metaEntry.find('#questGraph');
    $(".snoMeta").eq(0).after(metaEntry);

    let cy = cytoscape({
        container: cyDiv.get(0),
        boxSelectionEnabled: false,
        autoungrabify: true,
        style: [
            {
                selector: "node",
                css: {
                    label: "data(name)",
                    "text-valign": "center",
                    "text-halign": "center",
                    height: "50px",
                    width: "50px",
                    shape: "circle",
                    "background-color": "#343a40",
                    "border": "none",
                    "color": "#4c6ef5",
                    "font-size": "14px"
                }
            },
            {
                selector: "edge",
                css: {
                    label: "data(label)",
                    color: "white",
                    "curve-style": "bezier",
                    "target-arrow-shape": "data(shape)",
                    "line-color": "data(color)",
                    "target-arrow-color": "data(color)",
                    "line-opacity": "0.5",
                }
            }
        ],
        elements: [
            ...nodes,
            ...edges,
        ],
    });

    // Add node events
    cy.on('tap', 'node', function () {
        this.data('e').get(0).scrollIntoView({
            behavior: 'smooth',
        });
    });

    cy.on('mouseover', 'node', function (e) {
        cyDiv.css('cursor', 'pointer');
        const keyElem = this.data('eVal').closest('.f').find('.fk');
        reversePath(keyElem, pathHint);
        pathHint.show();
    })

    cy.on('mouseout', 'node', function () {
        cyDiv.css('cursor', '');
        pathHint.hide().empty();
    })

    // Add edge events
    cy.on('tap', 'edge', function () {
        const e = this.data('e')
        if (e) {
            e.get(0).scrollIntoView({
                behavior: 'smooth',
            });
        }
    });

    cy.on('mouseover', 'edge', function (e) {
        const eVal = this.data('eVal')
        if (eVal) {
            cyDiv.css('cursor', 'pointer');
            const keyElem = eVal.closest('.f').find('.fk');
            reversePath(keyElem, pathHint);
            pathHint.show();
        }
    })

    cy.on('mouseout', 'edge', function () {
        cyDiv.css('cursor', '');
        pathHint.hide().empty();
    })

    // Layout and render the graph
    cy.layout({
        name: 'dagre'
    }).run();
}
