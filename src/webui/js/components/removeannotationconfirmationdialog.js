RemoveAnnotationConfirmationDialogComponent = {
    template: "#remove-annotation-confirmation-dialog-template",
    delimiters: ['${', '}$'],
	data() {
        return {
            visible: false
        }
    },
    methods: {
        show: function() {
            this.visible = true;
        },
		hide: function() {
			this.visible = false;
		},
		onConfirmRemoveAnnotation: function() {
			EventBus.$emit("confirmRemoveAnnotation");
			this.hide();
		}
    },
    mounted: function() {
    }
};
