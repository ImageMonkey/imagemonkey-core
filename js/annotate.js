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

var Polygon = (function () {
  function Polygon(canvas) {
    var inst=this;
    this.canvas = canvas;
    this.polygonMode = true;
    this.pointArray = new Array();
    this.lineArray = new Array();
    this.activeLine = null;
    this.max = 999999;
    this.min = 99;
    this.activeShape = false;
  }

  Polygon.prototype.clear = function () {
    this.polygonMode = true;
    this.pointArray.length = 0;
    this.lineArray.length = 0;
    this.activeLine = null;
    this.activeShape = false;
  };

  Polygon.prototype.addPoint = function (options) {
    var random = Math.floor(Math.random() * (this.max - this.min + 1)) + this.min;
    var id = new Date().getTime() + random;
    var pointer = this.canvas.getPointer(options.e);
    var circle = new fabric.Circle({
      radius: 5,
      fill: '#ffffff',
      stroke: '#333333',
      strokeWidth: 0.5,
      left: pointer.x, //(options.e.layerX/this.canvas.getZoom()),
      top: pointer.y, //(options.e.layerY/this.canvas.getZoom()),
      selectable: false,
      hasBorders: false,
      hasControls: false,
      originX:'center',
      originY:'center',
      id:id
    });
    if(this.pointArray.length == 0){
      circle.set({
        fill:'red'
      })
    }
    //var points = [(options.e.layerX/this.canvas.getZoom()),(options.e.layerY/this.canvas.getZoom()),(options.e.layerX/this.canvas.getZoom()),(options.e.layerY/this.canvas.getZoom())];
    var points = [pointer.x, pointer.y, pointer.x, pointer.y];
    line = new fabric.Line(points, {
      strokeWidth: 2,
      fill: '#999999',
      stroke: '#999999',
      class:'line',
      originX:'center',
      originY:'center',
      selectable: false,
      hasBorders: false,
      hasControls: false,
      evented: false
    });
    if(this.activeShape){
      var pos = this.canvas.getPointer(options.e);
      var points = this.activeShape.get("points");
      points.push({
        x: pos.x,
        y: pos.y
      });
      var polygon = new fabric.Polygon(points,{
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
    }
    else{
      //var polyPoint = [{x:(options.e.layerX/this.canvas.getZoom()),y:(options.e.layerY/this.canvas.getZoom())}];
      var polyPoint = [{x: pointer.x, y: pointer.y}];
      var polygon = new fabric.Polygon(polyPoint,{
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

  Polygon.prototype.generatePolygon = function () {
    var points = new Array();
    $.each(this.pointArray,function(index,point){
      points.push({
        x:point.left,
        y:point.top
      });
      this.canvas.remove(point);
    });
    $.each(this.lineArray,function(index,line){
      this.canvas.remove(line);
    });
    this.canvas.remove(this.activeShape).remove(this.activeLine);
    var polygon = new fabric.Polygon(points,{
      stroke: 'red',
      strokeWidth: 5,
      fill: 'transparent',
      hasBorders: true,
      hasControls: true
    });
    this.canvas.add(polygon);

    this.activeLine = null;
    this.activeShape = null;
    this.polygonMode = false;
    this.canvas.selection = true;
  };

  Polygon.prototype.firstId = function () {
    if(this.pointArray.length === 0)
      return -1;
    return this.pointArray[0].id;
  };

  Polygon.prototype.move = function(pointer) {
    if(this.activeLine && this.activeLine.class == "line"){
      this.activeLine.set({ x2: pointer.x, y2: pointer.y });

      var points = this.activeShape.get("points");
      points[this.pointArray.length] = {
        x:pointer.x,
        y:pointer.y
      }
        
      this.activeShape.set({
        points: points
      });
      this.canvas.renderAll();
    }
  };


  return Polygon;
}());







var Annotator = (function () {
  function Annotator(canvas, objSelected) {
    var inst=this;
    this.canvas = canvas;
    this.className= "Rectangle";
    this.isDrawing = false;
    this.overObject = false;
    this.blocked = false; 
    this.type = "Rectangle";
    this.polygon = new Polygon(this.canvas);
    this.objSelected = objSelected;
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

    this.bindEvents();
  }

  Annotator.prototype.bindEvents = function() {
    var inst = this;
    inst.canvas.on('mouse:down', function(o) {
      inst.onMouseDown(o);
      if(inst.isPanMode)
        inst.panning = true;
    });
    inst.canvas.on('mouse:move', function(o) {
      inst.onMouseMove(o);
    });
    inst.canvas.on('mouse:up', function(o) {
      inst.onMouseUp(o);
      if(inst.isPanMode)
        inst.panning = false;
    });
    inst.canvas.on('object:moving', function(o) {
      inst.disable();
    });
    inst.canvas.on('object:selected', function(o) {
      inst.objSelected();
    });
    inst.canvas.on('object:modified', function(o) {
      
    });
    inst.canvas.on('object:added', function(o) {
      inst.saveState();
    });

    inst.canvas.on('mouse:over', function(o) {
      if(o.target && o.target.id == inst.polygon.firstId()){ //did we hove over the first polygon point?
        inst.canvas.hoverCursor = 'crosshair';
      }
      else if(o.target){
        if(inst.type !== "Blocks"){
          inst.over();
          inst.canvas.hoverCursor = 'move';
        }
      }
    })
    inst.canvas.on('mouse:out', function(o) {
      if(o.target)
        inst.out();
    })
  }
  Annotator.prototype.onMouseUp = function (o) {
    var inst = this;
    inst.disable();

    if(this.type === "Blocks"){
      this.markBlocks();
      this.createHull();
    }
  };

  Annotator.prototype.redo = function (o) {
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

  Annotator.prototype.undo = function (o) {
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

  Annotator.prototype.initHistory = function (o) {
    this.saveState();
  };


  Annotator.prototype.saveState = function (o) {
    /*if(!this.isRedoing){
      j = JSON.stringify(this.canvas.toObject());
      this.history.push(j);
    }*/
  };

  Annotator.prototype.handleBlocks = function (origX, origY) {
    var beginX = this.cellSize * Math.floor((origX/this.cellSize), 0);
    var beginY = this.cellSize * Math.floor((origY/this.cellSize), 0);

    var key = beginX.toString() + beginY.toString();

    if(key in this.selectedBlocks){
      var persistent = this.selectedBlocks[key];
      if(persistent){
        this.canvas.getItemByAttr("id", ("block" + key)).remove();
        this.selectedBlocks[key] = false;
        this.recentlyDeletedBlocks[key] = key;
        //delete this.selectedBlocks[key];

        delete this.selectedBlocksPoints[key];
      }
    }
    else{
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
      
      this.selectedBlocksPoints[key] = [{"x": beginX, "y": beginY}, {"x": (beginX + this.cellSize), "y": beginY}, 
                                        {"x": beginX, "y": (beginY + this.cellSize)}, {"x": (beginX + this.cellSize), "y": (beginY + this.cellSize)}];
      this.recentlyAddedBlocks[key] = key;
    }
  };

  Annotator.prototype.markBlocks = function () {
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

  Annotator.prototype.createHull = function () {
    var points = [];
    for (var key in this.selectedBlocksPoints) {
      if (this.selectedBlocksPoints.hasOwnProperty(key)) {
        var p = this.selectedBlocksPoints[key];
        for(var i = 0; i < p.length; i++){
          points.push(p[i]);
        }
      }
    }
    h = hull(points, 50, ['.x', '.y']);

    var existingHull = this.canvas.getItemByAttr("id", "hull");
    if(existingHull !== null)
      existingHull.remove();

    var polyline = new fabric.Polyline(h,{
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


  


  Annotator.prototype.onMouseMove = function (o) {
    var inst = this;

    if(!inst.isPanMode){
      if(!inst.isEnable()){ return; }
      var pointer = inst.canvas.getPointer(o.e);

      if((inst.type === 'Rectangle') || (inst.type === 'Circle')){
        var activeObj = inst.canvas.getActiveObject();
        activeObj.stroke= 'red',
        activeObj.strokeWidth= 5;
        activeObj.fill = 'transparent';

        if(origX > pointer.x){
          activeObj.set({ left: Math.abs(pointer.x) }); 
        }
        if(origY > pointer.y){
          activeObj.set({ top: Math.abs(pointer.y) });
        }

        inst.canvas.renderAll();
      }

      if(inst.type === 'Rectangle'){
        activeObj.set({ width: Math.abs(origX - pointer.x) });
        activeObj.set({ height: Math.abs(origY - pointer.y) });

        activeObj.setCoords();

        inst.canvas.renderAll();
      }
      if(inst.type === 'Circle'){   
        activeObj.set({ rx: Math.abs(origX - pointer.x) / 2 });
        activeObj.set({ ry: Math.abs(origY - pointer.y) / 2 });

        activeObj.setCoords();

        inst.canvas.renderAll();
      }
      if(inst.type === 'Polygon'){
        this.polygon.move(pointer);
        
        inst.canvas.renderAll();
      }
      if(inst.type === "Blocks"){
        this.handleBlocks(pointer.x, pointer.y);
      }
    }
    else{
      if(inst.panning && o && o.e){
        var units = 10;
        var delta = new fabric.Point(o.e.movementX, o.e.movementY);
        inst.canvas.relativePan(delta);
      }
    }
  };

  Annotator.prototype.deleteSelected = function (o) {
    this.canvas.getActiveObject().remove();
  };

  Annotator.prototype.objectsSelected = function (o) {
    var obj = this.canvas.getActiveObject();
    if(!obj) return false;
    return true;
  };

  Annotator.prototype.onMouseDown = function (o) {
    var inst = this;
    if(!inst.isOver() && !inst.isBlocked() && !inst.isPanMode){
      inst.enable();

      var pointer = inst.canvas.getPointer(o.e);
      origX = pointer.x;
      origY = pointer.y;

      if(inst.type === 'Rectangle'){
        var rect = new fabric.Rect({
          left: origX,
          top: origY,
          originX: 'left',
          originY: 'top',
          width: pointer.x-origX,
          height: pointer.y-origY,
          angle: 0,
          transparentCorners: false,
          hasBorders: true,
          hasControls: true
        });

        inst.canvas.add(rect).setActiveObject(rect);
      }
      if(inst.type === 'Circle'){
        var circle = new fabric.Ellipse({
          top: origY,
          left: origX,
          radius: 0,
          rx: 0,
          ry: 0,
          transparentCorners: false,
          hasBorders: true,
          hasControls: true
        });

        inst.canvas.add(circle).setActiveObject(circle);
      }

      if(inst.type === 'Polygon'){
        if(o.target && o.target.id == this.polygon.firstId()){
          this.polygon.generatePolygon();
          this.polygon.clear();
        }
        else{
          this.polygon.addPoint(o);
        }
      }

      if(inst.type === "Blocks"){
        this.selectedBlocksPoints = {}; //clear before we start a new drawing

        this.handleBlocks(origX, origY);
      }

    }
  };

  Annotator.prototype.isEnable = function(){
    return this.isDrawing;
  }

  Annotator.prototype.isBlocked = function(){
    return this.blocked;
  }

  Annotator.prototype.enable = function(){
    this.isDrawing = true;
  }

  Annotator.prototype.disable = function(){
    this.isDrawing = false;
  }

  Annotator.prototype.isOver = function(){
    return this.overObject;
  }

  Annotator.prototype.over = function(){
    this.overObject = true;
  }

  Annotator.prototype.out = function(){
    this.overObject = false;
  }

  Annotator.prototype.block = function(){
    this.blocked = true;
  }

  Annotator.prototype.unblock = function(){
    this.blocked = false;
  }

  Annotator.prototype.setShape = function(t){
    this.type = t;
  }

  Annotator.prototype.enablePanMode = function(){
    this.isPanMode = true;
    this.canvas.selection = false; //disable group selection in pan mode
    this.canvas.forEachObject(function(o) { //disable object selection in pan mode
      o.selectable = false;
    });
  }

  Annotator.prototype.disablePanMode = function(){
    this.isPanMode = false;
    this.canvas.selection = true; //enable group selection again when pan mode ends
    this.canvas.forEachObject(function(o) { //enable object selection again when pan mode ends
      o.selectable = true;
    });
  }

  Annotator.prototype.showGrid = function(){
    this.gridVisible = true;

    this.canvas.selection = false; //disable group selection when grid is shown
    this.selectedBlocks = {}; //clear selected blocks array
    this.recentlyDeletedBlocks = {};
    this.recentlyAddedBlocks = {};
    this.selectedBlocksPoints = {};

    if(this.canvas.height > this.canvas.width)
      this.cellSize = this.canvas.height/this.gridSize;
    else
      this.cellSize = this.canvas.width/this.gridSize;

    for(var x = 1; x < (this.canvas.width/this.gridSize); x++){
      this.canvas.add(new fabric.Line([this.cellSize * x, 0, this.cellSize * x, this.canvas.height],{ stroke: "#000000", strokeWidth: 1, selectable:false, strokeDashArray: [5, 5], id: "grid"}));
      this.canvas.add(new fabric.Line([0, this.cellSize * x, this.canvas.width, this.cellSize * x],{ stroke: "#000000", strokeWidth: 1, selectable:false, strokeDashArray: [5, 5], id: "grid"}));
    }
    this.canvas.renderAll();
  }

  Annotator.prototype.hideGrid = function(){
    this.gridVisible = false;
    this.canvas.removeItemsByAttr("id", "grid");
    this.canvas.selection = true; //enable group selection when grid is hidden
  }

  Annotator.prototype.toggleGrid = function(){
    if(this.gridVisible)
      this.hideGrid();
    else
      this.showGrid();
  }




  return Annotator;
}());