package tests

import (
	"database/sql"
	"bytes"
	"os/exec"
	"fmt"
	"errors"
	"math/rand"
	"time"
)

func random(min, max int) int {
    rand.Seed(time.Now().Unix())
    return rand.Intn(max - min) + min
}

func randomBool() bool {
    return rand.Float32() < 0.5
}


func loadSchema() error {
	var out, stderr bytes.Buffer
	schemaPath := "../env/postgres/schema.sql"

	//load schema
	cmd := exec.Command("psql", "-f", schemaPath, "-d", "imagemonkey", "-U", "postgres", "-h", "127.0.0.1")
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
	    fmt.Sprintf("Error executing query. Command Output: %+v\n: %+v, %v", out.String(), stderr.String(), err)
	    return err
	}

	return nil
}

func loadDefaults() error {
	var out, stderr bytes.Buffer
	defaultsPath := "../env/postgres/defaults.sql"

	//load defaults
	cmd := exec.Command("psql", "-f", defaultsPath, "-d", "imagemonkey", "-U", "postgres", "-h", "127.0.0.1")
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
	    fmt.Sprintf("Error executing query. Command Output: %+v\n: %+v, %v", out.String(), stderr.String(), err)
	    return err
	}

	return nil
} 


func installTriggers() error {
	var out, stderr bytes.Buffer
	triggersPath := "../env/postgres/triggers.sql"

	//load defaults
	cmd := exec.Command("psql", "-f", triggersPath, "-d", "imagemonkey", "-U", "postgres", "-h", "127.0.0.1")
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
	    fmt.Sprintf("Error executing query. Command Output: %+v\n: %+v, %v", out.String(), stderr.String(), err)
	    return err
	}

	return nil
} 

func populateLabels() error {
	var out, stderr bytes.Buffer
	cmd := exec.Command("go", "run", "populate_labels.go", "api_secrets.go", "--dryrun=false")
	cmd.Dir = "../src/"
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
	    fmt.Sprintf("Error executing query. Command Output: %+v\n: %+v, %v", out.String(), stderr.String(), err)
	    return err
	}

	return nil
}

func installUuidExtension() error {
	query := "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\""
	var out, stderr bytes.Buffer

	//load defaults
	cmd := exec.Command("psql", "-c", query, "-d", "imagemonkey", "-U", "postgres", "-h", "127.0.0.1")
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
	    fmt.Sprintf("Error executing query. Command Output: %+v\n: %+v, %v", out.String(), stderr.String(), err)
	    return err
	}

	return nil
} 

func installPostgisExtension() error {
	query := "CREATE EXTENSION IF NOT EXISTS \"postgis\""
	var out, stderr bytes.Buffer

	//load defaults
	cmd := exec.Command("psql", "-c", query, "-d", "imagemonkey", "-U", "postgres", "-h", "127.0.0.1")
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
	    fmt.Sprintf("Error executing query. Command Output: %+v\n: %+v, %v", out.String(), stderr.String(), err)
	    return err
	}

	return nil
} 

func installTemporalTablesExtension() error {
	query := "CREATE EXTENSION IF NOT EXISTS temporal_tables"
	var out, stderr bytes.Buffer

	//load defaults
	cmd := exec.Command("psql", "-c", query, "-d", "imagemonkey", "-U", "postgres", "-h", "127.0.0.1")
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
	    fmt.Sprintf("Error executing query. Command Output: %+v\n: %+v, %v", out.String(), stderr.String(), err)
	    return err
	}

	return nil
} 


