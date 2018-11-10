package tests

import (
	"testing"
	"gopkg.in/resty.v1"
	"../src/datastructures"
	"strconv"
)

type ImageDescriptionStateType int

const (
  ImageDescriptionStateUnknown ImageDescriptionStateType = 1 << iota
  ImageDescriptionStateLocked
  ImageDescriptionStateUnlocked
)

type ImageDescriptionSummary struct {
    Description string `json:"description"`
    NumOfValid int `json:"num_of_yes"`
    Uuid string `json:"uuid"`
    State ImageDescriptionStateType `json:"state"`
    Language string `json:"language"`
}

func testGetNumOfUnprocessedImageDescriptions(t *testing.T, expectedCount int, expectedStatusCode int) {
	url := BASE_URL + API_VERSION + "/donations/unprocessed-descriptions"
	req := resty.R().
			SetHeader("X-Total-Count", "")

	resp, err := req.Head(url)

	ok(t, err)
	equals(t, resp.StatusCode(), expectedStatusCode)

	count := resp.Header().Get("X-Total-Count")
	equals(t, strconv.Itoa(expectedCount), count)
}

func testGetUnprocessedImageDescriptions(t *testing.T, token string, expectedStatusCode int) []datastructures.DescriptionsPerImage {
	var imageDescriptions []datastructures.DescriptionsPerImage

	url := BASE_URL + API_VERSION + "/donations/unprocessed-descriptions"
	req := resty.R().
			SetResult(&imageDescriptions)

	if token != "" {
		req.SetAuthToken(token)
		req.SetHeader("X-Moderation", "true")
	}

	resp, err := req.Get(url)

	ok(t, err)
	equals(t, resp.StatusCode(), expectedStatusCode)

	return imageDescriptions
}

func testUnlockImageDescription(t *testing.T, imageId string, descriptionId string, token string, expectedStatusCode int) {
	url := BASE_URL + API_VERSION + "/donation/" + imageId + "/description/" + descriptionId + "/unlock"
	req := resty.R()

	if token != "" {
		req.SetAuthToken(token)
		req.SetHeader("X-Moderation", "true")
	}

	resp, err := req.Post(url)

	ok(t, err)
	equals(t, resp.StatusCode(), expectedStatusCode)
}

func testLockImageDescription(t *testing.T, imageId string, descriptionId string, token string, expectedStatusCode int) {
	url := BASE_URL + API_VERSION + "/donation/" + imageId + "/description/" + descriptionId + "/lock"
	req := resty.R()

	if token != "" {
		req.SetAuthToken(token)
		req.SetHeader("X-Moderation", "true")
	}

	resp, err := req.Post(url)

	ok(t, err)
	equals(t, resp.StatusCode(), expectedStatusCode)
}

func testGetImageDescriptions(t *testing.T, imageId string, token string, numOfDescriptions int) {
	var img datastructures.ImageToLabel

	url := BASE_URL + API_VERSION + "/donation/" + imageId + "/description"
	req := resty.R().
				SetResult(&img)

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(url)

	ok(t, err)
	equals(t, resp.StatusCode(), 200)
	equals(t, numOfDescriptions, len(img.ImageDescriptions))
}

func testAddImageDescriptions(t *testing.T, imageId string, descriptions []datastructures.ImageDescription) {
	url := BASE_URL + API_VERSION + "/donation/" + imageId + "/description"
	resp, err := resty.R().
			SetHeader("Content-Type", "application/json").
			SetBody(descriptions).
			Post(url)

	ok(t, err)
	equals(t, resp.StatusCode(), 201)
}

func TestGetImageDescription(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)

	var imageDescriptions []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDescriptions = append(imageDescriptions, imageDescription)
	testAddImageDescriptions(t, imageId, imageDescriptions)

	descriptions, err = db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 1)
	equals(t, descriptions[0].NumOfValid, 0)
	equals(t, descriptions[0].State, ImageDescriptionStateUnknown)
}

func TestGetImageDescriptionMultiple(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)


	var imageDescriptions []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDescriptions = append(imageDescriptions, imageDescription)

	testAddImageDescriptions(t, imageId, imageDescriptions)
	testAddImageDescriptions(t, imageId, imageDescriptions)

	descriptions, err = db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 1)
	equals(t, descriptions[0].NumOfValid, 1)
}

func TestGetImageDescriptionMultipleDifferent(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)

	var imageDescriptions []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDescriptions = append(imageDescriptions, imageDescription)

	testAddImageDescriptions(t, imageId, imageDescriptions)

	var imageDescriptions2 []datastructures.ImageDescription
	imageDescription2 := datastructures.ImageDescription{Description: "apple on the desk", Language: "en"}
	imageDescriptions2 = append(imageDescriptions2, imageDescription2)

	testAddImageDescriptions(t, imageId, imageDescriptions2)

	descriptions, err = db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 2)
	equals(t, descriptions[0].NumOfValid, 0)
	equals(t, descriptions[1].NumOfValid, 0)
}

