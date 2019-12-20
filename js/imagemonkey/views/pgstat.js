var PgStatView = (function() {
    function PgStatView(apiBaseUrl, clientId, clientSecret) {
        this.apiBaseUrl = apiBaseUrl;
        this.clientId = clientId;
        this.clientSecret = clientSecret;

        this.imageMonkeyApi = new ImageMonkeyApi(this.apiBaseUrl);
        this.imageMonkeyApi.setToken(getCookie("imagemonkey"));
        this.imageMonkeyApi.setClientId(clientId);
        this.imageMonkeyApi.setClientSecret(clientSecret);
    }

    PgStatView.prototype.setSentryDSN = function(sentryDSN) {
        try {
            Sentry.init({
                dsn: sentryDSN,
            });
        } catch (e) {}
    }

    PgStatView.prototype.exec = function() {
        $("#loadingSpinner").show();
        this.imageMonkeyApi.getPgStatStatements()
            .then(function(data) {
                for (var i = 0; i < data.length; i++) {
                    var cellColor = "#ffffff";


                    if (data[i].avg >= 1000)
                        cellColor = "#ff0000";
                    else if (data[i].avg >= 500)
                        cellColor = "#ffa500";

                    var elem = $(('<tr>' +
                        '<td bgcolor="' + cellColor + '">' + escapeHtml(data[i].total) + '</td>' +
                        '<td bgcolor="' + cellColor + '">' + escapeHtml(data[i].avg) + '</td>' +
                        '<td bgcolor="' + cellColor + '">' + escapeHtml(data[i].query) + '</td>' +
                        '</tr>'));
                    $("#pgStatTableContent").append(elem);
                }
            }).catch(function(e) {
                Sentry.captureException(e);
            });
        $("#loadingSpinner").hide();
    }

    return PgStatView;
}());