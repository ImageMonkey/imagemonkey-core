UnifiedAnnotationModeComponent = {
    template: "#unified-annotation-mode-template",
    delimiters: ['${', '}$'],
    data() {
        return {
            visible: false
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
        }
    },
    beforeDestroy: function() {
        EventBus.$off("removeLabel", this.onRemoveLabel);
        EventBus.$off("duplicateLabelAdded", this.onDuplicateLabelAdded);
        EventBus.$off("unauthenticatedAccess", this.onUnauthenticatedAccess);
    },
    mounted: function() {
        EventBus.$on("removeLabel", this.onRemoveLabel);
        EventBus.$on("duplicateLabelAdded", this.onDuplicateLabelAdded);
        EventBus.$on("unauthenticatedAccess", this.onUnauthenticatedAccess);
    }
}