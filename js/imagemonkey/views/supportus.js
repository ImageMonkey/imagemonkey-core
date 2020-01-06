var SupportUsView = (function() {
    function SupportUsView(apiBaseUrl) {
        this.apiBaseUrl = apiBaseUrl;
    }

    SupportUsView.prototype.setSentryDSN = function(sentryDSN) {
        try {
            Sentry.init({
                dsn: sentryDSN,
            });
        } catch (e) {}
    }

    SupportUsView.prototype.exec = function() {

        $("#selectPurposeDropdown").dropdown();

        $("#amount").val("25");
        $("#totalAmount").val("25");

        $("#amount").change(function() {
            var regex = /^[1-9]\d*(((,\d{3}){1})?(\.\d{0,2})?)$/;
            var val = $("#amount").val();

            if (!regex.test(val)) {
                $("#warningMsgText").text("Invalid input: " + val);
                $("#warningMsg").show();
            } else {
                $("#warningMsg").hide();
                $("#totalAmount").val(val);
            }
        });

        $("#customAmountButton").click(function(e) {
            $("#amount").val("");
            $("#amount").focus();
        });


    }

    return SupportUsView;
}());