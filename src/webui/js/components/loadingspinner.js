LoadingSpinnerComponent = {
    template: "#loadingspinner-template",
    data() {
        return {
            visible: false
        }
    },
    methods: {
        show: function() {
            this.visible = true;
        },
        hide: function() {
            this.visible = false;
        }
    },
    beforeDestroy: function() {
        EventBus.$off("hideLoadingSpinner", this.hide);
        EventBus.$off("showLoadingSpinner", this.show);
    },
    mounted: function() {
        EventBus.$on("hideLoadingSpinner", this.hide);
        EventBus.$on("showLoadingSpinner", this.show);
    }
};