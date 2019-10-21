package clients

import (
	"database/sql"
	"golang.org/x/oauth2"
	"errors"
	"context"
	"github.com/gofrs/uuid"
	"github.com/google/go-github/github"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	commons "github.com/bbernhard/imagemonkey-core/commons"
	imagemonkeydb "github.com/bbernhard/imagemonkey-core/database"
	log "github.com/sirupsen/logrus"
	"fmt"
)

type MakeLabelsProductiveClient struct {
	dbConnectionString string
	labelsPath string
	metalabelsPath string
	githubRepository string
	githubRepositoryOwner string
	githubApiToken string
	autoCloseIssue bool
	strict bool
	labels *commons.LabelRepository
	metalabels *commons.MetaLabels
	db *sql.DB
}

func NewMakeLabelsProductiveClient(dbConnectionString string, labelsPath string, metalabelsPath string, strict bool, autoCloseIssue bool) *MakeLabelsProductiveClient {
	return &MakeLabelsProductiveClient {
		dbConnectionString: dbConnectionString,
		autoCloseIssue: autoCloseIssue,
		labelsPath: labelsPath,
		metalabelsPath: metalabelsPath,
		strict: strict,
	}
}

func (p *MakeLabelsProductiveClient) SetGithubRepository(githubRepository string) {
	p.githubRepository = githubRepository
}

func (p *MakeLabelsProductiveClient) SetGithubRepositoryOwner(githubRepositoryOwner string) {
	p.githubRepositoryOwner = githubRepositoryOwner
}

func (p *MakeLabelsProductiveClient) SetGithubApiToken(githubApiToken string) {
	p.githubApiToken = githubApiToken
}

