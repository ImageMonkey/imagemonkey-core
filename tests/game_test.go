package tests

import (
	"testing"
	"gopkg.in/resty.v1"
	"../src/datastructures"
)

func testImageHuntDonate(t *testing.T, path string, label string, token string, expectedStatusCode int) {
	numBefore, err := db.GetNumberOfImages()
	ok(t, err)

	url := BASE_URL + API_VERSION + "/games/imagehunt/donate"

	req := resty.R()

	if label == "" {
		req.
	      SetFile("image", path)
	} else {
		req.
			SetFile("image", path).
			SetFormData(map[string]string{
	        "label": label,
	      	})
	}

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Post(url)

    equals(t, resp.StatusCode(), expectedStatusCode)

    if expectedStatusCode == 200 {
	    numAfter, err := db.GetNumberOfImages(); 
	    ok(t, err)

	    equals(t, numAfter, numBefore + 1)
	}
}

func testGetImageHuntStats(t *testing.T, username string, token string, expectedStatusCode int) datastructures.ImageHuntStats {
	var imageHuntStats datastructures.ImageHuntStats

	url := BASE_URL + API_VERSION + "/user/" + username + "/games/imagehunt/stats"

	req := resty.R().
			SetResult(&imageHuntStats)

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(url)

    equals(t, resp.StatusCode(), expectedStatusCode)
    ok(t, err)

    return imageHuntStats
}

func TestImageHuntGameDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "user", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "user", 200)

	testImageHuntDonate(t, "./images/apples/apple1.jpeg", "apple", userToken, 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	numImageHuntTasks, err := db.GetNumberOfImageHuntTasksForImageWithLabel(imageId, "apple")
	ok(t, err)
	equals(t, int(numImageHuntTasks), int(1))

	numImageUserEntries, err := db.GetNumberOfImageUserEntriesForImageAndUser(imageId, "user")
	ok(t, err)
	equals(t, int(numImageUserEntries), int(1))
}

func TestImageHuntGameDonationShouldFailOnNonProductiveLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "user", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "user", 200)

	testImageHuntDonate(t, "./images/apples/apple1.jpeg", "not-existing-label", userToken, 400)
}

func TestImageHuntGameDonationShouldFailDueToUnauthenticatedUser(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testImageHuntDonate(t, "./images/apples/apple1.jpeg", "apple", "", 403)
}

func TestImageHuntGameDonationShouldFailDueToWrongToken(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testImageHuntDonate(t, "./images/apples/apple1.jpeg", "apple", "not-existing-token", 403)
}

func TestGetImageHuntStats(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "user", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "user", 200)

	testImageHuntDonate(t, "./images/apples/apple1.jpeg", "apple", userToken, 200)

	stats := testGetImageHuntStats(t, "user", userToken, 200)
	equals(t, stats.Stars, 1)
}

func TestGetImageHuntStatsShouldFailDueToWrongToken(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "user", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "user", 200)

	testSignUp(t, "user1", "user1", "user1@imagemonkey.io")
	userToken1 := testLogin(t, "user1", "user1", 200)

	testImageHuntDonate(t, "./images/apples/apple1.jpeg", "apple", userToken, 200)

	testGetImageHuntStats(t, "user", userToken1, 403)
}