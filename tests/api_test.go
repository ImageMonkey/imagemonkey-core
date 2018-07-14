package tests

import (
	"testing"
	"encoding/json"
	"gopkg.in/resty.v1"
	//"errors"
	"io/ioutil"
	"reflect"
	"strconv"
	//"fmt"
)

type LoginResult struct {
	Token string `json:"token"`
}

type AnnotatedImage struct {
    Image struct {
        Id string `json:"uuid"`
        Provider string `json:"provider"`
        Width int32 `json:"width"`
        Height int32 `json:"height"`
    } `json:"image"`

    Validation struct {
        Label string `json:"label"`
        Sublabel string `json:"sublabel"`
    } `json:"validation"`
    

    Id string `json:"uuid"`
    NumOfValid int32 `json:"num_yes"`
    NumOfInvalid int32 `json:"num_no"`
    Annotations []json.RawMessage `json:"annotations"`
    NumRevisions int32 `json:"num_revisions"`
    Revision int32 `json:"revision"`
}

type ValidateResult struct {
    Id string `json:"uuid"`
    Url string `json:"url"`
    Label string `json:"label"`
    Provider string `json:"provider"`
    Probability float32 `json:"probability"`
    NumOfValid int32 `json:"num_yes"`
    NumOfInvalid int32 `json:"num_no"`
}

type AnnotationRow struct {
	Image struct {
		Id string `json:"uuid"`
	} `json:"image"`

	Validation struct {
		Id string `json:"uuid"`
		Label string `json:"label"`
		Sublabel string `json:"sublabel"`
	} `json:"validation"`
}

func testSignUp(t *testing.T, username string, password string, email string) {
	numBefore, err := db.GetNumberOfUsers()
	ok(t, err)


	url := BASE_URL + API_VERSION + "/signup"
	resp, err := resty.R().
    	SetHeader("Content-Type", "application/json").
     	SetBody(map[string]interface{}{"username": username, "password": password, "email": email}).
     	Post(url)

    equals(t, resp.StatusCode(), 201)

    numAfter, err := db.GetNumberOfUsers(); 
    ok(t, err)

	equals(t, numAfter, (numBefore + 1))
}

func testLogin(t *testing.T, username string, password string, requiredStatusCode int) string {
	url := BASE_URL + API_VERSION + "/login"
	resp, err := resty.R().
     	SetBasicAuth(username, password).
     	SetResult(&LoginResult{}).
     	Post(url)

    ok(t, err)
    equals(t, resp.StatusCode(), requiredStatusCode)

    return resp.Result().(*LoginResult).Token
}

func testAnnotate(t *testing.T, imageId string, label string, sublabel string, annotations string) {
	type Annotation struct {
		Annotations []json.RawMessage `json:"annotations"`
		Label string `json:"label"`
		Sublabel string `json:"sublabel"`
	}

	annotationEntry := Annotation{Label: label, Sublabel: sublabel}

	err := json.Unmarshal([]byte(annotations), &annotationEntry.Annotations)
	ok(t, err)

	url := BASE_URL + API_VERSION + "/annotate/" + imageId

	resp, err := resty.R().
					SetHeader("Content-Type", "application/json").
					SetBody(annotationEntry).
					Post(url)
	ok(t, err)

	equals(t, resp.StatusCode(), 201)

	//export annotations again
	url = resp.Header().Get("Location")
	resp, err = resty.R().
					SetHeader("Content-Type", "application/json").
					SetResult(&AnnotatedImage{}).
					Get(url)
	ok(t, err)

	equals(t, resp.StatusCode(), 200)

	j, err := json.Marshal(&resp.Result().(*AnnotatedImage).Annotations)
	ok(t, err)

	equal, err := equalJson(string(j), annotations)
	equals(t, equal, true)

	ok(t, err)
}


func testRandomAnnotate(t *testing.T, num int, annotations string) {
	for i := 0; i < num; i++ {
		annotationRow, err := db.GetRandomImageForAnnotation()
		ok(t, err)

		testAnnotate(t, annotationRow.Image.Id, annotationRow.Validation.Label, 
						annotationRow.Validation.Sublabel, annotations)
	}
}

func testDonate(t *testing.T, path string, label string) {
	numBefore, err := db.GetNumberOfImages()
	ok(t, err)

	url := BASE_URL + API_VERSION + "/donate"
	resp, err := resty.R().
      SetFile("image", path).
      SetFormData(map[string]string{
        "label": label,
      }).
      Post(url)

    equals(t, resp.StatusCode(), 200)

    numAfter, err := db.GetNumberOfImages(); 
    ok(t, err)

    equals(t, numAfter, numBefore + 1)

	//after image donation, unlock all images
    err = db.UnlockAllImages()
    ok(t, err)
}

