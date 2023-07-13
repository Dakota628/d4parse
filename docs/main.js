document.addEventListener(
    'DOMContentLoaded',
    () => loadJS("https://code.jquery.com/jquery-3.7.0.min.js", onLoad),
    false,
);

// We use this to prevent including an extra script tag in each html file
function loadJS(url, cb) {
    var scriptTag = document.createElement('script');
    scriptTag.src = url;
    scriptTag.onload = cb;
    scriptTag.onreadystatechange = cb;
    document.body.appendChild(scriptTag);
}

function onLoad() {
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

    // Field paths
    // fieldPath();
    // $(window).scroll(fieldPath);

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
    loadRefs()
}

function fieldPath() {
    let path = {};
    const windowTop = $(window).scrollTop();

    $(".fn").each(function (i, elem) {
        const elemTop = $(elem).offset().top;

        // Assure elem has been scrolled past and parent is in viewport
        if (elemTop > windowTop || !$(elem).closest('.t').isInViewport()) {
            return;
        }

        // Update path
        const elemLeft = $(elem).offset().left;
        if (path[elemLeft]) {
            const pathTop = $(path[elemLeft]).offset().top;
            if (elemTop > pathTop) {
                path[elemLeft] = elem;
            }
        } else {
            path[elemLeft] = elem;
        }
    });

    path = Object.entries(path)
        .sort((a, b) => a[0] - b[0])
        .map(x => $(x[1]).text());

    console.log(path);
}

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
