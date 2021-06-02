ImageCanvasComponent = {
  template: "#imagecanvas-template", 
  imageMonkeyApi: null,
  delimiters: ['${', '}$'],
  methods: {
	loadUnannotatedImage: function() {
		var that = this;
		console.log("id = " + that.$el.id);
		console.log("id = ", that);
		imageMonkeyApi.getUnannotatedImage(undefined, null)
		.then(function(data) {
			EventBus.$emit("unannotatedImageDataReceived", data);

			canvas = new CanvasDrawer(that.$el.id);

			let maxWidth = data.width;
			if(maxWidth > 800)
				maxWidth = 800;
				
			let scaleFactor = maxWidth / data.width;
			let width = scaleFactor * data.width;
			let height = scaleFactor * data.height;

			canvas.setWidth(width); 
        	canvas.setHeight(height);
			console.log("width = " +document.getElementById("annotation-area-container").clientWidth);
			console.log("height = " +document.getElementById("annotation-area-container").clientHeight);
			console.log("annotationLabelListContainer width: " +$("#annotation-label-list-container").width());
			
			let backgroundImageUrl = data.url;
			canvas.setCanvasBackgroundImageUrl(backgroundImageUrl, function() {
				EventBus.$emit("canvasCreated", canvas);
				EventBus.$emit("hideLoadingSpinner", null, null);
			});
		}).catch(function() {
			Sentry.captureException(e); 
		});
	}
  },
  mounted: function() {
		var that = this;
		EventBus.$on("loadUnannotatedImage", (item, response) => {
			that.loadUnannotatedImage();
		});
	}
};
