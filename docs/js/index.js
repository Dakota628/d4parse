requirejs.config({
    "baseUrl": "js/lib",
    "paths": {
        "jquery": "//code.jquery.com/jquery-3.7.0.min",
        "notie": "//unpkg.com/notie@4.3.1/dist/notie.min",
    }
});

define(['jquery', 'notie'], ($, notie) => {
    $(() => {
        const input = $("#search");
        input.focus();
        input.on("keydown", function search(e){
            notie.hideAlerts();
            if (e.keyCode === 13) {
                const url = `sno/${$(this).val()}.html`
                $.get({
                    url: `sno/${$(this).val()}.html`,
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
                    input.focus();
                })
            }
        });
    });
});
