AnnotationToolboxComponent = {
    template: "#annotation-toolbox-template",
    data() {
        return {
            canvas: null,
            annotator: null,
            visible: true,
            isDisabled: true
        }
    },
    computed: {
        rectAnnotationModeIconColor() {
            if (!this.isDisabled && this.annotator && this.annotator.getShape() === "Rectangle")
                return "color:blue";
            return "color: gray";
        },
        circleAnnotationModeIconColor() {
            if (!this.isDisabled && this.annotator && this.annotator.getShape() === "Circle")
                return "color:blue";
            return "color: gray";
        },
        polygonAnnotationModeIconColor() {
            if (!this.isDisabled && this.annotator && this.annotator.getShape() === "Polygon")
                return "color:blue";
            return "color: gray";
        },
        selectMoveAnnotationModeIconColor() {
            if (!this.isDisabled && this.annotator && this.annotator.isSelectMoveModeEnabled())
                return "color:blue";
            return "color: gray";
        },
        toolboxItemEnabled() {
            if (this.isDisabled)
                return "opacity-50 cursor-not-allowed";
            return "";
        }
    },
    methods: {
        enableTools: function() {
            this.isDisabled = false;
            this.annotator.unblock();
        },
        disableTools: function() {
            this.isDisabled = true;
            this.annotator.block();
        },
        zoomOut: function() {
            if (!this.isDisabled)
                this.canvas.fabric().setZoom(canvas.fabric().getZoom() / 1.1);
        },
        zoomIn: function() {
            if (!this.isDisabled)
                this.canvas.fabric().setZoom(canvas.fabric().getZoom() * 1.1);
        },
        onAnnotatorMouseUp: function() {},
        onAnnotatorObjectDeselected: function() {},
        onAnnotatorObjectSelected: function() {},
        rectAnnotationMode: function() {
            if (!this.isDisabled) {
                this.annotator.disablePanMode();
                this.annotator.disableSelectMoveMode();
                this.annotator.setShape("Rectangle");
            }
        },
        circleAnnotationMode: function() {
            if (!this.isDisabled) {
                this.annotator.disablePanMode();
                this.annotator.disableSelectMoveMode()
                this.annotator.setShape("Circle");
            }
        },
        polygonAnnotationMode: function() {
            if (!this.isDisabled) {
                this.annotator.disablePanMode();
                this.annotator.disableSelectMoveMode();
                this.annotator.setShape("Polygon");
            }
        },
        selectMoveAnnotationMode: function() {
            if (!this.isDisabled) {
                this.annotator.disablePanMode();
                this.annotator.enableSelectMoveMode();
                this.annotator.setShape("");
            }
        },
        onCanvasCreated: function(canvas) {
            this.canvas = canvas;
            this.annotator = new Annotator(this.canvas.fabric(), this.onAnnotatorObjectSelected.bind(this),
                this.onAnnotatorMouseUp.bind(this), this.onAnnotatorObjectDeselected.bind(this));
        }

    },
    beforeDestroy: function() {
        EventBus.$off("canvasCreated", this.onCanvasCreated);
    },
    mounted: function() {
        EventBus.$on("canvasCreated", this.onCanvasCreated);
    }

};