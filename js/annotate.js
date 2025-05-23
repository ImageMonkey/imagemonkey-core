fabric.Canvas.prototype.getItemByAttr = function(attr, name) {
    var object = null,
        objects = this.getObjects();
    for (var i = 0, len = this.size(); i < len; i++) {
        if (objects[i][attr] && objects[i][attr] === name) {
            object = objects[i];
            break;
        }
    }
    return object;
};

fabric.Canvas.prototype.removeItemsByAttr = function(attr, name) {
    var object = null,
        objects = this.getObjects();
    var i = this.size();
    while (i--) {
        if (objects[i][attr] && objects[i][attr] === name) {
            objects[i].remove();
        }
    }
};


function generateRandomId() {
    function s4() {
        return Math.floor((1 + Math.random()) * 0x10000)
            .toString(16)
            .substring(1);
    }
    return s4() + s4() + '-' + s4() + '-' + s4() + '-' +
        s4() + '-' + s4() + s4() + s4();
}


var Polygon = (function() {
    function Polygon(canvas, polygonVertexSize = 5) {
        var inst = this;
        this.canvas = canvas;
        this.polygonMode = true;
        this.pointArray = new Array();
        this.lineArray = new Array();
        this.activeLine = null;
        this.max = 999999;
        this.min = 99;
        this.activeShape = false;
        this.index = 0;
        this.currentId = "";
        this.polygons = {}
        this.currentlyShownPolygonId = "";
        this.polygonVertexSize = polygonVertexSize;
    }

    Polygon.prototype.setPolygonVertexSize = function(polygonVertexSize) {
        this.polygonVertexSize = polygonVertexSize;
    };

    Polygon.prototype.clear = function() {
        this.polygonMode = true;
        this.pointArray.length = 0;
        this.lineArray.length = 0;
        this.activeLine = null;
        this.activeShape = false;
        this.index = 0;
        this.currentId = "";
    };

    Polygon.prototype.reset = function() {
        this.clear();
        this.polygons = {}
    };

    Polygon.prototype.getCurrentId = function() {
        return this.currentId;
    }

    Polygon.prototype.addPoint = function(options) {
        /*if(this.currentlyShownPolygonId !== ""){
          this.getCurrentlyEditedPolygon().selectable = true;
          this.hidePolyPoints(this.currentlyShownPolygonId);
          return;
        }*/


        var random = Math.floor(Math.random() * (this.max - this.min + 1)) + this.min;
        var id = new Date().getTime() + random;
        var pointer = this.canvas.getPointer(options.e);
        var circle = new fabric.Circle({
            radius: this.polygonVertexSize,
            fill: '#ffffff',
            stroke: '#333333',
            strokeWidth: 0.5,
            left: pointer.x, //(options.e.layerX/this.canvas.getZoom()),
            top: pointer.y, //(options.e.layerY/this.canvas.getZoom()),
            selectable: true,
            hasBorders: false,
            hasControls: false,
            originX: 'center',
            originY: 'center',
            index: this.index,
            id: generateRandomId()
        });

        this.index += 1;

        //if it's the first point
        if (this.pointArray.length == 0) {
            circle.set({
                fill: 'red'
            })

            this.currentId = generateRandomId();
        }
        circle.set({
            'belongsToPolygon': this.currentId
        });


        //var points = [(options.e.layerX/this.canvas.getZoom()),(options.e.layerY/this.canvas.getZoom()),(options.e.layerX/this.canvas.getZoom()),(options.e.layerY/this.canvas.getZoom())];
        var points = [pointer.x, pointer.y, pointer.x, pointer.y];
        line = new fabric.Line(points, {
            strokeWidth: 2,
            fill: '#999999',
            stroke: '#999999',
            class: 'line',
            originX: 'center',
            originY: 'center',
            selectable: false,
            hasBorders: false,
            hasControls: false,
            evented: false
        });
        if (this.activeShape) {
            var pos = this.canvas.getPointer(options.e);
            var points = this.activeShape.get("points");
            points.push({
                x: pos.x,
                y: pos.y
            });
            var polygon = new fabric.Polygon(points, {
                stroke: 'red',
                strokeWidth: 5,
                fill: 'red',
                opacity: 0.5,
                selectable: false,
                hasBorders: false,
                hasControls: false,
                evented: false
            });
            this.canvas.remove(this.activeShape);
            this.canvas.add(polygon);
            this.activeShape = polygon;
            this.canvas.renderAll();
        } else {
            //var polyPoint = [{x:(options.e.layerX/this.canvas.getZoom()),y:(options.e.layerY/this.canvas.getZoom())}];
            var polyPoint = [{
                x: pointer.x,
                y: pointer.y
            }];
            var polygon = new fabric.Polygon(polyPoint, {
                stroke: 'red',
                strokeWidth: 5,
                fill: 'red',
                opacity: 0.5,
                selectable: false,
                hasBorders: false,
                hasControls: false,
                evented: false
            });
            this.activeShape = polygon;
            this.canvas.add(polygon);
        }
        this.activeLine = line;

        this.pointArray.push(circle);
        this.lineArray.push(line);

        this.canvas.add(line);
        this.canvas.add(circle);
        this.canvas.selection = false;
    };

    Polygon.prototype.showPolyPoints = function(polygonId) {
        this.currentlyShownPolygonId = polygonId;
        var points = this.polygons[polygonId];
        var inst = this;
        $.each(points, function(index, point) {
            inst.canvas.add(point);
            point.setCoords();
        });
        inst.canvas.renderAll();
    }

    Polygon.prototype.movePolyPoints = function(polygonId, moveX, moveY) {
        var points = this.polygons[polygonId];
        var inst = this;
        for (var i = 0; i < points.length; i++) {
            points[i].setLeft(points[i].left + moveX);
            points[i].setTop(points[i].top + moveY);
        }
        this.polygons[polygonId] = points;
    }

    Polygon.prototype.hidePolyPoints = function(polygonId) {
        this.currentlyShownPolygonId = "";
        var points = this.polygons[polygonId];
        var inst = this;
        $.each(points, function(index, point) {
            inst.canvas.remove(point);
        });
        inst.canvas.renderAll();
    }

    Polygon.prototype.removePolygonById = function(polygonId) {
        if (polygonId in this.polygons) {
            this.hidePolyPoints(polygonId);
            delete this.polygons[polygonId];
        }
    }

    Polygon.prototype.getPaintedPolygonById = function(id) {
        return this.canvas.getItemByAttr("id", id);
    }

    Polygon.prototype.isInPolyEditMode = function() {
        return (this.currentlyShownPolygonId === "") ? false : true;
    }

    Polygon.prototype.getCurrentlyShownPolyPoints = function() {
        return this.polygons[this.currentlyShownPolygonId];
    }

    Polygon.prototype.getCurrentlyEditedPolygon = function() {
        return this.canvas.getItemByAttr("id", this.currentlyShownPolygonId);
    }

    Polygon.prototype.updateCurrentlyEditedPolygon = function(points) {
        this.updatePolygonPoints(this.currentlyShownPolygonId, points);
    }

    Polygon.prototype.updatePolygonPoints = function(id, points) {
        var oldPolygon = this.canvas.getItemByAttr("id", id);
        oldPolygon.points = points;
        this.canvas.renderAll();
    }

    Polygon.prototype.addPolygon = function(polygon) {
        var points = polygon.points;
        var polyPoints = new Array();
        for (var i = 0; i < points.length; i++) {
            var circle = new fabric.Circle({
                radius: this.polygonVertexSize,
                fill: '#ffffff',
                stroke: '#333333',
                strokeWidth: 0.5,
                left: points[i].x,
                top: points[i].y,
                selectable: true,
                hasBorders: false,
                hasControls: false,
                originX: 'center',
                originY: 'center',
                index: i,
                isPolygonHandle: true,
                belongsToPolygon: polygon.id
            });
            polyPoints.push(circle);
        }
        this.polygons[polygon.id] = polyPoints;
    }

    Polygon.prototype.generatePolygon = function() {
        var points = new Array();
        var polyPoints = new Array();
        var inst = this;
        $.each(this.pointArray, function(index, point) {
            points.push({
                x: point.left,
                y: point.top
            });

            var circle = new fabric.Circle({
                radius: inst.polygonVertexSize,
                fill: '#ffffff',
                stroke: '#333333',
                strokeWidth: 0.5,
                left: point.left,
                top: point.top,
                selectable: true,
                hasBorders: false,
                hasControls: false,
                originX: 'center',
                originY: 'center',
                index: index,
                isPolygonHandle: true,
                belongsToPolygon: inst.currentId
            });

            polyPoints.push(circle);

            this.canvas.remove(point);
        });

        this.polygons[this.currentId] = polyPoints;


        $.each(this.lineArray, function(index, line) {
            this.canvas.remove(line);
        });
        this.canvas.remove(this.activeShape).remove(this.activeLine);
        var polygon = new fabric.Polygon(points, {
            stroke: '#F00',
            strokeWidth: 5,
            fill: 'transparent',
            hasBorders: true,
            hasControls: false,
            objectCaching: false,
            selectable: false,
            lockMovementX: true,
            lockMovementY: true,
            selected: false,
            evented: false,
            id: this.currentId
        });
        this.canvas.add(polygon);
        this.activeLine = null;
        this.activeShape = null;
        this.polygonMode = false;
        this.canvas.selection = true;
        this.index = 0;
        this.currentId = "";

        this.showPolyPoints(polygon.id);
    };

    Polygon.prototype.firstId = function() {
        if (this.pointArray.length === 0)
            return -1;
        return this.pointArray[0].id;
    };

    Polygon.prototype.move = function(pointer) {
        if (this.activeLine && this.activeLine.class == "line") {
            this.activeLine.set({
                x2: pointer.x,
                y2: pointer.y
            });

            var points = this.activeShape.get("points");
            points[this.pointArray.length] = {
                x: pointer.x,
                y: pointer.y
            }

            this.activeShape.set({
                points: points
            });
            this.canvas.renderAll();
        }
    };


    return Polygon;
}());




