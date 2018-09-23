package tests

import (
	"testing"
	"gopkg.in/resty.v1"
	"../src/datastructures"
)

type ImageDescriptionSummary struct {
    Description string `json:"description"`
    NumOfValid int `json:"num_of_yes"`
    Uuid string `json:"uuid"`
    Unlocked bool `json:"unlocked"`
}

func testGetLockedImageDescriptions(t *testing.T, token string, expectedStatusCode int) []datastructures.DescriptionsPerImage {
	var imageDescriptions []datastructures.DescriptionsPerImage

	url := BASE_URL + API_VERSION + "/donations/locked-descriptions"
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

func testAddImageDescription(t *testing.T, imageId string, description string) {
	var imageDescription datastructures.ImageDescription
	imageDescription.Description = description

	url := BASE_URL + API_VERSION + "/donation/" + imageId + "/description"
	resp, err := resty.R().
			SetHeader("Content-Type", "application/json").
			SetBody(imageDescription).
			Post(url)

	ok(t, err)
	equals(t, resp.StatusCode(), 201)
}

func TestGetImageDescription(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)

	testAddImageDescription(t, imageId, "apple on the floor")

	descriptions, err = db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 1)
	equals(t, descriptions[0].NumOfValid, 0)
	equals(t, descriptions[0].Unlocked, false)
}

func TestGetImageDescriptionMultiple(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)

	testAddImageDescription(t, imageId, "apple on the floor")
	testAddImageDescription(t, imageId, "apple on the floor")

	descriptions, err = db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 1)
	equals(t, descriptions[0].NumOfValid, 1)
}

func TestGetImageDescriptionMultipleDifferent(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)

	testAddImageDescription(t, imageId, "apple on the floor")
	testAddImageDescription(t, imageId, "apple on the desk")

	descriptions, err = db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 2)
	equals(t, descriptions[0].NumOfValid, 0)
	equals(t, descriptions[1].NumOfValid, 0)
}

func TestUnlockImageDescriptionNoModerator(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)

	testAddImageDescription(t, imageId, "apple on the floor")

	descriptions, err = db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 1)

	testUnlockImageDescription(t, imageId, descriptions[0].Uuid, "", 401)
}


func TestUnlockImageDescriptionFromModerator(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)

	testAddImageDescription(t, imageId, "apple on the floor")

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
	equals(t, descriptions[0].Unlocked, true)
}

func TestUnlockImageDescriptionFromModeratorButInvalidImageId(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)

	testAddImageDescription(t, imageId, "apple on the floor")

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	equals(t, len(descriptions), 0)

	testAddImageDescription(t, imageId, "apple on the floor")

	descriptions, err = db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)

	err = db.GiveUserModeratorRights("moderator")
	ok(t, err)

	equals(t, len(descriptions), 1)

	testUnlockImageDescription(t, imageId, "", moderatorToken, 404)
}

func TestGetLockedImageDescriptionsNoPermissions(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAddImageDescription(t, imageId, "apple on the floor")

	testGetLockedImageDescriptions(t, "", 401)
}

func TestGetLockedImageDescriptionsModeratorPermissions(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAddImageDescription(t, imageId, "apple on the floor")

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)

	err = db.GiveUserModeratorRights("moderator")
	ok(t, err)

	imageDescriptions := testGetLockedImageDescriptions(t, moderatorToken, 200)
	equals(t, len(imageDescriptions), 1)
	equals(t, len(imageDescriptions[0].Descriptions), 1)
}

func TestGetLockedImageDescriptionsModeratorPermissionsAndUnlock(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAddImageDescription(t, imageId, "apple on the floor")

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)

	err = db.GiveUserModeratorRights("moderator")
	ok(t, err)

	imageDescriptions := testGetLockedImageDescriptions(t, moderatorToken, 200)
	equals(t, len(imageDescriptions), 1)
	equals(t, len(imageDescriptions[0].Descriptions), 1)

	testUnlockImageDescription(t, imageId, imageDescriptions[0].Descriptions[0].Uuid, moderatorToken, 201)

	imageDescriptions = testGetLockedImageDescriptions(t, moderatorToken, 200)
	equals(t, len(imageDescriptions), 0)
}