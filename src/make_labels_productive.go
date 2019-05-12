package main

import (
	log "github.com/Sirupsen/logrus"
	"flag"
	"github.com/satori/go.uuid"
	"database/sql"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"context"
	datastructures "./datastructures"
	commons "./commons"
	imagemonkeydb "./database"

)

var db *sql.DB

func trendingLabelExists(label string, tx *sql.Tx) (bool, error) {
	var numOfRows int32
	err := tx.QueryRow(`SELECT COUNT(*) FROM trending_label_suggestion t 
			  			RIGHT JOIN label_suggestion l ON t.label_suggestion_id = l.id
			  			WHERE l.name = $1`, label).Scan(&numOfRows)
	if err != nil {
		return false, err
	}

	if numOfRows > 0 {
		return true, nil
	}

	return false, nil
}

func getLabelId(labelIdentifier string, tx *sql.Tx) (int64, error) {
	var labelId int64 = -1
	var err error
	var rows *sql.Rows

	_, err = uuid.FromString(labelIdentifier)
	if err == nil { //is UUID
		rows, err = tx.Query(`SELECT l.id 
						   FROM label l 
			  			   WHERE l.uuid::text = $1`, labelIdentifier)
	} else {
		rows, err = tx.Query(`SELECT l.id 
						   FROM label l 
			  			   WHERE l.name = $1 AND l.parent_id is null`, labelIdentifier)
	}

	
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&labelId)
		if err != nil {
			return -1, err
		}
	}
	return labelId, nil
}

func getLabelMeEntryFromUuid(tx *sql.Tx, uuid string) (datastructures.LabelMeEntry, error) {
	var labelMeEntry datastructures.LabelMeEntry

	rows, err := tx.Query(`SELECT l.name, COALESCE(pl.name, '')
			   				FROM label l
			   				LEFT JOIN label pl ON pl.id = l.parent_id
			   				WHERE l.uuid::text = $1`, uuid)
	if err != nil {
		return labelMeEntry, err
	}
	defer rows.Close()

	var label1 string
	var label2 string
	if rows.Next() {
		err = rows.Scan(&label1, &label2)
		if err != nil {
			return labelMeEntry, err
		}

		if label2 == "" {
            labelMeEntry.Label = label1
        } else {
            labelMeEntry.Label = label2
            labelMeEntry.Sublabels = append(labelMeEntry.Sublabels, 
            								datastructures.Sublabel{Name: label1})
        } 
	}

	return labelMeEntry, nil
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

func closeGithubIssue(trendingLabel string, repository string, tx *sql.Tx) error {
	rows, err := tx.Query(`SELECT t.github_issue_id, t.closed
							FROM trending_label_suggestion t
							JOIN label_suggestion l ON t.label_suggestion_id = l.id
							WHERE l.name = $1`, trendingLabel)
	if err != nil {
		return err
	}

	defer rows.Close()

	if rows.Next() {
		var githubIssueId int
		var issueClosed bool

		err = rows.Scan(&githubIssueId, &issueClosed)
		if err != nil {
			return err
		}

		//if !issueClosed {
			ctx := context.Background()
			ts := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: GITHUB_API_TOKEN},
			)
			tc := oauth2.NewClient(ctx, ts)

			client := github.NewClient(tc)

			body := "label is now productive."

			//create a new comment
			commentRequest := &github.IssueComment{
				Body:    github.String(body),
			}

			//we do not care whether we can successfully close the github issue..if it doesn't work, one can always close it
			//manually.
			_, _, err = client.Issues.CreateComment(ctx, GITHUB_PROJECT_OWNER, repository, githubIssueId, commentRequest)
			if err == nil { //if comment was successfully created, close issue
				issueRequest := &github.IssueRequest{
					State: github.String("closed"),
				}

				_, _, err = client.Issues.Edit(ctx, GITHUB_PROJECT_OWNER, repository, githubIssueId, issueRequest)
				if err != nil {
					log.Info("[Main] Couldn't close github issue, please close manually!")
				}
			} else {
				log.Info("[Main] Couldn't close github issue, please close manually!")
			}
		//}
	}

	return nil
}