var FreeDrawer = (function() {
    function FreeDrawer(canvas, closedPathMode = true) {
        var inst = this;
        this.canvas = canvas;
        this.pointArray = new Array();
        this.closedPathMode = closedPathMode;
    };

    FreeDrawer.prototype.clear = function() {
        this.pointArray.length = 0;
    };

    FreeDrawer.prototype.isClosedPathMode = function() {
        return this.closedPathMode;
    };

    FreeDrawer.prototype.enableClosedPathMode = function() {
        this.closedPathMode = true;
        this.clear();
    };

    FreeDrawer.prototype.disableClosedPathMode = function() {
        this.closedPathMode = false;
        this.clear();
    };

    FreeDrawer.prototype.addPoint = function(options) {
        var pointer = this.canvas.getPointer(options.e);
        var circle = new fabric.Circle({
            radius: 5,
            fill: '#ffffff',
            stroke: '#333333',
            strokeWidth: 0.5,
            left: pointer.x,
            top: pointer.y,
            selectable: false,
            hasBorders: false,
            hasControls: false,
            originX: 'center',
            originY: 'center'
        });
        this.pointArray.push(circle);
    };


    FreeDrawer.prototype.move = function(pointer) {
        var circle = new fabric.Circle({
            radius: 5,
            fill: '#ffffff',
            stroke: '#333333',
            strokeWidth: 0.5,
            left: pointer.x,
            top: pointer.y,
            selectable: false,
            hasBorders: false,
            hasControls: false,
            originX: 'center',
            originY: 'center'
        });
        this.pointArray.push(circle);
    };

    FreeDrawer.prototype.generatePolygon = function() {
        var simplifiedPoints = simplify(this.canvas.freeDrawingBrush._points, 0.8, false);


        this.canvas.freeDrawingBrush._points = simplifiedPoints;

        this.canvas.isDrawingMode = false;
        this.canvas.freeDrawingBrush.onMouseUp();
    };

    FreeDrawer.prototype.isPathClosed = function(pointer) {
        var margin = 5;
        if (this.pointArray.length > 30) {
            var left = this.pointArray[0].left - margin;
            var right = this.pointArray[0].left + margin;
            var top = this.pointArray[0].top + margin;
            var bottom = this.pointArray[0].top - margin;
            if (((pointer.x >= left) && (pointer.x <= right)) && ((pointer.y >= bottom) && (pointer.y <= top)))
                return true;
        }
        return false;
    };

    return FreeDrawer;
}());




