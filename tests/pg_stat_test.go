package tests

import (
	"testing"
	"gopkg.in/resty.v1"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
)

func testGetPgStats(t *testing.T, clientId string, clientSecret string, token string, requiredStatusCode int) []datastructures.PgStatStatementResult {
	u := BASE_URL + API_VERSION + "/internal/statistics/pg"
	var res []datastructures.PgStatStatementResult

	req := resty.R().
		    SetResult(&res)

	if token != "" {
		req.SetAuthToken(token)
	}

	if clientId != "" {
		req.SetHeader("X-Client-Id", clientId)
	}

	if clientSecret != "" {
		req.SetHeader("X-Client-Secret", clientSecret)
	}

	resp, err := req.Get(u)
	
	ok(t, err)
	equals(t, resp.StatusCode(), requiredStatusCode)
	
	return res
}

func TestGetPgStatsShouldFailDueToMissingPermissions(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
	
	testGetPgStats(t, X_CLIENT_ID, X_CLIENT_SECRET, "", 403)
}

func TestGetPgStatsShouldSucceed(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	moderatorToken := testLogin(t, "moderator", "moderator", 200)

	err := db.GiveUserModeratorRights("moderator")
	ok(t, err)
	
	testGetPgStats(t, X_CLIENT_ID, X_CLIENT_SECRET, moderatorToken, 200)
}
