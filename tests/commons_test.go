package tests

import (
	"testing"
)

func TestDatabaseEmpty(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)
}

func TestDatabaseEmptyWithUserThatHasUnlockImagePermission(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	testSignUp(t, "moderator", "moderator", "moderator@imagemonkey.io")
	testLogin(t, "moderator", "moderator", 200)

	err := db.GiveUserModeratorRights("moderator")
	ok(t, err)

	err = db.GiveUserUnlockImagePermissions("moderator")
	ok(t, err)
}
