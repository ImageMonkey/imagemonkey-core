AnnotationToolboxComponent = {
    template: "#annotation-toolbox-template",
    data() {
        return {
            canvas: null,
            annotator: null,
            visible: true,
            isDisabled: true,
            annotationHotkeyHandler: null,
            tooltipsEnabled: false,
            showAllAnnotations: false
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
        showHideAllAnnotationsIconColor() {
            if (this.showAllAnnotations)
                return "color:blue";
            return "color:gray";
        },
        toolboxItemEnabled() {
            if (this.isDisabled)
                return "opacity-50 cursor-not-allowed";
            return "";
        },
        showHideAllAnnotationsToolboxItemEnabled() {
            if (this.showAllAnnotations)
                return "";
            if (this.isDisabled)
                return "opacity-50 cursor-not-allowed";
            return "";
        },
        rectAnnotationModeTooltip: function() {
            return (this.tooltipsEnabled) ? "Rectangle (r)" : null;
        },
        circleAnnotationModeTooltip: function() {
            return (this.tooltipsEnabled) ? "Circle (c)" : null;
        },
        polygonAnnotationModeTooltip: function() {
            return (this.tooltipsEnabled) ? "Polygon (p)" : null;
        },
        selectMoveAnnotationModeTooltip: function() {
            return (this.tooltipsEnabled) ? "Select & Move (s)" : null;
        },
        zoomInTooltip: function() {
            return (this.tooltipsEnabled) ? "Zoom In (+)" : null;
        },
        zoomOutTooltip: function() {
            return (this.tooltipsEnabled) ? "Zoom Out (-)" : null;
        },
        removeAnnotationTooltip: function() {
            return (this.tooltipsEnabled) ? "Remove Annotation (del)" : null;
        },
        showHideAllAnnotationsTooltip: function() {
            return (this.showAllAnnotations) ? "Hide all Annotations" : "Show all Annotations";
        }
    },
    methods: {
        enableTools: function() {
            if (this.annotator !== null) {
                this.tooltipsEnabled = true;
                this.isDisabled = false;
                this.annotator.unblock();
            } else {
                console.error("Couldn't enable annotator tools as annotator is null");
            }
        },
        disableTools: function() {
            if (this.annotator !== null) {
                this.tooltipsEnabled = false;
                this.isDisabled = true;
                this.annotator.block();
            } else {
                console.error("Couldn't disable annotator tools as annotator is null");
            }
        },
        zoomOut: function() {
            if (!this.isDisabled)
                this.canvas.fabric().setZoom(this.canvas.fabric().getZoom() / 1.1);
        },
        zoomIn: function() {
            if (!this.isDisabled)
                this.canvas.fabric().setZoom(this.canvas.fabric().getZoom() * 1.1);
        },
        deleteAnnotation: function() {
            EventBus.$emit("deleteSelectedAnnotation");
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
            if (!this.annotator) {
                this.annotator = new Annotator(this.canvas.fabric(), this.onAnnotatorObjectSelected.bind(this),
                    this.onAnnotatorMouseUp.bind(this), this.onAnnotatorObjectDeselected.bind(this));
            } else {
                this.annotator.reset(false);
            }
            this.showAllAnnotations = false;
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
        },
        showHideAllAnnotations: function() {
            this.showAllAnnotations = !this.showAllAnnotations;
            if (this.showAllAnnotations) {
                this.disableTools();
                EventBus.$emit("showAllAnnotations");
            } else {
                this.enableTools();
                EventBus.$emit("hideAllAnnotations");
            }
        },
        removeAllAnnotations: function() {
            this.annotator.deleteAll();
        },
        allAnnotationsAreShown: function() {
            return this.showAllAnnotations;
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