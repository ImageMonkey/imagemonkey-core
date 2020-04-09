$(document)
    .ready(function() {

        // fix menu when passed
        $('.masthead')
            .visibility({
                once: false,
                onBottomPassed: function() {
                    $('.fixed.menu').transition('fade in');
                },
                onBottomPassedReverse: function() {
                    $('.fixed.menu').transition('fade out');
                }
            });

        // create sidebar and attach to menu open
        $('.ui.sidebar')
            .sidebar('attach events', '.toc.item');

    });

function escapeHtml(str) {
    return jQuery('<div/>').text(str).html();
}

function unescapeHtml(safe) {
    return $('<div />').html(safe).text();
}

function parseJwt(token) {
    var base64Url = token.split('.')[1];
    var base64 = base64Url.replace('-', '+').replace('_', '/');
    return JSON.parse(window.atob(base64));
};

function getCookie(s) {
    var cookie = Cookies.get(s);
    if (typeof cookie == "undefined") {
        return "";
    }

    //in case the token already expired, return an empty string
    //otherwise the backend fails the request due to an invalid token.
    //if no token is provided the backend will fall back to the (restricted)
    //unauthorized mode.
    var jwt = parseJwt(cookie);
    if (Math.round((new Date()).getTime() / 1000) > jwt["exp"]) {
        return "";
    }

    return cookie;
}

function isMobileDevice() {
    if (/Mobi/.test(navigator.userAgent)) {
        return true;
    }
    return false;
}
