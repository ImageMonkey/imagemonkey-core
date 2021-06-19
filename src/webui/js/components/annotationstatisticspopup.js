AnnotationStatisticsPopupComponent = {
    template: "#annotation-statistics-popup-template",
    delimiters: ['${', '}$'],
    data() {
        return {
            annotatedStatistics: []
        }
    },
    computed: {},
    methods: {
        loadAnnotatedStatistics: function() {
            var that = this;
            imageMonkeyApi.getAnnotatedStatistics()
                .then(function(data) {
                    for (const elem of data) {
                        let percentage = 0;
                        if (elem.num.total !== 0)
                            percentage = Math.round(((elem.num.completed / elem.num.total) * 100));
                        let labelUrl = elem.label.name + "(" + elem.num.completed + "/" + elem.num.total + ")";
                        that.annotatedStatistics.push({
                            labelUrl: labelUrl,
                            percentage: percentage,
                            labelName: elem.label.name
                        });
                    }
                    EventBus.$emit("annotatedStatisticsLoaded");
                }).catch(function(e) {
                    console.log(e)
                    Sentry.captureException(e);
                });
        },
        labelClicked: function(label) {
            $("#" + this.$el.id).modal("hide");
            EventBus.$emit("annotatedStatisticsPopupLabelClicked", label);
        },
        onShowAnnotatedStatisticsPopup: function() {
            $("#" + this.$el.id).modal("show");
        }
    },
    beforeDestroy: function() {
        EventBus.$off("showAnnotatedStatisticsPopup", this.onShowAnnotatedStatisticsPopup);
    },
    mounted: function() {
        this.loadAnnotatedStatistics();

        EventBus.$on("showAnnotatedStatisticsPopup", this.onShowAnnotatedStatisticsPopup);
    }
};
