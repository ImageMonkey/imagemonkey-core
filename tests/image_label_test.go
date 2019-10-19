package tests

import (
	"testing"
	"gopkg.in/resty.v1"
	"net/url"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	"encoding/json"
	"sort"
)

func testLabelImage(t *testing.T, imageId string, label string, sublabel string, token string, requiredStatusCode int) {
	oldNum, err := db.GetNumberOfImagesWithLabel(label)
	ok(t, err)

	var labelMeEntries []datastructures.LabelMeEntry
	labelMeEntry := datastructures.LabelMeEntry{Label: label}

	if sublabel != "" {
		labelMeEntry.Sublabels = append(labelMeEntry.Sublabels, datastructures.Sublabel{Name: sublabel})
	}

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
	equals(t, resp.StatusCode(), requiredStatusCode)

	newNum, err := db.GetNumberOfImagesWithLabel(label)
	ok(t, err)
	
	if requiredStatusCode == 200 {
		equals(t, oldNum+1, newNum)
	} else {
		equals(t, oldNum, newNum)
	}
}


func testSuggestLabelForImage(t *testing.T, imageId string, label string, annotatable bool, token string, requiredStatusCode int) {
	type LabelMeEntry struct {
		Label string `json:"label"`
		Annotatable bool `json:"annotatable"`
	}

	existingNumLabelsWithSameName, err := db.GetNumberOfLabelSuggestionsWithLabelForImage(imageId, label)
	ok(t, err)

	oldNum, err := db.GetNumberOfImagesWithLabelSuggestions(label)
	ok(t, err)

	var labelMeEntries []LabelMeEntry
	labelMeEntry := LabelMeEntry{Label: label}
	labelMeEntry.Annotatable = annotatable
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

	//the labelme endpoint behaves a bit special, as it returns always 200
	//no matter if the label already exists or not.
	//the only difference is, that it won't be inserted again if it already exists! 
	if requiredStatusCode == 200 {
		if existingNumLabelsWithSameName == 1 { 
			equals(t, oldNum, newNum)
		} else {
			equals(t, oldNum+1, newNum)
		}
	}
}

func testBrowseLabel(t *testing.T, query string, token string, requiredNumOfResults int, 
		requiredStatusCode int) []datastructures.ImageLabel {
	u := BASE_URL + API_VERSION + "/donations/labels"
	var labeledImages []datastructures.ImageLabel 

	req := resty.R().
			SetQueryParams(map[string]string{
				"query": url.QueryEscape(query),
		    }).
		    SetResult(&labeledImages)

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(u)
	
	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
	equals(t, len(labeledImages), requiredNumOfResults)

	return labeledImages
} 

func testGetLabelsForImage(t *testing.T, imageId string, token string, includeOnlyUnlockedLabels bool, requiredStatusCode int) []datastructures.LabelMeEntry {
	u := BASE_URL + API_VERSION + "/donation/" + imageId + "/labels"

	var labels []datastructures.LabelMeEntry

	urlParam := "false"
	if includeOnlyUnlockedLabels {
		urlParam = "true"
	}

	req := resty.R().
			SetQueryParams(map[string]string{
				"only_unlocked_labels": urlParam,
		    }).
		    SetResult(&labels)

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(u)
	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)

	return labels
}


func TestBrowseLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	testBrowseLabel(t, "apple", "", 1, 200)
}

func TestBrowseLabel1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "egg", "", "", 200)

	testBrowseLabel(t, "apple&egg", "", 1, 200)
	testBrowseLabel(t, "apple|egg", "", 1, 200)
	testBrowseLabel(t, "apple&~egg", "", 0, 200)
}

func TestBrowseLabel2(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	testBrowseLabel(t, "apple&egg", "", 0, 200)
}

func TestBrowseLabelLocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, "", "", 200)

	testBrowseLabel(t, "apple", "", 0, 200)
}

func TestBrowseLabelLockedAndOwnDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "", 200)

	testBrowseLabel(t, "apple", userToken, 1, 200)
}

func TestBrowseLabelLockedButForeignDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testSignUp(t, "user1", "pwd1", "user1@imagemonkey.io")
	userToken1 := testLogin(t, "user1", "pwd1", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "", 200)

	testBrowseLabel(t, "apple", userToken1, 0, 200)
}

