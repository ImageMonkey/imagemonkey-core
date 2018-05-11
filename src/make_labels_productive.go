package main

import (
	log "github.com/Sirupsen/logrus"
	"flag"
	"database/sql"
)

var db *sql.DB

func trendingLabelExists(label string, tx *sql.Tx) (bool, error) {
	var numOfRows int32
	err := tx.QueryRow(`SELECT COUNT(*) FROM trending_label_suggestion t 
			  			JOIN label_suggestion l ON t.label_suggestion_id = l.id`).Scan(&numOfRows)
	if err != nil {
		return false, err
	}

	if numOfRows > 0 {
		return true, nil
	}

	return false, nil
}

func labelExists(label string, tx *sql.Tx) (bool, error) {
	var numOfRows int32
	err := tx.QueryRow(`SELECT COUNT(*) FROM label l 
			  			WHERE l.name = $1 AND l.parent_id is null`, label).Scan(&numOfRows)
	if numOfRows > 0 {
		return true, err
	}

	return false, err
}

func removeTrendingLabelEntries(trendingLabel string, tx *sql.Tx) (error) {
	_, err := tx.Exec(`DELETE FROM image_label_suggestion s
					   WHERE s.label_suggestion_id IN (
					   	SELECT l.id FROM label_suggestion l WHERE l.name = $1
					   )`, trendingLabel)

	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE trending_label_suggestion t 
				 	  SET closed = true
				 	  FROM label_suggestion AS l 
				 	  WHERE label_suggestion_id = l.id AND l.name = $1`, trendingLabel)
	
	return err
}


func makeTrendingLabelProductive(label LabelMeEntry, tx *sql.Tx) error {
	type Result struct {
		ImageId string
    	Annotatable bool
	}

    rows, err := tx.Query(`SELECT i.key, annotatable FROM trending_label_suggestion t 
			  			   JOIN label_suggestion l ON t.label_suggestion_id = l.id
			  			   JOIN image_label_suggestion isg on isg.label_suggestion_id =l.id
			  			   JOIN image i on i.id = isg.image_id
			  			   WHERE l.name = $1`, label.Label)
    if err != nil {
    	return err
    }

    defer rows.Close()

    //due to a bug in the pq driver (see https://github.com/lib/pq/issues/81)
    //we need to first close the rows before we can execute another transaction.
    //that means we need to store the result set temporarily in a list
	var results []Result
    for rows.Next() {
    	var result Result

    	err = rows.Scan(&result.ImageId, &result.Annotatable)
    	if err != nil { 
    		return err
    	}

    	results = append(results, result)
    }

    rows.Close()

    var labels []LabelMeEntry
	labels = append(labels, label)
    for _, elem := range results {
    	if elem.Annotatable {
			_, err = _addLabelsToImage("", elem.ImageId, labels, 0, 0, tx)  
			if err != nil {
				return err
			} 	
		} else {
			//if label is not annotatable, set num_of_not_annotatable to 10
			_, err = _addLabelsToImage("", elem.ImageId, labels, 0, 10, tx)
			if err != nil {
				return err
			}
		}
	}

    return nil
}

func makeLabelMeEntry(name string, annotatable bool, sublabels []string) LabelMeEntry {
	var label LabelMeEntry
	label.Label = name 
    label.Annotatable = annotatable
    label.Sublabels = sublabels

    return label
}

func isLabelInLabelsMap(labelMap map[string]LabelMapEntry, label LabelMeEntry) bool {
	return isLabelValid(labelMap, label.Label, label.Sublabels)
}


func main() {
	log.SetLevel(log.DebugLevel)

	trendingLabel := flag.String("trendinglabel", "", "The name of the trending label that should be made productive")
	wordlistPath := flag.String("wordlist", "../wordlists/en/labels.json", "Path to label map")
	dryRun := flag.Bool("dryrun", true, "Specifies whether this is a dryrun or not")
	flag.Parse()

	labelMap, _, err := getLabelMap(*wordlistPath)
	if err != nil {
		log.Error("[Main] Couldn't read label map...terminating!")
		return
	}

	if *trendingLabel == "" {
		log.Error("[Main] Please provide a trending label!")
		return
	}

	db, err = sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
	if err != nil {
		log.Error(err)
		return
	}

	err = db.Ping()
	if err != nil {
		log.Error("[Main] Couldn't ping database: ", err.Error())
		return
	}
	defer db.Close()


	tx, err := db.Begin()
    if err != nil {
    	log.Error("[Main] Couldn't begin trensaction: ", err.Error())
        return
    }

	exists, err := trendingLabelExists(*trendingLabel, tx)
	if err != nil {
		tx.Rollback()
		log.Error("[Main] Couldn't determine whether trending label exists: ", err.Error())
		return
	}
	if !exists {
		tx.Rollback()
		log.Error("[Main] Trending label doesn't exist. Maybe a typo?")
		return
	}

	exists, err = labelExists(*trendingLabel, tx)
	if err != nil {
		tx.Rollback()
		log.Error("[Main] Couldn't determine whether label exists: ", err.Error())
		return
	}
	if !exists {
		tx.Rollback()
		log.Error("[Main] label doesn't exist in database - please add it via the populate_labels script.")
		return
	}


	labelMeEntry := makeLabelMeEntry(*trendingLabel, true, []string{})
	if !isLabelInLabelsMap(labelMap, labelMeEntry) {
		tx.Rollback()
		log.Error("[Main] Label doesn't exist in labels map - please add it first!")
		return
	}

	err = makeTrendingLabelProductive(labelMeEntry, tx)
	if err != nil {
		tx.Rollback()
		log.Error("[Main] Couldn't make trending label ", *trendingLabel, " productive: ", err.Error())
		return
	}

	err = removeTrendingLabelEntries(*trendingLabel, tx)
    if err != nil {
    	tx.Rollback()
    	log.Error("[Main] Couldn't remove trending label entries: ", err.Error())
    	return
    }

    if *dryRun {
    	err = tx.Rollback()
		if err != nil {
			log.Error("[Main] Couldn't rollback transaction: ", err.Error())
			return
		}

		log.Info("[Main] This was just a dry run - rolling back everything")
    } else {
		err = tx.Commit()
		if err != nil {
			log.Error("[Main] Couldn't commit transaction: ", err.Error())
			return
		}

		log.Info("[Main] Label ", *trendingLabel, " was successfully made productive. You can now close the corresponding github issue!")
	}
}