var Annotator = (function() {
    function Annotator(canvas, objSelected, mouseUp, objDeselected) {
        var inst = this;
        this.canvas = canvas;
        this.className = "Rectangle";
        this.isDrawing = false;
        this.overObject = false;
        this.blocked = false;
        this.type = "Rectangle";
        this.polygon = new Polygon(this.canvas);
        this.objSelected = objSelected;
        this.objDeselected = objDeselected;
        this.history = new Array();
        this.isRedoing = false;
        this.currentHistoryPosition = 0;
        this.isPanMode = false;
        this.panning = false;
        this.gridVisible = false;
        this.gridSize = 20;
        this.cellSize = 100;
        this.selectedBlocks = {};
        this.selectedBlocksPoints = {};
        this.recentlyAddedBlocks = {};
        this.recentlyDeletedBlocks = {};
        this.freeDrawing = new FreeDrawer(this.canvas);
        this.mouseUpCB = mouseUp;
        this.smartAnnotation = false;
        this.brushColor = "red";
        this.brushWidth = 1;
        this.brushType = "PencilBrush";
        this.smartAnnotationData = [];
        this.defaultStrokeWidth = 5;
        this.maxStrokeWidth = 5;
        this.minStrokeWidth = 2;
        this.isSelectMoveMode = false;
        this.refinementsPerAnnotation = {};
        this._refAnnotations = [];
        this._highlightOnMouseOver = false;
        this._annotationLabelOverviewMapping = {};

        this.setBrushType(this.brushType);
        this.setBrushColor(this.brushColor);
        this.setBrushWidth(this.brushWidth);

        this.bindEvents();
    }

    Annotator.prototype.setPolygonVertexSize = function(polygonVertexSize) {
        this.polygon.setPolygonVertexSize(polygonVertexSize);
    };

    Annotator.prototype._selectObjectByMouse = function(pointer) {
        var point = new fabric.Point(pointer.x, pointer.y);
        var objects = this.canvas.getObjects();
        var foundObj = null;
        var hasControls;
        for (var i = 0; i < objects.length; i++) {
            var boundingRect = objects[i].getBoundingRect();
            if ((pointer.x >= boundingRect.left && pointer.x <= (boundingRect.left + boundingRect.width)) &&
                (pointer.y >= boundingRect.top && pointer.y <= (boundingRect.top + boundingRect.height))) {
                if (foundObj) { //we already have found one object that lies within the position of the cursor
                    if (objects[i].isContainedWithinObject(foundObj)) { //is there another object that is even smaller? (i.e is fully contained with in existing one)
                        foundObj.set({
                            hasBorders: false,
                            hasControls: false,
                            evented: false,
                            selectable: false,
                            selected: false
                        }); //if so, remove the selected property again..it's not the object we are looking for
                    }
                }


                hasControls = true;
                if (objects[i]["type"] === "polygon")
                    hasControls = false; //currently we do not support controls on polygon objects

                objects[i].set({
                    hasBorders: true,
                    hasControls: hasControls,
                    evented: true,
                    selectable: true,
                    selected: true
                });
                foundObj = objects[i];
            } else {
                objects[i].set({
                    hasBorders: false,
                    hasControls: false,
                    evented: false,
                    selectable: false,
                    selected: false
                });
            }
        }
    }

    //de-select any selected objects + group and make it non-selectable
    Annotator.prototype._silenceAllObjects = function() {
        var objects = this.canvas.getObjects();
        for (var i = 0; i < objects.length; i++) {
            objects[i].set({
                evented: false,
                selectable: false,
                selected: false
            });
        }

        this.canvas.discardActiveObject();
        this.canvas.discardActiveGroup();
        this.canvas.renderAll();
    }

    Annotator.prototype.bindEvents = function() {
        var inst = this;

        //currently disabled
        /*fabric.util.addListener(this.canvas.upperCanvasEl, 'dblclick', function(e) {
          //only if in polygon mode and shape is closed
          if((inst.type === "Polygon") && (inst.polygon.getCurrentId() === "")){
            if (inst.canvas.findTarget(e)) {
                var obj = inst.canvas.findTarget(e);
                if (obj.type === 'polygon') {
                  obj.selectable = false;
                  inst.canvas.discardActiveObject();
                  inst.polygon.showPolyPoints(obj.id);
                }
            }
          }
        });*/

        inst.canvas.on('mouse:down', function(o) {
            if (o) {
                inst.onMouseDown(o);
                if (inst.isPanMode)
                    inst.panning = true;
                if (inst.isSelectMoveMode) {
                    var pointer = inst.canvas.getPointer(o.e);
                    inst._selectObjectByMouse(pointer);
                }
            }
        });

        inst.canvas.on('before:selection:cleared', function() {
            inst.objDeselected();
        });

        inst.canvas.on('mouse:move', function(o) {
            if (o)
                inst.onMouseMove(o);
        });
        inst.canvas.on('mouse:up', function(o) {
            inst.onMouseUp(o);
            if (inst.isPanMode)
                inst.panning = false;
        });
        inst.canvas.on('object:moving', function(o) {
            inst.disable();

            var p = o.target;
            if (("isPolygonHandle" in p) && p["isPolygonHandle"]) {
                if (inst.polygon.isInPolyEditMode()) {
                    var points = inst.polygon.getPaintedPolygonById(p.belongsToPolygon).points;

                    //check for the specific event type here. In case it's a mousemove event
                    //we directly have the 'movementX'/'movementY' information. In case it's
                    //not a mousemove event (as we are running on a mobile/tablet device),
                    //check the changedTouches list and calculate the relative x/y positions manually.
                    //caution: this code doesn't handle multi touch events properly!
                    let newX = null;
                    let newY = null;
                    if (o.e.type === "mousemove") {
                        newX = points[p.index].x + o.e.movementX;
                        newY = points[p.index].y + o.e.movementY;
                    } else {
                        let canvasBoundingRect = o.e.target.getBoundingClientRect();
                        newX = o.e.changedTouches.item(0).clientX - canvasBoundingRect.left;
                        newY = o.e.changedTouches.item(0).clientY - canvasBoundingRect.top;
                    }

                    points[p.index] = {
                        x: newX,
                        y: newY
                    }
                    //points[p.index] = {x: p.getCenterPoint().x, y: p.getCenterPoint().y};
                    inst.polygon.getPaintedPolygonById(p.belongsToPolygon).setCoords();
                    inst.polygon.updatePolygonPoints(p.belongsToPolygon, points);
                } else {
                    var obj = inst.canvas.getActiveObject();
                    if (obj.type === "polygon") {
                        inst.polygon.movePolyPoints(obj.id, o.e.movementX, o.e.movementY);
                        obj.setCoords();
                        inst.canvas.renderAll();
                    }
                }
            }
            /*else if(p.type === "polygon"){
              var points = p.points;
              for(var i = 0; i < points.length; i++){
                points[i] = {x: (points[i].x + o.e.movementX), y: (points[i].y + o.e.movementY)};
              }
              p.set({left: p.left - o.e.movementX});
              p.set({top: p.top - o.e.movementY});
              p.points = points;
              p.setCoords();
              inst.canvas.renderAll();
            }*/



        });
        inst.canvas.on('object:selected', function(o) {
            inst.objSelected();
        });

        inst.canvas.on('object:scaling', function(o) {
            var e = o.target;
            //in case stroke width gets bigger, than max stroke width
            //rescale it to max stroke width
            if (e.strokeWidth * e.scaleX > inst.maxStrokeWidth) {
                e.objectCaching = false;
                e.strokeWidth = inst.maxStrokeWidth / ((e.scaleX + e.scaleY) / 2);
            }
            //in case stroke width gets smaller, than min stroke width
            //rescale it to min stroke width
            if (e.strokeWidth * e.scaleX < inst.minStrokeWidth) {
                e.objectCaching = false;
                e.strokeWidth = inst.minStrokeWidth / ((e.scaleX + e.scaleY) / 2);
            }
        });

        inst.canvas.on('object:modified', function(o) {});
        inst.canvas.on('object:added', function(o) {
            inst.saveState();
        });

        inst.canvas.on('mouse:over', function(o) {
            if (o.target && o.target.id == inst.polygon.firstId()) { //did we hove over the first polygon point?
                inst.canvas.hoverCursor = 'crosshair';
            } else if (o.target) {
                if (inst.type !== "Blocks") {
                    inst.over();
                    inst.canvas.hoverCursor = 'default';
                } else {
                    inst.canvas.hoverCursor = 'default';
                }
            } else {
                inst.canvas.hoverCursor = 'default';
            }

            if (inst._highlightOnMouseOver) {
                inst.canvas.removeItemsByAttr("id", "annotationsoverviewlabeltext");
                if (o.target) {
                    o.target.set("fill", o.target.get("stroke"));
                    if (o.target.id !== undefined) {
                        if (o.target.id in inst._annotationLabelOverviewMapping) {
                            var label = inst._annotationLabelOverviewMapping[o.target.id];
                            inst.canvas.add(new fabric.Text(label, {
                                id: "annotationsoverviewlabeltext",
                                fontSize: 20,
                                left: 2,
                                top: 2,
                                fill: o.target.get("stroke"),
                                fontFamily: "Arial"
                            }));
                        }
                    }
                    inst.canvas.renderAll();
                }
            }
        })
        inst.canvas.on('mouse:out', function(o) {
            if (o.target) {
                inst.out();
                if (inst._highlightOnMouseOver) {
                    inst.canvas.removeItemsByAttr("id", "annotationsoverviewlabeltext");

                    o.target.set("fill", "");
                    inst.canvas.renderAll();
                }
            }
        })
    }
    Annotator.prototype.onMouseUp = function(o) {
        var inst = this;

        if (this.type === "Blocks") {
            this.markBlocks();
            this.createHull();
        }


        /*else if(this.type === "FreeDrawing"){
          if(this.isDrawing && !this.freeDrawing.isClosedPathMode())
            this.freeDrawing.generatePolygon();
        }*/

        inst.disable();

        typeof this.mouseUpCB === 'function' && this.mouseUpCB();
    };

    Annotator.prototype.redo = function(o) {
        /*if (this.currentHistoryPosition > 0) {
            this.isRedoing = true;
            this.currentHistoryPosition -= 1;
            this.canvas.clear().renderAll();
            this.canvas.loadFromJSON(this.history[this.history.length - this.currentHistoryPosition + 1], function() {
              this.isRedoing = false;
            });
            this.canvas.renderAll();


        }*/
    };

    Annotator.prototype.undo = function(o) {
        /*if (this.currentHistoryPosition < this.history.length) {
            this.isRedoing = true;
            this.canvas.clear().renderAll();
            this.canvas.loadFromJSON(this.history[this.history.length - 1 - this.currentHistoryPosition], function() {
              this.isRedoing = false;
            });
            this.canvas.renderAll();
            this.currentHistoryPosition += 1;
        }*/
    };

    Annotator.prototype.initHistory = function(o) {
        this.saveState();
    };


    Annotator.prototype.saveState = function(o) {
        /*if(!this.isRedoing){
          j = JSON.stringify(this.canvas.toObject());
          this.history.push(j);
        }*/
    };

    Annotator.prototype.handleBlocks = function(origX, origY) {
        var beginX = this.cellSize * Math.floor((origX / this.cellSize), 0);
        var beginY = this.cellSize * Math.floor((origY / this.cellSize), 0);

        var key = beginX.toString() + beginY.toString();

        if (key in this.selectedBlocks) {
            var persistent = this.selectedBlocks[key];
            if (persistent) {
                this.canvas.getItemByAttr("id", ("block" + key)).remove();
                this.selectedBlocks[key] = false;
                this.recentlyDeletedBlocks[key] = key;
                //delete this.selectedBlocks[key];

                delete this.selectedBlocksPoints[key];
            }
        } else {
            var block = new fabric.Rect({
                left: beginX,
                top: beginY,
                originX: 'left',
                originY: 'top',
                width: this.cellSize,
                height: this.cellSize,
                angle: 0,
                opacity: 0.5,
                fill: "red",
                transparentCorners: false,
                hasBorders: false,
                hasControls: false,
                selectable: false,
                id: ("block" + key)
            });

            this.selectedBlocks[key] = false;
            this.canvas.add(block);

            this.selectedBlocksPoints[key] = [{
                    "x": beginX,
                    "y": beginY
                }, {
                    "x": (beginX + this.cellSize),
                    "y": beginY
                },
                {
                    "x": beginX,
                    "y": (beginY + this.cellSize)
                }, {
                    "x": (beginX + this.cellSize),
                    "y": (beginY + this.cellSize)
                }
            ];
            this.recentlyAddedBlocks[key] = key;
        }
    };

    Annotator.prototype.markBlocks = function() {
        for (var key in this.recentlyAddedBlocks) {
            if (this.recentlyAddedBlocks.hasOwnProperty(key)) {
                this.selectedBlocks[this.recentlyAddedBlocks[key]] = true;
            }
        }
        this.recentlyAddedBlocks = {};

        for (var key in this.recentlyDeletedBlocks) {
            if (this.recentlyDeletedBlocks.hasOwnProperty(key)) {
                delete this.selectedBlocks[this.recentlyDeletedBlocks[key]];
            }
        }
        this.recentlyDeletedBlocks = {};
    };

    Annotator.prototype.createHull = function() {
        var points = [];
        for (var key in this.selectedBlocksPoints) {
            if (this.selectedBlocksPoints.hasOwnProperty(key)) {
                var p = this.selectedBlocksPoints[key];
                for (var i = 0; i < p.length; i++) {
                    points.push(p[i]);
                }
            }
        }
        h = hull(points, 50, ['.x', '.y']);

        var existingHull = this.canvas.getItemByAttr("id", "hull");
        if (existingHull !== null)
            existingHull.remove();

        var polyline = new fabric.Polyline(h, {
            stroke: 'blue',
            strokeWidth: 5,
            fill: 'transparent',
            opacity: 0.5,
            selectable: false,
            hasBorders: false,
            hasControls: false,
            evented: false,
            "id": "hull"
        });
        this.canvas.add(polyline);
        this.canvas.renderAll();
    };





    Annotator.prototype.onMouseMove = function(o) {
        var inst = this;

        if (!inst.isPanMode) {
            if (!inst.isEnable()) {
                return;
            }
            var pointer = inst.canvas.getPointer(o.e);
            var activeObj = inst.canvas.getActiveObject();
            if ((inst.type === 'Rectangle') || (inst.type === 'Circle')) {
                if (activeObj) {
                    if (origX > pointer.x) {
                        activeObj.set({
                            left: Math.abs(pointer.x)
                        });
                    }
                    if (origY > pointer.y) {
                        activeObj.set({
                            top: Math.abs(pointer.y)
                        });
                    }
                }

                inst.canvas.renderAll();
            }

            if (inst.type === 'Rectangle') {
                if (activeObj) {
                    activeObj.set({
                        width: Math.abs(origX - pointer.x)
                    });
                    activeObj.set({
                        height: Math.abs(origY - pointer.y)
                    });

                    activeObj.setCoords();
                }
                inst.canvas.renderAll();
            }
            if (inst.type === 'Circle') {
                if (activeObj) {
                    activeObj.set({
                        rx: Math.abs(origX - pointer.x) / 2
                    });
                    activeObj.set({
                        ry: Math.abs(origY - pointer.y) / 2
                    });

                    activeObj.setCoords();
                }

                inst.canvas.renderAll();
            }
            if (inst.type === 'Polygon') {
                this.polygon.move(pointer);

                inst.canvas.renderAll();
            }
            if (inst.type === "Blocks") {
                this.handleBlocks(pointer.x, pointer.y);
            }
            if (inst.type === "FreeDrawing") {

                if (this.canvas.isDrawingMode && this.freeDrawing.isClosedPathMode()) {
                    if (this.freeDrawing.isPathClosed(pointer)) {
                        this.freeDrawing.generatePolygon();
                    } else {
                        this.freeDrawing.move(pointer);
                    }
                    inst.canvas.renderAll();
                }
            }
        } else {
            if (inst.panning && o && o.e) {
                var units = 10;
                var delta = new fabric.Point(o.e.movementX, o.e.movementY);
                inst.canvas.relativePan(delta);
            }
        }
    };

    Annotator.prototype.getAbsoluteCanvasPosition = function() {
        var p = {
            x: this.canvas.width / 2,
            y: this.canvas.height
        };
        var invertedMatrix = fabric.util.invertTransform(this.canvas.viewportTransform);
        var transformedP = fabric.util.transformPoint(p, invertedMatrix);
        transformedP.x = transformedP.x - this.canvas.width / 2;
        transformedP.y = transformedP.y - this.canvas.height;
        return transformedP;
    }

    Annotator.prototype.reset = function(clearCanvas = true) {
        if (clearCanvas)
            this.canvas.clear();
        this.canvas.setZoom(1.0);
        //this.canvas.viewport.position.x = 0;
        //this.canvas.viewport.position.y = 0;
        this.polygon.reset();
        this.canvas.absolutePan(new fabric.Point(0, 0));
        this._refAnnotations = [];
        this.refinementsPerAnnotation = {};
        this._annotationLabelOverviewMapping = {};
    };

    Annotator.prototype.deleteAll = function() {
        //remove all objects from canvas
        var objects = this.canvas.getObjects();
        while (objects.length != 0) {
            this.canvas.remove(objects[0]);
            this.canvas.discardActiveGroup();
        }
        this.polygon.reset();
        this.canvas.renderAll();
        this._refAnnotations = [];
        this.refinementsPerAnnotation = {};
        this._annotationLabelOverviewMapping = {};
    };

    Annotator.prototype.disableHighlightOnMouseOver = function() {
        this._highlightOnMouseOver = false;
    }

    Annotator.prototype.deleteSelected = function(o) {
        var activeObj = this.canvas.getActiveObject();
        if ("id" in activeObj) {
            this.polygon.removePolygonById(activeObj["id"]);
        }
        activeObj.remove();
    };

    Annotator.prototype.objectsSelected = function(o) {
        var obj = this.canvas.getActiveObject();
        if (!obj) return false;
        return true;
    };

    Annotator.prototype.onMouseDown = function(o) {
        var inst = this;
        if (!inst.isOver() && !inst.isBlocked() && !inst.isPanMode) {
            inst.enable();

            var pointer = inst.canvas.getPointer(o.e);
            origX = pointer.x;
            origY = pointer.y;

            if (inst.type === 'Rectangle') {
                var rect = new fabric.Rect({
                    left: origX,
                    top: origY,
                    originX: 'left',
                    originY: 'top',
                    width: pointer.x - origX,
                    height: pointer.y - origY,
                    angle: 0,
                    stroke: "#F00",
                    fill: "transparent",
                    transparentCorners: false,
                    hasBorders: false,
                    hasControls: false,
                    selectable: false,
                    selected: false,
                    evented: false,
                    id: generateRandomId(),
                    strokeWidth: inst.defaultStrokeWidth
                });

                inst.canvas.add(rect).setActiveObject(rect);
            }
            if (inst.type === 'Circle') {
                var circle = new fabric.Ellipse({
                    top: origY,
                    left: origX,
                    radius: 0,
                    rx: 0,
                    ry: 0,
                    fill: "transparent",
                    stroke: "#F00",
                    transparentCorners: false,
                    hasBorders: false,
                    hasControls: false,
                    selectable: false,
                    selected: false,
                    evented: false,
                    id: generateRandomId(),
                    strokeWidth: inst.defaultStrokeWidth
                });

                inst.canvas.add(circle).setActiveObject(circle);
            }

            if (inst.type === 'Polygon') {
                if (o.target && o.target.id === this.polygon.firstId()) {
                    this.polygon.generatePolygon();
                    this.polygon.clear();
                } else {
                    this.polygon.addPoint(o);
                }
            }

            if (inst.type === "Blocks") {
                this.selectedBlocksPoints = {}; //clear before we start a new drawing

                this.handleBlocks(origX, origY);
            }

            if (inst.type === 'FreeDrawing') {
                if (this.freeDrawing.isPathClosed(pointer)) {
                    this.freeDrawing.generatePolygon();
                    this.freeDrawing.clear();
                } else {
                    this.freeDrawing.addPoint(o);
                }
            }

        }
    };

    Annotator.prototype.enableSmartAnnotation = function() {
        this.freeDrawing.disableClosedPathMode();
        this.setBrushType("PencilBrush");
        this.smartAnnotation = true;
    }

    Annotator.prototype.disableSmartAnnotation = function() {
        this.freeDrawing.enableClosedPathMode();
        this.setBrushType("PencilBrush");
        this.smartAnnotation = false;
    }

    Annotator.prototype.isEnable = function() {
        return this.isDrawing;
    }

    Annotator.prototype.isBlocked = function() {
        return this.blocked;
    }

    Annotator.prototype.enable = function() {
        this.isDrawing = true;
    }

    Annotator.prototype.disable = function() {
        this.isDrawing = false;
    }

    Annotator.prototype.isOver = function() {
        return this.overObject;
    }

    Annotator.prototype.over = function() {
        this.overObject = true;
    }

    Annotator.prototype.out = function() {
        this.overObject = false;
    }

    Annotator.prototype.block = function() {
        this.blocked = true;
    }

    Annotator.prototype.unblock = function() {
        this.blocked = false;
    }

    Annotator.prototype.setShape = function(t) {
        this.type = t;

        if (this.type === "FreeDrawing")
            this.canvas.isDrawingMode = true;
        else
            this.canvas.isDrawingMode = false;
    }

    Annotator.prototype.getShape = function(t) {
        return this.type;
    }

    Annotator.prototype.setBrushColor = function(brushColor) {
        this.brushColor = brushColor;
        this.canvas.freeDrawingBrush.color = this.brushColor;
    }

    Annotator.prototype.setBrushWidth = function(brushWidth) {
        this.brushWidth = brushWidth;
        this.canvas.freeDrawingBrush.width = this.brushWidth;
    }

    Annotator.prototype.setBrushType = function(brushType) {
        this.brushType = brushType;
        this.canvas.freeDrawingBrush = new fabric[this.brushType](this.canvas);
    }

    Annotator.prototype.enablePanMode = function() {
        this.isPanMode = true;
        this.canvas.selection = false; //disable group selection in pan mode
        this.canvas.forEachObject(function(o) { //disable object selection in pan mode
            o.selectable = false;
        });
    }

    Annotator.prototype.enableSelectMoveMode = function() {
        this.isSelectMoveMode = true;
    }

    Annotator.prototype.isSelectMoveModeEnabled = function() {
        return this.isSelectMoveMode;
    }

    Annotator.prototype.disableSelectMoveMode = function() {
        this._silenceAllObjects();
        this.objSelected();
        this.isSelectMoveMode = false;
    }

    Annotator.prototype.disablePanMode = function() {
        this.isPanMode = false;
        this.canvas.selection = true; //enable group selection again when pan mode ends
        this.canvas.forEachObject(function(o) { //enable object selection again when pan mode ends
            o.selectable = true;
        });
    }

    Annotator.prototype.isPanModeEnabled = function() {
        return this.isPanMode;
    }

    Annotator.prototype.getIdOfSelectedItem = function() {
        var activeObj = this.canvas.getActiveObject();
        if (activeObj !== undefined && activeObj !== null) {
            return activeObj.get("id");
        }
        return "";
    }

    Annotator.prototype.setRefinements = function(refinements) {
        var id = this.getIdOfSelectedItem();
        if (id !== "") {
            this.refinementsPerAnnotation[id] = refinements;
        }
    }

    Annotator.prototype.getRefinements = function() {
        var refs = [];
        for (var key in this.refinementsPerAnnotation) {
            if (this.refinementsPerAnnotation.hasOwnProperty(key)) {
                refs.push(this.refinementsPerAnnotation[key]);
            }
        }
        return refs;
    }

    Annotator.prototype.getRefinementsOfSelectedItem = function() {
        var id = this.getIdOfSelectedItem();
        if (id !== "") {
            if (id in this.refinementsPerAnnotation) {
                return this.refinementsPerAnnotation[id];
            }
        }
        return [];
    }

    Annotator.prototype.showGrid = function() {
        this.gridVisible = true;

        this.canvas.selection = false; //disable group selection when grid is shown
        this.selectedBlocks = {}; //clear selected blocks array
        this.recentlyDeletedBlocks = {};
        this.recentlyAddedBlocks = {};
        this.selectedBlocksPoints = {};

        if (this.canvas.height > this.canvas.width)
            this.cellSize = this.canvas.height / this.gridSize;
        else
            this.cellSize = this.canvas.width / this.gridSize;

        for (var x = 1; x < (this.canvas.width / this.gridSize); x++) {
            this.canvas.add(new fabric.Line([this.cellSize * x, 0, this.cellSize * x, this.canvas.height], {
                stroke: "#000000",
                strokeWidth: 1,
                selectable: false,
                strokeDashArray: [5, 5],
                id: "grid"
            }));
            this.canvas.add(new fabric.Line([0, this.cellSize * x, this.canvas.width, this.cellSize * x], {
                stroke: "#000000",
                strokeWidth: 1,
                selectable: false,
                strokeDashArray: [5, 5],
                id: "grid"
            }));
        }
        this.canvas.renderAll();
    }

    Annotator.prototype.hideGrid = function() {
        this.gridVisible = false;
        this.canvas.removeItemsByAttr("id", "grid");
        this.canvas.selection = true; //enable group selection when grid is hidden
    }

    Annotator.prototype.setStrokeWidthOfSelected = function(strokeWidth) {
        var activeObj = this.canvas.getActiveObject();
        if (activeObj !== undefined && activeObj !== null) {
            activeObj.set({
                strokeWidth: strokeWidth
            });
            this.canvas.renderAll();
        }
    }

    Annotator.prototype.setStrokeColorOfSelected = function(strokeColor) {
        var activeObj = this.canvas.getActiveObject();
        if (activeObj !== undefined && activeObj !== null) {
            activeObj.set({
                stroke: strokeColor
            });
            this.canvas.renderAll();
        }
    }

    Annotator.prototype.getStrokeColorOfSelected = function() {
        var activeObj = this.canvas.getActiveObject();
        if (activeObj !== undefined && activeObj !== null) {
            return activeObj.get("stroke");
        }
        return null;
    }

    Annotator.prototype.toggleGrid = function() {
        if (this.gridVisible)
            this.hideGrid();
        else
            this.showGrid();
    }

    Annotator.prototype.setSmartAnnotationData = function(smartAnnotationData) {
        this.smartAnnotationData = smartAnnotationData;
    }

    Annotator.prototype.toJSON = function(includeRefinementUuid = null) {
        var data = this.canvas.toJSON(["id"]); //include custom property "id" when converting to json
        var imgScaleX = data["backgroundImage"]["scaleX"];
        var imgScaleY = data["backgroundImage"]["scaleY"];
        var objs = data["objects"];
        var res = [];

        if (this.smartAnnotation) {
            if (this.smartAnnotationData.length > 0) {
                res = this.smartAnnotationData;
            }

        } else {
            var left, top, width, height, rx, ry, type, points, pointX, pointY, angle, color,
                pointX, pointY, pX, pY, strokeWidth, strokeColor, annotationId, refinements;

            for (var i = 0; i < objs.length; i++) {
                //skip polygon handles
                if (("isPolygonHandle" in objs[i]) && objs[i]["isPolygonHandle"])
                    continue;


                angle = objs[i]["angle"];
                type = objs[i]["type"];
                annotationId = objs[i]["id"];
                refinements = [];
                if (annotationId in this.refinementsPerAnnotation) {
                    ref = this.refinementsPerAnnotation[annotationId];
                    for (var j = 0; j < ref.length; j++) {
                        refinements.push({
                            "label_uuid": ref[j]
                        });
                    }
                }

                if (includeRefinementUuid != null) {
                    refinements.push({
                        "label_uuid": includeRefinementUuid
                    });
                }


                if (type === "rect") {
                    left = Math.round(((objs[i]["left"] / imgScaleX)), 0);
                    top = Math.round(((objs[i]["top"] / imgScaleY)), 0);
                    width = Math.round(((objs[i]["width"] / imgScaleX) * objs[i]["scaleX"]), 0);
                    height = Math.round(((objs[i]["height"] / imgScaleY) * objs[i]["scaleY"]), 0);
                    strokeWidth = Math.round((objs[i]["strokeWidth"] * ((objs[i]["scaleX"] + objs[i]["scaleY"]) / 2)), 0);
                    strokeColor = objs[i]["stroke"];

                    if ((width != 0) && (height != 0))
                        res.push({
                            "refinements": refinements,
                            "left": left,
                            "top": top,
                            "width": width,
                            "height": height,
                            "angle": angle,
                            "type": "rect",
                            "stroke": {
                                "width": strokeWidth,
                                "color": strokeColor
                            }
                        });

                } else if (type === "ellipse") {
                    left = Math.round(((objs[i]["left"] / imgScaleX)), 0);
                    top = Math.round(((objs[i]["top"] / imgScaleY)), 0);
                    rx = Math.round(((objs[i]["rx"] / imgScaleX) * objs[i]["scaleX"]), 0);
                    ry = Math.round(((objs[i]["ry"] / imgScaleY) * objs[i]["scaleY"]), 0);
                    strokeWidth = Math.round((objs[i]["strokeWidth"] * ((objs[i]["scaleX"] + objs[i]["scaleY"]) / 2)), 0);
                    strokeColor = objs[i]["stroke"];

                    if ((rx != 0) && (ry != 0))
                        res.push({
                            "refinements": refinements,
                            "left": left,
                            "top": top,
                            "rx": rx,
                            "ry": ry,
                            "angle": angle,
                            "type": "ellipse",
                            "stroke": {
                                "width": strokeWidth,
                                "color": strokeColor
                            }
                        });
                } else if (type === "polygon") {
                    left = Math.round(((objs[i]["left"] / imgScaleX)), 0);
                    top = Math.round(((objs[i]["top"] / imgScaleY)), 0);
                    width = Math.round(((objs[i]["width"] / imgScaleX)), 0);
                    height = Math.round(((objs[i]["height"] / imgScaleY)), 0);
                    strokeWidth = Math.round((objs[i]["strokeWidth"] * ((objs[i]["scaleX"] + objs[i]["scaleY"]) / 2)), 0);
                    strokeColor = objs[i]["stroke"];


                    points = objs[i]["points"];

                    var scaledPoints = [];
                    for (var j = 0; j < points.length; j++) {
                        pointX = Math.round(((points[j]["x"] / imgScaleX) * objs[i]["scaleX"]), 0);
                        pointY = Math.round(((points[j]["y"] / imgScaleY) * objs[i]["scaleY"]), 0);
                        pX = pointX;
                        pY = pointY;
                        scaledPoints.push({
                            "x": pX,
                            "y": pY
                        });
                    }

                    res.push({
                        "refinements": refinements,
                        "points": scaledPoints,
                        "angle": angle,
                        "type": "polygon",
                        "stroke": {
                            "width": strokeWidth,
                            "color": strokeColor
                        }
                    });
                }
            }
        }
        return res;
    }

    Annotator.prototype._handleLoadedAutoAnnotation = function(obj) {
        if (obj.type === "polygon") {
            obj.id = generateRandomId();
            obj.objectCaching = false;
            this.polygon.addPolygon(obj);
            this.polygon.showPolyPoints(obj.id);
        }
    }

    Annotator.prototype._handleLoadedAnnotation = function(annotation, obj) {
        obj.id = generateRandomId();
        if (obj.type === "polygon") {
            obj.objectCaching = false;
            this.polygon.addPolygon(obj);
            this.polygon.showPolyPoints(obj.id);
        }

        if ("refinements" in annotation) {
            var refinements = [];
            var ref = annotation["refinements"];
            for (var j = 0; j < ref.length; j++) {
                refinements.push(ref[j]["label_uuid"]);
            }
            this.refinementsPerAnnotation[obj.id] = refinements;
        }
    }

    Annotator.prototype._handleLoadedAnnotationOverview = function(label, obj) {
        obj.id = generateRandomId();
        obj.objectCaching = false;
        this.canvas.renderAll();
        this._annotationLabelOverviewMapping[obj.id] = label;
    }

    Annotator.prototype._simplifyAutoAnnotations = function(autoAnnotations) {
        for (var i = 0; i < autoAnnotations.length; i++) {
            autoAnnotations[i].points = simplify(autoAnnotations[i].points, 4.0, false);
        }
        return autoAnnotations;
    }

    Annotator.prototype.loadAnnotations = function(annotations, scaleFactor = 1.0) {
        this.deleteAll();
        for (var i = 0; i < annotations.length; i++) {
            drawAnnotations(this.canvas, [annotations[i]], scaleFactor, this._handleLoadedAnnotation.bind(this, annotations[i]));
        }

        this._refAnnotations = this.toJSON();
    }

    Annotator.prototype.loadAnnotationsOverview = function(annotationsWithLabel, scaleFactor = 1.0) {
        this.deleteAll();

        this._highlightOnMouseOver = true;
        for (var i = 0; i < annotationsWithLabel.length; i++) {
            for (var j = 0; j < annotationsWithLabel[i].annotations.length; j++) {
                drawAnnotations(this.canvas, [annotationsWithLabel[i].annotations[j]], scaleFactor,
                    this._handleLoadedAnnotationOverview.bind(this, annotationsWithLabel[i].label));
            }
        }
    }

    Annotator.prototype.loadAutoAnnotations = function(autoAnnotations, scaleFactor = 1.0) {
        var simplifiedAutoAnnotations = this._simplifyAutoAnnotations(autoAnnotations);
        drawAnnotations(this.canvas, simplifiedAutoAnnotations, scaleFactor, this._handleLoadedAutoAnnotation.bind(this));
    }

    Annotator.prototype.isDirty = function() {
        if (_.isEqual(this._refAnnotations, this.toJSON()))
            return false;
        return true;
    }


    Annotator.prototype.getMask = function() {
        var img = this.canvas.backgroundImage;
        this.canvas.backgroundImage = null;

        var oldBgColor = this.canvas.backgroundColor;
        this.canvas.backgroundColor = "black";

        //remember current canvas pos
        var oldPos = this.getAbsoluteCanvasPosition();

        //remember current canvas zoom
        var oldZoom = this.canvas.getZoom();

        //set canvas pos to (0,0)
        this.canvas.absolutePan(new fabric.Point(0, 0));

        //set canvas zoom to 1.0
        this.canvas.setZoom(1.0);

        var objects = this.canvas.getObjects();
        var old = [];
        var strokeColor;
        for (var i = 0; i < objects.length; i++) {
            strokeColor = objects[i].get("stroke");
            old.push([objects[i].get("fill"), strokeColor]);
            if (strokeColor === "white") {
                objects[i].set("fill", "white");
            } else if (strokeColor !== "black") {
                objects[i].set("fill", "grey");
                objects[i].set("stroke", "grey");
            }
        }

        var res = this.canvas.toDataURL({
            format: 'png'
        });

        for (var i = 0; i < objects.length; i++) {
            objects[i].set("fill", old[i][0]);
            objects[i].set("stroke", old[i][1]);
        }

        //restore old canvas pos
        this.canvas.absolutePan(oldPos);

        //restore old zoom
        this.canvas.setZoom(oldZoom);

        this.canvas.backgroundColor = oldBgColor;
        this.canvas.backgroundImage = img;
        this.canvas.renderAll();
        return res;
    }

    return Annotator;
}());