func installTruncateAllTablesFunction() error {
     query := `CREATE OR REPLACE FUNCTION truncate_tables(username IN VARCHAR) RETURNS void AS $$
				DECLARE
   					statements CURSOR FOR
        				SELECT tablename FROM pg_tables
        				WHERE tableowner = username AND schemaname = 'public';
				BEGIN
    				FOR stmt IN statements LOOP
        				EXECUTE 'TRUNCATE TABLE ' || quote_ident(stmt.tablename) || ' CASCADE;';
   				END LOOP;
			   END;
			   $$ LANGUAGE plpgsql`
     var out, stderr bytes.Buffer

     //load defaults
     cmd := exec.Command("psql", "-c", query, "-d", "imagemonkey", "-U", "postgres", "-h", "127.0.0.1")
     cmd.Stdout = &out
     cmd.Stderr = &stderr

     err := cmd.Run()
     if err != nil {
         fmt.Sprintf("Error executing query. Command Output: %+v\n: %+v, %v", out.String(), stderr.String(), err)
         return err
     }

     return nil
 }


type ImageMonkeyDatabase struct {
    db *sql.DB
}

func NewImageMonkeyDatabase() *ImageMonkeyDatabase {
    return &ImageMonkeyDatabase{} 
}

func (p *ImageMonkeyDatabase) Open() error {
	var err error
    p.db, err = sql.Open("postgres", DB_CONNECTION_STRING)
	if err != nil {
		return err
	}

	err = p.db.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (p *ImageMonkeyDatabase) Initialize() error {

	//connect as user postgres, in order to drop + re-create database imagemonkey
	localDb, err := sql.Open("postgres", "user=postgres sslmode=disable")
	if err != nil {
		return err
	}

	defer localDb.Close()

	//terminate any open database connections (we need to do this, before we can drop the database)
	_, err = localDb.Exec(`SELECT pg_terminate_backend(pid)
					  FROM pg_stat_activity
					  WHERE datname = 'imagemonkey'`)
	if err != nil {
		return err
	}

	_, err = localDb.Exec("DROP DATABASE IF EXISTS imagemonkey")
	if err != nil {
		return err
	}

	_, err = localDb.Exec("CREATE DATABASE imagemonkey OWNER monkey")
	if err != nil {
		return err
	}

	err = installTemporalTablesExtension()
	if err != nil {
		return err
	}

	err = installUuidExtension()
	if err != nil {
		return err
	}

	err = installPostgisExtension()
	if err != nil {
		return err
	}
	
	err = loadSchema()
	if err != nil {
		return err
	}

	err = loadDefaults()
	if err != nil {
		return err
	}

	err = populateLabels()
	if err != nil {
		return err
	}

	err = installTriggers()
	if err != nil {
		return err
	}

	err = installTruncateAllTablesFunction()
	if err != nil {
		return err
	}

	return nil
}

func (p *ImageMonkeyDatabase) ClearAll() error {
	_, err := p.db.Exec(`SELECT truncate_tables('monkey')`)
	if err != nil {
		return err
	}

	err = loadDefaults()
	if err != nil {
		return err
	}

	err = populateLabels()
	return err
}

func (p *ImageMonkeyDatabase) UnlockAllImages() error {
	_, err := p.db.Exec(`UPDATE image SET unlocked = true`)
	if err != nil {
		return err
	}

	return nil
}

func (p *ImageMonkeyDatabase) GiveUserModeratorRights(name string) error {
	_, err := p.db.Exec("UPDATE account SET is_moderator = true WHERE name = $1", name)
	if err != nil {
		return err
	}

	_, err = p.db.Exec(`INSERT INTO account_permission(account_id, can_remove_label, can_unlock_image_description) 
							SELECT a.id, true, true FROM account a WHERE a.name = $1`, name)
	if err != nil {
		return err
	}

	return nil
}

func (p *ImageMonkeyDatabase) GiveUserUnlockImagePermissions(name string) error {
	_, err := p.db.Exec(`UPDATE account_permission 
						 SET can_unlock_image = true
						 FROM account a
						 WHERE a.id = account_id AND a.name = $1`, name)
	return err
}

func (p *ImageMonkeyDatabase) GetNumberOfImages() (int32, error) {
	var numOfImages int32
	err := p.db.QueryRow("SELECT count(*) FROM image").Scan(&numOfImages)
	if err != nil {
		return 0, err
	}

	return numOfImages, err
}

func (p *ImageMonkeyDatabase) GetNumberOfUsers() (int32, error) {
	var numOfUsers int32
	err := p.db.QueryRow("SELECT count(*) FROM account").Scan(&numOfUsers)
	if err != nil {
		return 0, err
	}

	return numOfUsers, err
}

func (p *ImageMonkeyDatabase) GetAllValidationIds() ([]string, error) {
	var validationIds []string

	rows, err := p.db.Query("SELECT uuid FROM image_validation")
	if err != nil {
		return validationIds, err
	}

	defer rows.Close()

	for rows.Next() {
		var validationId string
		err = rows.Scan(&validationId)
		if err != nil {
			return validationIds, err
		}

		validationIds = append(validationIds, validationId)
	}

	return validationIds, nil
}

func (p *ImageMonkeyDatabase) GetAllValidationIdsForLabel(label string) ([]string, error) {
     var validationIds []string

     rows, err := p.db.Query(`SELECT v.uuid FROM image_validation v 
	 						  JOIN label l ON v.label_id = l.id 
							  WHERE l.name = $1 AND l.parent_id is null`, label)
     if err != nil {
         return validationIds, err
     }

     defer rows.Close()

     for rows.Next() {
         var validationId string
         err = rows.Scan(&validationId)
         if err != nil {
             return validationIds, err
         }

         validationIds = append(validationIds, validationId)
     }

     return validationIds, nil
 }


func (p *ImageMonkeyDatabase) GetRandomValidationId() (string, error) {
	validationIds, err := db.GetAllValidationIds()
	if err != nil {
		return "", err
	}

	if len(validationIds) == 0 {
		return "", errors.New("Fetching random validation id - got no result!")
	}

	randomIdx := random(0, len(validationIds) -1)

	return validationIds[randomIdx], nil
}


func (p *ImageMonkeyDatabase) GetValidationCount(uuid string) (int32, int32, error) {
	var numOfYes int32
	var numOfNo int32
	err := p.db.QueryRow(`SELECT num_of_valid, num_of_invalid 
						  FROM image_validation WHERE uuid = $1`, uuid).Scan(&numOfYes, &numOfNo)

	return numOfYes, numOfNo, err
}

func (p *ImageMonkeyDatabase) GetAnnotationRevision(annotationId string) (int32, error) {
	var revision int32
	err := p.db.QueryRow(`SELECT revision 
						  FROM image_annotation WHERE uuid = $1`, annotationId).Scan(&revision)

	return revision, err
}

func (p *ImageMonkeyDatabase) GetOldAnnotationDataIds(annotationId string, revision int32) ([]int64, error) {
	var ids []int64

	rows, err := p.db.Query(`SELECT d.id
							FROM annotation_data d
							JOIN image_annotation_revision r ON d.image_annotation_revision_id = r.id
							JOIN image_annotation a ON r.image_annotation_id = a.id
							WHERE a.uuid = $1 AND r.revision = $2`, annotationId, revision)
	if err != nil {
		return ids, err
	}

	defer rows.Close()

	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return ids, err
		}

		ids = append(ids, id) 
	}

	return ids, nil
}

