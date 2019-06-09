package tests

import (
	"testing"
	"gopkg.in/resty.v1"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	"net/url"
)

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


func testGetImageToValidate(t *testing.T, validationId string, token string, requiredStatusCode int) {
	url := BASE_URL + API_VERSION + "/validation"
	if validationId != "" {
		url += "/" + validationId
	}

	req := resty.R()

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(url)

	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
}

func testGetImagesForValidation(t *testing.T, query string, token string, requiredStatusCode int, numOfExpectedResults int) {
	var validations []datastructures.Validation

	u := BASE_URL + API_VERSION + "/validations"

	req := resty.R().
				SetQueryParams(map[string]string{
		          "query": url.QueryEscape(query),
		        }).
		        SetResult(&validations)

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(u)

	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
	equals(t, len(validations), numOfExpectedResults)
}


func testBatchValidation(t *testing.T, validations []datastructures.ImageValidation, requiredStatusCode int) {
	var imageValidationBatch datastructures.ImageValidationBatch

	imageValidationBatch.Validations = validations

	u := BASE_URL + API_VERSION + "/validation/validate"

	req := resty.R().
				SetBody(imageValidationBatch)

	resp, err := req.Patch(u)

	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
}


func TestRandomImageValidation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t, "apple")
	testRandomImageValidation(t, 100)
}


func TestRandomModeratedImageValidation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t, "apple")
	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)
	db.GiveUserModeratorRights("moderator") //give user moderator rights
	testRandomModeratedImageValidation(t, 100, moderatorToken)
}

func TestGetImageToValidate(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	testGetImageToValidate(t, "", "", 200)
}

func TestGetImageToValidateById(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)
	equals(t, len(validationIds), 1)

	testGetImageToValidate(t, validationIds[0], "", 200)
}

func TestGetImageToValidateByIdAuthenticatedWrongUser(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	userToken1 := testLogin(t, "user1", "pwd1", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "", 200)

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)
	equals(t, len(validationIds), 1)

	testGetImageToValidate(t, validationIds[0], userToken1, 422)
}

func TestGetImageToValidateAuthenticated(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)


	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, userToken, "", 200)

	testGetImageToValidate(t, "", userToken, 200)
}

func TestGetImageToValidateByIdAuthenticated(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)


	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, userToken, "", 200)

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)
	equals(t, len(validationIds), 1)

	testGetImageToValidate(t, validationIds[0], userToken, 200)
}

func TestGetImageToValidateLocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "", 200)

	testGetImageToValidate(t, "", "", 422)
}

func TestGetImageToValidateLockedButOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "", 200)

	testGetImageToValidate(t, "", userToken, 200)
}

func TestGetImagesForValidation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)
	testGetImagesForValidation(t, "apple", "", 200, 1)
}

func TestGetImagesForValidationStaticQueryAttributes1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)
	testGetImagesForValidation(t, "apple & image.width > 200px", "", 200, 1)
}

func TestGetImagesForValidationStaticQueryAttributes2(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)
	testGetImagesForValidation(t, "apple & annotation.coverage = 0%", "", 200, 1)
}

func TestGetImagesForValidationEmptyResultSet(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)
	testGetImagesForValidation(t, "car", "", 200, 0)
}

func TestGetImagesForValidationLockedEmptyResultSet(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "", 200)
	testGetImagesForValidation(t, "apple", "", 200, 0)
}

func TestGetImagesForValidationLockedAndOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "", 200)

	testGetImagesForValidation(t, "apple", userToken, 200, 1)
}

func TestGetImagesForValidationLockedButForeignDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	userToken1 := testLogin(t, "user1", "pwd1", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "", 200)

	testGetImagesForValidation(t, "apple", userToken1, 200, 0)
}

func TestBatchValidation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t, "apple")

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)

	validationId1NumOfValidBefore, validationId1NumOfInvalidBefore, err := db.GetValidationCount(validationIds[0])
	ok(t, err)
	validationId2NumOfValidBefore, validationId2NumOfInvalidBefore, err := db.GetValidationCount(validationIds[1])
	ok(t, err)

	var validations []datastructures.ImageValidation
	validation1 := datastructures.ImageValidation{Uuid: validationIds[0], Valid: "yes"}
	validations = append(validations, validation1)
	validation2 := datastructures.ImageValidation{Uuid: validationIds[1], Valid: "no"}
	validations = append(validations, validation2)
	testBatchValidation(t, validations, 204)


	validationId1NumOfValidAfter, validationId1NumOfInvalidAfter, err := db.GetValidationCount(validationIds[0])
	ok(t, err)
	validationId2NumOfValidAfter, validationId2NumOfInvalidAfter, err := db.GetValidationCount(validationIds[1])
	ok(t, err)

	equals(t, (validationId1NumOfValidBefore + 1), validationId1NumOfValidAfter)
	equals(t, validationId1NumOfInvalidBefore, validationId1NumOfInvalidAfter)
	equals(t, validationId2NumOfValidBefore, validationId2NumOfValidAfter)
	equals(t, (validationId2NumOfInvalidBefore + 1), validationId2NumOfInvalidAfter)
}
