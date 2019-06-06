package tests

import (
	"testing"
	"bytes"
	"image"
	"net/http"
	_ "image/png"
	 "image/jpeg"
	"io/ioutil"
	"strconv"
	"net/url"
)

func testGetImage(t *testing.T, imageId string, scaledWidth int, scaledHeight int, 
					highlight string, requiredStatusCode int) (image.Image, int, int) {
	u := BASE_URL + API_VERSION + "/donation/" + imageId + "?width=" + strconv.Itoa(scaledWidth) + "&height=" + strconv.Itoa(scaledHeight)

	if highlight != "" {
		u += "&highlight=" + url.QueryEscape(highlight)
	}
	
	resp, err := http.Get(u)
	ok(t, err)

	defer resp.Body.Close()

    ok(t, err)
    equals(t, resp.StatusCode, requiredStatusCode)

    buf, err := ioutil.ReadAll(resp.Body)
    ok(t, err)
    img, _, err := image.Decode(bytes.NewReader(buf))
    ok(t, err)

    imgConf, _, err := image.DecodeConfig(bytes.NewReader(buf))
    ok(t, err)

    return img, imgConf.Width, imgConf.Height
}

func TestGetOriginalImage(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	_, imgWidth, imgHeight := testGetImage(t, imageId, 0, 0, "", 200)

	equals(t, imgWidth, 1132)
	equals(t, imgHeight, 750)
}

func TestGetScaledImage(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	_, imgWidth, imgHeight := testGetImage(t, imageId, 500, 0, "", 200)

	equals(t, imgWidth, 500)
	equals(t, imgHeight, 331)
}

func TestGetExactlyScaledToImage(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	_, imgWidth, imgHeight := testGetImage(t, imageId, 500, 200, "", 200)

	equals(t, imgWidth, 500)
	equals(t, imgHeight, 200)
}

func TestGetImageWithNotYetAvailableHighlights(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	img1, _, _ := testGetImage(t, imageId, 500, 200, "", 200)
	img2, _, _ := testGetImage(t, imageId, 500, 200, "apple", 200)

	buf1 := new(bytes.Buffer)
	err = jpeg.Encode(buf1, img1, nil)
	ok(t, err)
	img1Bytes := buf1.Bytes()

	buf2 := new(bytes.Buffer)
	err = jpeg.Encode(buf2, img2, nil)
	ok(t, err)
	img2Bytes := buf2.Bytes()

	equals(t, img1Bytes, img2Bytes)
}


func TestGetImageWitHighlights(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
						`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	img1, _, _ := testGetImage(t, imageId, 500, 200, "", 200)
	img2, _, _ := testGetImage(t, imageId, 500, 200, "apple", 200)

	buf1 := new(bytes.Buffer)
	err = jpeg.Encode(buf1, img1, nil)
	ok(t, err)
	img1Bytes := buf1.Bytes()

	buf2 := new(bytes.Buffer)
	err = jpeg.Encode(buf2, img2, nil)
	ok(t, err)
	img2Bytes := buf2.Bytes()

	notEquals(t, img1Bytes, img2Bytes)
}

func TestDonateAndAddLabelSuggestion(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	imagesBefore, err := db.GetAllImageIds()
	ok(t, err)
	equals(t, int(len(imagesBefore)), int(0))
	
	testSignUp(t, "user", "user", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "user", 200)

	numBefore, err := db.GetNumberOfImagesWithLabelSuggestions("new-label-that-doesnt-exist-yet")
	ok(t, err)
	equals(t, int(numBefore), 0)

	testDonate(t, "./images/apples/apple1.jpeg", "new-label-that-doesnt-exist-yet", true, userToken, "", 200)

	imagesAfter, err := db.GetAllImageIds()
	ok(t, err)
	equals(t, int(len(imagesAfter)), int(1))

	numAfter, err := db.GetNumberOfImagesWithLabelSuggestions("new-label-that-doesnt-exist-yet")
	ok(t, err)
	equals(t, int(numAfter), 1)
}

func TestDonateAndAddLabelSuggestionShouldFailDueToUnauthenticatedUser(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "new-label-that-doesnt-exist-yet", true, "", "", 401)
}

func TestDonateAndAddLabelSuggestionShouldFailDueToWrongUsername(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "new-label-that-doesnt-exist-yet", true, "not-existing-token", "", 401)
}
