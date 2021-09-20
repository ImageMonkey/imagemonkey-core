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
        loadUnannotatedImage: function(validationId) {
            this.$refs.annotationArea.loadUnannotatedImage(validationId);
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
        onLabelSelected: function(currentSelectedLabelUuid, previousSelectedLabelUuid) {
            this.$refs.annotationToolBox.enableTools();
            let annotationsChanged = this.$refs.annotationToolBox.annotationsChanged();
            if (annotationsChanged) {
                let annotationsOnCanvas = this.$refs.annotationToolBox.getAnnotationsOnCanvas();
                if (previousSelectedLabelUuid !== null)
                    this.$refs.annotationLabelList.updateAnnotations(annotationsOnCanvas, previousSelectedLabelUuid);
            }

            let annotations = this.$refs.annotationLabelList.getAnnotationsForLabelUuid(currentSelectedLabelUuid);
            this.$refs.annotationToolBox.drawAnnotations(annotations);
        },
        onNoLabelSelected: function() {
            this.$refs.annotationToolBox.disableTools();
        },
        onHideUnifiedAnnotationMode: function() {
            this.visible = false;
        },
        onImageInImageGridClicked: function(imageId, validationId) {
            this.iamgeId = imageId;
            let url = new URL(window.location);
            url.searchParams.set('validation_id', validationId);
            url.searchParams.set('image_id', imageId);
            window.history.replaceState({}, null, url);

            this.show();
            this.loadUnannotatedImage(validationId);
        },
        getActiveImageId: function() {
            return this.imageId;
        },
        onSaveChangesInUnifiedMode: function() {
            let inst = this;
            this.$refs.annotationLabelList.persistNewlyAddedLabels().then(function() {
                inst.$refs.annotationLabelList.persistAnnotations().then(function() {
                    EventBus.$emit("hideUnifiedAnnotationMode");
                    EventBus.$emit("showAnnotationBrowseMode");
                    EventBus.$emit("greyOutImageInImageGrid", inst.imageId);
                }).catch(function(e) {
                    Sentry.captureException(e);
                    EventBus.$emit("showErrorPopup", "Couldn't save changes");
                });
            }).catch(function(e) {
                Sentry.captureException(e);
                EventBus.$emit("showErrorPopup", "Couldn't save changes");
            });
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
    }
}