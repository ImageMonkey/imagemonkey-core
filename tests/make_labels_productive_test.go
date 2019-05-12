package tests

import (
	"testing"
)

func TestMakeLabelsProductiveNormalLabelTest(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200)

	err := db.RemoveLabel("banana")
	ok(t, err)

	testMultipleDonate(t, "apple")
	imageIds, err := db.GetAllImageIds()
	ok(t, err)
	for _, imageId := range imageIds {
		testSuggestLabelForImage(t, imageId, "bananaa", true, token)
	}

	numBefore, err := db.GetNumberOfImagesWithLabel("banana")
	ok(t, err)
	equals(t, int(numBefore), int(0))
	
	err = populateLabels()
	ok(t, err)

	runMakeTrendingLabelsProductiveScript(t, "bananaa", "banana", true)

	numAfter, err := db.GetNumberOfImagesWithLabel("banana")
	ok(t, err)
	equals(t, int(numAfter), int(13))

	validationIds, err := db.GetAllValidationIds()
	ok(t, err)
	for _, validationId := range validationIds {
		num, err := db.GetNumOfNotAnnotatable(validationId)
		ok(t, err)
		equals(t, int(num), 0)
	}
}

func TestMakeLabelsProductiveMetaLabelTest(t *testing.T) {
     teardownTestCase := setupTestCase(t)
     defer teardownTestCase(t)
 
     testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
     token := testLogin(t, "testuser", "testpassword", 200)
 
     err := db.RemoveLabel("construction site")
     ok(t, err)
 
     testMultipleDonate(t, "apple")
     imageIds, err := db.GetAllImageIds()
     ok(t, err)
     for _, imageId := range imageIds {
         testSuggestLabelForImage(t, imageId, "construction sitee", true, token)
     }
 
     numBefore, err := db.GetNumberOfImagesWithLabel("construction site")
     ok(t, err)
     equals(t, int(numBefore), int(0))
 
     err = populateLabels()
     ok(t, err)
 
     runMakeTrendingLabelsProductiveScript(t, "construction sitee", "construction site", true)
 
     numAfter, err := db.GetNumberOfImagesWithLabel("construction site")
     ok(t, err)
     equals(t, int(numAfter), int(13))
 
     validationIds, err := db.GetAllValidationIdsForLabel("construction site")
     ok(t, err)
     for _, validationId := range validationIds {
         num, err := db.GetNumOfNotAnnotatable(validationId)
         ok(t, err)
         equals(t, int(num), 0)
     }
 }
