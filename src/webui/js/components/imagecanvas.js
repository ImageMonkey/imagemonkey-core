ImageCanvasComponent = {
    template: "#imagecanvas-template",
    imageMonkeyApi: null,
    delimiters: ['${', '}$'],
    data() {
        return {
            visible: true,
            canvas: null,
            imageUrl: null,
            imageHeight: null,
            imageWidth: null
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
            this.imageUrl = imageUrl;
            this.imageHeight = imageHeight;
            this.imageWidth = imageWidth;
            EventBus.$emit("imageInfoReceived");
        },
        loadUnannotatedImageByValidationId: function(validationId) {
            let that = this;
            imageMonkeyApi.getUnannotatedImage(validationId, null)
                .then(function(data) {
                    EventBus.$emit("unannotatedImageDataReceived", data.uuid, validationId, data.unlocked);

                    let imageUrl = imageMonkeyApi.getImageUrl(data.uuid, data.unlocked);

                    that.canvas.clear();
                    that.imageUrl = imageUrl;
                    that.imageHeight = data.height;
                    that.imageWidth = data.width;
                    EventBus.$emit("imageInfoReceived");
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
        loadImage: function(maxCanvasWidth) {
            this.canvas.clear();
            let maxWidth = maxCanvasWidth;
            if (maxWidth > maxCanvasWidth)
                maxWidth = maxCanvasWidth;

            let scaleFactor = maxWidth / this.imageWidth;
            let width = scaleFactor * this.imageWidth;
            let height = scaleFactor * this.imageHeight;
            this.canvas.setWidth(width);
            this.canvas.setHeight(height);

            let backgroundImageUrl = this.imageUrl;
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