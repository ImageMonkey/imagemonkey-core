package tests

import (
	"testing"
	"os/exec"	
	"time"
	"os"
)

func runLabelsDownloader(t *testing.T) {
	os.RemoveAll("/tmp/labels-unittest.bak")
	// Start a process
	cmd := exec.Command("go", "run", "-tags", "dev", "labels_downloader.go", "-autoclose_github_issue=false", 
						"-singleshot=true", "-labels_dir=/tmp/labels-unittest", "-backup_dir=/tmp/labels-unittest.bak")
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

func TestLabelsDownloaderSuccess(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200)

	err := db.GiveUserModeratorRights("testuser")
	ok(t, err)

	testMultipleDonate(t, "floor")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for _, imageId := range imageIds {
		testSuggestLabelForImage(t, imageId, "red apple", true, token)
	}
	runTrendingLabelsWorker(t)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 1)


	testAcceptTrendingLabel(t, "red apple", "", "red apples", "red apple", token, "normal", 201)
	trendingLabels = testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 1)

	equals(t, trendingLabels[0].Status, "accepted")

	runLabelBot(t, "cisuccess")
	trendingLabels = testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 1)
	equals(t, trendingLabels[0].Status, "merged")

	numberOfLabelsBefore, err := db.GetNumberOfLabels()
	ok(t, err)
	runLabelsDownloader(t)
	numberOfLabelsAfter, err := db.GetNumberOfLabels()
	ok(t, err)
	equals(t, numberOfLabelsBefore+1, numberOfLabelsAfter)
}

func TestLabelsDownloaderFailure(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200)

	err := db.GiveUserModeratorRights("testuser")
	ok(t, err)

	testMultipleDonate(t, "floor")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for _, imageId := range imageIds {
		testSuggestLabelForImage(t, imageId, "red apple", true, token)
	}
	runTrendingLabelsWorker(t)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 1)


	testAcceptTrendingLabel(t, "red apple", "", "red apples", "red apple", token, "normal", 201)
	trendingLabels = testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 1)

	equals(t, trendingLabels[0].Status, "accepted")

	runLabelBot(t, "cifailure")
	trendingLabels = testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 1)
	equals(t, trendingLabels[0].Status, "build-failed")

	numberOfLabelsBefore, err := db.GetNumberOfLabels()
	ok(t, err)
	runLabelsDownloader(t)
	numberOfLabelsAfter, err := db.GetNumberOfLabels()
	ok(t, err)
	equals(t, numberOfLabelsBefore, numberOfLabelsAfter)
}

func TestLabelsDownloaderFailureCanBeRetried(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200)

	err := db.GiveUserModeratorRights("testuser")
	ok(t, err)

	testMultipleDonate(t, "floor")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for _, imageId := range imageIds {
		testSuggestLabelForImage(t, imageId, "red apple", true, token)
	}
	runTrendingLabelsWorker(t)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 1)


	testAcceptTrendingLabel(t, "red apple", "", "red apples", "red apple", token, "normal", 201)
	trendingLabels = testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 1)

	equals(t, trendingLabels[0].Status, "accepted")

	runLabelBot(t, "cifailure")
	trendingLabels = testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 1)
	equals(t, trendingLabels[0].Status, "build-failed")

	numberOfLabelsBefore, err := db.GetNumberOfLabels()
	ok(t, err)
	runLabelsDownloader(t)
	numberOfLabelsAfter, err := db.GetNumberOfLabels()
	ok(t, err)
	equals(t, numberOfLabelsBefore, numberOfLabelsAfter)

	testAcceptTrendingLabel(t, "red apple", "", "red apples", "red apple", token, "normal", 201) 
	trendingLabels = testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 1)

	equals(t, trendingLabels[0].Status, "retry")
}

func TestLabelsDownloaderSuccessCannotBeRetried(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200)

	err := db.GiveUserModeratorRights("testuser")
	ok(t, err)

	testMultipleDonate(t, "floor")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for _, imageId := range imageIds {
		testSuggestLabelForImage(t, imageId, "red apple", true, token)
	}
	runTrendingLabelsWorker(t)

	trendingLabels := testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 1)


	testAcceptTrendingLabel(t, "red apple", "", "red apples", "red apple", token, "normal", 201)
	trendingLabels = testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 1)

	equals(t, trendingLabels[0].Status, "accepted")

	runLabelBot(t, "cisuccess")
	trendingLabels = testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 1)
	equals(t, trendingLabels[0].Status, "merged")

	numberOfLabelsBefore, err := db.GetNumberOfLabels()
	ok(t, err)
	runLabelsDownloader(t)
	numberOfLabelsAfter, err := db.GetNumberOfLabels()
	ok(t, err)
	equals(t, numberOfLabelsBefore+1, numberOfLabelsAfter)

	testAcceptTrendingLabel(t, "red apple", "", "red apples", "red apple", token, "normal", 201) 
	trendingLabels = testGetTrendingLabels(t, token, 200)
	equals(t, len(trendingLabels), 1)

	equals(t, trendingLabels[0].Status, "merged")
}

