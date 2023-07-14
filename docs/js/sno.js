requirejs.config({
    "baseUrl": "js/lib",
    "paths": {
        "jquery": "//code.jquery.com/jquery-3.7.0.min",
        "cytoscape": "//unpkg.com/cytoscape@3.25.0/dist/cytoscape.min",
    }
});

define(['jquery', 'cytoscape'], ($, cytoscape) => {
    $(() => {
        // Get SNO ID
        window.snoId = Number($('.snoMeta .fn:contains("ID")').closest('.f').find('.fv').text());

        // Add isInViewport func
        $.fn.isInViewport = function () {
            var elementTop = $(this).offset().top;
            var elementBottom = elementTop + $(this).outerHeight();

            var viewportTop = $(window).scrollTop();
            var viewportBottom = viewportTop + $(window).height();

            return elementBottom > viewportTop && elementTop < viewportBottom;
        };

        // Collapsable types
        $(".tn").on("click", function () {
            $(this).siblings(".f").toggle();
        });

        // Field hover
        const pathHint = $('<div class="pathHint"></div>');
        pathHint.hide();
        $('body').append(pathHint);

        $(".fk").hover(
            function () {
                reversePath($(this), pathHint);
                pathHint.show();
            }, function () {
                pathHint.hide().empty();
            }
        );

        //  Load refs map
        loadRefs($);

        // Generate quest graph
        generateQuestGraph($, cytoscape);
    })
});

function reversePath(elem, pathHint) {
    const path = [];
    elem = elem.parents('.f').eq(0);
    while (elem.length) {
        path.unshift(elem.find(".fk > .fn").first().text());
        elem = elem.parents('.f').eq(0); // TODO: add support for array indexes
    }
    pathHint.text("$." + path.join("."));
}

function loadRefs($) {
    const snoMeta = $(".snoMeta").eq(0);
    const metaEntry = $('<div class="f"><div class="fk"><div class="fn">Referenced By</div></div><div class="fv refs"></div></div>');
    const valNode = metaEntry.find('.fv');
    snoMeta.append(metaEntry);

    const req = new XMLHttpRequest();
    req.open("GET", "../refs.bin", true);
    req.responseType = "arraybuffer";
    req.onload = function (e) {
        const dv = new DataView(req.response);
        for (let p = 0; p < dv.byteLength; p += 8) { // TODO: if we sort the refs bin, we can binary search
            // Read data
            const to = dv.getInt32(p + 4, true);
            if (to !== snoId) {
                continue;
            }
            const from = dv.getInt32(p, true);

            // Append link
            const link = $("<a></a>").attr("href", `${from}.html`).text(from);
            valNode.append(link);
        }

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
    }).eq(0).text();
}

function generateQuestGraph($, cytoscape) {
    const qd = findType($('body'), 'QuestDefinition');
    if (qd.length === 0) {
        return
    }

    let nodes = [];
    let edges = [];

    findType(qd, 'QuestPhase').each(function () {
        const phase = $(this);
        const phaseUid = getFieldValue(phase, 'dwUID');
        nodes.push({
            group: 'nodes',
            data: {
                id: phaseUid,
                name: `Phase ${phaseUid}`,
                e: phase,
            },
        });

        findType(phase, 'QuestObjectiveSet').each(function () {
            const objectiveSet = $(this);
            const objectiveSetUid = getFieldValue(objectiveSet, 'dwUID');
            nodes.push({
                group: 'nodes',
                data: {
                    id: objectiveSetUid,
                    name: `Objective Set ${objectiveSetUid}`,
                    e: objectiveSet,
                },
            });
            edges.push({
                group: 'edges',
                data: {
                    id: `${phaseUid}:${objectiveSetUid}`,
                    source: phaseUid,
                    target: objectiveSetUid,
                },
            });

            findType(objectiveSet, 'QuestObjectiveSetLink').each(function () {
                const link = $(this);
                const linkDestinationPhaseUid = getFieldValue(link, 'dwDestinationPhaseUID');
                edges.push({
                    group: 'edges',
                    data: {
                        id: `${objectiveSetUid}:${linkDestinationPhaseUid}`,
                        source: objectiveSetUid,
                        target: linkDestinationPhaseUid,
                    },
                });
            });

            findType(objectiveSet, 'QuestCallback').each(function () {
                const callback = $(this);
                const callbackUid = getFieldValue(callback, 'dwUID');
                nodes.push({
                    group: 'nodes',
                    data: {
                        id: callbackUid,
                        name: `Callback ${callbackUid}`,
                        e: callback,
                    },
                });
                edges.push({
                    group: 'edges',
                    data: {
                        id: `${objectiveSetUid}:${callbackUid}`,
                        source: objectiveSetUid,
                        target: callbackUid,
                    },
                });
            });
        });
    });

    const metaEntry = $('<div class="f"><div class="fk"><div class="fn">Quest Graph</div></div><div class="fv"><div id="questGraph"></div></div></div></div>');
    const cyDiv = metaEntry.find('#questGraph');
    $(".snoMeta").eq(0).append(metaEntry);

    let cy = cytoscape({
        container: cyDiv.get(0),
        boxSelectionEnabled: false,
        autoungrabify: true,
        style: [{
            selector: "node",
            css: {
                label: "data(name)",
                "text-valign": "center",
                "text-halign": "center",
                height: "60px",
                width: "150px",
                shape: "rectangle",
                "background-color": "#343a40",
                "border": "none",
                "color": "#4c6ef5",
                "font-size": "14px"
            }
        },
            {
                selector: "edge",
                css: {
                    "curve-style": "bezier",
                    "target-arrow-shape": "triangle"
                }
            }
        ],
    });
    cy.add(nodes);
    cy.add(edges);

    cy.on('tap', 'node', function () {
        this.data('e').get(0).scrollIntoView({
            behavior: 'smooth',
        });
    });

    cy.layout({
        name: 'breadthfirst'
    }).run();
}
