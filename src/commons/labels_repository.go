package commons

import (
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	"github.com/gofrs/uuid"
)

type LabelsRepositoryInterface interface {
	SetToken(token string)
	Clone() error
	AddLabelAndPushToRepo(trendingLabel datastructures.TrendingLabelBotTask) (string, error) 
	MergeRemoteBranchIntoMaster(branchName string) error
	RemoveLocal() error 
}


func generateLabelEntry(name string, plural string, description string) (datastructures.LabelMapEntry, error) {
	var labelMapEntry datastructures.LabelMapEntry
	labelMapEntry.Plural = plural
	labelMapEntry.Description = description
	labelMapEntry.Accessors = append(labelMapEntry.Accessors, ".")

	var err error
	labelMapEntry.Uuid, err = getUuidV4() 
	if err != nil {
		return labelMapEntry, err
	}

	return labelMapEntry, nil
}

func generateMetaLabelEntry(name string, plural string, description string) (datastructures.MetaLabelMapEntry, error) {
	var metaLabelMapEntry datastructures.MetaLabelMapEntry
	metaLabelMapEntry.Description = description
	metaLabelMapEntry.Accessors = append(metaLabelMapEntry.Accessors, ".")	

	var err error
	metaLabelMapEntry.Uuid, err = getUuidV4()
	if err != nil {
		return metaLabelMapEntry, err
	}

	return metaLabelMapEntry, nil
}

func getUuidV4() (string, error) {
	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	return u.String(), nil

	/*rows, err := db.Query(`SELECT * from uuid_generate_v4()`)
	if err != nil {
		return "", err
	}

	defer rows.Close()

	if rows.Next() {
		var u string
		err = rows.Scan(&u)
		if err != nil {
			return "", err
		}
		return u, nil
	}

	return "", errors.New("Couldn't get UUID")*/
}
