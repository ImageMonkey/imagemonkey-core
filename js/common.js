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
    var entityMap = {
        "&": "&amp;",
        "<": "&lt;",
        ">": "&gt;",
        '"': '&quot;',
        "'": '&#39;',
        "/": '&#x2F;'
    };

    return String(str).replace(/[&<>"'\/]/g, function(s) {
        return entityMap[s];
    });
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

function isMobileOrTabletDevice() {
    //there seems to be no reliable way to detect the device type (mobile, tablet, desktop, etc.)
	//the most reliable way seems to be to use the media queries. If the max screen width is <= 1280px
	//it's a mobile/tablet device.
	var mq = window.matchMedia("(max-width: 1280px)");
	if(mq.matches)
		return true;
	return false;
}
