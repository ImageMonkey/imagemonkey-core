AnnotationNavbarComponent = {
    template: "#annotation-navigationbar-template",
    data() {
        return {
            visible: true
        }
    },
    methods: {
        save: function() {
            this.$parent.$refs.annotationLabelList.persistNewlyAddedLabels().then(function() {
               	//TODO: persist annotations 
            }).catch(function(e) {
                Sentry.captureException(e);
                EventBus.$emit("showErrorPopup", "Couldn't save changes");
            });
        }
    },
    beforeDestroy: function() {
        EventBus.$off("save", this.save);
    },
    mounted: function() {
        EventBus.$on("save", this.save);
    }
};
