ImageCanvasComponent = {
    template: "#imagecanvas-template",
    imageMonkeyApi: null,
    delimiters: ['${', '}$'],
    data() {
        return {
            visible: true 
        }
    },
    methods: {
        loadUnannotatedImage: function(validationId = undefined) {
            var that = this;
            imageMonkeyApi.getUnannotatedImage(validationId, null)
                .then(function(data) {
                    EventBus.$emit("unannotatedImageDataReceived", data, validationId);

                    canvas = new CanvasDrawer(that.$el.id);

                    //TODO: make max width configureable
                    let maxWidth = data.width;
                    if (maxWidth > 800)
                        maxWidth = 800;

                    let scaleFactor = maxWidth / data.width;
                    let width = scaleFactor * data.width;
                    let height = scaleFactor * data.height;

                    canvas.setWidth(width);
                    canvas.setHeight(height);

                    let backgroundImageUrl = data.url;
                    canvas.setCanvasBackgroundImageUrl(backgroundImageUrl, function() {
                        EventBus.$emit("canvasCreated", canvas);
                        EventBus.$emit("hideLoadingSpinner", null, null);
                    });
                }).catch(function() {
                    Sentry.captureException(e);
                });
        }
    },
    beforeDestroy: function() {
        EventBus.$off("loadUnannotatedImage", this.loadUnannotatedImage);
    },
    mounted: function() {
        EventBus.$on("loadUnannotatedImage", this.loadUnannotatedImage);
    }
};
