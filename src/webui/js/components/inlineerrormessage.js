InlineErrorMessageComponent = {
	template: "#inline-error-message-template",
    delimiters: ['${', '}$'],
    data() {
		return {
			errorMessage: ""
		}
	},
	computed: {
    },
    methods: {
		onShowInlineError: function() {
			this.errorMessage = errorMessage;
			setTimeout(() => this.errorMessage = "", 5000);
		}
	},
	beforeDestroy: function() {
		EventBus.$off("showInlineError", this.onShowInlineError);
	},
	mounted: function() {
        EventBus.$on("showInlineError", this.onShowInlineError);
	}
}
