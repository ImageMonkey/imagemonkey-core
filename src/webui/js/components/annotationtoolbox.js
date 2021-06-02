AnnotationToolboxComponent = {
	template: "#annotation-toolbox-template",
    data() {
		return {canvas: null,
		        annotator: null,
			   }
	},
	computed: {
		rectAnnotationModeIconColor() {
			if(this.annotator && this.annotator.getShape() === "Rectangle")
				return "color:blue";
			return "color: gray";
		},
		circleAnnotationModeIconColor() {
			if(this.annotator && this.annotator.getShape() === "Circle")
				return "color:blue";
			return "color: gray";
		},
		polygonAnnotationModeIconColor() {
			if(this.annotator && this.annotator.getShape() === "Polygon")
				return "color:blue";
			return "color: gray";
		},
		selectMoveAnnotationModeIconColor() {
			if(this.annotator && this.annotator.isSelectMoveModeEnabled())
				return "color:blue";
			return "color: gray";
		}
	},
	methods: {
        zoomOut: function() {
            this.canvas.fabric().setZoom(canvas.fabric().getZoom() / 1.1);
        },
		zoomIn: function() {
			this.canvas.fabric().setZoom(canvas.fabric().getZoom() * 1.1);
		},
		onAnnotatorMouseUp: function() {
		},
		onAnnotatorObjectDeselected: function() {
		},
		onAnnotatorObjectSelected: function() {
		},
		rectAnnotationMode: function() {
			this.annotator.disablePanMode();
            this.annotator.disableSelectMoveMode();
            this.annotator.setShape("Rectangle");
		},
		circleAnnotationMode: function() {
			this.annotator.disablePanMode();
            this.annotator.disableSelectMoveMode()
			this.annotator.setShape("Circle");	
		},
		polygonAnnotationMode: function() {
			this.annotator.disablePanMode();
            this.annotator.disableSelectMoveMode();
			this.annotator.setShape("Polygon");
		},
		selectMoveAnnotationMode: function() {
			this.annotator.disablePanMode();
			this.annotator.enableSelectMoveMode();
			this.annotator.setShape("");
		}

    },
	mounted: function() {
		var that = this;
		

		EventBus.$on("canvasCreated", canvas => {
			that.canvas = canvas;
			that.annotator = new Annotator(that.canvas.fabric(), that.onAnnotatorObjectSelected.bind(that), 
											that.onAnnotatorMouseUp.bind(that), that.onAnnotatorObjectDeselected.bind(that)); 
		});
	}

};