func (p *MakeLabelsProductiveClient) Load() error {
	p.labels = commons.NewLabelRepository(p.labelsPath)
	err := p.labels.Load()
	if err != nil {
		return err
	}
	//labelMap := labelRepository.GetMapping()

	p.metalabels = commons.NewMetaLabels(p.metalabelsPath)
	err = p.metalabels.Load()
	if err != nil {
		return err
	}

	p.db, err = sql.Open("postgres", p.dbConnectionString)
	if err != nil {
		return err
	}

	err = p.db.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (p *MakeLabelsProductiveClient) DoIt(trendingLabel string, renameTo string, dryRun bool) error {
	if p.autoCloseIssue && p.githubRepository == "" {
		return errors.New("Please set a valid repository!")
	}
	
	if trendingLabel == "" {
		return errors.New("Please provide a trending label!")
	}

	tx, err := p.db.Begin()
    if err != nil {
		return errors.New("Couldn't begin trensaction: " + err.Error())
    }

	trendingLabels := []string{}
	if p.strict {
		exists, err := trendingLabelExists(trendingLabel, tx)
		if err != nil {
			tx.Rollback()
			return errors.New("Couldn't determine whether trending label exists: " + err.Error())
		}

		if exists {
			trendingLabels = append(trendingLabels, trendingLabel)
		}
	} else {
		nonStrictLabels, err := getNonStrictLabels(tx, trendingLabel)
		if err != nil {
			tx.Rollback()
			return errors.New("Couldn't get non strict labels: " + err.Error())
		}
		trendingLabels = nonStrictLabels
	}

	if len(trendingLabels) == 0 {
		return errors.New("Trending label doesn't exist. Maybe a typo?")
	}


	labelToCheck := trendingLabel
	if renameTo != "" {
		labelToCheck = renameTo
	}


	labelId, err := _getLabelId(labelToCheck, tx)
	if err != nil {
		tx.Rollback()
		return errors.New("Couldn't determine whether label exists: " + err.Error())
	}
	if labelId == -1 {
		tx.Rollback()
		return errors.New("label " + labelToCheck + " doesn't exist in database - please add it via the populate_labels script.")
	}


	labelMeEntry, err := makeLabelMeEntry(tx, labelToCheck)
	if err != nil {
		tx.Rollback()
		return errors.New("Couldn't create label entry - is UUID valid?")
	}

	if !isLabelInLabelsMap(p.labels.GetMapping(), p.metalabels, labelMeEntry) && renameTo == "" {
		tx.Rollback()
		return errors.New("Label doesn't exist in labels map - please add it first!")
	}

	for _, trendingLabel := range trendingLabels {
		err = makeTrendingLabelProductive(trendingLabel, labelMeEntry, labelId, tx)
		if err != nil {
			tx.Rollback()
			return errors.New("Couldn't make trending label " + trendingLabel + " productive: " + err.Error())
		}

		err = removeTrendingLabelEntries(trendingLabel, tx)
		if err != nil {
			tx.Rollback()
			return errors.New("Couldn't remove trending label entries: " + err.Error())
		}
	}

	err = makeAnnotationsProductive(trendingLabel, labelId, tx)
	if err != nil {
		tx.Rollback()
		return errors.New("Couldn't make annotations productive: " + err.Error())
	}

    if dryRun {
		err = tx.Rollback()
		if err != nil {
			return errors.New("Couldn't rollback transaction: " + err.Error())
		}

		log.Info("This was just a dry run - rolling back everything")
    } else {
		//only handle autoclose, in case it's not a dry run
		if p.autoCloseIssue {
			for _, trendingLabel := range trendingLabels {
				err := closeGithubIssue(trendingLabel, p.githubRepository, p.githubRepositoryOwner, p.githubApiToken, tx)
				if err != nil {
					tx.Rollback()
					return errors.New("Couldn't get github issue id to close issue!")
				}
			}
		}

		err = tx.Commit()
		if err != nil {
			return errors.New("Couldn't commit transaction: " + err.Error())
		}

		autoCloseIssueStr := ""
		if !p.autoCloseIssue {
			autoCloseIssueStr = "You can now close the corresponding github issue!"
		}

		for _, trendingLabel := range trendingLabels {
			if trendingLabel == labelMeEntry.Label {
				log.Info("[Main] Label ", trendingLabel, " was successfully made productive. ", autoCloseIssueStr)
			} else {
				log.Info("[Main] Label ", trendingLabel, " was renamed to ", labelMeEntry.Label, " successfully made productive. ", autoCloseIssueStr)
			}
		}
	}

	return nil
}


func makeAnnotationsProductive(trendingLabel string, labelId int64, tx *sql.Tx) error {
	//safety check
	//the below code was written given a specific database table layout. If someone changes
	//the database schema (e.g add/remove a column), the code below needs to be adapted.
	//so in order to prevent that someone adds/removes a column to these tables, but forgets
	//to change the code below we strictly check for the number of columns here
	rows, err := tx.Query(`SELECT table_name, count(*) as columns 
							FROM information_schema.columns 
                  			WHERE table_name='image_annotation' OR 
				  				  table_name = 'image_annotation_revision' or
				  				  table_name = 'annotation_data' or
				  				  table_name = 'user_image_annotation' or
				  				  table_name = 'image_annotation_suggestion' or
				  				  table_name= 'image_annotation_suggestion_revision' or
				  				  table_name = 'annotation_suggestion_data' or
				  				  table_name = 'user_image_annotation_suggestion'
				  			GROUP BY table_name`)
	if err != nil {
		return err
	}

	defer rows.Close()

	tableToColumnsMapping := make(map[string]int)
	for rows.Next() {
		var name, cols string
		err = rows.Scan(&name, &cols)
		if err != nil {
			return err
		}
	}
	rows.Close()
	

	if tableToColumnsMapping["image_annotation"] != 10 || tableToColumnsMapping["image_annotation_suggestion"] != 10 {
		return errors.New("either the image_annotation or the image_annotation_suggestion table has more columns than expected!")
	}

	if tableToColumnsMapping["annotation_data"] != 6 || tableToColumnsMapping["annotation_suggestion_data"] != 6 {
		return errors.New("either the annotation_data or the annotation_suggestion_data table has more columns than expected!")
	} 

	if tableToColumnsMapping["user_image_annotation"] != 4 || tableToColumnsMapping["user_image_annotation_suggestion"] != 4 {
		return errors.New("either the user_image_annotation or the user_image_annotation_suggestion table has more columns than expected!")
	}
	
	if tableToColumnsMapping["image_annotation_revision"] != 3 || tableToColumnsMapping["image_annotation_suggestion_revision"] != 3 {
		return errors.New("either the image_annotation_revision or the image_annotation_suggestion_revision table has more columns than expected!")
	}


	
	tempTables := []string{"temp_image_annotation_mapping", "temp_annotation_data_mapping", "temp_image_annotation_revision_mapping",
							"temp_annotation_data_revision_mapping"}
	
	for _, tempTable := range tempTables {
		_, err := tx.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS %s`, tempTable)) //controlled input, so no sql injection possible
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`CREATE TEMPORARY TABLE temp_image_annotation_mapping(old_image_annotation_id bigint, new_image_annotation_id bigint)`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO temp_image_annotation_mapping(old_image_annotation_id, new_image_annotation_id)
						SELECT a.id as old_image_annotation_id, nextval('image_annotation_id_seq') as new_image_annotation_id
						FROM image_annotation_suggestion a
						JOIN label_suggestion s ON s.id = a.label_suggestion_id
						WHERE s.name = $1`, trendingLabel)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE TEMPORARY TABLE temp_annotation_data_mapping(old_annotation_data_id bigint, 
								old_image_annotation_id bigint, new_annotation_data_id bigint, new_image_annotation_id bigint,
								old_image_annotation_revision_id bigint)`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO temp_annotation_data_mapping(old_annotation_data_id, old_image_annotation_id, new_annotation_data_id, 
								new_image_annotation_id, old_image_annotation_revision_id) 
						SELECT d.id as old_annotation_data_id, m.old_image_annotation_id as old_image_annotation_id, 
						nextval('image_annotation_data_id_seq') as new_annotation_data_id, m.new_image_annotation_id as new_image_annotation_id,
						d.image_annotation_suggestion_revision_id as old_image_annotation_revision_id
						FROM annotation_suggestion_data d
						JOIN temp_image_annotation_mapping m ON m.old_image_annotation_id = d.image_annotation_suggestion_id
						WHERE d.image_annotation_suggestion_id is not null`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`CREATE TEMPORARY TABLE temp_image_annotation_revision_mapping(old_image_annotation_revision_id bigint, 
								old_image_annotation_id bigint, new_image_annotation_revision_id bigint, new_image_annotation_id bigint)`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO temp_image_annotation_revision_mapping(old_image_annotation_revision_id, old_image_annotation_id, new_image_annotation_revision_id,
							new_image_annotation_id)
						SELECT r.id as old_image_annotation_revision_id, r.image_annotation_suggestion_id as old_image_annotation_id, 
							nextval('image_annotation_revision_id_seq') as new_image_annotation_revision_id, m.new_image_annotation_id as new_image_annotation_id
						FROM image_annotation_suggestion_revision r
						JOIN temp_image_annotation_mapping m ON m.old_image_annotation_id = r.image_annotation_suggestion_id`)
	if err != nil {
		return err
	}


	_, err = tx.Exec(`CREATE TEMPORARY TABLE temp_annotation_data_revision_mapping(old_annotation_data_id bigint, 
								old_image_annotation_id bigint, new_annotation_data_id bigint, new_image_annotation_id bigint,
								old_image_annotation_revision_id bigint)`)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`INSERT INTO temp_annotation_data_revision_mapping(old_annotation_data_id, old_image_annotation_id, new_annotation_data_id, 
								new_image_annotation_id, old_image_annotation_revision_id) 
						SELECT d.id as old_annotation_data_id, m.old_image_annotation_id as old_image_annotation_id, 
						nextval('image_annotation_data_id_seq') as new_annotation_data_id, m.new_image_annotation_id as new_image_annotation_id,
						d.image_annotation_suggestion_revision_id as old_image_annotation_revision_id
						FROM annotation_suggestion_data d
						JOIN temp_image_annotation_revision_mapping m ON m.old_image_annotation_revision_id = d.image_annotation_suggestion_revision_id
						WHERE d.image_annotation_suggestion_id is null`)
	if err != nil {
		return err
	}

	//insert
	var insertedImageAnnotationRows int64 = 0
	err = tx.QueryRow(`WITH rows AS (
							INSERT INTO image_annotation (id, image_id, num_of_valid, num_of_invalid, 
								fingerprint_of_last_modification, sys_period, label_id, uuid, auto_generated, revision)
								SELECT m.new_image_annotation_id, a.image_id, a.num_of_valid, a.num_of_invalid, a.fingerprint_of_last_modification, a.sys_period,
						   		$1, a.uuid, a.auto_generated, a.revision
						   		FROM temp_image_annotation_mapping m
						   		JOIN image_annotation_suggestion a ON a.id = m.old_image_annotation_id
					  			
								RETURNING 1
					   )
					   SELECT count(*) from rows`, labelId).Scan(&insertedImageAnnotationRows)
	if err != nil {
		return err
	}

	var insertedAnnotationDataRows int64 = 0
	err = tx.QueryRow(`WITH rows AS (
							INSERT INTO annotation_data(id, image_annotation_id, annotation, annotation_type_id, image_annotation_revision_id, uuid)
								SELECT m.new_annotation_data_id, m.new_image_annotation_id, d.annotation, 
										d.annotation_type_id, m1.new_image_annotation_revision_id, d.uuid 
								FROM temp_annotation_data_mapping m
								LEFT JOIN temp_image_annotation_revision_mapping m1 ON m1.old_image_annotation_revision_id = m.old_image_annotation_revision_id 
								AND m1.old_image_annotation_id = m.old_image_annotation_id
								JOIN annotation_suggestion_data d ON m.old_annotation_data_id = d.id
								
								RETURNING 1
					   )
					   SELECT count(*) FROM rows`).Scan(&insertedAnnotationDataRows)
	if err != nil {
		return err
	}

	var insertedImageAnnotationRevisionRows int64 = 0
	err = tx.QueryRow(`WITH rows AS (
							INSERT INTO image_annotation_revision(id, image_annotation_id, revision)
								SELECT m.new_image_annotation_revision_id, m.new_image_annotation_id, r.revision 
								FROM temp_image_annotation_revision_mapping m
								JOIN image_annotation_suggestion_revision r ON r.id = m.old_image_annotation_revision_id
						  		
								RETURNING 1
					   )
					   SELECT count(*) FROM rows`).Scan(&insertedImageAnnotationRevisionRows)
	if err != nil {
		return err
	}

	var insertedAnnotationDataRevisionRows int64 = 0
	err = tx.QueryRow(`WITH rows AS (
							INSERT INTO annotation_data(id, image_annotation_id, annotation, annotation_type_id, image_annotation_revision_id, uuid)
								SELECT m.new_annotation_data_id, null, d.annotation, 
										d.annotation_type_id, m1.new_image_annotation_revision_id, d.uuid 
								FROM temp_annotation_data_revision_mapping m
								JOIN temp_image_annotation_revision_mapping m1 ON m1.old_image_annotation_revision_id = m.old_image_annotation_revision_id 
								AND m1.old_image_annotation_id = m.old_image_annotation_id
								JOIN annotation_suggestion_data d ON m.old_annotation_data_id = d.id
								
								RETURNING 1
					   )
					   SELECT count(*) FROM rows`).Scan(&insertedAnnotationDataRevisionRows)
	if err != nil {
		return err
	}

	var insertedUserImageAnnotationRows int64 = 0
	err = tx.QueryRow(`WITH rows AS (
						INSERT INTO user_image_annotation(image_annotation_id, account_id, timestamp)
							SELECT m.new_image_annotation_id, u.account_id, u.timestamp
							FROM temp_image_annotation_mapping m
							JOIN user_image_annotation_suggestion u ON u.image_annotation_suggestion_id = m.old_image_annotation_id
					  
					  		RETURNING 1
					   )
					   SELECT count(*) FROM rows`).Scan(&insertedUserImageAnnotationRows)

	//delete 
	var deletedUserImageAnnotationSuggestionRows int64 = 0
	err = tx.QueryRow(`WITH rows AS (
						DELETE FROM user_image_annotation_suggestion WHERE image_annotation_suggestion_id IN
							(SELECT old_image_annotation_id FROM temp_image_annotation_mapping)
					  	
						RETURNING 1
					   )
					   SELECT count(*) FROM rows`).Scan(&deletedUserImageAnnotationSuggestionRows)
	if err != nil {
		return err
	}

	if insertedUserImageAnnotationRows != deletedUserImageAnnotationSuggestionRows {
		return errors.New("inserted user_image_annotation rows differ from deleted user_image_annoation_suggestion rows!")
	}
	
	var deletedAnnotationSuggestionDataRows int64 = 0
	err = tx.QueryRow(`WITH rows AS (
							DELETE FROM annotation_suggestion_data WHERE id IN
								(SELECT old_annotation_data_id FROM temp_annotation_data_mapping)
						  	
							RETURNING 1
					   )
					   SELECT count(*) FROM rows`).Scan(&deletedAnnotationSuggestionDataRows)
	if err != nil {
		return err
	}

	if insertedAnnotationDataRows != deletedAnnotationSuggestionDataRows {
		return errors.New("inserted annotation_data rows differ from deleted annotation_suggestion_data rows!")
	}


	var deletedAnnotationSuggestionDataRevisionRows int64 = 0
	err = tx.QueryRow(`WITH rows AS (
							DELETE FROM annotation_suggestion_data WHERE id IN
								(SELECT old_annotation_data_id FROM temp_annotation_data_revision_mapping)
						  	
							RETURNING 1
					   )
					   SELECT count(*) FROM rows`).Scan(&deletedAnnotationSuggestionDataRevisionRows)
	if err != nil {
		return err
	}
	if insertedAnnotationDataRevisionRows != deletedAnnotationSuggestionDataRevisionRows {
		return errors.New("inserted annotation_data revision rows differ from deleted annotation_suggestion_data revision rows!")
	}


	var deletedImageAnnotationSuggestionRevisionRows int64 = 0
	err = tx.QueryRow(`WITH rows AS (
						DELETE FROM image_annotation_suggestion_revision WHERE id IN
							(SELECT old_image_annotation_revision_id FROM  temp_image_annotation_revision_mapping)
							
						RETURNING 1
					   )
					   SELECT count(*) FROM rows`).Scan(&deletedImageAnnotationSuggestionRevisionRows)
	if err != nil {
		return err
	}

	if insertedImageAnnotationRevisionRows != deletedImageAnnotationSuggestionRevisionRows {
		return errors.New("inserted image_annotation_revision rows differ from deleted image_annotation_suggestion_revision rows!")
	}

	var deletedImageAnnotationSuggestionRows int64 = 0
	err = tx.QueryRow(`WITH rows AS (
						DELETE FROM image_annotation_suggestion WHERE id IN
							(SELECT old_image_annotation_id FROM temp_image_annotation_mapping)
							
						RETURNING 1
					   )
					   SELECT count(*) FROM rows`).Scan(&deletedImageAnnotationSuggestionRows)
	if err != nil {
		return err
	}

	if insertedImageAnnotationRows != deletedImageAnnotationSuggestionRows {
		return errors.New("inserted image_annotation rows differ from deleted image_annotation_suggestion rows!")
	}


	for _, tempTable := range tempTables {
		_, err = tx.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS %s`, tempTable)) //controlled input, so no sql injection possible
		if err != nil {
			return err
		}
	}

	return nil
}


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

