requirejs.config({
    "baseUrl": "js/lib",
    "paths": {
        "jquery": "//code.jquery.com/jquery-3.7.0.min",
    }
});

define(['jquery'], ($) => {
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
        loadRefs();
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

function loadRefs() {
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
