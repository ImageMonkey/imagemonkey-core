package tests

import (
	"testing"
	"gopkg.in/resty.v1"
	"net/url"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	"encoding/json"
	"sort"
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

func testGetLabelsForImage(t *testing.T, imageId string, token string, includeOnlyUnlockedLabels bool, requiredStatusCode int) []datastructures.LabelMeEntry {
	u := BASE_URL + API_VERSION + "/donation/" + imageId + "/labels"

	var labels []datastructures.LabelMeEntry

	urlParam := "false"
	if includeOnlyUnlockedLabels {
		urlParam = "true"
	}

	req := resty.R().
			SetQueryParams(map[string]string{
				"only_unlocked_labels": urlParam,
		    }).
		    SetResult(&labels)

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(u)
	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)

	return labels
}


func TestBrowseLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	testBrowseLabel(t, "apple", "", 1, 200)
}

func TestBrowseLabel1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "egg", "")

	testBrowseLabel(t, "apple&egg", "", 1, 200)
	testBrowseLabel(t, "apple|egg", "", 1, 200)
	testBrowseLabel(t, "apple&~egg", "", 0, 200)
}

func TestBrowseLabel2(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	testBrowseLabel(t, "apple&egg", "", 0, 200)
}

func TestBrowseLabelLocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "", 200)

	testBrowseLabel(t, "apple", "", 0, 200)
}

func TestBrowseLabelLockedAndOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "", 200)

	testBrowseLabel(t, "apple", userToken, 1, 200)
}

func TestBrowseLabelLockedButForeignDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	userToken1 := testLogin(t, "user1", "pwd1", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "", 200)

	testBrowseLabel(t, "apple", userToken1, 0, 200)
}

func TestBrowseLabelLockedAndOwnDonationButInQuarantine(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	err = db.PutImageInQuarantine(imageId)
	ok(t, err)

	testBrowseLabel(t, "apple", userToken, 0, 200)
}

func TestBrowseLabelImageDimensions(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	testBrowseLabel(t, "apple & image.width > 15px & image.height > 15px", "", 1, 200)
}

func TestBrowseLabelImageDimensionsWithoutLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	testBrowseLabel(t, "image.width > 15px & image.height > 15px", "", 1, 200)
}

func TestBrowseLabelWrongImageDimensionsSyntax(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	testBrowseLabel(t, "apple & image.width > 15", "", 0, 422)
}


func TestBrowseLabelAnnotationCoverage(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	testBrowseLabel(t, "apple & annotation.coverage = 0%", "", 1, 200)
}

func TestBrowseLabelAnnotationCoverage2(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":60,"left":145,"type":"rect","angle":0,"width":836,"height":660,"stroke":{"color":"red","width":5}}]`, "", 201)

	runDataProcessor(t)

	testBrowseLabel(t, "apple & annotation.coverage > 0%", "", 1, 200)
}

func TestBrowseLabelNoImageDescription(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	images := testBrowseLabel(t, "apple", "", 1, 200)

	equals(t, len(images[0].Image.Descriptions), 0)
}

func TestBrowseLabelOneImageDescription(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

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

func TestGetLabelsForImage(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	labels := testGetLabelsForImage(t, imageId, "", false, 200)
	equals(t, len(labels), 1)
	equals(t, labels[0].Label, "apple")
	equals(t, labels[0].Unlocked, true)
	equals(t, labels[0].Annotatable, true)
	equals(t, int(labels[0].Validation.NumOfValid), int(0))
	equals(t, int(labels[0].Validation.NumOfInvalid), int(0))
}
func TestGetLabelsForImageMultipleLabels(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "table", "")

	labels := testGetLabelsForImage(t, imageId, "", false, 200)
	equals(t, len(labels), 2)

	sort.SliceStable(labels, func(i, j int) bool { return labels[i].Label < labels[j].Label })


	equals(t, labels[0].Label, "apple")
	equals(t, labels[0].Unlocked, true)
	equals(t, labels[0].Annotatable, true)
	equals(t, labels[0].Validation.NumOfValid, int32(0))
	equals(t, labels[0].Validation.NumOfInvalid, int32(0))

	equals(t, labels[1].Label, "table")
	equals(t, labels[1].Unlocked, true)
	equals(t, labels[1].Annotatable, true)
	equals(t, labels[1].Validation.NumOfValid, int32(0))
	equals(t, labels[1].Validation.NumOfInvalid, int32(0))
}

func TestGetLabelsForImageMultipleLabelsIncludingNonProductiveOnes(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "table", "")
	testSuggestLabelForImage(t, imageId, "not-existing", true, userToken)

	labels := testGetLabelsForImage(t, imageId, "", false, 200)

	sort.SliceStable(labels, func(i, j int) bool { return labels[i].Label < labels[j].Label })

	equals(t, len(labels), 3)
	equals(t, labels[0].Label, "apple")
	equals(t, labels[0].Unlocked, true)
	equals(t, labels[0].Annotatable, true)
	equals(t, labels[0].Validation.NumOfValid, int32(0))
	equals(t, labels[0].Validation.NumOfInvalid, int32(0))

	equals(t, labels[1].Label, "not-existing")
	equals(t, labels[1].Unlocked, false)
	equals(t, labels[1].Annotatable, true)
	equals(t, labels[1].Validation.NumOfValid, int32(0))
	equals(t, labels[1].Validation.NumOfInvalid, int32(0))

	equals(t, labels[2].Label, "table")
	equals(t, labels[2].Unlocked, true)
	equals(t, labels[2].Annotatable, true)
	equals(t, labels[2].Validation.NumOfValid, int32(0))
	equals(t, labels[2].Validation.NumOfInvalid, int32(0))
}

func TestGetLabelsForImageMultipleLabelsWithoutNonProductiveOnes(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "table", "")
	testSuggestLabelForImage(t, imageId, "not-existing", true, userToken)

	labels := testGetLabelsForImage(t, imageId, "", true, 200)

	sort.SliceStable(labels, func(i, j int) bool { return labels[i].Label < labels[j].Label })

	equals(t, len(labels), 2)
	equals(t, labels[0].Label, "apple")
	equals(t, labels[0].Unlocked, true)
	equals(t, labels[0].Annotatable, true)
	equals(t, labels[0].Validation.NumOfValid, int32(0))
	equals(t, labels[0].Validation.NumOfInvalid, int32(0))

	/*equals(t, labels[1].Label, "not-existing")
	equals(t, labels[1].Unlocked, false)
	equals(t, labels[1].Annotatable, true)
	equals(t, labels[1].Validation.NumOfValid, int32(0))
	equals(t, labels[1].Validation.NumOfInvalid, int32(0))*/

	equals(t, labels[1].Label, "table")
	equals(t, labels[1].Unlocked, true)
	equals(t, labels[1].Annotatable, true)
	equals(t, labels[1].Validation.NumOfValid, int32(0))
	equals(t, labels[1].Validation.NumOfInvalid, int32(0))
}
