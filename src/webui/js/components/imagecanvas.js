ImageCanvasComponent = {
    template: "#imagecanvas-template",
    imageMonkeyApi: null,
    delimiters: ['${', '}$'],
    data() {
        return {
            visible: false
        }
    },
    methods: {
        loadUnannotatedImage: function(validationId = undefined) {
            var that = this;
            imageMonkeyApi.getUnannotatedImage(validationId, null)
                .then(function(data) {
                    EventBus.$emit("unannotatedImageDataReceived", data);

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
        },
        onImageInImageGridClicked: function() {
            this.visible = true;
        }
    },
    beforeDestroy: function() {
        EventBus.$off("loadUnannotatedImage", this.loadUnannotatedImage);
        EventBus.$off("imageInImageGridClicked", this.onImageInImageGridClicked);

    },
    mounted: function() {
        EventBus.$on("loadUnannotatedImage", this.loadUnannotatedImage);
        EventBus.$on("imageInImageGridClicked", this.onImageInImageGridClicked);
    }
};