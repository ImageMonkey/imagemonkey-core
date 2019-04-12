package tests

import (
	"testing"
	commons "../src/commons"
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

func TestLabelUuidsShouldBeUnique(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	labels, _, err := commons.GetLabelMap("../wordlists/en/labels.jsonnet")
	ok(t, err)

	uuids := make(map[string]bool)
	for _, val := range labels {
		if _, ok := uuids[val.Uuid]; ok {
			t.Errorf("Found a duplicate UUID '%s'", val.Uuid)
		} else {
			uuids[val.Uuid] = true
		}

		for _, hasVal := range val.LabelMapEntries {
			if _, ok := uuids[hasVal.Uuid]; ok {
				t.Errorf("Found a duplicate UUID '%s'", hasVal.Uuid)
			} else {
				uuids[val.Uuid] = true
			}
		}
	}
} 
