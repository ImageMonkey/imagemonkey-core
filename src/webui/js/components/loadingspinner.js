LoadingSpinnerComponent = {
	template: "#loadingspinner-template",
    props: ["visible"],
    data() {
        return {
            isvisible: true
        }
    },
    methods: {
        show: function() {
            this.isvisible = true;
        },
        hide: function() {
            this.isvisible = false;
        }
    },
	mounted: function() {
		var that = this;
		EventBus.$on('hideLoadingSpinner', (item, response) => {
			that.hide();
		});
	}
};
