AnnotationLabelListComponent = {
	template: "#annotation-label-list-template",
	delimiters: ['${', '}$'],
	data() {
		return {
			labels: [] 
		  }
	},
	mounted: function() {
		var that = this;
		EventBus.$on("unannotatedImageDataReceived", (data) => {
			let onlyUnlockedLabels = false;
			//console.log(data);
			imageMonkeyApi.getLabelsForImage(data.uuid, onlyUnlockedLabels)
				.then(function(entries) {
					let composedLabels = []
					for(const entry of entries) {
						composedLabels.push(...buildComposedLabels(entry.label, entry.sublabels));
					}
					that.labels = composedLabels; 
				}).catch(function(e) {
					console.log(e.message);
					Sentry.captureException(e);
				});
		});
	}
};
