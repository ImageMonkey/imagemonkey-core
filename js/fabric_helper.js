function setCanvasBackgroundImageUrl(canvas, url, callback) {
    if (url && url.length > 0) {
        fabric.Image.fromURL(url, function (img) {
            scaleAndPositionImage(canvas, img, callback);
        });
    } else {
        canvas.backgroundImage = 0;
        canvas.setBackgroundImage('', canvas.renderAll.bind(canvas));

    }
}

function scaleAndPositionImage(canvas, img, callback) {
    var scaleFactor = calcScaleFactor(img);
    canvas.setBackgroundImage(img,
        canvas.renderAll.bind(canvas), {
            scaleX: scaleFactor,
            scaleY: scaleFactor
        }
    );
    canvas.setWidth(img.width * scaleFactor);
    canvas.setHeight(img.height * scaleFactor);
    canvas.calcOffset();

    canvas.renderAll();
    typeof callback === 'function' && callback();
}

function calcScaleFactor(img){
    var maxImageWidth = 600.0;
    
    //on mobile, make image full width
    var isMobile = window.matchMedia("only screen and (max-width: 760px)");
    if (isMobile.matches) {
        maxImageWidth = document.body.clientWidth - 70;
    }
    var scaleFactor = maxImageWidth/img.width;
    if(scaleFactor > 1.0)
        scaleFactor = 1.0;

    return scaleFactor;
}

function drawAnnotations(canvas, annotations, scaleFactor){
      for(var i = 0; i < annotations.length; i++){
        var top = (annotations[i]["top"] * scaleFactor);
        var left = (annotations[i]["left"] * scaleFactor);
        var height = (annotations[i]["height"] * scaleFactor);
        var width = (annotations[i]["width"] * scaleFactor);

        var rect = new fabric.Rect({
            left: left,
            top: top,
            originX: 'left',
            originY: 'top',
            width: width,
            height: height,
            angle: 0,
            stroke: 'red',
            strokeWidth: 5,
            fill: "transparent",
            transparentCorners: false,
            hasBorders: false,
            hasControls: false,
            selectable: false
        });
        canvas.add(rect);
        canvas.renderAll();
      }
    }