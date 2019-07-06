package tests

import (
	"testing"
	"gopkg.in/resty.v1"
	"encoding/json"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
)

func testImageAnnotationRefinement(t *testing.T, annotationId string, annotationDataId string, labelUuid string, expectedStatusResponse int) {
	type AnnotationRefinementEntry struct {
    	LabelUuid string `json:"label_uuid"`
	}
	var annotationRefinementEntries []AnnotationRefinementEntry

	annotationRefinementEntry := AnnotationRefinementEntry{LabelUuid:labelUuid}
	annotationRefinementEntries = append(annotationRefinementEntries, annotationRefinementEntry)

	url := BASE_URL + API_VERSION + "/annotation/" + annotationId + "/refine/" + annotationDataId
	resp, err := resty.R().
				SetBody(annotationRefinementEntries).
				Post(url)

	ok(t, err)
	equals(t, resp.StatusCode(), expectedStatusResponse)
}

func testRandomAnnotationRefinement(t *testing.T, num int) {
	for i := 0; i < num; i++ {
		annotationId, annotationDataId, err := db.GetRandomAnnotationData()
		ok(t, err)

		labelUuid, err := db.GetRandomLabelUuid()
		ok(t, err)

		testImageAnnotationRefinement(t, annotationId, annotationDataId, labelUuid, 201)
	}
}

func testBrowseRefinement(t *testing.T, query string, annotationDataId string, 
	requiredStatusCode int, requiredNumOfResults int) []datastructures.AnnotationRefinementTask {
	url := BASE_URL + API_VERSION + "/refine"

	var err error
	var resp *resty.Response
	var refinementEntries []datastructures.AnnotationRefinementTask

	req := resty.R()

	if query != "" {
		req.SetQueryParams(map[string]string{
			"query": query,
		}).
		SetResult(&refinementEntries)

		resp, err = req.Get(url)
	}else if annotationDataId != "" {
		var refinementEntry datastructures.AnnotationRefinementTask
		req.SetQueryParams(map[string]string{
			"annotation_data_id": annotationDataId,
		}).SetResult(&refinementEntry)

		resp, err = req.Get(url)

		refinementEntries = append(refinementEntries, refinementEntry)
	} else {
		resp, err = req.Get(url)
	}

	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
	equals(t, len(refinementEntries), requiredNumOfResults)

	return refinementEntries
}


func TestRandomImageAnnotationRefinement(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t, "apple")
	testRandomAnnotate(t, 5, `[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`)
	testRandomAnnotationRefinement(t, 4)
}


func TestGetRandomImageQuiz(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "dog", true, "", "", 200)
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "dog", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)
	
	annotationIds, err := db.GetAllAnnotationIds()
	ok(t, err)

	err = db.SetAnnotationValid(annotationIds[0], 5)
	ok(t, err)

	testGetRandomImageQuiz(t, 200)
}

func TestGetRandomImageQuizImageStillLocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "dog", false, userToken, "", 200)
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "dog", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, userToken, 201)
	
	annotationIds, err := db.GetAllAnnotationIds()
	ok(t, err)

	err = db.SetAnnotationValid(annotationIds[0], 5)
	ok(t, err)

	testGetRandomImageQuiz(t, 422)
}

func TestBrowseRefinement(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "dog", true, "", "", 200)
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "dog", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	testBrowseRefinement(t, "dog", "", 200, 1)
}

func TestBrowseRefinementNoResult(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "dog", true, "", "", 200)
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "dog", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	testBrowseRefinement(t, "apple", "", 200, 0)
}


func TestBrowseRefinementInvalidRequest(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	//donate image with some label + annotate
	testDonate(t, "./images/apples/apple1.jpeg", "dog", true, "", "", 200)
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "dog", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	testBrowseRefinement(t, "", "", 422, 0)
}

func TestBrowseRefinementByAnnotationDataId(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	//donate image with some label + annotate
	testDonate(t, "./images/apples/apple1.jpeg", "dog", true, "", "", 200)
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "dog", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	annotationDataId, err := db.GetLastAddedAnnotationDataId()
	ok(t, err)

	refinementEntries := testBrowseRefinement(t, "", annotationDataId, 200, 1)
	refinementEntry := refinementEntries[0]

	var f map[string]interface{}
    err = json.Unmarshal([]byte(refinementEntry.Annotation.Data), &f)
    ok(t, err)
	equals(t, f["uuid"], annotationDataId)
}

