package tests

import (
	"testing"
	"time"
	"os"
	"os/exec"
	"gopkg.in/resty.v1"
	"github.com/bbernhard/imagemonkey-core/datastructures"
	"github.com/bbernhard/imagemonkey-core/ioutils"
)

func getLabelFiles(t *testing.T) {
	os.RemoveAll("/tmp/labels-unittest")
	err := ioutils.CopyDirectory("../wordlists", "/tmp/labels-unittest")	
	ok(t, err)
}



func runLabelBot(t *testing.T, runType string) {
	getLabelFiles(t)

	// Start a process
	cmd := exec.Command("go", "run", "-tags", "dev " + runType, "bot.go", "-singleshot=true", 
						"-git_checkout_dir=/tmp/labels-unittest", "-use_sentry=false")
	cmd.Dir = "../src"
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	ok(t, err)

	// Wait for the process to finish or kill it after a timeout:
	done := make(chan error, 1)
	go func() {
	    done <- cmd.Wait()
	}()
	select {
	case <-time.After(60 * time.Second):
	    err := cmd.Process.Kill()
	    ok(t, err) //failed to kill process
	    t.Errorf("process killed as timeout reached")
	case err := <-done:
	    ok(t, err)
	}
}

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

func testAcceptTrendingLabel(t *testing.T, name string, description string, plural string, renameTo string, 
								token string, labelType string, requiredStatusCode int) {
	url := BASE_URL + API_VERSION + "/trendinglabels/" + name + "/accept" 	

	var acceptTrendingLabel datastructures.AcceptTrendingLabel
	acceptTrendingLabel.Label.Type = labelType
	acceptTrendingLabel.Label.Description = description 
	acceptTrendingLabel.Label.Plural = plural
	acceptTrendingLabel.Label.RenameTo = renameTo

	req := resty.R().
			SetBody(&acceptTrendingLabel)	
	
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
	runTrendingLabelsWorker(t, 5)

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
	runTrendingLabelsWorker(t, 5)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)

	testAcceptTrendingLabel(t, "hallowelt 1", "", "hallowelt 1s", "hallowelt 1", token, "normal", 201)
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
	runTrendingLabelsWorker(t, 5)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)

	beforeState, err := db.GetTrendingLabelBotTaskState("hallowelt 1")
	ok(t, err)
	equals(t, beforeState, "")

	testAcceptTrendingLabel(t, "hallowelt 1", "", "hallowelt 1s", "hallowelt 1", token, "normal", 201)

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
	runTrendingLabelsWorker(t, 5)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)

	beforeState, err := db.GetTrendingLabelBotTaskState("hallowelt 1")
	ok(t, err)
	equals(t, beforeState, "")

	err = db.GiveUserModeratorRights("testuser")
	ok(t, err)

	testAcceptTrendingLabel(t, "hallowelt 1", "", "hallowelt 1s", "hallowelt 1", token, "normal", 201)

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
	runTrendingLabelsWorker(t, 5)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)

	testAcceptTrendingLabel(t, "hallowelt 1", "", "hallowelt 1s", "hallowelt 1", "invalid token", "normal", 401)
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
	runTrendingLabelsWorker(t, 5)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)


	testAcceptTrendingLabel(t, "hallowelt 1", "", "hallowelt 1s", "hallowelt 1", token, "normal", 201)

	err = db.SetTrendingLabelBotTaskState("hallowelt 1", "build-failed")
	ok(t, err)

	beforeState, err := db.GetTrendingLabelBotTaskState("hallowelt 1")
	ok(t, err)
	equals(t, beforeState, "build-failed") 

	testAcceptTrendingLabel(t, "hallowelt 1", "", "hallowelt 1s", "hallowelt 1", token, "normal", 201)

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
	runTrendingLabelsWorker(t, 5)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)

	testAcceptTrendingLabel(t, "not-existing", "", "not-existing-plural", "not-existing", token, "normal", 404)
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
	runTrendingLabelsWorker(t, 5)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)


	testAcceptTrendingLabel(t, "hallowelt 1", "", "hallowelt 1s", "hallowelt 1", token, "normal", 201)

	err = db.SetTrendingLabelBotTaskState("hallowelt 1", "building")
	ok(t, err)

	beforeState, err := db.GetTrendingLabelBotTaskState("hallowelt 1")
	ok(t, err)
	equals(t, beforeState, "building") 

	testAcceptTrendingLabel(t, "hallowelt 1", "", "hallowelt 1s", "hallowelt 1", token, "normal", 201)

	afterState, err := db.GetTrendingLabelBotTaskState("hallowelt 1")
	ok(t, err)
	equals(t, afterState, "building") 
}

func TestAcceptTrendingLabelForBotFailsAsLabelAlreadyExists(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200)

	err := db.GiveUserModeratorRights("testuser")
	ok(t, err)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for _, imageId := range imageIds {
		testSuggestLabelForImage(t, imageId, "hallowelt", true, token)
		testSuggestLabelForImage(t, imageId, "hallowelt 1", true, token)
	}
	runTrendingLabelsWorker(t, 5)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)


	testAcceptTrendingLabel(t, "hallowelt 1", "", "apples", "apple", token, "normal", 201)

	runLabelBot(t, "cisuccess")

	beforeState, err := db.GetTrendingLabelBotTaskState("hallowelt 1")
	ok(t, err)
	equals(t, beforeState, "already exists")
}

func TestAcceptTrendingLabelForBotFailsAsLabelAlreadyExistsButNotProductive(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200)

	err := db.GiveUserModeratorRights("testuser")
	ok(t, err)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for _, imageId := range imageIds {
		testSuggestLabelForImage(t, imageId, "hallowelt", true, token)
		testSuggestLabelForImage(t, imageId, "hallowelt 1", true, token)
	}
	runTrendingLabelsWorker(t, 5)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 2)


	testAcceptTrendingLabel(t, "hallowelt 1", "", "hallowelt 1s", "hallowelt 1", token, "normal", 201)
	runLabelBot(t, "cisuccess")

	state, err := db.GetTrendingLabelBotTaskState("hallowelt 1")
	ok(t, err)
	equals(t, state, "merged")

	testAcceptTrendingLabel(t, "hallowelt", "", "hallowelt 1s", "hallowelt 1", token, "normal", 201)
	runLabelBot(t, "cisuccess")

	state, err = db.GetTrendingLabelBotTaskState("hallowelt")
	ok(t, err)
	equals(t, state, "already exists")
}

