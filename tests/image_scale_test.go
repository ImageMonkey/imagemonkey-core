package tests

import (
	"testing"
	"strconv"
	//"math/rand"
	//"time"
	"net/http"
	"image"
)

/*func random(min, max int) int {
    rand.Seed(time.Now().Unix())
    return rand.Intn(max - min) + min
}*/

func testFetchScaledDonation(t *testing.T, imageId string, width int, height int) {
	u := (BASE_URL + API_VERSION + "/donation/" + imageId + "?width=" + strconv.Itoa(width) + 
			"&height=" + strconv.Itoa(height)) 
	
	resp, err := http.Get(u)
    ok(t, err)

    defer resp.Body.Close()

    m, _, err := image.Decode(resp.Body)
    ok(t, err)
    g := m.Bounds()

    // Get height and width
    h := g.Dy()
    w := g.Dx()

	equals(t, h, height)
	equals(t, w, width)
}

func TestImageScale(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testDonate(t, "./images/apples/apple1.jpeg", "apple", true, "", "", 200)

	imageIds, err := db.GetAllImageIds()
	ok(t, err)
	equals(t, len(imageIds), 1)


	testFetchScaledDonation(t, imageIds[0], 100, 100)
}

func TestMultipleRandomImageScale(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testMultipleDonate(t, "apple")

	imageIds, err := db.GetAllImageIds()
	ok(t, err)

	i := 0
	for i < 200 {
		randomImageId := imageIds[random(0, len(imageIds) - 1)]
		randomWidth := random(1, 500)
		randomHeight := random(1, 500)
		testFetchScaledDonation(t, randomImageId, randomWidth, randomHeight)
		i += 1
	}
}