func (p *ImageMonkeyDatabase) GetAnnotationDataIds(annotationId string) ([]int64, error) {
	var ids []int64
	rows, err := p.db.Query(`SELECT d.id 
							FROM annotation_data d
							JOIN image_annotation a ON d.image_annotation_id = a.id
							WHERE a.uuid = $1`, annotationId)
	if err != nil {
		return ids, err
	}

	defer rows.Close()

	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return ids, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func (p *ImageMonkeyDatabase) GetRandomImageForAnnotation() (AnnotationRow, error) {
	var annotationRow AnnotationRow
	err := p.db.QueryRow(`SELECT i.key, v.uuid, l.name, COALESCE(pl.name, '')
			      	 		FROM image i 
				  	 		JOIN image_validation v ON v.image_id = i.id 
				  	 		JOIN label l ON v.label_id = l.id
				  	 		LEFT JOIN label pl ON pl.id = l.parent_id 
				  	 		WHERE NOT EXISTS (
				  	 			SELECT 1 FROM image_annotation a 
				  	 			WHERE a.label_id = l.id AND a.image_id = i.id
				  	 		) LIMIT 1`).Scan(&annotationRow.Image.Id, &annotationRow.Validation.Id,
				  	 			&annotationRow.Validation.Label, &annotationRow.Validation.Sublabel)

	return annotationRow, err
}

func (p *ImageMonkeyDatabase) GetRandomAnnotationId() (string, error) {
	var annotationId string
	err := p.db.QueryRow(`SELECT a.uuid FROM image_annotation a ORDER BY random() LIMIT 1`).Scan(&annotationId)
	return annotationId, err
}

func (p *ImageMonkeyDatabase) GetLastAddedAnnotationDataId() (string, error) {
	var annotationDataId string
	err := p.db.QueryRow(`SELECT d.uuid FROM annotation_data d ORDER BY d.id DESC LIMIT 1`).Scan(&annotationDataId)
	return annotationDataId, err
}

func (p *ImageMonkeyDatabase) GetLastAddedAnnotationId() (string, error) {
	var annotationId string
	err := p.db.QueryRow(`SELECT a.uuid FROM image_annotation a ORDER BY a.id DESC LIMIT 1`).Scan(&annotationId)
	return annotationId, err
}

func (p *ImageMonkeyDatabase) GetRandomLabelId() (int64, error) {
	var labelId int64
	err := p.db.QueryRow(`SELECT l.id FROM label l ORDER BY random() LIMIT 1`).Scan(&labelId)
	return labelId, err
}

func (p *ImageMonkeyDatabase) GetRandomLabelUuid() (string, error) {
	var labelUuid string
	err := p.db.QueryRow(`SELECT l.uuid FROM label l ORDER BY random() LIMIT 1`).Scan(&labelUuid)
	return labelUuid, err
}

func (p *ImageMonkeyDatabase) GetRandomAnnotationData() (string, string, error) {
	var annotationId string
	var annotationDataId string
	err := p.db.QueryRow(`SELECT a.uuid, d.uuid
						  FROM image_annotation a 
						  JOIN annotation_data d ON d.image_annotation_id = a.id
						  ORDER BY random() LIMIT 1`).Scan(&annotationId, &annotationDataId)
	return annotationId, annotationDataId, err
}

func (p *ImageMonkeyDatabase) GetLastAddedAnnotationData() (string, string, error) {
	var annotationId string
	var annotationDataId string
	err := p.db.QueryRow(`SELECT a.uuid, d.uuid
						  FROM image_annotation a 
						  JOIN annotation_data d ON d.image_annotation_id = a.id
						  ORDER BY a.id DESC LIMIT 1`).Scan(&annotationId, &annotationDataId)
	return annotationId, annotationDataId, err
}

func (p *ImageMonkeyDatabase) GetNumberOfImagesWithLabel(label string) (int32, error) {
	var num int32
	err := p.db.QueryRow(`SELECT count(*) 
						   FROM image_validation v 
						   JOIN label l ON v.label_id = l.id
						   WHERE l.name = $1 AND l.parent_id is null`, label).Scan(&num)
	return num, err
}

func (p *ImageMonkeyDatabase) GetNumberOfImagesWithLabelUuid(labelUuid string) (int32, error) {
	var num int32
	err := p.db.QueryRow(`SELECT count(*) 
						   FROM image_validation v 
						   JOIN label l ON v.label_id = l.id
						   WHERE l.uuid::text = $1`, labelUuid).Scan(&num)
	return num, err
}

func (p *ImageMonkeyDatabase) GetNumberOfImagesWithLabelSuggestions(label string) (int32, error) {
	var num int32
	err := p.db.QueryRow(`SELECT count(*) 
						   FROM image_label_suggestion ils
						   JOIN label_suggestion l ON l.id = ils.label_suggestion_id
						   WHERE l.name = $1
						 `, label).Scan(&num)
	return num, err
}

func (p *ImageMonkeyDatabase) GetNumberOfTrendingLabelSuggestions() (int32, error) {
	var num int32
	err := p.db.QueryRow(`SELECT count(*) 
						   FROM trending_label_suggestion`).Scan(&num)
	return num, err
}

func (p *ImageMonkeyDatabase) GetNumberOfImageHuntTasksForImageWithLabel(imageId string, label string) (int32, error) {
	var num int32
	err := p.db.QueryRow(`SELECT count(*) 
						   FROM imagehunt_task h
						   JOIN image_validation v ON v.id = h.image_validation_id
						   JOIN label l ON l.id = v.label_id
						   JOIN image i ON i.id = v.image_id
						   WHERE i.key = $1 AND l.name = $2`, imageId, label).Scan(&num)
	return num, err
}

func (p *ImageMonkeyDatabase) GetNumberOfImageUserEntriesForImageAndUser(imageId string, username string) (int32, error) {
	var num int32
	err := p.db.QueryRow(`SELECT count(*) 
						   FROM image i
						   JOIN user_image u ON u.image_id = i.id
						   JOIN account a ON a.id = u.account_id
						   WHERE i.key = $1 AND a.name = $2`, imageId, username).Scan(&num)
	return num, err
}

func (p *ImageMonkeyDatabase) GetProductiveLabelIdsForTrendingLabels() ([]int64, error) {
	productiveLabelIds := []int64{}

	rows, err := p.db.Query(`SELECT t.productive_label_id FROM trending_label_suggestion t 
							 WHERE t.productive_label_id is not null`)
	if err != nil {
		return productiveLabelIds, err
	}
	defer rows.Close()
	for rows.Next() {
		var productiveLabelId int64
		err = rows.Scan(&productiveLabelId)
		if err != nil {
			return productiveLabelIds, err
		}

		productiveLabelIds = append(productiveLabelIds, productiveLabelId)
	}

	return productiveLabelIds, nil
}

func (p *ImageMonkeyDatabase) GetRandomLabelName(skipLabel string) (string, error) {
	var queryParams []interface{}
	skipLabelStr := ""
	if skipLabel != "" {
		skipLabelStr = "AND l.name != $1"
		queryParams = append(queryParams, skipLabel)
	}
	
	query := fmt.Sprintf(`SELECT l.name
                			FROM label l
                			WHERE l.parent_id is null AND l.label_type = 'normal'
							%s
                			ORDER BY random() LIMIT 1`, skipLabelStr) 
	
	
	
	var label string
	err := p.db.QueryRow(query, queryParams...).Scan(&label)
	return label, err
}

func (p *ImageMonkeyDatabase) GetAllImageIds() ([]string, error) {
	var imageIds []string

	rows, err := p.db.Query(`SELECT i.key FROM image i ORDER BY random()`)
	if err != nil {
		return imageIds, err
	}

	defer rows.Close()

	for rows.Next() {
		var imageId string
		err = rows.Scan(&imageId)
		if err != nil {
			return imageIds, err
		}

		imageIds = append(imageIds, imageId)
	}

	return imageIds, nil
}

func (p *ImageMonkeyDatabase) GetLatestDonatedImageId() (string,error) {
	var imageId string 
	err := p.db.QueryRow(`SELECT i.key FROM image i ORDER BY id DESC LIMIT 1`).Scan(&imageId)
	return imageId, err
}

func (p *ImageMonkeyDatabase) PutImageInQuarantine(imageId string) error { 
	_, err := p.db.Exec(`INSERT INTO image_quarantine(image_id)
							SELECT id FROM image WHERE key = $1`, imageId)
	return err
}

func (p *ImageMonkeyDatabase) GetLabelUuidFromName(label string) (string, error) {
	var uuid string 
	err := p.db.QueryRow(`SELECT l.uuid 
							FROM label l 
							WHERE l.name = $1 and l.parent_id is null`, label).Scan(&uuid)
	return uuid, err
}

func (p *ImageMonkeyDatabase) GetLabelIdFromName(label string) (int64, error) {
	var labelId int64 
	err := p.db.QueryRow(`SELECT l.id 
							FROM label l 
							WHERE l.name = $1 and l.parent_id is null`, label).Scan(&labelId)
	return labelId, err
}

func (p *ImageMonkeyDatabase) GetLabelIdFromUuid(labelUuid string) (int64, error) {
	var labelId int64 
	err := p.db.QueryRow(`SELECT l.id 
							FROM label l 
							WHERE l.uuid::text = $1`, labelUuid).Scan(&labelId)
	return labelId, err
}

func (p *ImageMonkeyDatabase) GetNumOfSentOfTrendingLabel(tendingLabel string) (int, error) {
	var tendingLabelId int 
	err := p.db.QueryRow(`SELECT t.num_of_last_sent 
						  FROM trending_label_suggestion t
						  JOIN label_suggestion l ON t.label_suggestion_id = l.id
						  WHERE l.name = $1`, tendingLabel).Scan(&tendingLabelId)
	return tendingLabelId, err
}

func (p *ImageMonkeyDatabase) SetValidationValid(validationId string, num int) error {
	_, err := p.db.Exec(`UPDATE image_validation 
							SET num_of_valid = $2 
							WHERE uuid = $1`, validationId, num)
	return err
}

func (p *ImageMonkeyDatabase) GetNumOfRefinements() (int, error) {
	var num int 
	err := p.db.QueryRow(`SELECT count(*) FROM image_annotation_refinement`).Scan(&num)
	return num, err
}

func (p *ImageMonkeyDatabase) GetAllAnnotationIds() ([]string, error) {
	var annotationIds []string

	rows, err := p.db.Query("SELECT uuid FROM image_annotation")
	if err != nil {
		return annotationIds, err
	}

	defer rows.Close()

	for rows.Next() {
		var annotationId string
		err = rows.Scan(&annotationId)
		if err != nil {
			return annotationIds, err
		}

		annotationIds = append(annotationIds, annotationId)
	}

	return annotationIds, nil
}

func (p *ImageMonkeyDatabase) SetAnnotationValid(annotationId string, num int) error {
	_, err := p.db.Exec(`UPDATE image_annotation 
							SET num_of_valid = $2 
							WHERE uuid = $1`, annotationId, num)
	return err
}

func (p *ImageMonkeyDatabase) GetImageAnnotationCoverageForImageId(imageId string) (int, error) {
	rows, err := p.db.Query(`SELECT annotated_percentage 
							  FROM image_annotation_coverage c
							  JOIN image i ON i.id = c.image_id
							  WHERE i.key = $1`, imageId)
	if err != nil {
		return 0, err
	}

	defer rows.Close()

	if rows.Next() {
		var coverage int
		err = rows.Scan(&coverage)
		if err != nil {
			return 0, err
		}

		return coverage, nil
	}
	return 0, errors.New("missing result set")
}

func (p *ImageMonkeyDatabase) GetImageDescriptionForImageId(imageId string) ([]ImageDescriptionSummary, error) {
	var descriptionSummaries []ImageDescriptionSummary

	rows, err := p.db.Query(`SELECT dsc.description, dsc.num_of_valid, dsc.uuid, dsc.state, l.name
							 FROM image_description dsc
							 JOIN language l ON l.id = dsc.language_id
							 JOIN image i ON i.id = dsc.image_id
							 WHERE i.key = $1
							 ORDER BY dsc.id asc`, imageId)
	if err != nil {
		return descriptionSummaries, err
	}

	defer rows.Close()

	for rows.Next() {
		var dsc ImageDescriptionSummary
		var state string
		err = rows.Scan(&dsc.Description, &dsc.NumOfValid, &dsc.Uuid, &state, &dsc.Language)
		if err != nil {
			return descriptionSummaries, err
		}

		if state == "unknown" {
			dsc.State = ImageDescriptionStateUnknown
		} else if state == "locked" {
			dsc.State = ImageDescriptionStateLocked
		} else if state == "unlocked" {
			dsc.State = ImageDescriptionStateUnlocked
		}

		descriptionSummaries = append(descriptionSummaries, dsc)
	}

	return descriptionSummaries, nil
}


func (p *ImageMonkeyDatabase) GetModeratorWhoProcessedImageDescription(imageId string, imageDescription string) (string, error) {
	rows, err := p.db.Query(`SELECT a.name
							 FROM image_description dsc
							 JOIN image i ON i.id = dsc.image_id
							 JOIN account a ON a.id = dsc.processed_by
							 WHERE i.key = $1 AND dsc.description = $2`, imageId, imageDescription)
	if err != nil {
		return "", err
	}

	defer rows.Close()

	if rows.Next() {
		var moderator string
		err = rows.Scan(&moderator)
		if err != nil {
			return "", err
		}

		return moderator, nil
	}
	return "", errors.New("missing result set")
}

func (p *ImageMonkeyDatabase) IsImageUnlocked(imageId string) (bool, error) {
	rows, err := p.db.Query(`SELECT unlocked FROM image i WHERE i.key = $1`, imageId)
	if err != nil {
		return false, err
	}

	defer rows.Close()

	if rows.Next() {
		var unlocked bool
		err = rows.Scan(&unlocked)
		if err != nil {
			return false, err
		}

		return unlocked, nil
	}
	return false, errors.New("missing result set")
}

func (p *ImageMonkeyDatabase) IsImageInQuarantine(imageId string) (bool, error) {
	rows, err := p.db.Query(`SELECT CASE 
									 WHEN COUNT(*) <> 0 THEN true 
									 ELSE false
									END as in_quarantine
							 FROM image_quarantine q 
							 WHERE q.image_id IN (
							 	SELECT i.id FROM image i WHERE i.key = $1
							 )`, imageId)
	if err != nil {
		return false, err
	}

	defer rows.Close()

	if rows.Next() {
		var inQuarantine bool
		err = rows.Scan(&inQuarantine)
		if err != nil {
			return false, err
		}

		return inQuarantine, nil
	}
	return false, errors.New("missing result set")
}

func (p *ImageMonkeyDatabase) DoLabelAccessorsBelongToMoreThanOneLabelId() (bool, error) {
	rows, err := p.db.Query(`SELECT label_id 
								FROM label_accessor
								GROUP BY label_id
								HAVING COUNT(label_id) > 1`)
	if err != nil {
		return false, err
	}

	defer rows.Close()

	if rows.Next() {
		return true, nil
	}

	return false, nil
}

func (p *ImageMonkeyDatabase) GetNumOfMetaLabelImageValidations() (int, error) {
	var num int 
	err := p.db.QueryRow(`SELECT count(*) FROM 
							image_validation v 
							JOIN label l ON l.id = v.label_id
							WHERE l.label_type = 'meta'`).Scan(&num)
	return num, err
}

func (p *ImageMonkeyDatabase) GetNumOfDatesFromNowTilOneMonthAgo() (int, error) {
	var num int
	err := p.db.QueryRow(`SELECT COUNT(*)
                            FROM generate_series((CURRENT_DATE - interval '1 month'), CURRENT_DATE, '1 day')`).Scan(&num)
    return num, err
}

func (p *ImageMonkeyDatabase) RemoveLabel(labelName string) error {
	tx, err := p.db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}
	
	_, err = tx.Exec("DELETE FROM label_accessor a WHERE a.label_id IN (SELECT l.id FROM label l WHERE l.name = $1 AND l.parent_id is null)", labelName)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM label l WHERE l.name = $1 AND l.parent_id is null", labelName)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
} 

func (p *ImageMonkeyDatabase) GetNumOfNotAnnotatable(uuid string) (int, error) {
	rows, err := p.db.Query("SELECT num_of_not_annotatable FROM image_validation WHERE uuid = $1", uuid)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	
	var num int
	if rows.Next() {
		err = rows.Scan(&num)
		if err != nil {
			return 0, err
		}
	}
	return num, nil
} 

func (p *ImageMonkeyDatabase) Close() {
	p.db.Close()
}