var AnnotationSettings = (function() {
    function AnnotationSettings() {
        var inst = this;
    }

    AnnotationSettings.prototype.getPreferedAnnotationTool = function() {
        var radioButtonId = $("#preferedAnnotationToolCheckboxes :radio:checked").attr('id');
        if (radioButtonId === "preferedRectangleAnnotationToolCheckboxInput") {
            return "Rectangle";
        }
        if (radioButtonId === "preferedCircleAnnotationToolCheckboxInput") {
            return "Circle";
        }
        if (radioButtonId === "preferedPolygonAnnotationToolCheckboxInput") {
            return "Polygon";
        }
        return "Rectangle";
    }

    AnnotationSettings.prototype.getWorkspaceSize = function() {
        var radioButtonId = $("#annotationWorkspaceSizeCheckboxes :radio:checked").attr('id');
        if (radioButtonId === "annotationWorkspaceSizeSmallCheckboxInput") {
            return "small";
        }
        if (radioButtonId === "annotationWorkspaceSizeMediumCheckboxInput") {
            return "medium";
        }
        if (radioButtonId === "annotationWorkspaceSizeBigCheckboxInput") {
            return "big";
        }
        return "small";
    }

    AnnotationSettings.prototype.getAnnotationMode = function() {
        var radioButtonId = $("#annotationModeCheckboxes :radio:checked").attr('id');
        if (radioButtonId === "annotationDefaultModeCheckboxInput") {
            return "default";
        }
        if (radioButtonId === "annotationBrowseModeCheckboxInput") {
            return "browse";
        }
        return "default";
    }

    AnnotationSettings.prototype.getPolygonVertexSize = function() {
        var polygonVertexSize = $("#annotationPolygonVertexSizeInput").val();
        return polygonVertexSize;
    }

    AnnotationSettings.prototype.persistAll = function() {
        var settings = new Settings();

        var preferedAnnotationTool = this.getPreferedAnnotationTool();
        localStorage.setItem('preferedAnnotationTool', preferedAnnotationTool); //store in local storage
        var workspaceSize = this.getWorkspaceSize();
        localStorage.setItem('annotationWorkspaceSize', workspaceSize);
        var annotationMode = this.getAnnotationMode();
        settings.setAnnotationMode(annotationMode);
        var polygonVertexSize = this.getPolygonVertexSize();
        settings.setPolygonVertexSize(polygonVertexSize);
    }

    AnnotationSettings.prototype.setAll = function() {
        this.setPreferedAnnotationTool();
        this.setWorkspaceSize();
        this.setAnnotationMode();
        this.setPolygonVertexSize();
    }

    AnnotationSettings.prototype.setWorkspaceSize = function() {
        var workspaceSize = localStorage.getItem("annotationWorkspaceSize");
        if (workspaceSize === "small") {
            $("#annotationWorkspaceSizeMediumCheckbox").checkbox("set unchecked");
            $("#annotationWorkspaceSizeBigCheckbox").checkbox("set unchecked");
            $("#annotationWorkspaceSizeSmallCheckbox").checkbox("set checked");
        } else if (workspaceSize === "medium") {
            $("#annotationWorkspaceSizeSmallCheckbox").checkbox("set unchecked");
            $("#annotationWorkspaceSizeBigCheckbox").checkbox("set unchecked");
            $("#annotationWorkspaceSizeMediumCheckbox").checkbox("check");
        } else if (workspaceSize === "big") {
            $("#annotationWorkspaceSizeSmallCheckbox").checkbox("set unchecked");
            $("#annotationWorkspaceSizeMediumCheckbox").checkbox("set unchecked");
            $("#annotationWorkspaceSizeBigCheckbox").checkbox("set checked");
        }
    }

    AnnotationSettings.prototype.setPreferedAnnotationTool = function() {
        var preferedAnnotationTool = localStorage.getItem("preferedAnnotationTool");
        if (preferedAnnotationTool === "Rectangle") {
            $("#preferedCircleAnnotationToolCheckbox").checkbox("set unchecked");
            $("#preferedPolygonAnnotationToolCheckbox").checkbox("set unchecked");
            $("#preferedRectangleAnnotationToolCheckbox").checkbox("set checked");
        } else if (preferedAnnotationTool === "Circle") {
            $("#preferedPolygonAnnotationToolCheckbox").checkbox("set unchecked");
            $("#preferedRectangleAnnotationToolCheckbox").checkbox("set unchecked");
            $("#preferedCircleAnnotationToolCheckbox").checkbox("check");
        } else if (preferedAnnotationTool === "Polygon") {
            $("#preferedPolygonAnnotationToolCheckbox").checkbox("set unchecked");
            $("#preferedCircleAnnotationToolCheckbox").checkbox("set unchecked");
            $("#preferedPolygonAnnotationToolCheckbox").checkbox("set checked");
        }
    }

    AnnotationSettings.prototype.setAnnotationMode = function() {
        var settings = new Settings();
        annotationMode = settings.getAnnotationMode();
        if (annotationMode === "default") {
            $("#annotationBrowseModeCheckbox").checkbox("set unchecked");
            $("#annotationDefaultModeCheckbox").checkbox("set checked");
        } else if (annotationMode === "browse") {
            $("#annotationDefaultModeCheckbox").checkbox("set unchecked");
            $("#annotationBrowseModeCheckbox").checkbox("check");
        }
    }

    AnnotationSettings.prototype.setPolygonVertexSize = function() {
        var settings = new Settings();
        $("#annotationPolygonVertexSizeInput").val(settings.getPolygonVertexSize());
    }

    AnnotationSettings.prototype.loadPreferedAnnotationTool = function(annotationView, annotator) {
        var preferedAnnotationTool = localStorage.getItem("preferedAnnotationTool");
        if ((preferedAnnotationTool === "Rectangle") || (preferedAnnotationTool === "Circle") || (preferedAnnotationTool === "Polygon")) {
            annotationView.changeMenuItem(preferedAnnotationTool);
            annotator.setShape(preferedAnnotationTool);
        }
    }

    AnnotationSettings.prototype.loadWorkspaceSize = function() {
        var workspaceSize = localStorage.getItem("annotationWorkspaceSize");
        if ((workspaceSize === "small") || (workspaceSize === "medium") || (workspaceSize === "big")) {
            return workspaceSize;
        }
        return "small";
    }

    return AnnotationSettings;
}());