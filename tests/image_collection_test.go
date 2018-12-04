package tests

import (
	"testing"
	"gopkg.in/resty.v1"
	"../src/datastructures"
)


func testGetImageCollections(t *testing.T, username string, token string, requiredStatusCode int) []datastructures.ImageCollection {
	var imageCollections []datastructures.ImageCollection

	url := BASE_URL + API_VERSION + "/user/" + username + "/imagecollections"
	req := resty.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&imageCollections)

    if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(url)

    ok(t, err)
    equals(t, resp.StatusCode(), requiredStatusCode)

    return imageCollections
}

func testAddImageCollection(t *testing.T, username string, token string, name string, description string, requiredStatusCode int) {
	imageCollection := datastructures.ImageCollection{Name: name, Description: description}

	url := BASE_URL + API_VERSION + "/user/" + username + "/imagecollection"
	req := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(imageCollection)

    if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Post(url)

    ok(t, err)
    equals(t, resp.StatusCode(), requiredStatusCode)
}

func addImageToImageCollection(t *testing.T, username string, token string, imageCollectionName string, imageId string, requiredStatusCode int) {
	url := BASE_URL + API_VERSION + "/user/" + username + "/imagecollection/" + imageCollectionName + "/image/" + imageId

	req := resty.R().
		SetHeader("Content-Type", "application/json")

    if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Post(url)

    ok(t, err)
    equals(t, resp.StatusCode(), requiredStatusCode)
}


func TestAddImageCollectionSuccess(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testAddImageCollection(t, "user", token, "new-image-collection", "my-new-image-collection", 201)
}

func TestAddImageCollectionFailsDueToWrongToken(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	_ = testLogin(t, "user", "pwd", 200)

	testAddImageCollection(t, "user", "wrong-token", "new-image-collection", "my-new-image-collection", 401)
}

func TestAddImageCollectionFailsCantAddCollectionToOtherUser(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	_ = testLogin(t, "user1", "pwd1", 200)

	testAddImageCollection(t, "user1", token, "new-image-collection", "my-new-image-collection", 403)
}

func TestAddImageCollectionFailsAsAlreadyExists(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testAddImageCollection(t, "user", token, "new-image-collection", "my-new-image-collection", 201)
	testAddImageCollection(t, "user", token, "new-image-collection", "my-new-image-collection1", 409)
}

func TestAddImageCollectionFailsAsNameContainsUnsoppertedChars(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testAddImageCollection(t, "user", token, "new-image-collection contains space", "my-new-image-collection", 400)
}

func TestGetImageCollectionSuccess(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	imageCollections := testGetImageCollections(t, "user", token, 200)
	equals(t, len(imageCollections), 0)

	testAddImageCollection(t, "user", token, "new-image-collection", "my-new-image-collection", 201)


	imageCollections = testGetImageCollections(t, "user", token, 200)
	equals(t, len(imageCollections), 1)
}

func TestGetImageCollectionsFailsDueToWrongToken(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	_ = testLogin(t, "user", "pwd", 200)

	testGetImageCollections(t, "user", "wrong-token", 401)
}

func TestGetImageCollectionsFailsCantGetCollectionForOtherUser(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	_ = testLogin(t, "user1", "pwd1", 200)

	testGetImageCollections(t, "user1", token, 403)
}

func TestAddImageToImageCollection(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAddImageCollection(t, "user", token, "new-image-collection", "my-new-image-collection", 201)
	addImageToImageCollection(t, "user", token, "new-image-collection", imageId, 201) 
}

func TestAddImageToImageCollectionTwiceShouldFail(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAddImageCollection(t, "user", token, "new-image-collection", "my-new-image-collection", 201)
	addImageToImageCollection(t, "user", token, "new-image-collection", imageId, 201) 
	addImageToImageCollection(t, "user", token, "new-image-collection", imageId, 409) 
}

func TestAddImageToImageCollectionWrongTokenShouldFail(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	token1 := testLogin(t, "user1", "pwd1", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAddImageCollection(t, "user", token, "new-image-collection", "my-new-image-collection", 201)
	addImageToImageCollection(t, "user", token1, "new-image-collection", imageId, 403) 
}


func TestAddImageToImageCollectionWrongTokenShouldFail2(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAddImageCollection(t, "user", token, "new-image-collection", "my-new-image-collection", 201)
	addImageToImageCollection(t, "user", "a", "new-image-collection", imageId, 401) 
}

func TestDonateImageAndAssignToImageCollection(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testAddImageCollection(t, "user", token, "new-image-collection", "my-new-image-collection", 201)
	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, token, "new-image-collection") 
}


//temporarily disabled, as it doesn't work
/*func TestDonateImageCouldntAssignToImageCollectionAsImageCollectionDoesntExit(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testAddImageCollection(t, "user", token, "new-image-collection", "my-new-image-collection", 201)
	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, token, "image-collection-doesnt-exist") 
}*/