func makeTrendingLabelProductive(trendingLabel string, label datastructures.LabelMeEntry,
									labelId int64, tx *sql.Tx) error {
	type Result struct {
		ImageId string
    	Annotatable bool
	}

    rows, err := tx.Query(`SELECT i.key, annotatable 
			  			   FROM label_suggestion l
			  			   JOIN image_label_suggestion isg on isg.label_suggestion_id =l.id
			  			   JOIN image i on i.id = isg.image_id
			  			   WHERE l.name = $1`, trendingLabel)
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

    var labels []datastructures.LabelMeEntry
    labels = append(labels, label)

    for _, elem := range results {
		numNonAnnotatable := 0
    	if !elem.Annotatable {
			numNonAnnotatable = 10
		}

		_, err = imagemonkeydb.AddLabelsToImageInTransaction("", elem.ImageId, labels, 0, numNonAnnotatable, tx)  
		if err != nil {
			return err
		} 	
	}

	_, err = tx.Exec(`UPDATE trending_label_suggestion t
						SET productive_label_id = $2
						FROM label_suggestion l
						WHERE t.label_suggestion_id = l.id AND l.name = $1`, trendingLabel, labelId)
	if err != nil {
		return err
	}

    return nil
}

func makeLabelMeEntry(tx *sql.Tx, name string) (datastructures.LabelMeEntry, error) {
    _, err := uuid.FromString(name)
    if err == nil { //is UUID
    	entry, err := getLabelMeEntryFromUuid(tx, name)
    	if err != nil {
    		return datastructures.LabelMeEntry{}, err //UUID is not in database
    	}
    	return entry, nil
    }
    
	//not a UUID	
	var label datastructures.LabelMeEntry
	label.Label = name 
	label.Sublabels = []datastructures.Sublabel{}

    return label, nil
}

func isLabelInLabelsMap(labelMap map[string]datastructures.LabelMapEntry, metalabels *commons.MetaLabels, label datastructures.LabelMeEntry) bool {
	return commons.IsLabelValid(labelMap, metalabels, label.Label, label.Sublabels)
}


func main() {
	log.SetLevel(log.DebugLevel)

	trendingLabel := flag.String("trendinglabel", "", "The name of the trending label that should be made productive")
	renameTo := flag.String("renameto", "", "Rename the label")
	wordlistPath := flag.String("wordlist", "../wordlists/en/labels.jsonnet", "Path to label map")
	metalabelsPath := flag.String("metalabels", "../wordlists/en/metalabels.json", "Path to metalabels map")
	dryRun := flag.Bool("dryrun", true, "Specifies whether this is a dryrun or not")
	autoCloseIssue := flag.Bool("autoclose", true, "Automatically close issue")
	githubRepository := flag.String("repository", "", "Github repository")

	flag.Parse()

	if *autoCloseIssue && *githubRepository == "" {
		log.Fatal("Please set a valid repository!")
	}

	labelRepository := commons.NewLabelRepository()
	err := labelRepository.Load(*wordlistPath)
	if err != nil {
		log.Error("[Main] Couldn't read label map...terminating!")
		return
	}
	labelMap := labelRepository.GetMapping()

	metaLabels := commons.NewMetaLabels(*metalabelsPath)
	err = metaLabels.Load()
	if err != nil {
		log.Error("[Main] Couldn't read metalabel map...terminating!")
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

	
	labelToCheck := *trendingLabel
	if *renameTo != "" {
		labelToCheck = *renameTo
	}


	labelId, err := getLabelId(labelToCheck, tx)
	if err != nil {
		tx.Rollback()
		log.Error("[Main] Couldn't determine whether label exists: ", err.Error())
		return
	}
	if labelId == -1 {
		tx.Rollback()
		log.Error("[Main] label doesn't exist in database - please add it via the populate_labels script.")
		return
	}


	labelMeEntry, err := makeLabelMeEntry(tx, labelToCheck)
	if err != nil {
		tx.Rollback()
		log.Error("[Main] Couldn't create label entry - is UUID valid?")
		return
	}

	if !isLabelInLabelsMap(labelMap, metaLabels, labelMeEntry) && *renameTo == "" {
		tx.Rollback()
		log.Error("[Main] Label doesn't exist in labels map - please add it first!")
		return
	}

	err = makeTrendingLabelProductive(*trendingLabel, labelMeEntry, labelId, tx)
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
    	//only handle autoclose, in case it's not a dry run
    	if *autoCloseIssue {
    		err := closeGithubIssue(*trendingLabel, *githubRepository, tx)
    		if err != nil {
    			log.Error("[Main] Couldn't get github issue id to close issue!")
    			tx.Rollback()
    			return
    		}
    	}


		err = tx.Commit()
		if err != nil {
			log.Error("[Main] Couldn't commit transaction: ", err.Error())
			return
		}

		if *renameTo == "" {
			log.Info("[Main] Label ", *trendingLabel, " was successfully made productive. You can now close the corresponding github issue!")
		} else {
			log.Info("[Main] Label ", *trendingLabel, " was renamed to ", *renameTo  ," successfully made productive. You can now close the corresponding github issue!")
		}
	}
}
