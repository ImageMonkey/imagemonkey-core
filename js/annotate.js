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







var Shape = (function () {
  function Shape(canvas, objSelected) {
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

    this.bindEvents();
  }

  Shape.prototype.bindEvents = function() {
    var inst = this;
    inst.canvas.on('mouse:down', function(o) {
      inst.onMouseDown(o);
    });
    inst.canvas.on('mouse:move', function(o) {
      inst.onMouseMove(o);
    });
    inst.canvas.on('mouse:up', function(o) {
      inst.onMouseUp(o);
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
        inst.over();
        inst.canvas.hoverCursor = 'move';
      }
    })
    inst.canvas.on('mouse:out', function(o) {
      if(o.target)
        inst.out();
    })
  }
  Shape.prototype.onMouseUp = function (o) {
    var inst = this;
    inst.disable();
  };

  Shape.prototype.redo = function (o) {
    if (this.currentHistoryPosition > 0) {
        this.isRedoing = true;
        this.currentHistoryPosition -= 1;
        this.canvas.clear().renderAll();
        this.canvas.loadFromJSON(this.history[this.history.length - this.currentHistoryPosition + 1], function() {
          this.isRedoing = false;
        });
        this.canvas.renderAll();
        
        
    }
  };

  Shape.prototype.undo = function (o) {
    if (this.currentHistoryPosition < this.history.length) {
        this.isRedoing = true;
        this.canvas.clear().renderAll();
        this.canvas.loadFromJSON(this.history[this.history.length - 1 - this.currentHistoryPosition], function() {
          this.isRedoing = false;
        });
        this.canvas.renderAll();
        this.currentHistoryPosition += 1;
    }
  };

  Shape.prototype.initHistory = function (o) {
    this.saveState();
  };


  Shape.prototype.saveState = function (o) {
    if(!this.isRedoing){
      j = JSON.stringify(this.canvas.toObject());
      this.history.push(j);
    }
  };
  


  Shape.prototype.onMouseMove = function (o) {
    var inst = this;


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
    }

    if(inst.type === 'Rectangle'){
      activeObj.set({ width: Math.abs(origX - pointer.x) });
      activeObj.set({ height: Math.abs(origY - pointer.y) });

      activeObj.setCoords();
    }
    if(inst.type === 'Circle'){   
      activeObj.set({ rx: Math.abs(origX - pointer.x) / 2 });
      activeObj.set({ ry: Math.abs(origY - pointer.y) / 2 });

      activeObj.setCoords();
    }
    if(inst.type === 'Polygon'){
      this.polygon.move(pointer);
      
      inst.canvas.renderAll();
    }
 
    inst.canvas.renderAll();
  };

  Shape.prototype.deleteSelected = function (o) {
    this.canvas.getActiveObject().remove();
  };

  Shape.prototype.objectsSelected = function (o) {
    var obj = this.canvas.getActiveObject();
    if(!obj) return false;
    return true;
  };

  Shape.prototype.onMouseDown = function (o) {
    var inst = this;
    if(!inst.isOver() && !inst.isBlocked()){
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

    }
  };

  Shape.prototype.isEnable = function(){
    return this.isDrawing;
  }

  Shape.prototype.isBlocked = function(){
    return this.blocked;
  }

  Shape.prototype.enable = function(){
    this.isDrawing = true;
  }

  Shape.prototype.disable = function(){
    this.isDrawing = false;
  }

  Shape.prototype.isOver = function(){
    return this.overObject;
  }

  Shape.prototype.over = function(){
    this.overObject = true;
  }

  Shape.prototype.out = function(){
    this.overObject = false;
  }

  Shape.prototype.block = function(){
    this.blocked = true;
  }

  Shape.prototype.unblock = function(){
    this.blocked = false;
  }

  Shape.prototype.setShape = function(t){
    this.type = t;
  }


  return Shape;
}());