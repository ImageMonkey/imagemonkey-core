package tests

import (
	"testing"
	"encoding/json"
	"gopkg.in/resty.v1"
	"io/ioutil"
	"reflect"
	"os"
	"../src/datastructures"
)

const UNVERIFIED_DONATIONS_DIR string = "../unverified_donations/"
const DONATIONS_DIR string = "../donations/"

type LoginResult struct {
	Token string `json:"token"`
}

/*type ImageLabel struct {
    Image struct {
        Id string `json:"uuid"`
        Unlocked bool `json:"unlocked"`
        Url string `json:"url"`
        Provider string `json:"provider"`
        Width int32 `json:"width"`
        Height int32 `json:"height"`
    } `json:"image"`

    Labels[] struct {
        Name string `json:"name"`
        Unlocked bool `json:"unlocked"`
        Sublabels[] struct {
            Name string `json:"name"`
        } `json:"sublabels"`
    } `json:"labels"`
}

type AnnotatedImage struct {
    Image struct {
        Id string `json:"uuid"`
        Unlocked bool `json:"unlocked"`
        Url string `json:"url"`
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
}*/

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

func testAnnotate(t *testing.T, imageId string, label string, sublabel string, annotations string, token string) {
	type Annotation struct {
		Annotations []json.RawMessage `json:"annotations"`
		Label string `json:"label"`
		Sublabel string `json:"sublabel"`
	}

	annotationEntry := Annotation{Label: label, Sublabel: sublabel}

	err := json.Unmarshal([]byte(annotations), &annotationEntry.Annotations)
	ok(t, err)

	url := BASE_URL + API_VERSION + "/annotate/" + imageId

	req := resty.R().
			SetHeader("Content-Type", "application/json").
			SetBody(annotationEntry)

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Post(url)

	ok(t, err)

	equals(t, resp.StatusCode(), 201)

	//export annotations again
	url = resp.Header().Get("Location")
	req = resty.R().
					SetHeader("Content-Type", "application/json").
					SetResult(&datastructures.AnnotatedImage{})
					
	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err = req.Get(url)
	ok(t, err)

	equals(t, resp.StatusCode(), 200)

	j, err := json.Marshal(&resp.Result().(*datastructures.AnnotatedImage).Annotations)
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
						annotationRow.Validation.Sublabel, annotations, "")
	}
}

