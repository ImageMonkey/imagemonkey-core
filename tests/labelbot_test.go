package tests

import (
	"testing"
	"gopkg.in/resty.v1"
	"github.com/bbernhard/imagemonkey-core/datastructures"
)

func testGetTrendingLabels(t *testing.T, token string, requiredStatusCode int) []datastructures.TrendingLabel {
	var trendingLabels []datastructures.TrendingLabel
	
	url := BASE_URL + API_VERSION + "/trendinglabels"
	req := resty.R().
     	SetResult(&trendingLabels)
	
	if token != "" {
		req.SetAuthToken(token)	
	}

	resp, err := req.Get(url)

    ok(t, err)
    equals(t, resp.StatusCode(), requiredStatusCode)

    return trendingLabels
}

func testAcceptTrendingLabel(t *testing.T, name string, token string, requiredStatusCode int) {
	url := BASE_URL + API_VERSION + "/trendinglabels/" + name + "/accept" 
	req := resty.R()
	
	if token != "" {
		req.SetAuthToken(token)	
	}

	resp, err := req.Post(url)

    ok(t, err)
    equals(t, resp.StatusCode(), requiredStatusCode)
}

func TestGetTrendingLabelsForBot(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for _, imageId := range imageIds {
		testSuggestLabelForImage(t, imageId, "hallowelt", true, token)
		testSuggestLabelForImage(t, imageId, "hallowelt 1", true, token)
	}
	runTrendingLabelsWorker(t)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)
}

func TestAcceptTrendingLabelForBot(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for _, imageId := range imageIds {
		testSuggestLabelForImage(t, imageId, "hallowelt", true, token)
		testSuggestLabelForImage(t, imageId, "hallowelt 1", true, token)
	}
	runTrendingLabelsWorker(t)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)

	testAcceptTrendingLabel(t, "hallowelt 1", token, 201)
}

func TestAcceptTrendingLabelForBot2(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for _, imageId := range imageIds {
		testSuggestLabelForImage(t, imageId, "hallowelt", true, token)
		testSuggestLabelForImage(t, imageId, "hallowelt 1", true, token)
	}
	runTrendingLabelsWorker(t)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)

	beforeState, err := db.GetTrendingLabelBotTaskState("hallowelt 1")
	ok(t, err)
	equals(t, beforeState, "")

	testAcceptTrendingLabel(t, "hallowelt 1", token, 201)

	afterState, err := db.GetTrendingLabelBotTaskState("hallowelt 1")
	ok(t, err)
	equals(t, afterState, "waiting for moderator approval") 
}

func TestAcceptTrendingLabelForBotWithModeratorPermissions(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for _, imageId := range imageIds {
		testSuggestLabelForImage(t, imageId, "hallowelt", true, token)
		testSuggestLabelForImage(t, imageId, "hallowelt 1", true, token)
	}
	runTrendingLabelsWorker(t)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)

	beforeState, err := db.GetTrendingLabelBotTaskState("hallowelt 1")
	ok(t, err)
	equals(t, beforeState, "")

	err = db.GiveUserModeratorRights("testuser")
	ok(t, err)

	testAcceptTrendingLabel(t, "hallowelt 1", token, 201)

	afterState, err := db.GetTrendingLabelBotTaskState("hallowelt 1")
	ok(t, err)
	equals(t, afterState, "accepted") 
}

func TestCouldntAcceptTrendingLabelForBotAsNotAuthenticated(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for _, imageId := range imageIds {
		testSuggestLabelForImage(t, imageId, "hallowelt", true, token)
		testSuggestLabelForImage(t, imageId, "hallowelt 1", true, token)
	}
	runTrendingLabelsWorker(t)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)

	testAcceptTrendingLabel(t, "hallowelt 1", "invalid token", 401)
}

func TestAcceptTrendingLabelForBotRetryFailedBuild(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for _, imageId := range imageIds {
		testSuggestLabelForImage(t, imageId, "hallowelt", true, token)
		testSuggestLabelForImage(t, imageId, "hallowelt 1", true, token)
	}
	runTrendingLabelsWorker(t)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)


	testAcceptTrendingLabel(t, "hallowelt 1", token, 201)

	err = db.SetTrendingLabelBotTaskState("hallowelt 1", "build-failed")
	ok(t, err)

	beforeState, err := db.GetTrendingLabelBotTaskState("hallowelt 1")
	ok(t, err)
	equals(t, beforeState, "build-failed") 

	testAcceptTrendingLabel(t, "hallowelt 1", token, 201)

	afterState, err := db.GetTrendingLabelBotTaskState("hallowelt 1")
	ok(t, err)
	equals(t, afterState, "retry") 
}

func TestCouldntAcceptTrendingLabelForBotAsWrongLabelName(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for _, imageId := range imageIds {
		testSuggestLabelForImage(t, imageId, "hallowelt", true, token)
		testSuggestLabelForImage(t, imageId, "hallowelt 1", true, token)
	}
	runTrendingLabelsWorker(t)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)

	testAcceptTrendingLabel(t, "not-existing", token, 404)
}


func TestCannotChangeTrendingLabelBotTaskStateWhileBuilding(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for _, imageId := range imageIds {
		testSuggestLabelForImage(t, imageId, "hallowelt", true, token)
		testSuggestLabelForImage(t, imageId, "hallowelt 1", true, token)
	}
	runTrendingLabelsWorker(t)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)


	testAcceptTrendingLabel(t, "hallowelt 1", token, 201)

	err = db.SetTrendingLabelBotTaskState("hallowelt 1", "building")
	ok(t, err)

	beforeState, err := db.GetTrendingLabelBotTaskState("hallowelt 1")
	ok(t, err)
	equals(t, beforeState, "building") 

	testAcceptTrendingLabel(t, "hallowelt 1", token, 201)

	afterState, err := db.GetTrendingLabelBotTaskState("hallowelt 1")
	ok(t, err)
	equals(t, afterState, "building") 
}