func TestBrowseLabelLockedAndOwnDonationButInQuarantine(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", false, userToken, "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	err = db.PutImageInQuarantine(imageId)
	ok(t, err)

	testBrowseLabel(t, "apple", userToken, 0, 200)
}

func TestBrowseLabelImageDimensions(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	testBrowseLabel(t, "apple & image.width > 15px & image.height > 15px", "", 1, 200)
}

func TestBrowseLabelImageDimensionsWithoutLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	testBrowseLabel(t, "image.width > 15px & image.height > 15px", "", 1, 200)
}

func TestBrowseLabelWrongImageDimensionsSyntax(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	testBrowseLabel(t, "apple & image.width > 15", "", 0, 422)
}


func TestBrowseLabelAnnotationCoverage(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	testBrowseLabel(t, "apple & annotation.coverage = 0%", "", 1, 200)
}

func TestBrowseLabelAnnotationCoverage2(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":60,"left":145,"type":"rect","angle":0,"width":836,"height":660,"stroke":{"color":"red","width":5}}]`, "", 201)

	runDataProcessor(t)

	testBrowseLabel(t, "apple & annotation.coverage > 0%", "", 1, 200)
}

func TestBrowseLabelNoImageDescription(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	images := testBrowseLabel(t, "apple", "", 1, 200)

	equals(t, len(images[0].Image.Descriptions), 0)
}

func TestBrowseLabelOneImageDescription(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	var imageDscs []datastructures.ImageDescription
	imageDsc := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDscs = append(imageDscs, imageDsc)

	testAddImageDescriptions(t, imageId, imageDscs)

	images := testBrowseLabel(t, "apple", "", 1, 200)

	equals(t, len(images[0].Image.Descriptions), 1)

	rawImageDescriptions := images[0].Image.Descriptions

	type ImageDescription struct {
		Text string `json:"text"`
	}

	var imageDescription ImageDescription

	err = json.Unmarshal(rawImageDescriptions[0], &imageDescription)
	ok(t, err)
	equals(t, imageDescription.Text, "apple on the floor")
}


func TestBrowseLabelOneImageDescriptionLocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	var imageDescriptions []datastructures.ImageDescription
	imageDescription := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDescriptions = append(imageDescriptions, imageDescription)

	testAddImageDescriptions(t, imageId, imageDescriptions)

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)

	err = db.GiveUserModeratorRights("moderator")
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	testLockImageDescription(t, imageId, descriptions[0].Uuid, moderatorToken, 201)

	images := testBrowseLabel(t, "apple", "", 1, 200)

	equals(t, len(images[0].Image.Descriptions), 0)
}


func TestBrowseLabelOneImageDescriptionUnlocked(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	var imageDscs []datastructures.ImageDescription
	imageDsc := datastructures.ImageDescription{Description: "apple on the floor", Language: "en"}
	imageDscs = append(imageDscs, imageDsc)

	testAddImageDescriptions(t, imageId, imageDscs)

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)

	err = db.GiveUserModeratorRights("moderator")
	ok(t, err)

	descriptions, err := db.GetImageDescriptionForImageId(imageId)
	ok(t, err)

	testUnlockImageDescription(t, imageId, descriptions[0].Uuid, moderatorToken, 201)

	images := testBrowseLabel(t, "apple", "", 1, 200)

	equals(t, len(images[0].Image.Descriptions), 1)

	rawImageDescriptions := images[0].Image.Descriptions

	type ImageDescription struct {
		Text string `json:"text"`
	}

	var imageDescription ImageDescription

	err = json.Unmarshal(rawImageDescriptions[0], &imageDescription)
	ok(t, err)
	equals(t, imageDescription.Text, "apple on the floor")
}

func TestOnlyOneLabelAccessorPerLabelId(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	moreThanOneLabelId, err := db.DoLabelAccessorsBelongToMoreThanOneLabelId()
	ok(t, err)
	equals(t, moreThanOneLabelId, false)
}

