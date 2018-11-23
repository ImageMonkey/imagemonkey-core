package tests

import (
	"testing"
	"gopkg.in/resty.v1"
	"net/url"
	"../src/datastructures"
	"encoding/json"
)

func testBrowseLabel(t *testing.T, query string, token string, requiredNumOfResults int, 
		requiredStatusCode int) []datastructures.ImageLabel {
	u := BASE_URL + API_VERSION + "/donations/labels"
	var labeledImages []datastructures.ImageLabel 

	req := resty.R().
			SetQueryParams(map[string]string{
				"query": url.QueryEscape(query),
		    }).
		    SetResult(&labeledImages)

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(u)
	
	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
	equals(t, len(labeledImages), requiredNumOfResults)

	return labeledImages
} 


func TestBrowseLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	testBrowseLabel(t, "apple", "", 1, 200)
}

func TestBrowseLabel1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "egg")

	testBrowseLabel(t, "apple&egg", "", 1, 200)
	testBrowseLabel(t, "apple|egg", "", 1, 200)
	testBrowseLabel(t, "apple&~egg", "", 0, 200)
}

func TestBrowseLabel2(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	testBrowseLabel(t, "apple&egg", "", 0, 200)
}

func TestBrowseLabelLocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "")

	testBrowseLabel(t, "apple", "", 0, 200)
}

func TestBrowseLabelLockedAndOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "")

	testBrowseLabel(t, "apple", userToken, 1, 200)
}

func TestBrowseLabelLockedButForeignDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	userToken1 := testLogin(t, "user1", "pwd1", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "")

	testBrowseLabel(t, "apple", userToken1, 0, 200)
}

func TestBrowseLabelLockedAndOwnDonationButInQuarantine(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	err = db.PutImageInQuarantine(imageId)
	ok(t, err)

	testBrowseLabel(t, "apple", userToken, 0, 200)
}

func TestBrowseLabelImageDimensions(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	testBrowseLabel(t, "apple & image.width > 15px & image.height > 15px", "", 1, 200)
}

func TestBrowseLabelImageDimensionsWithoutLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	testBrowseLabel(t, "image.width > 15px & image.height > 15px", "", 1, 200)
}

func TestBrowseLabelWrongImageDimensionsSyntax(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	testBrowseLabel(t, "apple & image.width > 15", "", 0, 422)
}


func TestBrowseLabelAnnotationCoverage(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	testBrowseLabel(t, "apple & annotation.coverage = 0%", "", 1, 200)
}

func TestBrowseLabelAnnotationCoverage2(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":60,"left":145,"type":"rect","angle":0,"width":836,"height":660,"stroke":{"color":"red","width":5}}]`, "")

	runDataProcessor(t)

	testBrowseLabel(t, "apple & annotation.coverage > 0%", "", 1, 200)
}

func TestBrowseLabelNoImageDescription(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	images := testBrowseLabel(t, "apple", "", 1, 200)

	equals(t, len(images[0].Image.Descriptions), 0)
}

func TestBrowseLabelOneImageDescription(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	var imageDscs []datastructures.ImageDescription
	imageDsc := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDscs = append(imageDscs, imageDsc)

	testAddImageDescriptions(t, imageId, imageDscs)

	images := testBrowseLabel(t, "apple", "", 1, 200)

	equals(t, len(images[0].Image.Descriptions), 1)

	rawImageDescriptions := images[0].Image.Descriptions

	type ImageDescription struct {
		Text string `json:"text"`
	}

	var imageDescription ImageDescription

	err = json.Unmarshal(rawImageDescriptions[0], &imageDescription)
	ok(t, err)
	equals(t, imageDescription.Text, "apple on the floor")
}


func TestBrowseLabelOneImageDescriptionLocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	var imageDescriptions []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDescriptions = append(imageDescriptions, imageDescription)

	testAddImageDescriptions(t, imageId, imageDescriptions)

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)

	err = db.GiveUserModeratorRights("moderator")
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	testLockImageDescription(t, imageId, descriptions[0].Uuid, moderatorToken, 201)

	images := testBrowseLabel(t, "apple", "", 1, 200)

	equals(t, len(images[0].Image.Descriptions), 0)
}


func TestBrowseLabelOneImageDescriptionUnlocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	var imageDscs []datastructures.ImageDescription
	imageDsc := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDscs = append(imageDscs, imageDsc)

	testAddImageDescriptions(t, imageId, imageDscs)

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)

	err = db.GiveUserModeratorRights("moderator")
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	testUnlockImageDescription(t, imageId, descriptions[0].Uuid, moderatorToken, 201)

	images := testBrowseLabel(t, "apple", "", 1, 200)

	equals(t, len(images[0].Image.Descriptions), 1)

	rawImageDescriptions := images[0].Image.Descriptions

	type ImageDescription struct {
		Text string `json:"text"`
	}

	var imageDescription ImageDescription

	err = json.Unmarshal(rawImageDescriptions[0], &imageDescription)
	ok(t, err)
	equals(t, imageDescription.Text, "apple on the floor")
}

func TestOnlyOneLabelAccessorPerLabelId(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	moreThanOneLabelId, err := db.DoLabelAccessorsBelongToMoreThanOneLabelId()
	ok(t, err)
	equals(t, moreThanOneLabelId, false)
}