AnnotationBrowseFormContainerComponent = {
    template: "#annotation-browse-form-container-template",
    delimiters: ['${', '}$'],
    data() {
        return {
			visible: true
		}
	},
	methods: {
		hide: function() {
			this.visible = false;
		}
	}
}