func TestGetLabelsForImage(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	labels := testGetLabelsForImage(t, imageId, "", false, 200)
	equals(t, len(labels), 1)
	equals(t, labels[0].Label, "apple")
	equals(t, labels[0].Unlocked, true)
	equals(t, labels[0].Annotatable, true)
	equals(t, int(labels[0].Validation.NumOfValid), int(0))
	equals(t, int(labels[0].Validation.NumOfInvalid), int(0))
}
func TestGetLabelsForImageMultipleLabels(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "table", "", "", 200)

	labels := testGetLabelsForImage(t, imageId, "", false, 200)
	equals(t, len(labels), 2)

	sort.SliceStable(labels, func(i, j int) bool { return labels[i].Label < labels[j].Label })


	equals(t, labels[0].Label, "apple")
	equals(t, labels[0].Unlocked, true)
	equals(t, labels[0].Annotatable, true)
	equals(t, labels[0].Validation.NumOfValid, int32(0))
	equals(t, labels[0].Validation.NumOfInvalid, int32(0))

	equals(t, labels[1].Label, "table")
	equals(t, labels[1].Unlocked, true)
	equals(t, labels[1].Annotatable, true)
	equals(t, labels[1].Validation.NumOfValid, int32(0))
	equals(t, labels[1].Validation.NumOfInvalid, int32(0))
}

func TestGetLabelsForImageMultipleLabelsIncludingNonProductiveOnes(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "table", "", "", 200)
	testSuggestLabelForImage(t, imageId, "not-existing", true, userToken, 200)

	labels := testGetLabelsForImage(t, imageId, "", false, 200)

	sort.SliceStable(labels, func(i, j int) bool { return labels[i].Label < labels[j].Label })

	equals(t, len(labels), 3)
	equals(t, labels[0].Label, "apple")
	equals(t, labels[0].Unlocked, true)
	equals(t, labels[0].Annotatable, true)
	equals(t, labels[0].Validation.NumOfValid, int32(0))
	equals(t, labels[0].Validation.NumOfInvalid, int32(0))

	equals(t, labels[1].Label, "not-existing")
	equals(t, labels[1].Unlocked, false)
	equals(t, labels[1].Annotatable, true)
	equals(t, labels[1].Validation.NumOfValid, int32(0))
	equals(t, labels[1].Validation.NumOfInvalid, int32(0))

	equals(t, labels[2].Label, "table")
	equals(t, labels[2].Unlocked, true)
	equals(t, labels[2].Annotatable, true)
	equals(t, labels[2].Validation.NumOfValid, int32(0))
	equals(t, labels[2].Validation.NumOfInvalid, int32(0))
}

func TestGetLabelsForImageMultipleLabelsWithoutNonProductiveOnes(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "table", "", "", 200)
	testSuggestLabelForImage(t, imageId, "not-existing", true, userToken, 200)

	labels := testGetLabelsForImage(t, imageId, "", true, 200)

	sort.SliceStable(labels, func(i, j int) bool { return labels[i].Label < labels[j].Label })

	equals(t, len(labels), 2)
	equals(t, labels[0].Label, "apple")
	equals(t, labels[0].Unlocked, true)
	equals(t, labels[0].Annotatable, true)
	equals(t, labels[0].Validation.NumOfValid, int32(0))
	equals(t, labels[0].Validation.NumOfInvalid, int32(0))

	/*equals(t, labels[1].Label, "not-existing")
	equals(t, labels[1].Unlocked, false)
	equals(t, labels[1].Annotatable, true)
	equals(t, labels[1].Validation.NumOfValid, int32(0))
	equals(t, labels[1].Validation.NumOfInvalid, int32(0))*/

	equals(t, labels[1].Label, "table")
	equals(t, labels[1].Unlocked, true)
	equals(t, labels[1].Annotatable, true)
	equals(t, labels[1].Validation.NumOfValid, int32(0))
	equals(t, labels[1].Validation.NumOfInvalid, int32(0))
}


func TestGetImagesToLabelByImageCollectionName(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAddImageCollection(t, "user", token, "mycoll", "my-new-image-collection", 201)

	addImageToImageCollection(t, "user", token, "mycoll", imageId, 201)

	imgs := testBrowseLabel(t, "image.collection='mycoll'", token, 1, 200)
	equals(t, imgs[0].Image.Unlocked, true)
	equals(t, int(imgs[0].Image.Width), int(1132))
	equals(t, int(imgs[0].Image.Height), int(750))
	equals(t, len(imgs[0].Labels), 1)	
	equals(t, imgs[0].Labels[0].Name, "apple")
}

