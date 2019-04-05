package tests

import (
	"testing"
	"encoding/json"
	"reflect"
	"gopkg.in/resty.v1"
	"../src/datastructures"
	"strconv"
)

func testGetExistingAnnotationsForAnnotationId(t *testing.T, token string, annotationId string, revision int, expectedAnnotations string) {
	url := BASE_URL + API_VERSION + "/annotation?annotation_id=" + annotationId

	if revision != -1 {
		url += "&rev=" + strconv.Itoa(revision)
	}

	req := resty.R().
				SetHeader("Content-Type", "application/json").
				SetResult(&datastructures.AnnotatedImage{})
	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(url)
	ok(t, err)

	equals(t, resp.StatusCode(), 200)

	j, err := json.Marshal(&resp.Result().(*datastructures.AnnotatedImage).Annotations)
	ok(t, err)

	equal, err := equalJson(string(j), expectedAnnotations)
	equals(t, equal, true)
}

func testAnnotationRework(t *testing.T, annotationId string, annotations string) {
	type Annotation struct {
		Annotations []json.RawMessage `json:"annotations"`
	}

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

func testRandomAnnotationRework(t *testing.T, num int, annotations string) {
	for i := 0; i < num; i++ {
		annotationId, err := db.GetRandomAnnotationId()
		ok(t, err)

		testAnnotationRework(t, annotationId, annotations)
	}
}

func TestRandomAnnotationRework(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t, "apple")
	testRandomAnnotate(t, 2, `[{"top":100,"left":200,"type":"rect","angle":0,"width":40,"height":60,"stroke":{"color":"red","width":1}}]`)
	testRandomAnnotationRework(t, 2, `[{"top":200,"left":300,"type":"rect","angle":10,"width":50,"height":30,"stroke":{"color":"blue","width":3}}]`)
}


//this test is flaky and fails sometimes due to the order in which the data gets returned by PostgreSQL
/*func TestAnnotationRework(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)
	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
						`[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`, "", 201)

	annotationId, err := db.GetLastAddedAnnotationId()
	ok(t, err)

	newAnnotations := `[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"refinements":[{"label_uuid":"86485bae-04a1-43ef-a191-5f2a0464595a"},{"label_uuid":"07d13f17-3757-45c5-ba20-4f53c8a46334"}]},{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`
	expectedAnnotations := `[{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"refinements":[{"label_uuid":"07d13f17-3757-45c5-ba20-4f53c8a46334"},{"label_uuid":"86485bae-04a1-43ef-a191-5f2a0464595a"}]},{"top":50,"left":300,"type":"rect","angle":15,"width":240,"height":100,"stroke":{"color":"red","width":1}}]`

	testAnnotationRework(t, annotationId, newAnnotations)

	testGetExistingAnnotationsForAnnotationId(t, "", annotationId, -1, expectedAnnotations)

	testGetExistingAnnotationsForAnnotationId(t, "", annotationId, 2, newAnnotations)
}*/