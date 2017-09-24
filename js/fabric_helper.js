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
    var scaleFactor = maxImageWidth/img.width;
    if(scaleFactor > 1.0)
        scaleFactor = 1.0;

    return scaleFactor;
}