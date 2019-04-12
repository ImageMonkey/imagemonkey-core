package tests

import (
	"testing"
	"gopkg.in/resty.v1"
	"../src/datastructures"
	"encoding/json"
)

func testInternalAutoDonate(t *testing.T, imageId string, label string, sublabel string, annotations string, 
							clientId string, clientSecret string, requiredStatusCode int) {
	annotationEntry := datastructures.Annotations{Label: label, Sublabel: sublabel}

	err := json.Unmarshal([]byte(annotations), &annotationEntry.Annotations)
	ok(t, err)

	url := BASE_URL + API_VERSION + "/internal/auto-annotate/" + imageId
	req := resty.R().
			SetHeader("Content-Type", "application/json").
			SetBody(annotationEntry)
			

    if clientId != "" {
		req.SetHeader("X-Client-Id", clientId)
	}

	if clientSecret != "" {
		req.SetHeader("X-Client-Secret", clientSecret)
	}
	
	resp, err := req.Post(url)

	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
}

func TestInternalAutoDonate(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testInternalAutoDonate(t, imageId, "apple", "", `[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, 
							X_CLIENT_ID, X_CLIENT_SECRET, 201)
}

func TestInternalAutoDonateShouldFailDueToWrongClientSecret(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testInternalAutoDonate(t, imageId, "apple", "", `[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, 
							X_CLIENT_ID, "wrong-secret", 401)
}

func TestInternalAutoDonateShouldFailDueToWrongClientId(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testInternalAutoDonate(t, imageId, "apple", "", `[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, 
							"wrong-client-id", X_CLIENT_SECRET, 401)

}