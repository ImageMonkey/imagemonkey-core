package tests

import (
	"testing"
	"gopkg.in/resty.v1"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	"time"
	"os/exec"
	"net/url"
	"os"
)


func runDataProcessor(t *testing.T) {
	// Start a process:
	cmd := exec.Command("go", "run", "data_processor.go", "-singleshot", "true")
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = "../src"
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

func testBrowseAnnotation(t *testing.T, query string, requiredNumOfResults int, token string) []datastructures.AnnotationTask {
	var annotationTasks []datastructures.AnnotationTask

	u := BASE_URL + API_VERSION + "/validations/unannotated"
	req := resty.R().
			    SetQueryParams(map[string]string{
		          "query": url.QueryEscape(query),
		        }).
				SetResult(&annotationTasks)

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(u)

	ok(t, err)
    equals(t, resp.StatusCode(), 200)

    equals(t, len(annotationTasks), requiredNumOfResults)

	return annotationTasks
}


func testGetExistingAnnotations(t *testing.T, query string, token string, requiredStatusCode int, requiredNumOfResults int) []datastructures.AnnotatedImage {
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

	return annotatedImages
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

	testMultipleDonate(t, "apple")

	testGetExistingAnnotations(t, "apple", "", 200, 0)
}

func TestGetExistingAnnotations1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for i := 0; i < len(imageIds); i++ {
		//annotate image with label apple
		testAnnotate(t, imageIds[i], "apple", "", 
						`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	}

	testGetExistingAnnotations(t, "apple", "", 200, 13)
}

func TestGetExistingAnnotationsLockedAndAnnotatedByForeignUser(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, userToken, 201)

	testGetExistingAnnotations(t, "apple", "", 200, 0)
}

func TestGetExistingAnnotationsLockedAndAnnotatedByOwnUser(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, userToken, 201)

	testGetExistingAnnotations(t, "apple", userToken, 200, 1)
}

func TestGetImageAnnotations(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	testGetAnnotatedImage(t, imageId, "",  200)
}

func TestGetImageAnnotationsInvalidImageId(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	testGetAnnotatedImage(t, "this-is-an-invalid-image-id", "",  422)
}

func TestGetImageAnnotationsImageLockedForeignDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token, "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, token, 201)

	testGetAnnotatedImage(t, imageId, "",  422)
}

func TestGetImageAnnotationsImageLockedButOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token, "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, token, 201)

	testGetAnnotatedImage(t, imageId, token,  200)
}


func TestGetImageAnnotationsImageLockedOwnDonationButQuarantine(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token, "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, token, 201)

	err = db.PutImageInQuarantine(imageId)
	ok(t, err)

	testGetAnnotatedImage(t, imageId, token,  422)
}

func TestBrowseByCoverage(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":60,"left":145,"type":"rect","angle":0,"width":836,"height":660,"stroke":{"color":"red","width":5}}]`, "", 201)


	runDataProcessor(t)

	//expected coverage = annotation area / image area (i.e: (836*660)/(1132*750) = ~65%)
	coverage, err := db.GetImageAnnotationCoverageForImageId(imageId)
	ok(t, err)
	equals(t, coverage, 65)
}


func TestBrowseByCoverageFullyContained(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	//in case there is another rect that is fully contained within the bigger rect, the coverage should still be the same
	testAnnotate(t, imageId, "apple", "", 
					`[{"top":60,"left":145,"type":"rect","angle":0,"width":836,"height":660,"stroke":{"color":"red","width":5}},
					  {"top":67,"left":150,"type":"rect","angle":0,"width":500,"height":500,"stroke":{"color":"red","width":5}}]`, "", 201)


	runDataProcessor(t)

	//expected coverage = annotation area / image area (i.e: (836*660)/(1132*750) = ~65%)
	coverage, err := db.GetImageAnnotationCoverageForImageId(imageId)
	ok(t, err)
	equals(t, coverage, 65)
}

func TestBrowseAnnotationQuery(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	//give first image the labels cat and dog
	testLabelImage(t, imageIds[0], "dog", "", "", 200)
	testLabelImage(t, imageIds[0], "cat", "", "", 200)

	//add label 'cat' to second image
	testLabelImage(t, imageIds[1], "cat", "", "", 200)

	testBrowseAnnotation(t, "cat&dog", 2, "")
	testBrowseAnnotation(t, "cat|dog", 3, "")
	testBrowseAnnotation(t, "cat|cat", 2, "")

	//annotate image with label dog
	testAnnotate(t, imageIds[0], "dog", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	//dog is already annotated, so cat&dog should return 0
	testBrowseAnnotation(t, "cat&dog", 0, "")
	testBrowseAnnotation(t, "cat", 2, "")

	//but dog|cat should return 2
	testBrowseAnnotation(t, "cat|dog", 2, "")

	//annotate image with label cat
	testAnnotate(t, imageIds[0], "cat", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	//dog&cat should still return 0
	testBrowseAnnotation(t, "cat&dog", 0, "")
	testBrowseAnnotation(t, "dog", 0, "")

	//there is still one cat left
	testBrowseAnnotation(t, "cat", 1, "")

}

func TestBrowseAnnotationQueryNonProductiveLabelsShouldFailDueToNotAuthenticated(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	//give first image the label 'newlabel'
	testLabelImage(t, imageIds[0], "newlabel", "", "", 401)

	testBrowseAnnotation(t, "newlabel", 0, "")
}

func TestBrowseAnnotationQueryNonProductiveLabelsNoResultsAsNotAuthenticated(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	//give first image the label 'newlabel'
	testSuggestLabelForImage(t, imageIds[0], "newlabel", true, token, 200)
	
	testBrowseAnnotation(t, "newlabel", 0, "")
}

func TestBrowseAnnotationQueryNonProductiveLabelsAnnotationFailsDueToNotAuthenticated(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	//give first image the label 'newlabel'
	testSuggestLabelForImage(t, imageIds[0], "newlabel", true, token, 200)
	
	testBrowseAnnotation(t, "newlabel", 1, token)

	testAnnotate(t, imageIds[0], "newlabel", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 401)

	testBrowseAnnotation(t, "newlabel", 1, token)
}

func TestBrowseAnnotationQueryNonProductiveLabels(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	//give first image the label 'newlabel'
	testSuggestLabelForImage(t, imageIds[0], "newlabel", true, token, 200)
	
	testBrowseAnnotation(t, "newlabel", 1, token)

	testAnnotate(t, imageIds[0], "newlabel", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, token, 201)

	testBrowseAnnotation(t, "newlabel", 0, token)
}


func TestBrowseAnnotationQueryLockedButOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token, "", 200)
	testDonate(t, "./images/apples/apple2.jpeg", "apple", false, token, "", 200)

	testBrowseAnnotation(t, "apple", 2, token)
}

func TestBrowseAnnotationQueryLockedButForeignDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	token1 := testLogin(t, "user1", "pwd1", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token1, "", 200)
	testDonate(t, "./images/apples/apple2.jpeg", "apple", false, token, "", 200)

	testBrowseAnnotation(t, "apple", 1, token)
}

