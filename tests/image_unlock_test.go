package tests


import (
	"testing"
	"gopkg.in/resty.v1"
	"../src/datastructures"
)

func testGetAllUnverifiedDonations(t *testing.T, clientId string, clientSecret string, 
									requiredStatusCode int, requiredNumOfResults int) []datastructures.LockedImage {
	u := BASE_URL + API_VERSION + "/internal/unverified-donations"
	var images []datastructures.LockedImage

	req := resty.R().
		    SetResult(&images)

	if clientId != "" {
		req.SetHeader("X-Client-Id", clientId)
	}

	if clientSecret != "" {
		req.SetHeader("X-Client-Secret", clientSecret)
	}

	resp, err := req.Get(u)
	
	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
	equals(t, len(images), requiredNumOfResults)

	return images
}


func testUnlockUnverifiedDonation(t *testing.T, clientId string, clientSecret string, imageId string,
									requiredStatusCode int) {
	u := BASE_URL + API_VERSION + "/unverified/donation/" + imageId + "/good"

	req := resty.R()

	if clientId != "" {
		req.SetHeader("X-Client-Id", clientId)
	}

	if clientSecret != "" {
		req.SetHeader("X-Client-Secret", clientSecret)
	}

	resp, err := req.Post(u)
	
	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
}


func testDeleteUnverifiedDonation(t *testing.T, clientId string, clientSecret string, imageId string,
									requiredStatusCode int) {
	u := BASE_URL + API_VERSION + "/unverified/donation/" + imageId + "/delete"

	req := resty.R()

	if clientId != "" {
		req.SetHeader("X-Client-Id", clientId)
	}

	if clientSecret != "" {
		req.SetHeader("X-Client-Secret", clientSecret)
	}

	resp, err := req.Post(u)
	
	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
}

func testQuarantineUnverifiedDonation(t *testing.T, clientId string, clientSecret string, imageId string,
									requiredStatusCode int) {
	u := BASE_URL + API_VERSION + "/unverified/donation/" + imageId + "/quarantine"

	req := resty.R()

	if clientId != "" {
		req.SetHeader("X-Client-Id", clientId)
	}

	if clientSecret != "" {
		req.SetHeader("X-Client-Secret", clientSecret)
	}

	resp, err := req.Post(u)
	
	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
}

func testBatchHandleUnverifiedDonations(t *testing.T, clientId string, clientSecret string, 
										imageBatch []datastructures.LockedImageAction, requiredStatusCode int) {
	u := BASE_URL + API_VERSION + "/unverified/donation"

	req := resty.R().
			SetBody(imageBatch)

	if clientId != "" {
		req.SetHeader("X-Client-Id", clientId)
	}

	if clientSecret != "" {
		req.SetHeader("X-Client-Secret", clientSecret)
	}

	resp, err := req.Patch(u)
	
	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
}

func TestGetAllUnverifiedDonations(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "")
	testGetAllUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, 200, 1)
}

func TestGetAllUnverifiedDonationsNotAuthenticated(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "")
	testGetAllUnverifiedDonations(t, "", "", 401, 0)
}

func TestGetAllUnverifiedDonationsNothingToDo(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")
	testGetAllUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, 200, 0)
}

func TestUnlockImage(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "")
	testGetAllUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, 200, 1)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testUnlockUnverifiedDonation(t, X_CLIENT_ID, X_CLIENT_SECRET, imageId, 201)
	testGetAllUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, 200, 0)
}

func TestUnlockImageFailsDueToWrongClientAuth(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "")
	testGetAllUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, 200, 1)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testUnlockUnverifiedDonation(t, "wrong-client-id", "wrong-client-secret", imageId, 401)
	testGetAllUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, 200, 1)
}

func TestDeleteDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "")
	testGetAllUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, 200, 1)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testDeleteUnverifiedDonation(t, X_CLIENT_ID, X_CLIENT_SECRET, imageId, 201)
	testGetAllUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, 200, 0)
}

func TestDeleteDonationFailsDueToWrongClientAuth(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "")
	testGetAllUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, 200, 1)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testDeleteUnverifiedDonation(t, "wrong-client-id", "wrong-client-secret", imageId, 401)
	testGetAllUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, 200, 1)
}

func TestQuarantineDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "")
	testGetAllUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, 200, 1)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testQuarantineUnverifiedDonation(t, X_CLIENT_ID, X_CLIENT_SECRET, imageId, 201)
	testGetAllUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, 200, 0)
}

func TestQuarantineDonationFailsDueToWrongClientAuth(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "")
	testGetAllUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, 200, 1)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testDeleteUnverifiedDonation(t, "wrong-client-id", "wrong-client-secret", imageId, 401)
	testGetAllUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, 200, 1)
}

func TestBatchHandleLockedImages(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "")
	testDonate(t, "./images/apples/apple2.jpeg", "apple", false, "", "")
	testDonate(t, "./images/apples/apple3.jpeg", "apple", false, "", "")
	testGetAllUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, 200, 3)

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	var imageBatch []datastructures.LockedImageAction

	for i, imageId := range imageIds {
		if i == 0 {
			lockedImg1 := datastructures.LockedImageAction{Action: "quarantine", ImageId: imageId}
			imageBatch = append(imageBatch, lockedImg1)
		} else if i == 1 {
			lockedImg2 := datastructures.LockedImageAction{Action: "good", ImageId: imageId}
			imageBatch = append(imageBatch, lockedImg2)
		} else if i == 2 {
			lockedImg3 := datastructures.LockedImageAction{Action: "good", ImageId: imageId}
			imageBatch = append(imageBatch, lockedImg3)
		}
	}

	testBatchHandleUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, imageBatch, 204)
	testGetAllUnverifiedDonations(t, X_CLIENT_ID, X_CLIENT_SECRET, 200, 0)

	//first image
	unlocked, err := db.IsImageUnlocked(imageIds[0])
	ok(t, err)
	equals(t, unlocked, false)

	inQuarantine, err := db.IsImageInQuarantine(imageIds[0])
	ok(t, err)
	equals(t, inQuarantine, true)

	//second image
	unlocked, err = db.IsImageUnlocked(imageIds[1])
	ok(t, err)
	equals(t, unlocked, true)

	inQuarantine, err = db.IsImageInQuarantine(imageIds[1])
	ok(t, err)
	equals(t, inQuarantine, false)

	//third image
	unlocked, err = db.IsImageUnlocked(imageIds[2])
	ok(t, err)
	equals(t, unlocked, true)

	inQuarantine, err = db.IsImageInQuarantine(imageIds[2])
	ok(t, err)
	equals(t, inQuarantine, false)
}