func testDonate(t *testing.T, path string, label string, unlockImage bool, token string) {
	numBefore, err := db.GetNumberOfImages()
	ok(t, err)

	url := BASE_URL + API_VERSION + "/donate"

	req := resty.R()

	if label == "" {
		req.
	      SetFile("image", path)
	} else {
		req.
	      SetFile("image", path).
	      SetFormData(map[string]string{
	        "label": label,
	      })
	}

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Post(url)

    equals(t, resp.StatusCode(), 200)

    numAfter, err := db.GetNumberOfImages(); 
    ok(t, err)

    equals(t, numAfter, numBefore + 1)

    if(unlockImage) {
		//after image donation, unlock all images
	    err = db.UnlockAllImages()
	    ok(t, err)

	    imageId, err := db.GetLatestDonatedImageId()
	    ok(t, err)

	    err = os.Rename(UNVERIFIED_DONATIONS_DIR + imageId, DONATIONS_DIR + imageId)
	    ok(t, err)
	}
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

func testBrowseLabel(t *testing.T, query string, token string, requiredNumOfResults int, requiredStatusCode int) {
	url := BASE_URL + API_VERSION + "/donations/labels"
	var labeledImages []datastructures.ImageLabel

	req := resty.R().
			SetQueryParams(map[string]string{
				"query": query,
		    }).
		    SetResult(&labeledImages)

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(url)
	
	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
	equals(t, len(labeledImages), requiredNumOfResults)
} 

func testMultipleDonate(t *testing.T) int {
	dirname := "./images/apples/"
	files, err := ioutil.ReadDir(dirname)
    ok(t, err)

    num := 0
    for _, f := range files {
        testDonate(t, dirname + f.Name(), "apple", true, "")
        num += 1
    }

    return num
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

func testGetImageForAnnotation(t *testing.T, imageId string, token string, validationId string, requiredStatusCode int) {
	url := BASE_URL + API_VERSION + "/annotate"
	req := resty.R()

	if token != "" {
		req.SetAuthToken(token)
	}

	if imageId != "" {
		req.SetQueryParams(map[string]string{
		    "image_id": imageId,
		})
	}

	if validationId != "" {
		req.SetQueryParams(map[string]string{
		    "validation_id": validationId,
		})
	}

	resp, err := req.Get(url)

	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
}

func testImageAnnotationRefinement(t *testing.T, annotationId string, annotationDataId string, labelUuid string) {
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
	equals(t, resp.StatusCode(), 201)
}

func testRandomAnnotationRefinement(t *testing.T, num int) {
	for i := 0; i < num; i++ {
		annotationId, annotationDataId, err := db.GetRandomAnnotationData()
		ok(t, err)

		labelUuid, err := db.GetRandomLabelUuid()
		ok(t, err)

		testImageAnnotationRefinement(t, annotationId, annotationDataId, labelUuid)
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

func testBrowseAnnotation(t *testing.T, query string, requiredNumOfResults int, token string) {
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
	req := resty.R().
			    SetQueryParams(map[string]string{
		          "query": query,
		        }).
				SetResult(&annotationTasks)

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(url)

	ok(t, err)
    equals(t, resp.StatusCode(), 200)

    //fmt.Printf("b=%d\n",len(annotationTasks))
    //fmt.Printf("a=%d\n",requiredNumOfResults)
    equals(t, len(annotationTasks), requiredNumOfResults)
}

func testGetImageToLabel(t *testing.T, imageId string, token string, requiredStatusCode int) {
	url := BASE_URL + API_VERSION + "/labelme"

	req := resty.R()

	if token != "" {
		req.SetAuthToken(token)
	}

	if imageId != "" {
		req.SetQueryParams(map[string]string{
		      "image_id": imageId,
		    })
	}

	resp, err := req.
				Get(url)

	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
}

func testGetImageToValidate(t *testing.T, imageId string, token string, labelId string, requiredStatusCode int) {
	url := BASE_URL + API_VERSION + "/validation"

	req := resty.R()

	if token != "" {
		req.SetAuthToken(token)
	}

	if imageId != "" {
		req.SetQueryParams(map[string]string{
		      "image_id": imageId,
		    })
	}

	if labelId != "" {
		req.SetQueryParams(map[string]string{
		      "label_id": labelId,
		    })
	}

	resp, err := req.Get(url)

	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
}

func testGetImageDonation(t *testing.T, imageId string, imageUnlocked bool, token string, requiredStatusCode int) {
	url := "" 

	if imageUnlocked {
		url = BASE_URL + API_VERSION + "/donation/" + imageId
	} else {
		url = BASE_URL + API_VERSION + "/unverified-donation/" + imageId

		if token != "" {
			url += "?token=" + token
		}
	}

	req := resty.R()

	resp, err := req.Get(url)

	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
}

func testMarkValidationAsNotAnnotatable(t *testing.T, validationId string, num int) {
	for i := 0; i < num; i++ {
		url := BASE_URL + API_VERSION + "/validation/" + validationId + "/not-annotatable"

		resp, err := resty.R().
						Post(url)

		ok(t, err)
		equals(t, resp.StatusCode(), 200)
	}
}

func testBlacklistAnnotation(t *testing.T, validationId string, token string) {
	url := BASE_URL + API_VERSION + "/validation/" + validationId + "/blacklist-annotation"

	resp, err := resty.R().
					SetAuthToken(token).
					Post(url)

	ok(t, err)
	equals(t, resp.StatusCode(), 200)
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

	testBrowseAnnotation(t, "cat&dog", 2, "")
	testBrowseAnnotation(t, "cat|dog", 3, "")
	testBrowseAnnotation(t, "cat|cat", 2, "")

	//annotate image with label dog
	testAnnotate(t, imageIds[0], "dog", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "")

	//now we expect just one result 
	testBrowseAnnotation(t, "cat&dog", 1, "")
	testBrowseAnnotation(t, "cat", 2, "")

	//annotate image with label cat
	testAnnotate(t, imageIds[0], "cat", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "")

	//now we should get no result
	testBrowseAnnotation(t, "cat&dog", 0, "")
	testBrowseAnnotation(t, "dog", 0, "")

	//there is still one cat left
	testBrowseAnnotation(t, "cat", 1, "")

}

func TestBrowseAnnotationQueryLockedButOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token)
	testDonate(t, "./images/apples/apple2.jpeg", "apple", false, token)

	testBrowseAnnotation(t, "apple", 2, token)
}

func TestBrowseAnnotationQueryLockedButForeignDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	token1 := testLogin(t, "user1", "pwd1", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token1)
	testDonate(t, "./images/apples/apple2.jpeg", "apple", false, token)

	testBrowseAnnotation(t, "apple", 1, token)
}

func TestBrowseAnnotationQueryLockedOwnDonationButQuarantine(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, token)
	testDonate(t, "./images/apples/apple2.jpeg", "apple", false, token)

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

	num := testMultipleDonate(t)

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	testBrowseAnnotation(t, "~tree", num, "")
	testBrowseAnnotation(t, "apple", num, "")

	testBrowseAnnotation(t, "~tree | apple", num, "")
	testBrowseAnnotation(t, "~tree & apple", num, "")
	testBrowseAnnotation(t, "~tree & car", 0, "")

	
	testAnnotate(t, imageIds[0], "apple", "", 
					`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "")

	testBrowseAnnotation(t, "~tree", num-1, "")
	testBrowseAnnotation(t, "apple", num-1, "")	

}

func TestGetImageToLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")
	testGetImageToLabel(t, "", "", 200)
}

func TestGetUnlabeledImageToLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "", true, "")
	testGetImageToLabel(t, "", "", 200)
}

func TestGetImageToLabel1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")
	testDonate(t, "./images/apples/apple2.jpeg", "", true, "")
	testGetImageToLabel(t, "", "", 200)
}

func TestGetImageToLabelNotUnlocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "")
	testGetImageToLabel(t, "", "", 422)
}

func TestGetImageToLabelNotUnlocked1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "", false, "")
	testDonate(t, "./images/apples/apple2.jpeg", "apple", false, "")
	testGetImageToLabel(t, "", "", 422)
}

func TestGetImageToLabelLockedButOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "", false, token)

	testGetImageToLabel(t, "", token, 200)
}

func TestGetImageToLabelLockedButOwnDonation1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "", false, token)
	testDonate(t, "./images/apples/apple2.jpeg", "", true, token)

	testGetImageToLabel(t, "", token, 200)
}


func TestGetImageToLabelLockedAndForeignDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)


	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	userToken1 := testLogin(t, "user1", "pwd1", 200)


	testDonate(t, "./images/apples/apple1.jpeg", "", false, userToken)

	testGetImageToLabel(t, "", userToken1, 422)
}

func TestGetImageByImageId(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	//testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	//userToken := testLogin(t, "user", "pwd", 200)


	testDonate(t, "./images/apples/apple1.jpeg", "", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetImageToLabel(t, imageId, "", 200)

}

func TestGetImageByImageIdForeignDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)


	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	userToken1 := testLogin(t, "user1", "pwd1", 200)


	testDonate(t, "./images/apples/apple1.jpeg", "", false, userToken)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetImageToLabel(t, imageId, userToken1, 422)

}

func TestGetImageByImageIdOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)


	testDonate(t, "./images/apples/apple1.jpeg", "", false, userToken)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetImageToLabel(t, imageId, userToken, 200)
}


func TestGetImageOwnDonationButPutInQuarantine(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)


	testDonate(t, "./images/apples/apple1.jpeg", "", false, userToken)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	err = db.PutImageInQuarantine(imageId)
	ok(t, err)

	testGetImageToLabel(t, "", userToken, 422)
}

func TestGetImageByIdOwnDonationButPutInQuarantine(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)


	testDonate(t, "./images/apples/apple1.jpeg", "", false, userToken)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	err = db.PutImageInQuarantine(imageId)
	ok(t, err)

	testGetImageToLabel(t, imageId, userToken, 422)
}

func TestGetImageToValidate(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	testGetImageToValidate(t, "", "", "", 200)
}

func TestGetImageToValidateById(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetImageToValidate(t, imageId, "", "", 200)
}

func TestGetImageToValidateAuthenticated(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)


	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, userToken)

	testGetImageToValidate(t, "", userToken, "", 200)
}

func TestGetImageToValidateByIdAuthenticated(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)


	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, userToken)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)


	testGetImageToValidate(t, imageId, userToken, "", 200)
}

func TestGetImageToValidateLocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "")

	testGetImageToValidate(t, "", "", "", 422)
}

func TestGetImageToValidateLockedButOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	testGetImageToValidate(t, "", userToken, "", 200)
}

func TestGetImageToValidateByLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	uuid, err := db.GetLabelUuidFromName("apple")
	ok(t, err)

	testGetImageToValidate(t, "", "", uuid, 200)
}

func TestGetImageToValidateByLabelMoreThanOne(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")
	testDonate(t, "./images/apples/apple2.jpeg", "apple", true, "")

	uuid, err := db.GetLabelUuidFromName("apple")
	ok(t, err)

	testGetImageToValidate(t, "", "", uuid, 200)
}

func TestGetImageToValidateByLabelLocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "")

	uuid, err := db.GetLabelUuidFromName("apple")
	ok(t, err)

	testGetImageToValidate(t, "", "", uuid, 422)
}

func TestGetImageToValidateByLabelLockedButOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	uuid, err := db.GetLabelUuidFromName("apple")
	ok(t, err)

	testGetImageToValidate(t, "", userToken, uuid, 200)
}

func TestGetImageToValidateByLabelLockedButForeignDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	userToken1 := testLogin(t, "user1", "pwd1", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	uuid, err := db.GetLabelUuidFromName("apple")
	ok(t, err)

	testGetImageToValidate(t, "", userToken1, uuid, 422)
}


func TestGetImageToAnnotateButNotEnoughValidation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	testGetImageForAnnotation(t, "", "", "", 422)
}

func TestGetImageToAnnotate(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)
	

	testGetImageForAnnotation(t, "", "", "", 200)
}

func TestGetImageToAnnotateButLocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "")

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	testGetImageForAnnotation(t, "", "", "", 422)
}

func TestGetImageToAnnotateUnlockedButBlacklistedBySignedUpUser(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testBlacklistAnnotation(t, validationIds[0], userToken)

	testGetImageForAnnotation(t, "", "", "", 200)
}

func TestGetImageToAnnotateUnlockedButBlacklistedByOtherUser(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	userToken1 := testLogin(t, "user1", "pwd1", 200)

	testBlacklistAnnotation(t, validationIds[0], userToken1)

	testGetImageForAnnotation(t, "", userToken, "", 200)
}

func TestGetImageToAnnotateUnlockedButBlacklistedByOwnUser(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testBlacklistAnnotation(t, validationIds[0], userToken)

	testGetImageForAnnotation(t, "", userToken, "", 422)
}


func TestGetImageToAnnotateUnlockedButNotAnnotateable(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	//we need 3 not-annotatable votes until a annotation task won't show up anymore
	//(that's an arbitrary number that's set in the imagemonkey sourcecode)
	testMarkValidationAsNotAnnotatable(t, validationIds[0], 3)

	testGetImageForAnnotation(t, "", userToken, "", 422)
}

func TestGetImageToAnnotateLockedButOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	testGetImageForAnnotation(t, "", userToken, "", 200)
}

func TestGetImageToAnnotateLockedButForeignDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	userToken1 := testLogin(t, "user1", "pwd1", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	testGetImageForAnnotation(t, "", userToken1, "", 422)
}


func TestGetImageToAnnotateLockedOwnDonationButQuarantine(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	err = db.PutImageInQuarantine(imageId)
	ok(t, err)

	testGetImageForAnnotation(t, "", userToken, "", 422)
}


func TestGetImageToAnnotateById(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetImageForAnnotation(t, imageId, "", "", 200)
}

func TestGetImageToAnnotateByValidationId(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	testGetImageForAnnotation(t, "", "", validationIds[0], 200)
}

func TestGetImageToAnnotateLockedButOwnDonationByValidationId(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	testGetImageForAnnotation(t, "", userToken, validationIds[0], 200)
}

func TestGetImageToAnnotateByIdAuthenticated(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetImageForAnnotation(t, imageId, userToken, "", 200)
}


func TestGetImageToAnnotateByIdButLocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "")

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetImageForAnnotation(t, imageId, "", "", 422)
}


func TestGetImageToAnnotateByIdLockedButOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetImageForAnnotation(t, imageId, userToken, "", 200)
}

func TestGetImageToAnnotateByIdLockedAndForeignDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	userToken1 := testLogin(t, "user1", "pwd1", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetImageForAnnotation(t, imageId, userToken1, "", 422)
}


func TestGetImageToAnnotateByIdLockedOwnDonationButQuarantine(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	err = db.PutImageInQuarantine(imageId)
	ok(t, err)

	testGetImageForAnnotation(t, imageId, userToken, "", 422)
}

func TestGetUnlockedImageDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t)

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for i := 0; i < len(imageIds); i++ {
		testGetImageDonation(t, imageIds[i], true, "", 200)
	}
}

func TestGetLockedImageDonationWithoutValidToken(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetImageDonation(t, imageId, false, "", 403)
}

func TestGetLockedImageDonationOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetImageDonation(t, imageId, false, userToken, 200)
}


func TestGetLockedImageDonationForeignDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	userToken1 := testLogin(t, "user1", "pwd1", 200)


	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetImageDonation(t, imageId, false, userToken1, 403)
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

func TestBrowseLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	testBrowseLabel(t, "apple", "", 1, 200)
}

func TestBrowseLabel1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "egg")

	testBrowseLabel(t, "apple&egg", "", 1, 200)
	testBrowseLabel(t, "apple|egg", "", 1, 200)
	testBrowseLabel(t, "apple&~egg", "", 0, 200)
}

func TestBrowseLabel2(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	testBrowseLabel(t, "apple&egg", "", 0, 200)
}

func TestBrowseLabelLocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "")

	testBrowseLabel(t, "apple", "", 0, 200)
}

func TestBrowseLabelLockedAndOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	testBrowseLabel(t, "apple", userToken, 1, 200)
}

func TestBrowseLabelLockedButForeignDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	userToken1 := testLogin(t, "user1", "pwd1", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	testBrowseLabel(t, "apple", userToken1, 0, 200)
}

func TestBrowseLabelLockedAndOwnDonationButInQuarantine(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	err = db.PutImageInQuarantine(imageId)
	ok(t, err)

	testBrowseLabel(t, "apple", userToken, 0, 200)
}