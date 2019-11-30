package tests

import (
	"testing"
	"time"
	"gopkg.in/resty.v1"
	commons "github.com/bbernhard/imagemonkey-core/commons"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
)

func testImageHuntDonate(t *testing.T, path string, label string, token string, expectedStatusCode int) {
	numBefore, err := db.GetNumberOfImages()
	ok(t, err)

	url := BASE_URL + API_VERSION + "/games/imagehunt/donate"

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

    equals(t, resp.StatusCode(), expectedStatusCode)

    if expectedStatusCode == 200 {
	    numAfter, err := db.GetNumberOfImages(); 
	    ok(t, err)

	    equals(t, numAfter, numBefore + 1)
	}
}

func testGetImageHuntStats(t *testing.T, username string, token string, expectedStatusCode int) datastructures.ImageHuntStats {
	var imageHuntStats datastructures.ImageHuntStats

	url := BASE_URL + API_VERSION + "/user/" + username + "/games/imagehunt/stats"

	req := resty.R().
			SetResult(&imageHuntStats)

	if token != "" {
		req.SetAuthToken(token)
	}

	resp, err := req.Get(url)

    equals(t, resp.StatusCode(), expectedStatusCode)
    ok(t, err)

    return imageHuntStats
}

func TestImageHuntGameDonation(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "user", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "user", 200)

	testImageHuntDonate(t, "./images/apples/apple1.jpeg", "apple", userToken, 200)

	imageId, err := db.GetLatestDonatedImageId()
	ok(t, err)

	numImageHuntTasks, err := db.GetNumberOfImageHuntTasksForImageWithLabel(imageId, "apple")
	ok(t, err)
	equals(t, int(numImageHuntTasks), int(1))

	numImageUserEntries, err := db.GetNumberOfImageUserEntriesForImageAndUser(imageId, "user")
	ok(t, err)
	equals(t, int(numImageUserEntries), int(1))
}

func TestImageHuntGameDonationShouldFailOnNonProductiveLabel(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "user", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "user", 200)

	testImageHuntDonate(t, "./images/apples/apple1.jpeg", "not-existing-label", userToken, 400)
}

func TestImageHuntGameDonationShouldFailDueToUnauthenticatedUser(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testImageHuntDonate(t, "./images/apples/apple1.jpeg", "apple", "", 403)
}

func TestImageHuntGameDonationShouldFailDueToWrongToken(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testImageHuntDonate(t, "./images/apples/apple1.jpeg", "apple", "not-existing-token", 403)
}

func TestGetImageHuntStats(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "user", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "user", 200)

	testImageHuntDonate(t, "./images/apples/apple1.jpeg", "apple", userToken, 200)

	stats := testGetImageHuntStats(t, "user", userToken, 200)
	equals(t, stats.Stars, 1)
}

func TestGetImageHuntStatsShouldFailDueToWrongToken(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "user", "user", "user@imagemonkey.io")
	userToken := testLogin(t, "user", "user", 200)

	testSignUp(t, "user1", "user1", "user1@imagemonkey.io")
	userToken1 := testLogin(t, "user1", "user1", 200)

	testImageHuntDonate(t, "./images/apples/apple1.jpeg", "apple", userToken, 200)

	testGetImageHuntStats(t, "user", userToken1, 403)
}

func TestAchievementBadgesNoAchievements(t *testing.T) {
	achievementsGenerator := commons.NewAchievementsGenerator()
	achievementsGenerator.Add(time.Now())

	achievements, err := achievementsGenerator.GetAchievements("")
	ok(t, err)
	for _, achievement := range achievements {
		equals(t, achievement.Accomplished, false)
	}
}

