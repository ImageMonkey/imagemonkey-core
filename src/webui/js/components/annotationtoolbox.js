AnnotationToolboxComponent = {
    template: "#annotation-toolbox-template",
    data() {
        return {
            canvas: null,
            annotator: null,
            visible: false
        }
    },
    computed: {
        rectAnnotationModeIconColor() {
            if (this.annotator && this.annotator.getShape() === "Rectangle")
                return "color:blue";
            return "color: gray";
        },
        circleAnnotationModeIconColor() {
            if (this.annotator && this.annotator.getShape() === "Circle")
                return "color:blue";
            return "color: gray";
        },
        polygonAnnotationModeIconColor() {
            if (this.annotator && this.annotator.getShape() === "Polygon")
                return "color:blue";
            return "color: gray";
        },
        selectMoveAnnotationModeIconColor() {
            if (this.annotator && this.annotator.isSelectMoveModeEnabled())
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
        onAnnotatorMouseUp: function() {},
        onAnnotatorObjectDeselected: function() {},
        onAnnotatorObjectSelected: function() {},
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
        },
        onCanvasCreated: function(canvas) {
            this.canvas = canvas;
            this.annotator = new Annotator(this.canvas.fabric(), this.onAnnotatorObjectSelected.bind(this),
                this.onAnnotatorMouseUp.bind(this), this.onAnnotatorObjectDeselected.bind(this));
        },
        onImageInImageGridClicked: function(imageId) {
            this.visible = false;
        }

    },
    beforeDestroy: function() {
        EventBus.$off("canvasCreated", this.onCanvasCreated);
        EventBus.$off("imageInImageGridClicked", this.onImageInImageGridClicked);
    },
    mounted: function() {
        EventBus.$on("canvasCreated", this.onCanvasCreated);
        EventBus.$on("imageInImageGridClicked", this.onImageInImageGridClicked);
    }

};