func TestBrowseRefinementByInvalidAnnotationDataId(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	//donate image with some label + annotate
	testDonate(t, "./images/apples/apple1.jpeg", "dog", true, "", "", 200)
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "dog", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	testBrowseRefinement(t, "", "invalid-annotation-data-id", 422, 1)
}


func TestBrowseRefinementByCategory(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	//donate image with some label + annotate
	testDonate(t, "./images/apples/apple1.jpeg", "person", true, "", "", 200)
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "person", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)


	testBrowseRefinement(t, "person & ~gender", "", 200, 1)

	annotationId, annotationDataId, err := db.GetLastAddedAnnotationData()
	ok(t, err)

	labelUuid := "1eaa891f-9e5c-448d-ac90-78d5a4a189e9" //this is the uuid of the label "male" (see label-refinements.json)

	testImageAnnotationRefinement(t, annotationId, annotationDataId, labelUuid, 201)

	testBrowseRefinement(t, "person & ~gender", "", 200, 0)
	testBrowseRefinement(t, "person & gender='male'", "", 200, 1)
}

func TestImageAnnotationRefinementShouldFailWhenLabelIdIsInvalid(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	//donate image with some label + annotate
	testDonate(t, "./images/apples/apple1.jpeg", "person", true, "", "", 200)
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "person", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)


	testBrowseRefinement(t, "person & ~gender", "", 200, 1)

	annotationId, annotationDataId, err := db.GetLastAddedAnnotationData()
	ok(t, err)

	testImageAnnotationRefinement(t, annotationId, annotationDataId, "not-a-valid-label-uuid", 400)
}

func TestAnnotateAndRefine(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	//donate image with some label + annotate
	testDonate(t, "./images/apples/apple1.jpeg", "person", true, "", "", 200)
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "person", "", 
							`[{"refinements": [{"label_uuid": "86485bae-04a1-43ef-a191-5f2a0464595a"}],"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	annotationIds, err := db.GetAllAnnotationIds()
	ok(t, err)
	equals(t, len(annotationIds), 1)

	numOfRefinements, err := db.GetNumOfRefinements()
	ok(t, err)
	equals(t, numOfRefinements, 1)
}

func TestAnnotateAndRefineShouldFailDueToWrongSyntax(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	//donate image with some label + annotate
	testDonate(t, "./images/apples/apple1.jpeg", "person", true, "", "", 200)
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "person", "", 
							`[{"refinements": [{"invalid-key": "86485bae-04a1-43ef-a191-5f2a0464595a"}],"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 422)

	annotationIds, err := db.GetAllAnnotationIds()
	ok(t, err)
	equals(t, len(annotationIds), 0)

	numOfRefinements, err := db.GetNumOfRefinements()
	ok(t, err)
	equals(t, numOfRefinements, 0)
}

func TestAnnotateAndRefineMultiple(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	//donate image with some label + annotate
	testDonate(t, "./images/apples/apple1.jpeg", "person", true, "", "", 200)
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "person", "", 
							`[{"refinements": [{"label_uuid": "86485bae-04a1-43ef-a191-5f2a0464595a", "label_uuid": "07d13f17-3757-45c5-ba20-4f53c8a46334"}],"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	annotationIds, err := db.GetAllAnnotationIds()
	ok(t, err)
	equals(t, len(annotationIds), 1)

	numOfRefinements, err := db.GetNumOfRefinements()
	ok(t, err)
	equals(t, numOfRefinements, 1)
}

func TestBrowseAnnotationsWithRefinements(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	//donate image with some label + annotate
	testDonate(t, "./images/apples/apple1.jpeg", "person", true, "", "", 200)
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	annotations := `[{"refinements": [{"label_uuid": "86485bae-04a1-43ef-a191-5f2a0464595a", "label_uuid": "07d13f17-3757-45c5-ba20-4f53c8a46334"}],"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`

	testAnnotate(t, imageId, "person", "", annotations, "", 201)

	annotationIds, err := db.GetAllAnnotationIds()
	ok(t, err)
	equals(t, len(annotationIds), 1)

	numOfRefinements, err := db.GetNumOfRefinements()
	ok(t, err)
	equals(t, numOfRefinements, 1)

	annotatedImage := testGetExistingAnnotations(t, "person", "", 200, 1)
	
	storedAnnotations, err := json.Marshal(&annotatedImage[0].Annotations)
    ok(t, err)

    eq, err := equalJson(string(storedAnnotations), annotations)
    ok(t, err)
    equals(t, eq, true)
}