func TestUnlockImageDescriptionNoModerator(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)

	var imageDescriptions []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDescriptions = append(imageDescriptions, imageDescription)
	testAddImageDescriptions(t, imageId, imageDescriptions)

	descriptions, err = db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 1)

	testUnlockImageDescription(t, imageId, descriptions[0].Uuid, "", 401)
}


func TestUnlockImageDescriptionFromModerator(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)

	var imageDescriptions []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDescriptions = append(imageDescriptions, imageDescription)

	testAddImageDescriptions(t, imageId, imageDescriptions)

	descriptions, err = db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)

	err = db.GiveUserModeratorRights("moderator")
	ok(t, err)

	equals(t, len(descriptions), 1)

	testUnlockImageDescription(t, imageId, descriptions[0].Uuid, moderatorToken, 201)

	descriptions, err = db.GetImageDescriptionForImageId(imageId)
	ok(t, err)
	equals(t, descriptions[0].State, ImageDescriptionStateUnlocked)
}

func TestUnlockImageDescriptionFromModeratorButInvalidImageId(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)

	var imageDescriptions []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDescriptions = append(imageDescriptions, imageDescription)

	testAddImageDescriptions(t, imageId, imageDescriptions)

	descriptions, err = db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)

	err = db.GiveUserModeratorRights("moderator")
	ok(t, err)

	equals(t, len(descriptions), 1)

	testUnlockImageDescription(t, "", descriptions[0].Uuid, moderatorToken, 404)
}

func TestUnlockImageDescriptionFromModeratorButInvalidDescriptionId(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)

	var imageDescriptions []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDescriptions = append(imageDescriptions, imageDescription)

	testAddImageDescriptions(t, imageId, imageDescriptions)

	descriptions, err = db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)

	err = db.GiveUserModeratorRights("moderator")
	ok(t, err)

	equals(t, len(descriptions), 1)

	testUnlockImageDescription(t, imageId, "", moderatorToken, 404)
}

func TestGetUnprocessedImageDescriptionsNoPermissions(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)


	var imageDescriptions []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDescriptions = append(imageDescriptions, imageDescription)

	testAddImageDescriptions(t, imageId, imageDescriptions)

	testGetUnprocessedImageDescriptions(t, "", 401)
}

func TestGetUnprocessedImageDescriptionsModeratorPermissions(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	var imageDscs []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDscs = append(imageDscs, imageDescription)

	testAddImageDescriptions(t, imageId, imageDscs)

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)

	err = db.GiveUserModeratorRights("moderator")
	ok(t, err)

	imageDescriptions := testGetUnprocessedImageDescriptions(t, moderatorToken, 200)
	equals(t, len(imageDescriptions), 1)
	equals(t, len(imageDescriptions[0].Image.Descriptions), 1)
}

func TestGetUnprocessedImageDescriptionsModeratorPermissionsAndUnlock(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	var imageDscs []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDscs = append(imageDscs, imageDescription)

	testAddImageDescriptions(t, imageId, imageDscs)

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)

	err = db.GiveUserModeratorRights("moderator")
	ok(t, err)

	imageDescriptions := testGetUnprocessedImageDescriptions(t, moderatorToken, 200)
	equals(t, len(imageDescriptions), 1)
	equals(t, len(imageDescriptions[0].Image.Descriptions), 1)

	testUnlockImageDescription(t, imageId, imageDescriptions[0].Image.Descriptions[0].Uuid, moderatorToken, 201)

	imageDescriptions = testGetUnprocessedImageDescriptions(t, moderatorToken, 200)
	equals(t, len(imageDescriptions), 0)
}



func TestLockImageDescriptionFromModeratorButInvalidImageId(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)

	var imageDescriptions []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDescriptions = append(imageDescriptions, imageDescription)

	testAddImageDescriptions(t, imageId, imageDescriptions)

	descriptions, err = db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)

	err = db.GiveUserModeratorRights("moderator")
	ok(t, err)

	equals(t, len(descriptions), 1)

	testLockImageDescription(t, "", descriptions[0].Uuid, moderatorToken, 404)
}




