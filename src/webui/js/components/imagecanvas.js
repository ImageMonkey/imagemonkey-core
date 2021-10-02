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
        loadUnannotatedImage: function(imageAnnotationInfo) {
            let validationId = imageAnnotationInfo.validationId;
            if (validationId !== undefined && validationId !== null) {
                this.loadUnannotatedImageByValidationId(validationId);
            } else {
                this.loadUnannotatedImageByImageId(imageAnnotationInfo.imageId, imageAnnotationInfo.fullImageWidth,
                    imageAnnotationInfo.fullImageHeight, imageAnnotationInfo.imageUrl, imageAnnotationInfo.imageUnlocked);
            }
        },
        loadUnannotatedImageByImageId: function(imageId, imageWidth, imageHeight, imageUrl, imageUnlocked) {
            EventBus.$emit("unannotatedImageDataReceived", imageId, null, imageUnlocked);

            this.canvas.clear();
            this.loadImage(imageUrl, imageWidth, imageHeight);
        },
        loadUnannotatedImageByValidationId: function(validationId) {
            let that = this;
            imageMonkeyApi.getUnannotatedImage(validationId, null)
                .then(function(data) {
                    EventBus.$emit("unannotatedImageDataReceived", data.uuid, validationId, data.unlocked);

                    let imageUrl = imageMonkeyApi.getImageUrl(data.uuid, data.unlocked);

                    that.canvas.clear();
                    that.loadImage(imageUrl, data.width, data.height);
                }).catch(function(e) {
                    let err = e.message;
                    if (err.includes("missing result set")) { //IMPROVE ME: not particularily nice to check for a specific string here, as the actual error message is likely to change.
                        EventBus.$emit("showErrorPopup", "The requested resource either doesn't exist or you do not have the appropriate permissions to access this page.", false);
                        EventBus.$emit("hideLoadingSpinner");
                    } else {
                        console.log(e);
                        Sentry.captureException(e);
                    }
                });
        },
        loadImage: function(imageUrl, imageWidth, imageHeight) {
            //TODO: make max width configureable
            let maxWidth = imageWidth;
            if (maxWidth > 800)
                maxWidth = 800;

            let scaleFactor = maxWidth / imageWidth;
            let width = scaleFactor * imageWidth;
            let height = scaleFactor * imageHeight;
            this.canvas.setWidth(width);
            this.canvas.setHeight(height);

            let backgroundImageUrl = imageUrl;
            let that = this;
            this.canvas.setCanvasBackgroundImageUrl(backgroundImageUrl, function() {
                EventBus.$emit("canvasCreated", that.canvas);
                EventBus.$emit("imageLoaded");
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