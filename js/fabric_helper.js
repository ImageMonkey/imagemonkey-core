function setCanvasBackgroundImageUrl(canvas, url) {
            if (url && url.length > 0) {
                fabric.Image.fromURL(url, function (img) {
                    scaleAndPositionImage(canvas, img);
                });
            } else {
                canvas.backgroundImage = 0;
                canvas.setBackgroundImage('', canvas.renderAll.bind(canvas));

            }
          }

function scaleAndPositionImage(canvas, img) {
        var scaleFactor = 0.5;

        canvas.setBackgroundImage(img,
            canvas.renderAll.bind(canvas), {
                scaleX: scaleFactor,
                scaleY: scaleFactor
        });
        canvas.setWidth(img.width * scaleFactor);
        canvas.setHeight(img.height * scaleFactor);
        canvas.calcOffset();

        canvas.renderAll();
      }