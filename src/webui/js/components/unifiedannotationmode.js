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
		}
	},
	beforeDestroy: function() {
        EventBus.$off("removeLabel", this.onRemoveLabel);
    },
    mounted: function() {
        EventBus.$on("removeLabel", this.onRemoveLabel);
    }
}
