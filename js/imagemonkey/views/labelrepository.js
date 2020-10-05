var LabelRepositoryView = (function() {
    function LabelRepositoryView(apiBaseUrl, clientId, clientSecret, username, canAcceptTrendingLabelPermission, trendingLabelsRepositoryUrl, labelsRepositoryUrl) {
        this.apiBaseUrl = apiBaseUrl;
        this.clientId = clientId;
        this.clientSecret = clientSecret;
        this.username = username;
        this.canAcceptTrendingLabelPermission = canAcceptTrendingLabelPermission;
        this.trendingLabelsRepositoryUrl = trendingLabelsRepositoryUrl;
        this.labelsRepositoryUrl = labelsRepositoryUrl;

        this.imageMonkeyApi = new ImageMonkeyApi(this.apiBaseUrl);
        this.imageMonkeyApi.setToken(getCookie("imagemonkey"));
        this.imageMonkeyApi.setClientId(clientId);
        this.imageMonkeyApi.setClientSecret(clientSecret);
    }

    LabelRepositoryView.prototype.setSentryDSN = function(sentryDSN) {
        try {
            Sentry.init({
                dsn: sentryDSN,
            });
        } catch (e) {}
    }

    LabelRepositoryView.prototype.getRenameToLabel = function() {
        return $("#addTrendingLabelDlgRenameToLabelInput").val();
    }

    LabelRepositoryView.prototype.getLabelDescription = function() {
        return $("#addTrendingLabelDlgDescriptionInput").val();
    }

    LabelRepositoryView.prototype.getLabelPlural = function() {
        return $("#addTrendingLabelDlgPluralFormInput").val();
    }

    LabelRepositoryView.prototype.getSelectedLabelType = function() {
        var radioButtonId = $("#labelTypeRadioButtons :radio:checked").attr("id");
        if (radioButtonId === "labelTypeNormalRadioButtonInput") {
            return "normal";
        }
        if (radioButtonId === "labelTypeMetaRadioButtonInput") {
            return "meta";
        }
        return "";
    }

    LabelRepositoryView.prototype.onAddTrendingLabel = function(elem) {
        $("#addTrendingLabelDlg").attr("data-label-name", $(elem).attr("data-label"));
        $("#addTrendingLabelDlgPluralFormInput").val($(elem).attr("data-label-plural"));
        $("#addTrendingLabelDlgRenameToLabelInput").val($(elem).attr("data-label-renameto"));
        $("#addTrendingLabelDlgDescriptionInput").val($(elem).attr("data-label-description"));
        var labelType = this.getSelectedLabelType();
        if (labelType === "normal") {
            $("#addTrendingLabelDlgPluralFormInput").prop("disabled", false);
        } else {
            $("#addTrendingLabelDlgPluralFormInput").prop("disabled", true);
        }

        $("#addTrendingLabelDlg").modal("show");
    }

    LabelRepositoryView.prototype.onAcceptTrendingLabel = function(elem) {
        $("#addTrendingLabelDlg").attr("data-label-name", $(elem).attr("data-label"));
        $("#addTrendingLabelDlgPluralFormInput").val($(elem).attr("data-label-plural"));
        $("#addTrendingLabelDlgDescriptionInput").val($(elem).attr("data-label-description"));
        $("#addTrendingLabelDlgRenameToLabelInput").val($(elem).attr("data-label-renameto"));
        var labelType = $(elem).attr("data-label-type");
        if (labelType === "normal") {
            $("#addTrendingLabelDlgPluralFormInput").prop("disabled", false);
            $("#labelTypeNormalRadioButton").checkbox("set checked");
        } else {
            $("#addTrendingLabelDlgPluralFormInput").prop("disabled", true);
            $("#labelTypeMetaRadioButton").checkbox("set checked");
        }
        $("#addTrendingLabelDlg").modal("show");
    }

    LabelRepositoryView.prototype.acceptTrendingLabel = function(labelName, labelType, labelDescription, labelPlural, labelRenameTo) {
        let inst = this;
        $("#loadingIndicator").show();
        this.imageMonkeyApi.acceptTrendingLabel(labelName, labelType, labelDescription, labelPlural, labelRenameTo)
            .then(function() {
                $("#loadingIndicator").hide();
                inst.getTrendingLabels();
            }).catch(function(msg = "Couldn't accept trending label - please try again later") {
                $("#loadingIndicator").hide();
                $("#warningMessageBoxContent").text(msg);
                $("#warningMessageBox").show(200).delay(1500).hide(200);
            });
    }

    LabelRepositoryView.prototype.getTrendingLabels = function() {
        $("#loadingIndicator").show();
        $("#labelRepositoryTableContent").empty();
        let inst = this;
		this.imageMonkeyApi.getTrendingLabels()
            .then(function(data) {
                $("#loadingIndicator").hide();
                inst.populateLabelRepositoryTable(data);
            }).catch(function() {
                $("#loadingIndicator").hide();
                $("#warningMessageBoxContent").text("Couldn't get trending label - please try again later");
                $("#warningMessageBox").show(200).delay(1500).hide(200);
            });
    }

    LabelRepositoryView.prototype.populateLabelRepositoryTable = function(data) {
        $("#loadingIndicator").show();
        $("#labelRepositoryTable").hide();

        if (this.username === "")
            $("#authenticationNeededInfoMessage").show();

        for (var i = 0; i < data.length; i++) {
            var githubIssueUrl = '';
            if (data[i].github.issue.id !== -1)
                githubIssueUrl = ('<a href="' + this.trendingLabelsRepositoryUrl +
                    '/issues/' + data[i].github.issue.id + '">' +
                    '#' + data[i].github.issue.id + '</a>');
            var githubBranchUrl = ('<a href="' + this.labelsRepositoryUrl +
                '/tree/' + data[i].github.branch_name + '">' +
                data[i].github.branch_name + '</a>');
            var cellColor = data[i].github.issue.closed ? "#76ff03" : "#ffffff";
            var status = data[i].status;
            var button = '';

            var disabledStr = '';
            if (this.username === "")
                disabledStr = 'disabled ';

            var ciJobUrl = data[i].ci.job_url;
            var entryId = "trendingLabel" + i;

            var escapedTrendingLabel = escapeHtml(data[i].name);
            var escapedTrendingLabelDescription = escapeHtml(data[i].label.description);
            var escapedTrendingLabelPlural = escapeHtml(data[i].label.plural);
            var escapedTrendingLabelRenameTo = escapeHtml(data[i].rename_to);
            let labelCount = data[i].count;
            if (data[i].github.issue.closed) {
                status = 'closed';
                button = '';
            } else if (status === '') {
                status = 'open';
                button = ('<div class="ui fluid ' + disabledStr +
                    'button" data-label="' + escapedTrendingLabel +
                    '" data-label-plural="' + escapedTrendingLabel.trim() + 's' +
                    '" data-label-description="' + '' +
                    '" data-label-renameto="' + escapedTrendingLabel.trim() +
                    '" onclick="this.onAddTrendingLabel(this);">Add</div>');
            } else if (status === 'waiting for moderator approval') {
                var labelType = data[i].label.type;
                if (this.canAcceptTrendingLabelPermission) {
                    button = ('<div class="ui fluid button" data-label-description="' + escapedTrendingLabelDescription +
                        '" data-label-type="' + labelType + '" data-label="' + escapedTrendingLabel +
                        '" data-label-plural="' + escapedTrendingLabelPlural +
                        '" data-label-renameto="' + escapedTrendingLabelRenameTo +
                        '" onclick="this.onAcceptTrendingLabel(this);">Accept</div>');
                }
            } else if (status === 'building') {
                status = '<a href="' + ciJobUrl + '"><img src="/img/ci-build-in-progress.svg"></img></a>';
            } else if (status === 'build-passed') {
                status = '<a href="' + ciJobUrl + '"><img src="img/ci-build-passing.svg"></img></a>';
            } else if (status === 'build-canceled') {
                status = '<a href="' + ciJobUrl + '"><img src="img/ci-build-canceled.svg"></img></a>';
                button = ('<div class="ui fluid ' + disabledStr + ' button"' +
                    ' data-label="' + escapedTrendingLabel +
                    '" data-label-plural="' + escapedTrendingLabelPlural +
                    '" data-label-description="' + escapedTrendingLabelDescription +
                    '" data-label-renameto="' + escapedTrendingLabelRenameTo

                    +
                    '" onclick="this.onAddTrendingLabel(this);">Try again</div>');
            } else if (status === 'build-failed') {
                status = '<a href="' + ciJobUrl + '"><img src="img/ci-build-failed.svg"></img></a>';
                button = ('<div class="ui fluid ' + disabledStr + ' button"' +
                    ' data-label="' + escapedTrendingLabel +
                    '" data-label-plural="' + escapedTrendingLabelPlural +
                    '" data-label-description="' + escapedTrendingLabelDescription +
                    '" data-label-renameto="' + escapedTrendingLabelRenameTo +
                    '" onclick="this.onAddTrendingLabel(this);">Try again</div>');
            }

            elem = $(('<tr>' +
                '<td>' + escapedTrendingLabel + '</td>' +
                '<td class="center aligned">' + labelCount + '</td>' +
                '<td class="center aligned">' + githubIssueUrl + '</td>' +
                '<td class="center aligned">' + githubBranchUrl + '</td>' +
                '<td class="center aligned" bgcolor="' + cellColor + '">' + status + '</td>' +
                '<td class="">' + button + '</td>' +
                '</tr>'));
            $("#labelRepositoryTableContent").append(elem);
        }
        $("#labelRepositoryTable").tablesort();
        $("#loadingIndicator").hide();
        $("#labelRepositoryTable").show();
    }


    LabelRepositoryView.prototype.exec = function() {
        $("thead th.count").data("sortBy", function(th, td, tablesort) {
            return parseInt(td.text(), 10);
        });

        this.getTrendingLabels();

        let inst = this;
        $("#addTrendingLabelDlgYesButton").click(function(e) {
            e.preventDefault();
            inst.acceptTrendingLabel($("#addTrendingLabelDlg").attr("data-label-name"),
                inst.getSelectedLabelType(), inst.getLabelDescription(), inst.getLabelPlural(),
                inst.getRenameToLabel());

        });

        $("#labelTypeNormalRadioButton").click(function(e) {
            $("#addTrendingLabelDlgPluralFormInput").prop("disabled", false);
        });
        $("#labelTypeMetaRadioButton").click(function(e) {
            $("#addTrendingLabelDlgPluralFormInput").prop("disabled", true);
        });
    }

    return LabelRepositoryView;
}());
