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
    },
    mounted: function() {
        EventBus.$on("hideLoadingSpinner", this.hide);
    }
};