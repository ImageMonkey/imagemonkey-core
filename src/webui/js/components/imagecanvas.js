ImageCanvasComponent = {
    template: "#imagecanvas-template",
    imageMonkeyApi: null,
    delimiters: ['${', '}$'],
    data() {
        return {
            visible: true,
            canvas: null
        }
    },
    methods: {
        loadUnannotatedImage: function(validationId = undefined) {
            var that = this;
            imageMonkeyApi.getUnannotatedImage(validationId, null)
                .then(function(data) {
                    EventBus.$emit("unannotatedImageDataReceived", data, validationId);

                    that.canvas.clear();

                    //TODO: make max width configureable
                    let maxWidth = data.width;
                    if (maxWidth > 800)
                        maxWidth = 800;

                    let scaleFactor = maxWidth / data.width;
                    let width = scaleFactor * data.width;
                    let height = scaleFactor * data.height;

                    that.canvas.setWidth(width);
                    that.canvas.setHeight(height);

                    let backgroundImageUrl = data.url;
                    that.canvas.setCanvasBackgroundImageUrl(backgroundImageUrl, function() {
                        EventBus.$emit("canvasCreated", that.canvas);
                        EventBus.$emit("hideLoadingSpinner", null, null);
                    });
                }).catch(function() {
                    Sentry.captureException(e);
                });
        },
        canvas: function() {
            return this.canvas;
        },
        onClearImageAnnotationCanvas: function() {
            this.canvas.clear();

        }
    },
    beforeDestroy: function() {
        EventBus.$off("loadUnannotatedImage", this.loadUnannotatedImage);
        EventBus.$off("clearImageAnnotationCanvas", this.onClearImageAnnotationCanvas);
    },
    mounted: function() {
        EventBus.$on("loadUnannotatedImage", this.loadUnannotatedImage);
        EventBus.$on("clearImageAnnotationCanvas", this.onClearImageAnnotationCanvas);
        this.canvas = new CanvasDrawer(this.$el.id);
    }
};