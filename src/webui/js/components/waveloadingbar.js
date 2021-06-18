WaveLoadingBarComponent = {
    template: "#wave-loading-bar-template",
    delimiters: ['${', '}$'],
    data() {
        return {
            visible: false
        }
    },
    computed: {},
    methods: {
        showWaveLoadingIndicator: function() {
            this.visible = true;
        },
        hideWaveLoadingIndicator: function() {
            this.visible = false;
        }
    },
    beforeDestroy: function() {
        EventBus.$off("showWaveLoadingIndicator", this.showWaveLoadingIndicator);
        EventBus.$off("hideWaveLoadingIndicator", this.hideWaveLoadingIndicator);
    },
    mounted: function() {
        EventBus.$on("showWaveLoadingIndicator", this.showWaveLoadingIndicator);
        EventBus.$on("hideWaveLoadingIndicator", this.hideWaveLoadingIndicator);
    }
};