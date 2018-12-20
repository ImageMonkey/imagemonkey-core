package tests

import (
	"testing"
)

func TestAddMetaLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	numBefore, err := db.GetNumOfMetaLabelImageValidations()
	ok(t, err)
	equals(t, numBefore, 0)

	testDonate(t, "./images/apples/apple1.jpeg", "", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "kitchen", "")

	numAfter, err := db.GetNumOfMetaLabelImageValidations()
	ok(t, err)
	equals(t, numAfter, 1)
}

func TestMetaLabelValidationShouldntBeReturnedInRandomAnnotation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "kitchen", "")

	testGetImageForAnnotation(t, "", "", "", 422)
}

func TestMetaLabelShouldntBeAnnotatable(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "", true, "", "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "kitchen", "")

	testAnnotate(t, imageId, "kitchen", "", 
						`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 400)

}

