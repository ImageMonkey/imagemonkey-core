package tests

import (
	"testing"
	"gopkg.in/resty.v1"
	"../src/datastructures"
)

func testGetUnannotatedValidations(t *testing.T, imageId string, token string, requiredStatusCode int, requiredNumOfResults int) {
	url := BASE_URL +API_VERSION + "/donation/" + imageId + "/validations/unannotated"

	var unannotatedValidations []datastructures.UnannotatedValidation

	req := resty.R().
		   SetResult(&unannotatedValidations)

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(url)
	
	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
	equals(t, len(unannotatedValidations), requiredNumOfResults)
}

func TestGetUnannotatedValidations(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	
	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetUnannotatedValidations(t, imageId, "", 200, 1)
}

func TestMultipleGetUnannotatedValidations(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	
	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "dog")
	testLabelImage(t, imageId, "cat")

	testGetUnannotatedValidations(t, imageId, "", 200, 3)
}

func TestMultipleGetUnannotatedValidations2(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	
	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "dog")
	testLabelImage(t, imageId, "cat")

	testAnnotate(t, imageId, "dog", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "")

	testGetUnannotatedValidations(t, imageId, "", 200, 2)
}


func TestGetUnannotatedValidationsButLocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	
	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetUnannotatedValidations(t, imageId, "", 200, 0)
}

func TestGetUnannotatedValidationsLockedButOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)
	
	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetUnannotatedValidations(t, imageId, token, 200, 1)
}

func TestGetUnannotatedValidationsLockedOwnDonationButQuarantine(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)
	
	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	err = db.PutImageInQuarantine(imageId)
	ok(t, err)

	testGetUnannotatedValidations(t, imageId, token, 200, 0)
}

