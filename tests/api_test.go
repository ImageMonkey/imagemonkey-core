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

//const UNVERIFIED_DONATIONS_DIR string = "../unverified_donations/"
//const DONATIONS_DIR string = "../donations/"

type LoginResult struct {
	Token string `json:"token"`
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

func testDonate(t *testing.T, path string, label string, unlockImage bool, token string, imageCollectionName string) {
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



func testMultipleDonate(t *testing.T, label string) int {
	dirname := "./images/apples/"
	files, err := ioutil.ReadDir(dirname)
    ok(t, err)

    num := 0
    for _, f := range files {
        testDonate(t, dirname + f.Name(), label, true, "", "")
        num += 1
    }

    return num
}

/*func testMultipleDonateWithToken(t *testing.T, label string, token string) int {
	dirname := "./images/apples/"
	files, err := ioutil.ReadDir(dirname)
    ok(t, err)

    num := 0
    for _, f := range files {
        testDonate(t, dirname + f.Name(), label, true, token, "")
        num += 1
    }

    return num
}*/

func testLabelImage(t *testing.T, imageId string, label string, token string) {
	type LabelMeEntry struct {
		Label string `json:"label"`
	}

	oldNum, err := db.GetNumberOfImagesWithLabel(label)
	ok(t, err)

	var labelMeEntries []LabelMeEntry
	labelMeEntry := LabelMeEntry{Label: label}
	labelMeEntries = append(labelMeEntries, labelMeEntry)

	url := BASE_URL + API_VERSION + "/donation/" + imageId + "/labelme"
	req := resty.R().
			SetHeader("Content-Type", "application/json").
			SetBody(labelMeEntries)

	if token != "" {
		req.SetAuthToken(token)
	}
			
	resp, err := req.Post(url)

	ok(t, err)
	equals(t, resp.StatusCode(), 200)

	newNum, err := db.GetNumberOfImagesWithLabel(label)
	ok(t, err)

	equals(t, oldNum+1, newNum)
}


func testSuggestLabelForImage(t *testing.T, imageId string, label string, token string) {
	type LabelMeEntry struct {
		Label string `json:"label"`
	}

	oldNum, err := db.GetNumberOfImagesWithLabelSuggestions(label)
	ok(t, err)

	var labelMeEntries []LabelMeEntry
	labelMeEntry := LabelMeEntry{Label: label}
	labelMeEntries = append(labelMeEntries, labelMeEntry)

	url := BASE_URL + API_VERSION + "/donation/" + imageId + "/labelme"
	req := resty.R().
			SetHeader("Content-Type", "application/json").
			SetBody(labelMeEntries)

	if token != "" {
		req.SetAuthToken(token)
	}
			
	resp, err := req.Post(url)

	ok(t, err)
	equals(t, resp.StatusCode(), 200)

	newNum, err := db.GetNumberOfImagesWithLabelSuggestions(label)
	ok(t, err)

	equals(t, oldNum+1, newNum)
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

		testLabelImage(t, image, label, "")
	}
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

func testGetRandomImageQuiz(t *testing.T, requiredStatusCode int) {
	url := BASE_URL + API_VERSION + "/quiz-refine"

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

	testMultipleDonate(t, "apple")
}

func TestRandomAnnotate(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t, "apple")
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

func TestRandomAnnotationRework(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t, "apple")
	testRandomAnnotate(t, 2, `[{"top":100,"left":200,"type":"rect","angle":0,"width":40,"height":60,"stroke":{"color":"red","width":1}}]`)
	testRandomAnnotationRework(t, 2, `[{"top":200,"left":300,"type":"rect","angle":10,"width":50,"height":30,"stroke":{"color":"blue","width":3}}]`)
}


func TestRandomLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t, "apple")
	testRandomLabel(t, 7)
}

func TestGetImageToLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")
	testGetImageToLabel(t, "", "", 200)
}

func TestGetUnlabeledImageToLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "", true, "", "")
	testGetImageToLabel(t, "", "", 200)
}

func TestGetImageToLabel1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")
	testDonate(t, "./images/apples/apple2.jpeg", "", true, "", "")
	testGetImageToLabel(t, "", "", 200)
}

func TestGetImageToLabelNotUnlocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "")
	testGetImageToLabel(t, "", "", 422)
}

func TestGetImageToLabelNotUnlocked1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "", false, "", "")
	testDonate(t, "./images/apples/apple2.jpeg", "apple", false, "", "")
	testGetImageToLabel(t, "", "", 422)
}

func TestGetImageToLabelLockedButOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "", false, token, "")

	testGetImageToLabel(t, "", token, 200)
}

func TestGetImageToLabelLockedButOwnDonation1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "", false, token, "")
	testDonate(t, "./images/apples/apple2.jpeg", "", true, token, "")

	testGetImageToLabel(t, "", token, 200)
}


func TestGetImageToLabelLockedAndForeignDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)


	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	userToken1 := testLogin(t, "user1", "pwd1", 200)


	testDonate(t, "./images/apples/apple1.jpeg", "", false, userToken, "")

	testGetImageToLabel(t, "", userToken1, 422)
}

func TestGetImageByImageId(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	//testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	//userToken := testLogin(t, "user", "pwd", 200)


	testDonate(t, "./images/apples/apple1.jpeg", "", true, "", "")

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


	testDonate(t, "./images/apples/apple1.jpeg", "", false, userToken, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetImageToLabel(t, imageId, userToken1, 422)

}

func TestGetImageByImageIdOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)


	testDonate(t, "./images/apples/apple1.jpeg", "", false, userToken, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetImageToLabel(t, imageId, userToken, 200)
}


func TestGetImageOwnDonationButPutInQuarantine(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)


	testDonate(t, "./images/apples/apple1.jpeg", "", false, userToken, "")

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


	testDonate(t, "./images/apples/apple1.jpeg", "", false, userToken, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	err = db.PutImageInQuarantine(imageId)
	ok(t, err)

	testGetImageToLabel(t, imageId, userToken, 422)
}




func TestGetImageToAnnotateButNotEnoughValidation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	testGetImageForAnnotation(t, "", "", "", 422)
}

func TestGetImageToAnnotate(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)
	

	testGetImageForAnnotation(t, "", "", "", 200)
}

func TestGetImageToAnnotateButLocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "")

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	err = db.SetValidationValid(validationIds[0], 5)
	ok(t, err)

	testGetImageForAnnotation(t, "", "", "", 422)
}

func TestGetImageToAnnotateUnlockedButBlacklistedBySignedUpUser(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "")

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "")

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "")

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "")

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "")

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "")

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "")

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "")

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

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "")

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

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	for i := 0; i < len(imageIds); i++ {
		testGetImageDonation(t, imageIds[i], true, "", 200)
	}
}

func TestGetLockedImageDonationWithoutValidToken(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetImageDonation(t, imageId, false, "", 403)
}

func TestGetLockedImageDonationOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "")

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


	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testGetImageDonation(t, imageId, false, userToken1, 403)
}