func TestAchievementBadgesWeekendWarrior(t *testing.T) {
	achievementsGenerator := commons.NewAchievementsGenerator()
	achievementsGenerator.SetNumOfAvailableLabels(15)

	//three consecutive weekends in a row	
	t1 := time.Date(2019, time.November, 2, 23, 0, 0, 0, time.UTC)
	t2 := time.Date(2019, time.November, 9, 14, 0, 0, 0, time.UTC)
	t3 := time.Date(2019, time.November, 16, 17, 0, 0, 0, time.UTC)
	achievementsGenerator.Add(t1)
	achievementsGenerator.Add(t2)
	achievementsGenerator.Add(t3)

	achievements, err := achievementsGenerator.GetAchievements("")
	ok(t, err)
	for _, achievement := range achievements {
		if achievement.Name == "Weekend Warrior" {
			equals(t, achievement.Accomplished, true)
		} else {
			equals(t, achievement.Accomplished, false)
		}
	}
}


func TestAchievementBadgesWeekendWarrior1(t *testing.T) {
	achievementsGenerator := commons.NewAchievementsGenerator()
	achievementsGenerator.SetNumOfAvailableLabels(15)

	//two consecutive weekends in a row, then a gap
	t1 := time.Date(2019, time.November, 2, 23, 0, 0, 0, time.UTC)
	t2 := time.Date(2019, time.November, 9, 14, 0, 0, 0, time.UTC)
	t3 := time.Date(2019, time.November, 15, 17, 0, 0, 0, time.UTC)
	achievementsGenerator.Add(t1)
	achievementsGenerator.Add(t2)
	achievementsGenerator.Add(t3)

	achievements, err := achievementsGenerator.GetAchievements("")
	ok(t, err)
	for _, achievement := range achievements {
		equals(t, achievement.Accomplished, false)
	}
}

func TestAchievementBadgesWeekendWarrior2(t *testing.T) {
	achievementsGenerator := commons.NewAchievementsGenerator()
	achievementsGenerator.SetNumOfAvailableLabels(15)

	//two consecutive weekends in a row, then a gap and three consecutive weekens in a row
	t1 := time.Date(2019, time.November, 2, 23, 0, 0, 0, time.UTC)
	t2 := time.Date(2019, time.November, 9, 14, 0, 0, 0, time.UTC)
	t3 := time.Date(2019, time.November, 15, 17, 0, 0, 0, time.UTC)
	t4 := time.Date(2019, time.November, 23, 16, 15, 18, 0, time.UTC)
	t5 := time.Date(2019, time.November, 30, 17, 10, 0, 0, time.UTC)
	t6 := time.Date(2019, time.December, 7, 17, 20, 0, 0, time.UTC)
	achievementsGenerator.Add(t1)
	achievementsGenerator.Add(t2)
	achievementsGenerator.Add(t3)
	achievementsGenerator.Add(t4)
	achievementsGenerator.Add(t5)
	achievementsGenerator.Add(t6)

	achievements, err := achievementsGenerator.GetAchievements("")
	ok(t, err)
	for _, achievement := range achievements {
		if achievement.Name == "Weekend Warrior" {
			equals(t, achievement.Accomplished, true)
		} else {
			equals(t, achievement.Accomplished, false)
		}
	}
}

func TestAchievementBadgesWeekendWarrior3(t *testing.T) {
	achievementsGenerator := commons.NewAchievementsGenerator()
	achievementsGenerator.SetNumOfAvailableLabels(15)

	//three consecutive  weekends in a row, then a gap
	t1 := time.Date(2019, time.November, 2, 23, 0, 0, 0, time.UTC)
	t2 := time.Date(2019, time.November, 9, 14, 0, 0, 0, time.UTC)
	t3 := time.Date(2019, time.November, 16, 17, 0, 0, 0, time.UTC)
	t4 := time.Date(2019, time.December, 10, 16, 15, 18, 0, time.UTC)
	achievementsGenerator.Add(t1)
	achievementsGenerator.Add(t2)
	achievementsGenerator.Add(t3)
	achievementsGenerator.Add(t4)

	achievements, err := achievementsGenerator.GetAchievements("")
	ok(t, err)
	for _, achievement := range achievements {
		if achievement.Name == "Weekend Warrior" {
			equals(t, achievement.Accomplished, true)
		} else {
			equals(t, achievement.Accomplished, false)
		}
	}
}

