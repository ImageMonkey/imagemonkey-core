package tests

import (
	"testing"
	commons "github.com/bbernhard/imagemonkey-core/commons"
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

func TestLabelAndMetalabelUuidsShouldBeUnique(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	labelRepository := commons.NewLabelRepository()
	err := labelRepository.Load("../wordlists/en/labels.jsonnet")
	ok(t, err)
	labels := labelRepository.GetMapping()

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

	metalabels := commons.NewMetaLabels("../wordlists/en/metalabels.jsonnet")
	err = metalabels.Load()
	ok(t, err)
	m := metalabels.GetMapping()
	for _, val := range m.MetaLabelMapEntries {
		if _, ok := uuids[val.Uuid]; ok {
			t.Errorf("Found a duplicate UUID '%s'", val.Uuid)
		} else {
			uuids[val.Uuid] = true
		}
	}
} 