func TestGetImagesToLabelByImageCollectionNameNoLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAddImageCollection(t, "user", token, "mycoll", "my-new-image-collection", 201)

	addImageToImageCollection(t, "user", token, "mycoll", imageId, 201)

	imgs := testBrowseLabel(t, "image.collection='mycoll'", token, 1, 200)
	equals(t, imgs[0].Image.Unlocked, true)
	equals(t, int(imgs[0].Image.Width), int(1132))
	equals(t, int(imgs[0].Image.Height), int(750))
	equals(t, len(imgs[0].Labels), 0)	
}

func TestGetImagesToLabelByImageCollectionNameButForeignCollection(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAddImageCollection(t, "user", token, "mycoll", "my-new-image-collection", 201)

	addImageToImageCollection(t, "user", token, "mycoll", imageId, 201)

	testBrowseLabel(t, "image.collection='mycoll'", "", 0, 200)
}

func TestGetImagesToLabelByImageCollectionNameMultipleImages(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testAddImageCollection(t, "user", token, "mycoll", "my-new-image-collection", 201)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)
	addImageToImageCollection(t, "user", token, "mycoll", imageId, 201)

	testDonate(t, "./images/apples/apple2.jpeg", "apple", true, "", "", 200)
	imageId, err = db.GetLatestDonatedImageId()
	ok(t, err)
	addImageToImageCollection(t, "user", token, "mycoll", imageId, 201)

	testDonate(t, "./images/apples/apple3.jpeg", "", true, "", "", 200)
	imageId, err = db.GetLatestDonatedImageId()
	ok(t, err)
	addImageToImageCollection(t, "user", token, "mycoll", imageId, 201)

	imgs := testBrowseLabel(t, "image.collection='mycoll'", token, 3, 200)

	sort.SliceStable(imgs, func(i, j int) bool { return imgs[i].Image.Width < imgs[j].Image.Width })

	equals(t, imgs[0].Image.Unlocked, true)
	equals(t, int(imgs[0].Image.Width), int(750))
	equals(t, int(imgs[0].Image.Height), int(750))
	equals(t, len(imgs[0].Labels), 0)

	equals(t, imgs[1].Image.Unlocked, true)
	equals(t, int(imgs[1].Image.Width), int(1125))
	equals(t, int(imgs[1].Image.Height), int(750))
	equals(t, len(imgs[1].Labels), 1)
	equals(t, imgs[1].Labels[0].Name, "apple")
	equals(t, len(imgs[1].Labels[0].Sublabels), 0)


	equals(t, imgs[2].Image.Unlocked, true)
	equals(t, int(imgs[2].Image.Width), int(1132))
	equals(t, int(imgs[2].Image.Height), int(750))
	equals(t, len(imgs[2].Labels), 1)
	equals(t, imgs[2].Labels[0].Name, "apple")
	equals(t, len(imgs[2].Labels[0].Sublabels), 0)
}