func TestAchievementBadgesWorkerBee(t *testing.T) {
	achievementsGenerator := commons.NewAchievementsGenerator()
	achievementsGenerator.SetNumOfAvailableLabels(15)

	//worker bee = at least 7 consecutive days
	for i := 0; i < 7; i++ {
		t := time.Date(2019, time.February, 25+i, 23, 0, 0, 0, time.UTC)
		achievementsGenerator.Add(t)
	}

	achievements, err := achievementsGenerator.GetAchievements("")
	ok(t, err)
	for _, achievement := range achievements {
		if achievement.Name == "Worker Bee" {
			equals(t, achievement.Accomplished, true)
		} else {
			equals(t, achievement.Accomplished, false)
		}
	}
}

func TestAchievementBadgesWorkerBee1(t *testing.T) {
	achievementsGenerator := commons.NewAchievementsGenerator()
	achievementsGenerator.SetNumOfAvailableLabels(15)

	for i := 0; i < 6; i++ {
		t := time.Date(2019, time.February, 25+i, 23, 0, 0, 0, time.UTC)
		achievementsGenerator.Add(t)
	}

	//no worker bee if < 7 consecutive days
	achievements, err := achievementsGenerator.GetAchievements("")
	ok(t, err)
	for _, achievement := range achievements {
		equals(t, achievement.Accomplished, false)
	}
}

func TestAchievementBadgesWorkerBee2(t *testing.T) {
	achievementsGenerator := commons.NewAchievementsGenerator()
	achievementsGenerator.SetNumOfAvailableLabels(15)

	//3 consecutive days
	for i := 0; i < 3; i++ {
		t := time.Date(2019, time.February, 25+i, 23, 0, 0, 0, time.UTC)
		achievementsGenerator.Add(t)
	}

	//5 consecutive days
	for i := 0; i < 5; i++ {
		t := time.Date(2019, time.February, 10+i, 23, 0, 0, 0, time.UTC)
		achievementsGenerator.Add(t)
	}

	//but still no worker bee
	achievements, err := achievementsGenerator.GetAchievements("")
	ok(t, err)
	for _, achievement := range achievements {
		equals(t, achievement.Accomplished, false)
	}
}

func TestAchievementBadgesEarlyBird1(t *testing.T) {
	achievementsGenerator := commons.NewAchievementsGenerator()
	achievementsGenerator.SetNumOfAvailableLabels(15)

	//3 consecutive days (but wrong time)
	for i := 0; i < 3; i++ {
		t := time.Date(2019, time.February, 25+i, 23, 0, 0, 0, time.UTC)
		achievementsGenerator.Add(t)
	}

	//so still no early bird
	achievements, err := achievementsGenerator.GetAchievements("")
	ok(t, err)
	for _, achievement := range achievements {
		equals(t, achievement.Accomplished, false)
	}
}

func TestAchievementBadgesEarlyBird2(t *testing.T) {
	achievementsGenerator := commons.NewAchievementsGenerator()
	achievementsGenerator.SetNumOfAvailableLabels(15)

	//3 consecutive days (right time)
	for i := 0; i < 3; i++ {
		t := time.Date(2019, time.February, 25+i, 5, 10+i, 0, 0, time.UTC)
		achievementsGenerator.Add(t)
	}

	//so we are an early bird
	achievements, err := achievementsGenerator.GetAchievements("")
	ok(t, err)
	for _, achievement := range achievements {
		if achievement.Name == "Early Bird" {
			equals(t, achievement.Accomplished, true)
		} else {
			equals(t, achievement.Accomplished, false)
		}
	}
}