func testRandomAnnotationRework(t *testing.T, num int, annotations string) {
	type Annotation struct {
		Annotations []json.RawMessage `json:"annotations"`
	}

	for i := 0; i < num; i++ {
		annotationId, err := db.GetRandomAnnotationId()
		ok(t, err)

		oldAnnotationRevision, err := db.GetAnnotationRevision(annotationId)
		ok(t, err)

		oldAnnotationDataIds, err := db.GetAnnotationDataIds(annotationId)
		ok(t, err)

		annotationEntry := Annotation{}

		err = json.Unmarshal([]byte(annotations), &annotationEntry.Annotations)
		ok(t, err)

		url := BASE_URL + API_VERSION + "/annotation/" + annotationId
		resp, err := resty.R().
					SetHeader("Content-Type", "application/json").
					SetBody(annotationEntry).
					Put(url)
		ok(t, err)

		equals(t, resp.StatusCode(), 201)

		newAnnotationRevision, err := db.GetAnnotationRevision(annotationId)
		ok(t, err)

		newAnnotationDataIds, err := db.GetOldAnnotationDataIds(annotationId, oldAnnotationRevision)
		ok(t, err)

		equals(t, newAnnotationRevision, (oldAnnotationRevision + 1))

		equal := reflect.DeepEqual(oldAnnotationDataIds, newAnnotationDataIds)
		equals(t, equal, true)
	}
}

func testMultipleDonate(t *testing.T) {
	dirname := "./images/apples/"
	files, err := ioutil.ReadDir(dirname)
    ok(t, err)

    for _, f := range files {
        testDonate(t, dirname + f.Name(), "apple")
    }
}

func testLabelImage(t *testing.T, imageId string, label string) {
	type LabelMeEntry struct {
		Label string `json:"label"`
	}

	oldNum, err := db.GetNumberOfImagesWithLabel(label)
	ok(t, err)

	var labelMeEntries []LabelMeEntry
	labelMeEntry := LabelMeEntry{Label: label}
	labelMeEntries = append(labelMeEntries, labelMeEntry)

	url := BASE_URL + API_VERSION + "/donation/" + imageId + "/labelme"
	resp, err := resty.R().
			SetHeader("Content-Type", "application/json").
			SetBody(labelMeEntries).
			Post(url)

	ok(t, err)
	equals(t, resp.StatusCode(), 200)

	newNum, err := db.GetNumberOfImagesWithLabel(label)
	ok(t, err)

	equals(t, oldNum+1, newNum)
}


func testValidate(t *testing.T) {
	url := BASE_URL + API_VERSION + "/validate"
	_, err := resty.R().
			SetResult(&ValidateResult{}).
			Get(url)
	ok(t, err)
}

func testImageValidation(t *testing.T, uuid string, param string, moderated bool, token string) {
	url := BASE_URL + API_VERSION + "/validation/" + uuid + "/validate/" + param

	var resp *resty.Response
	var err error
	if moderated {
		resp, err = resty.R().
				SetHeader("X-Moderation", "true").
				SetAuthToken(token).
				Post(url)
	} else {
		resp, err = resty.R().
				Post(url)
	}

	equals(t, resp.StatusCode(), 200)
	ok(t, err)
}

func testRandomImageValidation(t *testing.T, num int) {
	for i := 0; i < num; i++ {
		param := ""
		randomBool := randomBool()
		if randomBool {
			param = "yes"
		} else {
			param = "no"
		}

		randomValidationId, err := db.GetRandomValidationId()
		ok(t, err)

		beforeChangeNumValid, beforeChangeNumInvalid, err := db.GetValidationCount(randomValidationId)
		ok(t, err)

		testImageValidation(t, randomValidationId, param, false, "")

		afterChangeNumValid, afterChangeNumInvalid, err := db.GetValidationCount(randomValidationId)
		ok(t, err)

		if param == "yes" {
			equals(t, afterChangeNumValid, (beforeChangeNumValid + 1))
		} else {
			equals(t, afterChangeNumInvalid, (beforeChangeNumInvalid + 1))
		}
	}
}

func testRandomModeratedImageValidation(t *testing.T, num int, token string) {
	for i := 0; i < num; i++ {
		randomValidationId, err := db.GetRandomValidationId()
		ok(t, err)

		_, beforeChangeNumInvalid, err := db.GetValidationCount(randomValidationId)
		ok(t, err)

		testImageValidation(t, randomValidationId, "no", true, token)

		_, afterChangeNumInvalid, err := db.GetValidationCount(randomValidationId)
		ok(t, err)


		equals(t, afterChangeNumInvalid, (beforeChangeNumInvalid + 5))
	}
}

