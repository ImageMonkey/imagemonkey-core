var ContributionStatisticsView = (function() {


    var totalContributionsChartConfig = {
        type: "line",
        data: {
            datasets: [{
                    data: [],
                    backgroundColor: "black",
                    label: "Donations",
                    fill: false,
                    borderColor: "black",
                    borderWidth: 3,
                    pointRadius: 0
                },
                {
                    data: [],
                    backgroundColor: "violet",
                    label: "Labeled Objects",
                    fill: false,
                    borderColor: "violet",
                    borderWidth: 3,
                    pointRadius: 0
                },
                {
                    data: [],
                    backgroundColor: "red",
                    label: "Validations",
                    fill: false,
                    borderColor: "red",
                    borderWidth: 3,
                    pointRadius: 0
                },
                {
                    data: [],
                    backgroundColor: "blue",
                    label: "Annotations",
                    fill: false,
                    borderColor: "blue",
                    borderWidth: 3,
                    pointRadius: 0
                }
            ],

        },
        options: {
            responsive: true,
            maintainAspectRatio: !isMobileDevice(),
            title: {
                display: true,
                text: "Total Activity",
                fontColor: "black",
                fontSize: 17
            },
            scales: {
                xAxes: [{
                    type: "time",
                    distribution: 'series',
                    time: {
                        unit: "quarter"
                    },
                    ticks: {
                        fontColor: "black",
                    },
                }],
            },
            legend: {
                position: "bottom",
                labels: {
                    fontColor: "black"
                }
            }
        }
    };


    function ContributionStatisticsView(apiBaseUrl) {
        this.apiBaseUrl = apiBaseUrl;
        this.imageMonkeyApi = new ImageMonkeyApi(this.apiBaseUrl);
        this.imageMonkeyApi.setToken(getCookie("imagemonkey"));
    }

    ContributionStatisticsView.prototype.setSentryDSN = function(sentryDSN) {
        try {
            Sentry.init({
                dsn: sentryDSN,
            });
        } catch (e) {}
    }

    ContributionStatisticsView.prototype.exec = function() {
        $("#loadingSpinner").show();
        this.imageMonkeyApi.getContributionStatistics()
            .then(function(data) {
                const imageDonations = data["donations"];
                const imageLabels = data["labels"];
                const imageValidations = data["validations"];
                const imageAnnotations = data["annotations"];
                for (var i = 0; i < imageDonations.length; i++) {
                    totalContributionsChartConfig.data.datasets[0].data.push({
                        x: moment(imageDonations[i].date, "YYYY-MM-DD"),
                        y: imageDonations[i].count
                    });
                }
                totalContributionsChartConfig.data.datasets[0].data = simplify(totalContributionsChartConfig.data.datasets[0].data);

                for (var i = 0; i < imageLabels.length; i++) {
                    totalContributionsChartConfig.data.datasets[1].data.push({
                        x: moment(imageLabels[i].date, "YYYY-MM-DD"),
                        y: imageLabels[i].count
                    });
                }
                totalContributionsChartConfig.data.datasets[1].data = simplify(totalContributionsChartConfig.data.datasets[1].data);

                for (var i = 0; i < imageValidations.length; i++) {
                    totalContributionsChartConfig.data.datasets[2].data.push({
                        x: moment(imageValidations[i].date, "YYYY-MM-DD"),
                        y: imageValidations[i].count
                    });
                }
                totalContributionsChartConfig.data.datasets[2].data = simplify(totalContributionsChartConfig.data.datasets[2].data);

                for (var i = 0; i < imageAnnotations.length; i++) {
                    totalContributionsChartConfig.data.datasets[3].data.push({
                        x: moment(imageAnnotations[i].date, "YYYY-MM-DD"),
                        y: imageAnnotations[i].count
                    });
                }
                totalContributionsChartConfig.data.datasets[3].data = simplify(totalContributionsChartConfig.data.datasets[3].data);

                var totalContributionsChartCtx = document.getElementById("totalContributionsChart").getContext("2d");
                window.totalContributionsChart = new Chart(totalContributionsChartCtx, totalContributionsChartConfig);

            }).catch(function(e) {
                Sentry.captureException(e);
            });
        $("#loadingSpinner").hide();

    }

    return ContributionStatisticsView;
}());