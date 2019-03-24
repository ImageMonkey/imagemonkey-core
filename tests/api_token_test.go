package tests

import (
	"testing"
	"../src/datastructures"
	"gopkg.in/resty.v1"
	"os"
)

func testDonateWithApiToken(t *testing.T, path string, label string, unlockImage bool, apiToken string, imageCollectionName string, expectedStatusCode int) {
    numBefore, err := db.GetNumberOfImages()
    ok(t, err)

    url := BASE_URL + API_VERSION + "/donate"

    req := resty.R()

    if label == "" {
        req.
          SetFile("image", path)
    } else {
        req.
          SetFile("image", path)
        
        if imageCollectionName == "" {
            req.SetFormData(map[string]string{
            "label": label,
            })
        } else {
            req.SetFormData(map[string]string{
            "label": label,
            "image_collection": imageCollectionName,
            })
        }
          
    }

    if apiToken != "" {
        req.SetHeader("X-Api-Token", apiToken)
    }

    resp, err := req.Post(url)

    equals(t, resp.StatusCode(), expectedStatusCode)

    if expectedStatusCode == 200 {

        numAfter, err := db.GetNumberOfImages(); 
        ok(t, err)

        equals(t, numAfter, numBefore + 1)

        if unlockImage {
            //after image donation, unlock all images
            err = db.UnlockAllImages()
            ok(t, err)

            imageId, err := db.GetLatestDonatedImageId()
            ok(t, err)

            err = os.Rename(UNVERIFIED_DONATIONS_DIR + imageId, DONATIONS_DIR + imageId)
            ok(t, err)
        }
    }
}

func testRevokeApiToken(t *testing.T, username string, apiToken string, token string, expectedStatusCode int) {
	u := BASE_URL + API_VERSION + "/user/" + username + "/api-token/" + apiToken + "/revoke"

	req := resty.R()

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Post(u)
	ok(t, err)
	equals(t, resp.StatusCode(), expectedStatusCode)
}

func testCreateApiToken(t *testing.T, username string, token string, description string, expectedStatusCode int) datastructures.APIToken {
	var apiToken datastructures.APIToken
	apiTokenRequest := datastructures.ApiTokenRequest{Description: description}
	u := BASE_URL + API_VERSION + "/user/" + username + "/api-token"
	req := resty.R().
			    SetHeader("Content-Type", "application/json").
				SetBody(apiTokenRequest).
				SetResult(&apiToken)

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Post(u)
	ok(t, err)
	equals(t, resp.StatusCode(), expectedStatusCode)

	return apiToken
} 

func TestApiTokenCreate(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testCreateApiToken(t, "user", userToken, "my-api-token", 201)
}

func TestApiTokenRevoke(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	apiToken := testCreateApiToken(t, "user", userToken, "my-api-token", 201)

	testRevokeApiToken(t, "user", apiToken.Token, userToken, 200)
}

func TestCouldntRevokeApiTokenInvalidUserToken(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	apiToken := testCreateApiToken(t, "user", userToken, "my-api-token", 201)

	testRevokeApiToken(t, "user", apiToken.Token, "invalid-user-token", 422)
}

func TestCouldntRevokeApiTokenInvalidUsername(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	apiToken := testCreateApiToken(t, "user", userToken, "my-api-token", 201)

	testRevokeApiToken(t, "invalid-username", apiToken.Token, userToken, 422)
}

func TestCouldntDonateImageDueToRevokedApiToken(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	apiToken := testCreateApiToken(t, "user", userToken, "my-api-token", 201)

	testRevokeApiToken(t, "user", apiToken.Token, userToken, 200)

	testDonateWithApiToken(t, "./images/apples/apple1.jpeg", "apple", true, apiToken.Token, "", 401)
}

func TestDonateImageWithApiToken(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	apiToken := testCreateApiToken(t, "user", userToken, "my-api-token", 201)

	testDonateWithApiToken(t, "./images/apples/apple1.jpeg", "apple", true, apiToken.Token, "", 200)
}