func TestAchievementBadgesEarlyBird3(t *testing.T) {
	achievementsGenerator := commons.NewAchievementsGenerator()
	achievementsGenerator.SetNumOfAvailableLabels(15)

	//4 consecutive days (but last two have wrong time)
	for i := 0; i < 4; i++ {
		var tt time.Time
		if i == 2 || i == 3 {
			tt = time.Date(2019, time.February, 25+i, 10, 0, 0, 0, time.UTC)
		} else {
			tt = time.Date(2019, time.February, 25+i, 5, 10+i, 0, 0, time.UTC)
		}
		achievementsGenerator.Add(tt)
	}

	//so still no early bird
	achievements, err := achievementsGenerator.GetAchievements("")
	ok(t, err)
	for _, achievement := range achievements {
		equals(t, achievement.Accomplished, false)
	}
}

func TestAchievementBadgesNightOwl1(t *testing.T) {
	achievementsGenerator := commons.NewAchievementsGenerator()
	achievementsGenerator.SetNumOfAvailableLabels(15)

	//3 consecutive days (but wrong time)
	for i := 0; i < 3; i++ {
		t := time.Date(2019, time.February, 25+i, 15, 0, 0, 0, time.UTC)
		achievementsGenerator.Add(t)
	}

	//so still no night owl
	achievements, err := achievementsGenerator.GetAchievements("")
	ok(t, err)
	for _, achievement := range achievements {
		equals(t, achievement.Accomplished, false)
	}
}

func TestAchievementBadgesNightOwl2(t *testing.T) {
	achievementsGenerator := commons.NewAchievementsGenerator()
	achievementsGenerator.SetNumOfAvailableLabels(15)

	//3 consecutive days (right time)
	for i := 0; i < 3; i++ {
		t := time.Date(2019, time.February, 25+i, 0, 10+i, 0, 0, time.UTC)
		achievementsGenerator.Add(t)
	}

	//so we are an night owl
	achievements, err := achievementsGenerator.GetAchievements("")
	ok(t, err)
	for _, achievement := range achievements {
		if achievement.Name == "Night Owl" {
			equals(t, achievement.Accomplished, true)
		} else {
			equals(t, achievement.Accomplished, false)
		}
	}
}

func TestAchievementBadgesNightOwl3(t *testing.T) {
	achievementsGenerator := commons.NewAchievementsGenerator()
	achievementsGenerator.SetNumOfAvailableLabels(15)

	//4 consecutive days (but last two have wrong time)
	for i := 0; i < 4; i++ {
		var tt time.Time
		if i == 2 || i == 3 {
			tt = time.Date(2019, time.February, 25+i, 0, 40+i, 0, 0, time.UTC)
		} else {
			tt = time.Date(2019, time.February, 25+i, 5, 10+i, 0, 0, time.UTC)
		}
		achievementsGenerator.Add(tt)
	}

	//so still no night owl
	achievements, err := achievementsGenerator.GetAchievements("")
	ok(t, err)
	for _, achievement := range achievements {
		equals(t, achievement.Accomplished, false)
	}
}

func TestAchievementBadges1(t *testing.T) {
	achievementsGenerator := commons.NewAchievementsGenerator()
	achievementsGenerator.SetNumOfAvailableLabels(80)


	for i := 0; i < 60; i++ {
		var tt time.Time
		if i < 10 {
			tt = time.Date(2019, time.February, 25, 0, 0+i, 0, 0, time.UTC)
		} else {
			tt = time.Date(2019, time.February, 25+i, 15, 0+i, 0, 0, time.UTC)
		}
		achievementsGenerator.Add(tt)
	}

	achievements, err := achievementsGenerator.GetAchievements("")
	ok(t, err)
	for _, achievement := range achievements {
		if achievement.Name == "Ant Power" || achievement.Name == "Worker Bee" {
			equals(t, achievement.Accomplished, true)
		} else {
			equals(t, achievement.Accomplished, false)
		}
	}
}


