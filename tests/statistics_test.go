package tests

import (
	"testing"
	"gopkg.in/resty.v1"
	"../src/datastructures"
)

func testGetAnnotatedStatistics(t *testing.T, token string, requiredNumOfResults int) []datastructures.AnnotatedStat {
	var annotationStatistics []datastructures.AnnotatedStat

	u := BASE_URL + API_VERSION + "/statistics/annotated"
	req := resty.R().
				SetResult(&annotationStatistics)

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(u)

	ok(t, err)
    equals(t, resp.StatusCode(), 200)

    equals(t, len(annotationStatistics), requiredNumOfResults)

    return annotationStatistics
}

func TestGetAnnotatedStatistics(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":60,"left":145,"type":"rect","angle":0,"width":836,"height":660,"stroke":{"color":"red","width":5}}]`, "")


	annotationStatistics := testGetAnnotatedStatistics(t, "", 1)

	annotationStatistics[0].Label.Name = "apple"
	annotationStatistics[0].Num.Completed = 1
	annotationStatistics[0].Num.Total = 1
}

func TestGetAnnotatedStatistics1(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "")

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	testLabelImage(t, imageId, "orange")

	testAnnotate(t, imageId, "apple", "", 
					`[{"top":60,"left":145,"type":"rect","angle":0,"width":836,"height":660,"stroke":{"color":"red","width":5}}]`, "")


	annotationStatistics := testGetAnnotatedStatistics(t, "", 2)

	annotationStatistics[0].Label.Name = "orange"
	annotationStatistics[0].Num.Completed = 0
	annotationStatistics[0].Num.Total = 1


	annotationStatistics[1].Label.Name = "apple"
	annotationStatistics[1].Num.Completed = 1
	annotationStatistics[1].Num.Total = 1

}