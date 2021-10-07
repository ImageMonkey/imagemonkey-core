AnnotationToolboxComponent = {
    template: "#annotation-toolbox-template",
    data() {
        return {
            canvas: null,
            annotator: null,
            visible: true,
            isDisabled: true,
            annotationHotkeyHandler: null
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
            if (this.annotator !== null) {
                this.isDisabled = false;
                this.annotator.unblock();
            } else {
                console.error("Couldn't enable annotator tools as annotator is null");
            }
        },
        disableTools: function() {
            if (this.annotator !== null) {
                this.isDisabled = true;
                this.annotator.block();
            } else {
                console.error("Couldn't disable annotator tools as annotator is null");
            }
        },
        zoomOut: function() {
            if (!this.isDisabled)
                this.canvas.fabric().setZoom(canvas.fabric().getZoom() / 1.1);
        },
        zoomIn: function() {
            if (!this.isDisabled)
                this.canvas.fabric().setZoom(canvas.fabric().getZoom() * 1.1);
        },
        deleteAnnotation: function() {
            this.annotator.deleteSelected();
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
            EventBus.$emit("annotatorInitialized");
        },
        drawAnnotations: function(annotations) {
            if (this.annotator !== null)
                this.annotator.loadAnnotations(annotations, this.canvas.fabric().backgroundImage.scaleX);
            else
                console.error("Couldn't draw annotations as annotator is null");
        },
        annotationsChanged: function() {
            if (this.annotator && this.annotator.isDirty())
                return true;
            return false;
        },
        getAnnotationsOnCanvas: function() {
            if (this.annotator) {
                return this.annotator.toJSON();
            }
            return [];
        },
        removeSelectedAnnotation: function() {
            this.deleteAnnotation();
        }

    },
    beforeDestroy: function() {
        EventBus.$off("canvasCreated", this.onCanvasCreated);
    },
    mounted: function() {
        EventBus.$on("canvasCreated", this.onCanvasCreated);

        let inst = this;
        this.annotationHotkeyHandler = new AnnotationHotkeyHandler();
        this.annotationHotkeyHandler.drawRectangle(function() {
            inst.rectAnnotationMode();
        });
        this.annotationHotkeyHandler.drawCircle(function() {
            inst.circleAnnotationMode();
        });
        this.annotationHotkeyHandler.drawPolygon(function() {
            inst.polygonAnnotationMode();
        });
        this.annotationHotkeyHandler.selectMove(function() {
            inst.selectMoveAnnotationMode();
        });
        this.annotationHotkeyHandler.zoomOut(function() {
            inst.zoomOut();
        });
        this.annotationHotkeyHandler.zoomIn(function() {
            inst.zoomIn();
        });
        this.annotationHotkeyHandler.deleteAnnotation(function() {
            EventBus.$emit("deleteSelectedAnnotation");
        });
    }

};