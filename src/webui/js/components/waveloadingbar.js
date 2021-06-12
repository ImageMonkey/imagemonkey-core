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
    mounted: function() {
        var that = this;
        EventBus.$on("showWaveLoadingIndicator", () => {
            console.log("showwww");
            that.showWaveLoadingIndicator();
        });

        EventBus.$on("hideWaveLoadingIndicator", () => {
            that.hideWaveLoadingIndicator();
        });
    }
};