func TestGetUnprocessedImageDescriptionsModeratorPermissionsAndLock(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	var imageDscs []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDscs = append(imageDscs, imageDescription)

	testAddImageDescriptions(t, imageId, imageDscs)

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)

	err = db.GiveUserModeratorRights("moderator")
	ok(t, err)

	imageDescriptions := testGetUnprocessedImageDescriptions(t, moderatorToken, 200)
	equals(t, len(imageDescriptions), 1)
	equals(t, len(imageDescriptions[0].Image.Descriptions), 1)

	testLockImageDescription(t, imageId, imageDescriptions[0].Image.Descriptions[0].Uuid, moderatorToken, 201)

	imageDescriptions = testGetUnprocessedImageDescriptions(t, moderatorToken, 200)
	equals(t, len(imageDescriptions), 0)
}


func TestGetUnprocessedImageDescriptionsModeratorPermissionsAndLockCheckProcessedBy(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	var imageDscs []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDscs = append(imageDscs, imageDescription)

	testAddImageDescriptions(t, imageId, imageDscs)

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)

	err = db.GiveUserModeratorRights("moderator")
	ok(t, err)

	imageDescriptions := testGetUnprocessedImageDescriptions(t, moderatorToken, 200)
	equals(t, len(imageDescriptions), 1)
	equals(t, len(imageDescriptions[0].Image.Descriptions), 1)

	testLockImageDescription(t, imageId, imageDescriptions[0].Image.Descriptions[0].Uuid, moderatorToken, 201)

	imageDescriptions = testGetUnprocessedImageDescriptions(t, moderatorToken, 200)
	equals(t, len(imageDescriptions), 0)

	processedBy, err := db.GetModeratorWhoProcessedImageDescription(imageId, "apple on the floor")
	ok(t, err)
	equals(t, processedBy, "moderator")
}

func TestGetUnprocessedImageDescriptionsModeratorPermissionsAndUnlockCheckProcessedBy(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	var imageDscs []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDscs = append(imageDscs, imageDescription)

	testAddImageDescriptions(t, imageId, imageDscs)

	testSignUp(t, "nicemoderator", "nice-moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "nicemoderator", "nice-moderator", 200)

	err = db.GiveUserModeratorRights("nicemoderator")
	ok(t, err)

	imageDescriptions := testGetUnprocessedImageDescriptions(t, moderatorToken, 200)
	equals(t, len(imageDescriptions), 1)
	equals(t, len(imageDescriptions[0].Image.Descriptions), 1)

	testUnlockImageDescription(t, imageId, imageDescriptions[0].Image.Descriptions[0].Uuid, moderatorToken, 201)

	imageDescriptions = testGetUnprocessedImageDescriptions(t, moderatorToken, 200)
	equals(t, len(imageDescriptions), 0)

	processedBy, err := db.GetModeratorWhoProcessedImageDescription(imageId, "apple on the floor")
	ok(t, err)
	equals(t, processedBy, "nicemoderator")
}

func TestGetNumOfUnprocessedImageDescriptions(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	var imageDscs []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDscs = append(imageDscs, imageDescription)

	testAddImageDescriptions(t, imageId, imageDscs)
	testGetNumOfUnprocessedImageDescriptions(t, 1, 200)

	testSignUp(t, "nicemoderator", "nice-moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "nicemoderator", "nice-moderator", 200)

	err = db.GiveUserModeratorRights("nicemoderator")
	ok(t, err)

	imageDescriptions := testGetUnprocessedImageDescriptions(t, moderatorToken, 200)
	equals(t, len(imageDescriptions), 1)
	equals(t, len(imageDescriptions[0].Image.Descriptions), 1)

	testUnlockImageDescription(t, imageId, imageDescriptions[0].Image.Descriptions[0].Uuid, moderatorToken, 201)
	testGetNumOfUnprocessedImageDescriptions(t, 0, 200)
}


func TestAddImageDescriptionGermanLanguage(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)


	var imageDescriptions []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "Ein Apfel, der am Boden liegt.", Language: "ger"}
	imageDescriptions = append(imageDescriptions, imageDescription)

	testAddImageDescriptions(t, imageId, imageDescriptions)

	descriptions, err = db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 1)
	equals(t, descriptions[0].Language, "ger")
}

func TestAddImageDescriptionMultipleLanguages(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)


	var imageDescriptions []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "Ein Apfel, der am Boden liegt.", Language: "ger"}
	imageDescription2 := datastructures.ImageDescription{Description: "apple on the floor.", Language: "en"}
	imageDescriptions = append(imageDescriptions, imageDescription)
	imageDescriptions = append(imageDescriptions, imageDescription2)

	testAddImageDescriptions(t, imageId, imageDescriptions)

	descriptions, err = db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 2)
	equals(t, descriptions[0].Language, "ger")
	equals(t, descriptions[0].Description, "Ein Apfel, der am Boden liegt.")
	equals(t, descriptions[1].Language, "en")
	equals(t, descriptions[1].Description, "apple on the floor.")
}