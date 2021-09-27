UnifiedAnnotationModeComponent = {
    template: "#unified-annotation-mode-template",
    delimiters: ['${', '}$'],
    data() {
        return {
            visible: false,
            imageId: null
        }
    },
    methods: {
        hide: function() {
            this.visible = false;
        },
        show: function() {
            this.visible = true;
        },
        loadUnannotatedImage: function(imageAnnotationInfo) {
            this.$refs.annotationArea.loadUnannotatedImage(imageAnnotationInfo);
        },
        onRemoveLabel: function(label) {
            this.$refs.removeLabelConfirmationDialog.show(label);
        },
        onDuplicateLabelAdded: function(label) {
            this.$refs.simpleErrorPopup.show("Label " + label + " already exists");
        },
        onUnauthenticatedAccess: function() {
            this.$refs.simpleErrorPopup.show("Please log in first");
        },
        updateAnnotationsInAnnotationLabelList: function(labelUuid) {
            let annotationsChanged = this.$refs.annotationToolBox.annotationsChanged();
            if (annotationsChanged) {
                let annotationsOnCanvas = this.$refs.annotationToolBox.getAnnotationsOnCanvas();
                if (labelUuid !== null)
                    this.$refs.annotationLabelList.updateAnnotations(annotationsOnCanvas, labelUuid);
            }

        },
        onLabelSelected: function(currentSelectedLabelUuid, previousSelectedLabelUuid) {
            this.$refs.annotationToolBox.enableTools();
            this.updateAnnotationsInAnnotationLabelList(previousSelectedLabelUuid);

            let annotations = this.$refs.annotationLabelList.getAnnotationsForLabelUuid(currentSelectedLabelUuid);
            this.$refs.annotationToolBox.drawAnnotations(annotations);
        },
        onNoLabelSelected: function() {
            this.$refs.annotationToolBox.disableTools();
        },
        onHideUnifiedAnnotationMode: function() {
            this.visible = false;
        },
        onImageInImageGridClicked: function(imageAnnotationInfo) {
            this.imageId = imageAnnotationInfo.imageId;
            let url = new URL(window.location);

            if (imageAnnotationInfo.validationId !== null)
                url.searchParams.set("validation_id", imageAnnotationInfo.validationId);
            url.searchParams.set("image_id", imageAnnotationInfo.imageId);
            window.history.replaceState({}, null, url);

            this.show();
            this.loadUnannotatedImage(imageAnnotationInfo);
        },
        getActiveImageId: function() {
            return this.imageId;
        },
        onSaveChangesInUnifiedMode: function() {
            let currentSelectedLabelUuid = this.$refs.annotationLabelList.getCurrentSelectedLabelUuid();
            this.updateAnnotationsInAnnotationLabelList(currentSelectedLabelUuid);

            let inst = this;
            this.$refs.annotationLabelList.persistNewlyAddedLabels().then(function() {
                inst.$refs.annotationLabelList.persistAnnotations().then(function() {
                    EventBus.$emit("hideUnifiedAnnotationMode");
                    EventBus.$emit("showAnnotationBrowseMode");
                    EventBus.$emit("greyOutImageInImageGrid", inst.imageId);
                    EventBus.$emit("clearImageAnnotationCanvas");

                    let url = new URL(window.location);
                    let query = url.searchParams.get("query");
                    url.searchParams.delete("validation_id");
                    url.searchParams.delete("image_id");
                    window.history.replaceState({}, null, url);

                    EventBus.$emit("callSearchIfUrlOpenedInStandaloneMode", query);
                }).catch(function(e) {
                    Sentry.captureException(e);
                    EventBus.$emit("showErrorPopup", "Couldn't save changes");
                });
            }).catch(function(e) {
                Sentry.captureException(e);
                EventBus.$emit("showErrorPopup", "Couldn't save changes");
            });
        },
        onDiscardChangesInUnifiedMode: function() {
            EventBus.$emit("hideUnifiedAnnotationMode");
            EventBus.$emit("showAnnotationBrowseMode");
            EventBus.$emit("clearImageAnnotationCanvas");

            let url = new URL(window.location);
            let query = url.searchParams.get("query");
            url.searchParams.delete("validation_id");
            url.searchParams.delete("image_id");
            window.history.replaceState({}, null, url);

            EventBus.$emit("callSearchIfUrlOpenedInStandaloneMode", query);
        },
        onDeleteSelectedAnnotation: function() {
            this.$refs.removeAnnotationConfirmationDialog.show();
        },
        onConfirmRemoveAnnotation: function() {
            this.$refs.annotationToolBox.removeSelectedAnnotation();
        },
        onLoadImage: function(imageId, validationId = null) {
            let imageAnnotationInfo = new ImageAnnotationInfo();
            imageAnnotationInfo.imageId = imageId;
            imageAnnotationInfo.validationId = validationId;
            if (validationId !== null) {
                this.onImageInImageGridClicked(imageAnnotationInfo);
            } else {
                let inst = this;
                imageMonkeyApi.getImageDetails(imageId).then(function(data) {
                    let imageUrl = imageMonkeyApi.getImageUrl(imageId, data.unlocked);
                    imageAnnotationInfo.fullImageWidth = data.width;
                    imageAnnotationInfo.fullImageHeight = data.height;
                    imageAnnotationInfo.imageUnlocked = data.unlocked;
                    imageAnnotationInfo.imageUrl = imageUrl;
                    inst.onImageInImageGridClicked(imageAnnotationInfo);
                }).catch(function(e) {
                    Sentry.captureException(e);
                    EventBus.$emit("showErrorPopup", "Couldn't get image details");
                });
            }
        }
    },
    beforeDestroy: function() {
        EventBus.$off("removeLabel", this.onRemoveLabel);
        EventBus.$off("duplicateLabelAdded", this.onDuplicateLabelAdded);
        EventBus.$off("unauthenticatedAccess", this.onUnauthenticatedAccess);
        EventBus.$off("labelSelected", this.onLabelSelected);
        EventBus.$off("noLabelSelected", this.onNoLabelSelected);
        EventBus.$off("hideUnifiedAnnotationMode", this.onHideUnifiedAnnotationMode);
        EventBus.$off("imageInImageGridClicked", this.onImageInImageGridClicked);
        EventBus.$off("saveChangesInUnifiedMode", this.onSaveChangesInUnifiedMode);
        EventBus.$off("discardChangesInUnifiedMode", this.onDiscardChangesInUnifiedMode);
        EventBus.$off("deleteSelectedAnnotation", this.onDeleteSelectedAnnotation);
        EventBus.$off("confirmRemoveAnnotation", this.onConfirmRemoveAnnotation);
        EventBus.$off("loadImage", this.onLoadImage);
    },
    mounted: function() {
        EventBus.$on("removeLabel", this.onRemoveLabel);
        EventBus.$on("duplicateLabelAdded", this.onDuplicateLabelAdded);
        EventBus.$on("unauthenticatedAccess", this.onUnauthenticatedAccess);
        EventBus.$on("labelSelected", this.onLabelSelected);
        EventBus.$on("noLabelSelected", this.onNoLabelSelected);
        EventBus.$on("hideUnifiedAnnotationMode", this.onHideUnifiedAnnotationMode);
        EventBus.$on("imageInImageGridClicked", this.onImageInImageGridClicked);
        EventBus.$on("saveChangesInUnifiedMode", this.onSaveChangesInUnifiedMode);
        EventBus.$on("discardChangesInUnifiedMode", this.onDiscardChangesInUnifiedMode);
        EventBus.$on("deleteSelectedAnnotation", this.onDeleteSelectedAnnotation);
        EventBus.$on("confirmRemoveAnnotation", this.onConfirmRemoveAnnotation);
        EventBus.$on("loadImage", this.onLoadImage);
    }
}
