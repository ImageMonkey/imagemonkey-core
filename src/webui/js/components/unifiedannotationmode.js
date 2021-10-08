UnifiedAnnotationModeComponent = {
    template: "#unified-annotation-mode-template",
    delimiters: ['${', '}$'],
    data() {
        return {
            visible: false,
            imageId: null,

            imageLoaded: false,
            labelsAndLabelSuggestionsLoaded: false,
            imageSpecificLabelsAndAnnotationsLoaded: false,
            imageInfoReceived: false,
            annotatorInitialized: false,
            labelListMarginTop: "mt-20"
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
            this.$refs.inlineInfoMessage.hide();
            this.labelListMarginTop = "mt-10";
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
            this.showLoadingSpinner();

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
        },
        getCanvasWidth: function() {
            let annotationToolBoxSidebarWidth = $("#annotation-toolbox-sidebar").width();
            let annotationLabelListWidth = $("#annotation-label-list").width();
            let windowWidth = $(window).width();
            let canvasWidth = windowWidth - (annotationToolBoxSidebarWidth + annotationLabelListWidth + 25);
            return canvasWidth;
        },


        showLoadingSpinner: function() {
            this.resetLoadedStates();
            this.$refs.loadingSpinner.show();
        },
        hideLoadingSpinner: function() {
            this.$refs.loadingSpinner.hide();
        },
        onImageLoaded: function() {
            this.imageLoaded = true;
            this.hideLoadingSpinnerIfEverythingIsLoaded();
        },
        onLabelsAndLabelSuggestionsLoaded: function() {
            this.labelsAndLabelSuggestionsLoaded = true;
            this.hideLoadingSpinnerIfEverythingIsLoaded();
        },
        onImageSpecificLabelsAndAnnotationsLoaded: function() {
            this.imageSpecificLabelsAndAnnotationsLoaded = true;
            if (this.imageInfoReceived && this.labelsAndLabelSuggestionsLoaded) {
                let maxCanvasWidth = this.getCanvasWidth();
                this.$refs.annotationArea.loadImage(maxCanvasWidth);
            }
            this.hideLoadingSpinnerIfEverythingIsLoaded();
        },
        hideLoadingSpinnerIfEverythingIsLoaded: function() {
            if (this.isEverythingLoaded()) {
                this.hideLoadingSpinner();
            }
        },
        onImageInfoReceived: function() {
            this.imageInfoReceived = true;
            if (this.imageInfoReceived && this.imageSpecificLabelsAndAnnotationsLoaded) {
                let maxCanvasWidth = this.getCanvasWidth();
                this.$refs.annotationArea.loadImage(maxCanvasWidth);
            }
            this.hideLoadingSpinnerIfEverythingIsLoaded();
        },
        onAnnotatorInitialized: function() {
            if (this.$refs.annotationLabelList.getCurrentSelectedLabelUuid() !== null) {
                this.$refs.annotationToolBox.enableTools();
                this.$refs.inlineInfoMessage.hide();
                this.labelListMarginTop = "mt-10";
            } else {
                this.$refs.annotationToolBox.disableTools();
                this.$refs.inlineInfoMessage.show("Add a label to start annotating");
                this.labelListMarginTop = "mt-20";
            }
            this.annotatorInitialized = true;
            this.hideLoadingSpinnerIfEverythingIsLoaded();
        },
        isEverythingLoaded: function() {
            if (this.imageInfoReceived && this.imageLoaded && this.labelsAndLabelSuggestionsLoaded && this.imageSpecificLabelsAndAnnotationsLoaded &&
                this.imageSpecificLabelsAndAnnotationsLoaded && this.annotatorInitialized) {
                return true;
            }
        },
        resetLoadedStates: function() {
            this.imageLoaded = false;
            //this.labelsAndLabelSuggestionsLoaded = false; //labels and label suggestions are only populated once, so do not reset them.
            this.imageSpecificLabelsAndAnnotationsLoaded = false;
            this.imageInfoReceived = false;
            this.annotatorInitialized = false;
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
        EventBus.$off("imageInfoReceived", this.onImageInfoReceived);
        EventBus.$off("ctrl+sPressed", this.onSaveChangesInUnifiedMode);
        EventBus.$off("ctrl+dPressed", this.onDiscardChangesInUnifiedMode);


        EventBus.$off("imageLoaded", this.onImageLoaded);
        EventBus.$off("labelsAndLabelSuggestionsLoaded", this.onLabelsAndLabelSuggestionsLoaded);
        EventBus.$off("imageSpecificLabelsAndAnnotationsLoaded", this.onImageSpecificLabelsAndAnnotationsLoaded);
        EventBus.$off("annotatorInitialized", this.onAnnotatorInitialized);
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
        EventBus.$on("imageInfoReceived", this.onImageInfoReceived);
        EventBus.$on("ctrl+sPressed", this.onSaveChangesInUnifiedMode);
        EventBus.$on("ctrl+dPressed", this.onDiscardChangesInUnifiedMode);


        EventBus.$on("imageLoaded", this.onImageLoaded);
        EventBus.$on("labelsAndLabelSuggestionsLoaded", this.onLabelsAndLabelSuggestionsLoaded);
        EventBus.$on("imageSpecificLabelsAndAnnotationsLoaded", this.onImageSpecificLabelsAndAnnotationsLoaded);
        EventBus.$on("annotatorInitialized", this.onAnnotatorInitialized);

    }
}