func _getLabelId(labelIdentifier string, tx *sql.Tx) (int64, error) {
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

func closeGithubIssue(trendingLabel string, repository string, githubProjectOwner string, githubApiToken string, tx *sql.Tx) error {
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
				&oauth2.Token{AccessToken: githubApiToken},
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
			_, _, err = client.Issues.CreateComment(ctx, githubProjectOwner, repository, githubIssueId, commentRequest)
			if err == nil { //if comment was successfully created, close issue
				issueRequest := &github.IssueRequest{
					State: github.String("closed"),
				}

				_, _, err = client.Issues.Edit(ctx, githubProjectOwner, repository, githubIssueId, issueRequest)
				if err != nil {
					log.Info("[Main] Couldn't close github issue, please close manually. ", err.Error())
				}
			} else {
				log.Info("[Main] Couldn't close github issue, please close manually. ", err.Error())
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


func getNonStrictLabels(tx *sql.Tx, name string) ([]string, error) {
	names := []string{}

	rows, err := tx.Query("SELECT l.name FROM label_suggestion l WHERE l.name ~  ('^[ ]*'||$1||'[ ]*$')", name)
	if err != nil {
		return names, err
	}

	defer rows.Close()

	for rows.Next() {
		var n string
		err = rows.Scan(&n)
		if err != nil {
			return names, err
		}

		names = append(names, n)
	}

	return names, nil
}

