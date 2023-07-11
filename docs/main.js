document.addEventListener(
    'DOMContentLoaded',
    () => loadJS("https://code.jquery.com/jquery-3.7.0.min.js", onLoad),
    false,
);

// We use this to prevent including an extra script tag in each html file
function loadJS(url, cb){
    var scriptTag = document.createElement('script');
    scriptTag.src = url;
    scriptTag.onload = cb;
    scriptTag.onreadystatechange = cb;
    document.body.appendChild(scriptTag);
}

function onLoad() {
    // Add isInViewport func
    $.fn.isInViewport = function() {
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
    $('body').append(pathHint);

    $(".fk").hover(
        function() {
            reversePath($(this), pathHint);
            pathHint.show();
        }, function() {
            pathHint.hide().empty();
        }
    );
}

function fieldPath() {
    let path = {};
    const windowTop = $(window).scrollTop();

    $(".fn").each(function(i, elem) {
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
        elem = elem.parents('.f').eq(0);
    }
    pathHint.text("$." + path.join("."));
}