func TestGetImagesToLabelByImageCollectionNameMultipleImagesWithSublabels(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "pwd", "user@imagemonkey.io")
	token := testLogin(t, "user", "pwd", 200)

	testAddImageCollection(t, "user", token, "mycoll", "my-new-image-collection", 201)

	testDonate(t, "./images/apples/apple1.jpeg", "", true, "", "", 200)
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)
	testLabelImage(t, imageId, "dog", "ear", "", 200)
	addImageToImageCollection(t, "user", token, "mycoll", imageId, 201)

	testDonate(t, "./images/apples/apple2.jpeg", "", true, "", "", 200)
	imageId, err = db.GetLatestDonatedImageId()
	ok(t, err)
	testLabelImage(t, imageId, "dog", "mouth", "", 200)
	addImageToImageCollection(t, "user", token, "mycoll", imageId, 201)

	testDonate(t, "./images/apples/apple3.jpeg", "", true, "", "", 200)
	imageId, err = db.GetLatestDonatedImageId()
	ok(t, err)
	testLabelImage(t, imageId, "dog", "eye", "", 200)
	addImageToImageCollection(t, "user", token, "mycoll", imageId, 201)

	imgs := testBrowseLabel(t, "image.collection='mycoll'", token, 3, 200)

	sort.SliceStable(imgs, func(i, j int) bool { return imgs[i].Image.Width < imgs[j].Image.Width })

	equals(t, imgs[0].Image.Unlocked, true)
	equals(t, int(imgs[0].Image.Width), int(750))
	equals(t, int(imgs[0].Image.Height), int(750))
	equals(t, len(imgs[0].Labels), 1)
	equals(t, imgs[0].Labels[0].Name, "dog")
	equals(t, len(imgs[0].Labels[0].Sublabels), 1)
	equals(t, imgs[0].Labels[0].Sublabels[0].Name, "eye")

	equals(t, imgs[1].Image.Unlocked, true)
	equals(t, int(imgs[1].Image.Width), int(1125))
	equals(t, int(imgs[1].Image.Height), int(750))
	equals(t, len(imgs[1].Labels), 1)
	equals(t, imgs[1].Labels[0].Name, "dog")
	equals(t, len(imgs[1].Labels[0].Sublabels), 1)
	equals(t, imgs[1].Labels[0].Sublabels[0].Name, "mouth")


	equals(t, imgs[2].Image.Unlocked, true)
	equals(t, int(imgs[2].Image.Width), int(1132))
	equals(t, int(imgs[2].Image.Height), int(750))
	equals(t, len(imgs[2].Labels), 1)
	equals(t, imgs[2].Labels[0].Name, "dog")
	equals(t, len(imgs[2].Labels[0].Sublabels), 1)
	equals(t, imgs[2].Labels[0].Sublabels[0].Name, "ear")
}

func TestGetImagesToLabelUnlabeledImages(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "", true, "", "", 200)

	imgs := testBrowseLabel(t, "image.unlabeled='true'", "", 1, 200)

	equals(t, imgs[0].Image.Unlocked, true)
	equals(t, int(imgs[0].Image.Width), int(1132))
	equals(t, int(imgs[0].Image.Height), int(750))
	equals(t, len(imgs[0].Labels), 0)
}

func TestGetImagesToLabelUnlabeledAndLabeledImages(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "", true, "", "", 200)
	testDonate(t, "./images/apples/apple2.jpeg", "apple", true, "", "", 200)

	imgs := testBrowseLabel(t, "image.unlabeled='true'", "", 1, 200)

	equals(t, imgs[0].Image.Unlocked, true)
	equals(t, int(imgs[0].Image.Width), int(1132))
	equals(t, int(imgs[0].Image.Height), int(750))
	equals(t, len(imgs[0].Labels), 0)
}

func TestGetImagesToLabelMultipleUnlabeledImages(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "", true, "", "", 200)
	testDonate(t, "./images/apples/apple2.jpeg", "", true, "", "", 200)

	imgs := testBrowseLabel(t, "image.unlabeled='true'", "", 2, 200)

	sort.SliceStable(imgs, func(i, j int) bool { return imgs[i].Image.Width < imgs[j].Image.Width })

	equals(t, imgs[0].Image.Unlocked, true)
	equals(t, int(imgs[0].Image.Width), int(1125))
	equals(t, int(imgs[0].Image.Height), int(750))
	equals(t, len(imgs[0].Labels), 0)

	equals(t, imgs[1].Image.Unlocked, true)
	equals(t, int(imgs[1].Image.Width), int(1132))
	equals(t, int(imgs[1].Image.Height), int(750))
	equals(t, len(imgs[1].Labels), 0)
}

func TestLabelImageNonProductiveLabelsOnlyOnce(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testSignUp(t, "testuser", "testpwd", "testuser@imagemonkey.io")
	token := testLogin(t, "testuser", "testpwd", 200)

	testSuggestLabelForImage(t, imageId, "test", true, token, 200)
	numBefore, err := db.GetNumberOfLabelSuggestionsForImage(imageId)
	ok(t, err)
	equals(t, numBefore, 1)

	testSuggestLabelForImage(t, imageId, "test", true, token, 200)
	numAfter, err := db.GetNumberOfLabelSuggestionsForImage(imageId)
	ok(t, err)
	equals(t, numAfter, 1)
}
