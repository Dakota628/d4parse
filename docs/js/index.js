requirejs.config({
    "baseUrl": "js/lib",
    "paths": {
        "jquery": "//code.jquery.com/jquery-3.7.0.min",
        "jquery-ui": "//code.jquery.com/ui/1.13.2/jquery-ui.min",
        "notie": "//unpkg.com/notie@4.3.1/dist/notie.min",
        "msgpack": "//unpkg.com/@msgpack/msgpack@3.0.0-beta2/dist.es5+umd/msgpack.min",
    }
});

define(['jquery', 'jquery-ui', 'notie', 'msgpack'], ($, ui, notie, msgpack) => {
    loadNames($, msgpack, notie);

    $(() => {
        const input = $("#search");
        input.focus();
    });
});

const queryRegex = /(\[(?<group>\w+)\] )?(?<query>.*)/g;

function subSearch(groups, query, max, group, m, results) {
    for (const [id, name] of Object.entries(m)) {
        if (name.toLowerCase().includes(query)) {
            results.push({
                label: entryName(groups, {group, name}),
                group,
                name,
                id,
            });
            if (results.length >= max) {
                return results
            }
        }
    }
    return results
}

function search(groups, groupsByName, names, query, max) {
    const queryParts = query.matchAll(queryRegex).next();
    if (!queryParts) {
        return []
    }
    const group = queryParts.value.groups.group;
    query = queryParts.value.groups.query.toLowerCase();

    // Group specified
    if (group) {
        const groupId = groupsByName[group];
        return subSearch(groups, query, max, groupId, names[groupId] ?? {}, []);
    }

    // No group specified
    let results = [];
    for (const [groupId, m] of Object.entries(names)) {
        results = subSearch(groups, query, max, groupId, m, results);
    }
    return results;
}

function loadNames($, msgpack, notie) {
    Promise.all([
        binaryRequest($, 'GET', 'groups.mpk'),
        binaryRequest($, 'GET', 'names.mpk'),
    ]).then((values) => {
        window.groups = msgpack.decode(values[0]);
        window.groupsByName = Object.fromEntries(Object.entries(groups).map(a => a.reverse()))
        window.names = msgpack.decode(values[1]);

        $("#search").autocomplete({
            source: function (request, response) {
                response(search(groups, groupsByName, names, request.term, 100))
            },
            select: function (event, ui) {
                const url = `sno/${ui.item.id}.html`
                $.get({
                    url: `sno/${ui.item.id}.html`,
                    cache: true,
                }).done(() => {
                    $(location).prop('href', url);
                }).fail(() => {
                    notie.alert({
                        type: 'error',
                        text: 'SNO not found!',
                        buttonText: 'Moo!',
                        time: 100,
                    });
                    $("#search").focus();
                })
            },
            focus: function (event, ui) {
                this.value = entryName(groups, names, ui.item);
                event.preventDefault();
            },
        });
    }, console.error);
}

function binaryRequest($, method, url) {
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

function entryName(groups, item) {
    return `[${groups[item.group]}] ${item.name}`
}