func TestBrowseAnnotationQueryLockedOwnDonationButQuarantine(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token, "", 200)
	testDonate(t, "./images/apples/apple2.jpeg", "apple", false, token, "", 200)

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	err = db.PutImageInQuarantine(imageIds[0])
	ok(t, err)

	err = db.PutImageInQuarantine(imageIds[1])
	ok(t, err)

	testBrowseAnnotation(t, "apple", 0, token)
}


func TestBrowseAnnotationQuery1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	num := testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	testBrowseAnnotation(t, "~tree", num, "")
	testBrowseAnnotation(t, "apple", num, "")

	testBrowseAnnotation(t, "~tree | apple", num, "")
	testBrowseAnnotation(t, "~tree & apple", num, "")
	testBrowseAnnotation(t, "~tree & car", 0, "")

	
	testAnnotate(t, imageIds[0], "apple", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	testBrowseAnnotation(t, "~tree", num-1, "")
	testBrowseAnnotation(t, "apple", num-1, "")	

}


func TestBrowseAnnotationQueryAnnotationCoverage(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "orange", "", "", 200)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":60,"left":145,"type":"rect","angle":0,"width":836,"height":660,"stroke":{"color":"red","width":5}}]`, "", 201)

	runDataProcessor(t)

	testBrowseAnnotation(t, "orange & annotation.coverage > 0%", 1, "")
}

func TestBrowseAnnotationQueryImageDimensions(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "orange", "", "", 200)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":60,"left":145,"type":"rect","angle":0,"width":836,"height":660,"stroke":{"color":"red","width":5}}]`, "", 201)

	runDataProcessor(t)

	testBrowseAnnotation(t, "orange & annotation.coverage > 0% & image.width > 100px & image.height > 100px", 1, "")
}

