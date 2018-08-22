package tests

import (
	"testing"
	"gopkg.in/resty.v1"
	"../src/datastructures"
	"time"
	"os/exec"
)


func runDataProcessor(t *testing.T) {
	// Start a process:
	cmd := exec.Command("go", "run", "data_processor.go", "api_secrets.go", "shared_secrets.go", "-singleshot", "true")
	cmd.Dir = "../src"
	err := cmd.Start()
	ok(t, err)

	// Wait for the process to finish or kill it after a timeout:
	done := make(chan error, 1)
	go func() {
	    done <- cmd.Wait()
	}()
	select {
	case <-time.After(5 * time.Second):
	    err := cmd.Process.Kill()
	    ok(t, err) //failed to kill process
	    t.Errorf("process killed as timeout reached")
	case err := <-done:
	    ok(t, err)
	}
}


func testGetExistingAnnotations(t *testing.T, query string, token string, requiredStatusCode int, requiredNumOfResults int) {
	url := BASE_URL +API_VERSION + "/annotations"

	var annotatedImages []datastructures.AnnotatedImage

	req := resty.R().
			SetQueryParams(map[string]string{
				"query": query,
		   }).
		   SetResult(&annotatedImages)
	
	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(url)
	
	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
	equals(t, len(annotatedImages), requiredNumOfResults)
}

func testGetAnnotatedImage(t *testing.T, imageId string, token string, requiredStatusCode int) {
	url := BASE_URL +API_VERSION + "/donation/" + imageId + "/annotations"

	var annotatedImages []datastructures.AnnotatedImage

	req := resty.R().
			SetQueryParams(map[string]string{
				"image_id": imageId,
		    }).
		    SetResult(&annotatedImages)
	
	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(url)
	
	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
} 


func TestGetExistingAnnotations(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t)

	testGetExistingAnnotations(t, "apple", "", 200, 0)
}

func TestGetExistingAnnotations1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t)

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for i := 0; i < len(imageIds); i++ {
		//annotate image with label apple
		testAnnotate(t, imageIds[i], "apple", "", 
						`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "")

	}

	testGetExistingAnnotations(t, "apple", "", 200, 13)
}

func TestGetExistingAnnotationsLockedAndAnnotatedByForeignUser(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, userToken)

	testGetExistingAnnotations(t, "apple", "", 200, 0)
}

func TestGetExistingAnnotationsLockedAndAnnotatedByOwnUser(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, userToken)

	testGetExistingAnnotations(t, "apple", userToken, 200, 1)
}

func TestGetImageAnnotations(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "")

	testGetAnnotatedImage(t, imageId, "",  200)
}

func TestGetImageAnnotationsInvalidImageId(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "")

	testGetAnnotatedImage(t, "this-is-an-invalid-image-id", "",  422)
}

func TestGetImageAnnotationsImageLockedForeignDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, token)

	testGetAnnotatedImage(t, imageId, "",  422)
}

func TestGetImageAnnotationsImageLockedButOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, token)

	testGetAnnotatedImage(t, imageId, token,  200)
}


func TestGetImageAnnotationsImageLockedOwnDonationButQuarantine(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, token)

	err = db.PutImageInQuarantine(imageId)
	ok(t, err)

	testGetAnnotatedImage(t, imageId, token,  422)
}

func TestBrowseByCoverage(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":60,"left":145,"type":"rect","angle":0,"width":836,"height":660,"stroke":{"color":"red","width":5}}]`, "")


	runDataProcessor(t)

	//expected coverage = annotation area / image area (i.e: (836*660)/(1132*750) = ~65%)
	coverage, err := db.GetImageAnnotationCoverageForImageId(imageId)
	ok(t, err)
	equals(t, coverage, 65)
}


func TestBrowseByCoverageFullyContained(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	//in case there is another rect that is fully contained within the bigger rect, the coverage should still be the same
	testAnnotate(t, imageId, "apple", "", 
					`[{"top":60,"left":145,"type":"rect","angle":0,"width":836,"height":660,"stroke":{"color":"red","width":5}},
					  {"top":67,"left":150,"type":"rect","angle":0,"width":500,"height":500,"stroke":{"color":"red","width":5}}]`, "")


	runDataProcessor(t)

	//expected coverage = annotation area / image area (i.e: (836*660)/(1132*750) = ~65%)
	coverage, err := db.GetImageAnnotationCoverageForImageId(imageId)
	ok(t, err)
	equals(t, coverage, 65)
}

