package commons

import (
    log "github.com/Sirupsen/logrus"
    "gopkg.in/h2non/bimg.v1"
    //"os"
    //"github.com/nfnt/resize"
    "gocv.io/x/gocv"
    "image/color"
    _"image/gif"
    "image/jpeg"
    _"image/png"
    "math"
    "bytes"
    "image"
    "errors"
)

func ResizeImage(path string, scaleToWidth int, scaleToHeight int) ([]byte, string, error) {
    imgFormat := ""

    buffer, err := bimg.Read(path)
    if err != nil {
      log.Error("[Resize Image Handler] Couldn't open image: ", err.Error())
      return []byte{}, imgFormat, err
    }

    img := bimg.NewImage(buffer)
    imgSize, err := img.Size()
    if err != nil {
        log.Error("[Resize Image Handler] Couldn't determine image size: ", err.Error())
        return []byte{}, imgFormat, err
    }

    width := imgSize.Width
    height := imgSize.Height

    scaleFactor := 1.0
    if scaleToWidth != 0 && scaleToHeight != 0 {
        width = scaleToWidth
        height = scaleToHeight
    } else {
        if scaleToWidth != 0 {
            width = scaleToWidth
            scaleFactor = float64(scaleToWidth) / float64(imgSize.Width)
            height = int(math.Round(float64(scaleFactor) * float64(height)))
        } else {
            height = scaleToHeight
            scaleFactor = float64(scaleToHeight) / float64(imgSize.Height)
            width = int(math.Round(float64(scaleFactor) * float64(width)))
        }
    }

    if width != imgSize.Width && height != imgSize.Height {
        newImage, err := img.Resize(width, height)
        if err != nil {
          log.Error("[Resize Image Handler] Couldn't resize image: ", err.Error())
          return []byte{}, imgFormat, err
        }
        return newImage, img.Type(), nil
    }

    return img.Image(), img.Type(), nil
}

/*func ResizeImage(path string, width uint, height uint) ([]byte, string, error){
    buf := new(bytes.Buffer) 
    imgFormat := ""

    file, err := os.Open(path)
    if err != nil {
        log.Debug("[Resize Image Handler] Couldn't open image: ", err.Error())
        return buf.Bytes(), imgFormat, err
    }

    // decode jpeg into image.Image
    img, format, err := image.Decode(file)
    if err != nil {
        log.Debug("[Resize Image Handler] Couldn't decode image: ", err.Error())
        return buf.Bytes(), imgFormat, err
    }
    file.Close()

    resizedImg := resize.Resize(width, height, img, resize.NearestNeighbor)


    if format == "png" {
        err = png.Encode(buf, resizedImg)
        if err != nil {
            log.Debug("[Resize Image Handler] Couldn't encode image: ", err.Error())
            return buf.Bytes(), imgFormat, err
        }
    } else if format == "gif" {
        err = gif.Encode(buf, resizedImg, nil)
        if err != nil {
            log.Debug("[Resize Image Handler] Couldn't encode image: ", err.Error())
            return buf.Bytes(), imgFormat, err
        }
    } else {
        err = jpeg.Encode(buf, resizedImg, nil)
        if err != nil {
            log.Debug("[Resize Image Handler] Couldn't encode image: ", err.Error())
            return buf.Bytes(), imgFormat, err
        }
    }
    imgFormat = format

    return buf.Bytes(), imgFormat, nil
}*/


func HighlightAnnotationsInImage(path string, regions []image.Rectangle, scaleToWidth int, scaleToHeight int) ([]byte, error) {
    img := gocv.IMRead(path, gocv.IMReadColor)
    defer img.Close()
    if img.Empty() {
        return []byte{}, errors.New("")
    }

    mask := gocv.NewMatWithSize(img.Rows(), img.Cols(), gocv.MatTypeCV8UC3)
    defer mask.Close()

    //gocv.Rectangle(&mask, image.Rect(0, 0, img.Size()[1], img.Size()[0]), color.RGBA{255, 0, 0, 0}, -1)

    for _, region := range regions {
        gocv.Rectangle(&mask, region, color.RGBA{255, 255, 255, 0}, -1)
    }

    dstImage := gocv.NewMatWithSize(img.Rows(), img.Cols(), img.Type())
    defer dstImage.Close()

    img.CopyToWithMask(&dstImage, mask)


    imgSize := img.Size()
    height := imgSize[0]
    width := imgSize[1]

    scaleFactor := 1.0
    if scaleToWidth != 0 && scaleToHeight != 0 {
        width = scaleToWidth
        height = scaleToHeight
    } else {
        if scaleToWidth != 0 {
            width = scaleToWidth
            scaleFactor = float64(scaleToWidth) / float64(imgSize[1])
            height = int(math.Round(float64(scaleFactor) * float64(height)))
        } else {
            height = scaleToHeight
            scaleFactor = float64(scaleToHeight) / float64(imgSize[0])
            width = int(math.Round(float64(scaleFactor) * float64(width)))
        }
    }

    if width != imgSize[1] && height != imgSize[0] {
        gocv.Resize(dstImage, &dstImage, image.Point{X: width, Y:height}, 0, 0, 1)
    }


    i, err := dstImage.ToImage()

    buf := new(bytes.Buffer) 
    err = jpeg.Encode(buf, i, nil)
    if err != nil {
        log.Error("[Extract ROI from Image] Couldn't encode image: ", err.Error())
        return []byte{}, err
    }

    return buf.Bytes(), nil
}