requirejs.config({
    "baseUrl": "js/lib",
    "paths": {
        "jquery": "//code.jquery.com/jquery-3.7.0.min",
    }
});

define(['jquery'], ($) => {
    //  Load refs map
    loadRefs();

    $(() => {
        // Read SNO info from HTML
        window.sno = {
            group:  getSnoInfo("Group"),
            id: Number(getSnoInfo("ID")),
            name:  getSnoInfo("Name"),
            file:  getSnoInfo("File"),
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

        // Collapsable types
        $(".tn").on("click", function () {
            $(this).siblings(".f").toggle();
        });

        // Field hover
        const pathHint = $('<div class="pathHint"></div>').hide();
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

function loadRefs() {
    const metaEntry = $('<div class="f"><div class="fk"><div class="fn">Referenced By</div></div><div class="fv refs"></div></div>');
    const valNode = metaEntry.find('.fv');

    const req = new XMLHttpRequest();
    req.open("GET", "../refs.bin", true);
    req.responseType = "arraybuffer";
    req.onload = function (e) {
        const dv = new DataView(req.response);
        for (let p = 0; p < dv.byteLength; p += 8) { // TODO: if we sort the refs bin, we can binary search
            // Read data
            const to = dv.getInt32(p + 4, true);
            if (to !== sno.id) {
                continue;
            }
            const from = dv.getInt32(p, true);

            // Append link
            const link = $("<a></a>").attr("href", `${from}.html`).text(from);
            valNode.append(link);
        }

        $(() => {
            const snoMeta = $(".snoMeta").eq(0);
            snoMeta.append(metaEntry);
        })
    };
    req.send();
}