func testImageAnnotationRefinement(t *testing.T, annotationId string, annotationDataId int64, labelId int64) {
	type AnnotationRefinementEntry struct {
    	LabelId int64 `json:"label_id"`
	}
	var annotationRefinementEntries []AnnotationRefinementEntry

	annotationRefinementEntry := AnnotationRefinementEntry{LabelId:labelId}
	annotationRefinementEntries = append(annotationRefinementEntries, annotationRefinementEntry)

	url := BASE_URL + API_VERSION + "/annotation/" + annotationId + "/refine/" + strconv.Itoa(int(annotationDataId))
	resp, err := resty.R().
				SetBody(annotationRefinementEntries).
				Post(url)

	ok(t, err)
	equals(t, resp.StatusCode(), 201)
}

func testRandomAnnotationRefinement(t *testing.T, num int) {
	for i := 0; i < num; i++ {
		annotationId, annotationDataId, err := db.GetRandomAnnotationData()
		ok(t, err)

		labelId, err := db.GetRandomLabelId()
		ok(t, err)

		testImageAnnotationRefinement(t, annotationId, annotationDataId, labelId)
	}
}

func testRandomLabel(t *testing.T, num int) {
	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	if num > len(imageIds) {
		t.Errorf("num can't be greater than the number of available images!")
	}


	for i := 0; i < num; i++ {
		image := imageIds[i]
		label, err := db.GetRandomLabelName()
		ok(t, err)

		testLabelImage(t, image, label)
	}
}

func testBrowseAnnotation(t *testing.T, query string, requiredNumOfResults int) {
	type AnnotationTask struct {
	    Image struct {
	        Id string `json:"uuid"`
	        Width int32 `json:"width"`
	        Height int32 `json:"height"`
	    } `json:"image"`

	    Id string `json:"uuid"`
	}

	var annotationTasks []AnnotationTask

	url := BASE_URL + API_VERSION + "/validations/unannotated"
	resp, err := resty.R().
			    SetQueryParams(map[string]string{
		          "query": query,
		        }).
				SetResult(&annotationTasks).
				Get(url)

	ok(t, err)
    equals(t, resp.StatusCode(), 200)

    //fmt.Printf("b=%d\n",len(annotationTasks))
    //fmt.Printf("a=%d\n",requiredNumOfResults)
    equals(t, len(annotationTasks), requiredNumOfResults)
}



func TestMultipleDonate(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t)
}

func TestRandomAnnotate(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t)
	testRandomAnnotate(t, 2, `[{"top":100,"left":200,"type":"rect","angle":0,"width":40,"height":60,"stroke":{"color":"red","width":1}}]`)
}

func TestSignUp(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
}

func TestLogin(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	testLogin(t, "testuser", "testpassword", 200)
}

func TestLoginShouldFailDueToWrongPassword(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	testLogin(t, "testuser", "wrongpassword", 401)
}

func TestLoginShouldFailDueToWrongUsername(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	testLogin(t, "wronguser", "testpassword", 401)
}

func TestRandomImageValidation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t)
	testRandomImageValidation(t, 100)
}

func TestRandomAnnotationRework(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t)
	testRandomAnnotate(t, 2, `[{"top":100,"left":200,"type":"rect","angle":0,"width":40,"height":60,"stroke":{"color":"red","width":1}}]`)
	testRandomAnnotationRework(t, 2, `[{"top":200,"left":300,"type":"rect","angle":10,"width":50,"height":30,"stroke":{"color":"blue","width":3}}]`)
}


func TestRandomModeratedImageValidation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t)
	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)
	db.GiveUserModeratorRights("moderator") //give user moderator rights
	testRandomModeratedImageValidation(t, 100, moderatorToken)
}

func TestRandomImageAnnotationRefinement(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t)
	testRandomAnnotate(t, 5, `[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`)
	testRandomAnnotationRefinement(t, 4)
}


func TestRandomLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t)
	testRandomLabel(t, 7)
}

func TestBrowseAnnotationQuery(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t)

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	//give first image the labels cat and dog
	testLabelImage(t, imageIds[0], "dog")
	testLabelImage(t, imageIds[0], "cat")

	//add label 'cat' to second image
	testLabelImage(t, imageIds[1], "cat")

	testBrowseAnnotation(t, "cat&dog", 2)
	testBrowseAnnotation(t, "cat|dog", 3)
	testBrowseAnnotation(t, "cat|cat", 2)

	//annotate image with label dog
	testAnnotate(t, imageIds[0], "dog", "", `[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`)

	//now we expect just one result 
	testBrowseAnnotation(t, "cat&dog", 1)
	testBrowseAnnotation(t, "cat", 2)

	//annotate image with label cat
	testAnnotate(t, imageIds[0], "cat", "", `[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`)

	//now we should get no result
	testBrowseAnnotation(t, "cat&dog", 0)
	testBrowseAnnotation(t, "dog", 0)

	//there is still one cat left
	testBrowseAnnotation(t, "cat", 1)

}