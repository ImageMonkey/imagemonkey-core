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
		testSuggestLabelForImage(t, imageId, "bananaa", true, token, 200)
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
         testSuggestLabelForImage(t, imageId, "construction sitee", true, token, 200)
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

 func TestMakeLabelWithLeadingAndTrailingWhitespacesProductive(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200)
 
 	err := db.RemoveLabel("car")
    ok(t, err)

	testMultipleDonate(t, "carpet")
	imageIds, err := db.GetAllImageIds()
	ok(t, err)
	for i, imageId := range imageIds {
		if i == 0 || i == 4 {
			testSuggestLabelForImage(t, imageId, " car", true, token, 200)
		} else if i == 2 || i == 3 {
			testSuggestLabelForImage(t, imageId, "car ", true, token, 200)
		} else {
			testSuggestLabelForImage(t, imageId, " car ", true, token, 200)
		}
	}

	numCarpetBefore, err := db.GetNumberOfImagesWithLabel("carpet")
	ok(t, err)
	equals(t, int(numCarpetBefore), int(13))

	numCarBefore, err := db.GetNumberOfImagesWithLabel("car")
	ok(t, err)
	equals(t, int(numCarBefore), int(0))

	err = populateLabels()
	ok(t, err)
	
	runMakeTrendingLabelsProductiveScript(t, "car", "car", true)

	numCarAfter, err := db.GetNumberOfImagesWithLabel("car")
	ok(t, err)
	equals(t, int(numCarAfter), int(13))

	numCarpetAfter, err := db.GetNumberOfImagesWithLabel("carpet")
	ok(t, err)
	equals(t, int(numCarpetAfter), int(13))
 }

 func TestMakeLabelWithLeadingAndTrailingWhitespacesProductive2(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200) 

	testMultipleDonate(t, "carpet")
	imageIds, err := db.GetAllImageIds()
	ok(t, err)
	for i, imageId := range imageIds {
		if i == 0 || i == 4 {
			testSuggestLabelForImage(t, imageId, " car   ", true, token, 200)
		} else if i == 2 || i == 3 {
			testSuggestLabelForImage(t, imageId, "  car  ", true, token, 200)
		} else {
			testSuggestLabelForImage(t, imageId, " car    ", true, token, 200)
		}
	}

	numCarpetBefore, err := db.GetNumberOfImagesWithLabel("carpet")
	ok(t, err)
	equals(t, int(numCarpetBefore), int(13))

	numCarBefore, err := db.GetNumberOfImagesWithLabel("car")
	ok(t, err)
	equals(t, int(numCarBefore), int(0))
	
	runMakeTrendingLabelsProductiveScript(t, "car", "car", true)

	numCarAfter, err := db.GetNumberOfImagesWithLabel("car")
	ok(t, err)
	equals(t, int(numCarAfter), int(13))

	numCarpetAfter, err := db.GetNumberOfImagesWithLabel("carpet")
	ok(t, err)
	equals(t, int(numCarpetAfter), int(13))
 }

 func TestMakeLabelWithLeadingAndTrailingWhitespacesProductive3(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "testuser", "testpassword", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpassword", 200) 

	testMultipleDonate(t, "carpet")
	imageIds, err := db.GetAllImageIds()
	ok(t, err)
	for i, imageId := range imageIds {
		testSuggestLabelForImage(t, imageId, "car seat", true, token, 200)
		testSuggestLabelForImage(t, imageId, " car seat", true, token, 200)
		testSuggestLabelForImage(t, imageId, "car seat ", true, token, 200)
		testSuggestLabelForImage(t, imageId, " car seat ", true, token, 200)
		if i == 0 || i == 4 {
			testSuggestLabelForImage(t, imageId, " car   ", true, token, 200)
		} else if i == 2 || i == 3 {
			testSuggestLabelForImage(t, imageId, "  car  ", true, token, 200)
		} else {
			testSuggestLabelForImage(t, imageId, " car    ", true, token, 200)
		}
	}

	numCarpetBefore, err := db.GetNumberOfImagesWithLabel("carpet")
	ok(t, err)
	equals(t, int(numCarpetBefore), int(13))

	numCarBefore, err := db.GetNumberOfImagesWithLabel("car")
	ok(t, err)
	equals(t, int(numCarBefore), int(0))
	
	runMakeTrendingLabelsProductiveScript(t, "car", "car", true)

	numCarAfter, err := db.GetNumberOfImagesWithLabel("car")
	ok(t, err)
	equals(t, int(numCarAfter), int(13))

	numCarpetAfter, err := db.GetNumberOfImagesWithLabel("carpet")
	ok(t, err)
	equals(t, int(numCarpetAfter), int(13))

	numCarseatAfter, err := db.GetNumberOfImagesWithLabelSuggestions("car seat")
	ok(t, err)
	equals(t, int(numCarseatAfter), int(13))

	numCarseatAfter, err = db.GetNumberOfImagesWithLabelSuggestions(" car seat")
	ok(t, err)
	equals(t, int(numCarseatAfter), int(13))

	numCarseatAfter, err = db.GetNumberOfImagesWithLabelSuggestions("car seat ")
	ok(t, err)
	equals(t, int(numCarseatAfter), int(13))
 
 	numCarseatAfter, err = db.GetNumberOfImagesWithLabelSuggestions(" car seat ")
	ok(t, err)
	equals(t, int(numCarseatAfter), int(13))
 }