func TestGetExistingAnnotationsNotBelongingToImageCollection(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testAddImageCollection(t, "user", token, "mycollection", "", 201)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for i := 0; i < len(imageIds); i++ {
		//annotate image with label apple
		testAnnotate(t, imageIds[i], "apple", "", 
						`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	}

	testGetExistingAnnotations(t, "image.collection='mycollection'", token, 200, 0)
}

func TestGetExistingAnnotationsBelongingToImageCollection(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testAddImageCollection(t, "user", token, "mycollection", "", 201)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, token, "mycollection", 200)
	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for i := 0; i < len(imageIds); i++ {
		//annotate image with label apple
		testAnnotate(t, imageIds[i], "apple", "", 
						`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	}

	testGetExistingAnnotations(t, "image.collection='mycollection'", token, 200, 1)
}

func TestGetExistingAnnotationsBelongingToOtherImageCollection(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user99", "pwd99", "user99@imagemonkey.io")
	token2 := testLogin(t, "user99", "pwd99", 200)

	testAddImageCollection(t, "user", token, "mycollection", "", 201)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, token, "mycollection", 200)
	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for i := 0; i < len(imageIds); i++ {
		//annotate image with label apple
		testAnnotate(t, imageIds[i], "apple", "", 
						`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	}

	testGetExistingAnnotations(t, "image.collection='mycollection'", token2, 200, 0)
}

func TestBrowseAnnotationNotBelongingToImageCollection(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token, "", 200)
	testDonate(t, "./images/apples/apple2.jpeg", "apple", false, token, "", 200)

	testBrowseAnnotation(t, "image.collection='test'", 0, token)
}

func TestBrowseAnnotationBelongingToImageCollection(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testAddImageCollection(t, "user", token, "mycollection", "", 201)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token, "mycollection", 200)
	testDonate(t, "./images/apples/apple2.jpeg", "apple", false, token, "mycollection", 200)

	testBrowseAnnotation(t, "image.collection='mycollection'", 2, token)
}

func TestBrowseAnnotationBelongingToImageCollection2(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testAddImageCollection(t, "user", token, "mycollection", "", 201)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token, "mycollection", 200)
	
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "orange", "", "", 200)

	testBrowseAnnotation(t, "image.collection='mycollection'", 2, token)
}

func TestBrowseAnnotationBelongingToImageCollection3(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testAddImageCollection(t, "user", token, "mycollection", "", 201)

	testDonate(t, "./images/apples/apple2.jpeg", "apple", false, token, "", 200)
	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token, "mycollection", 200)	

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "orange", "", "", 200)

	testBrowseAnnotation(t, "image.collection='mycollection'", 2, token)
}

func TestBrowseAnnotationBelongingToOtherImageCollection(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user2", "pwd2", "user2@imagemonkey.io")
	token2 := testLogin(t, "user2", "pwd2", 200)

	testAddImageCollection(t, "user", token, "mycollection", "", 201)

	testDonate(t, "./images/apples/apple2.jpeg", "apple", false, token, "", 200)
	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token, "mycollection", 200)	

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "orange", "", "", 200)

	testBrowseAnnotation(t, "image.collection='mycollection'", 0, token2)
}

func TestBrowseAnnotationQueryNoLabelsImage(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple2.jpeg", "", true, token, "", 200)

	testBrowseAnnotation(t, "image.unlabeled='true'", 1, "")
}

func TestBrowseAnnotationQueryNoLabelsImage1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple2.jpeg", "", true, token, "", 200)
	testDonate(t, "./images/apples/apple3.jpeg", "apple", true, token, "", 200)

	testBrowseAnnotation(t, "apple & image.unlabeled='true'", 0, "")
}

func TestBrowseAnnotationQueryNoLabelsImage2(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple2.jpeg", "", true, token, "", 200)
	testDonate(t, "./images/apples/apple3.jpeg", "apple", true, token, "", 200)

	testBrowseAnnotation(t, "apple | image.unlabeled='true'", 2, "")
}

func TestBrowseImageNumAnnotations(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	testBrowseAnnotation(t, "image.num_annotations < 2", 1, "")
	testBrowseAnnotation(t, "image.num_annotations = 1", 1, "")
	testBrowseAnnotation(t, "image.num_annotations > 1", 0, "")
	testBrowseAnnotation(t, "apple & image.num_annotations = 1", 1, "")
	testBrowseAnnotation(t, "egg & image.num_annotations = 1", 0, "")
	testBrowseAnnotation(t, "apple | image.num_annotations = 1", 1, "")
}


func TestBrowseImageNumAnnotations1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "",
						`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	testBrowseAnnotation(t, "image.num_annotations < 2", 0, "")
	testBrowseAnnotation(t, "image.num_annotations = 1", 0, "")
	testBrowseAnnotation(t, "image.num_annotations > 1", 0, "")
	testBrowseAnnotation(t, "apple & image.num_annotations = 1", 0, "")
	testBrowseAnnotation(t, "egg & image.num_annotations = 1", 0, "")
	testBrowseAnnotation(t, "apple | image.num_annotations = 1", 0, "")
}


