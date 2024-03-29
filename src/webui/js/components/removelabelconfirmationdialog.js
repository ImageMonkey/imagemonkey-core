RemoveLabelConfirmationDialogComponent = {
    template: "#remove-label-confirmation-dialog-template",
    delimiters: ['${', '}$'],
	data() {
        return {
            visible: false,
			labelToBeRemoved: null
        }
    },
    methods: {
        show: function(labelToBeRemoved) {
			this.labelToBeRemoved = labelToBeRemoved;
            this.visible = true;
        },
		hide: function() {
			this.visible = false;
		},
		onConfirmRemoveLabel: function() {
			EventBus.$emit("confirmRemoveLabel", this.labelToBeRemoved);
			this.hide();
		}
    },
    mounted: function() {
    }
};
