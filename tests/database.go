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
	cmd := exec.Command("psql", "-f", schemaPath, "-d", "imagemonkey", "-U", "postgres")
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
	cmd := exec.Command("psql", "-f", defaultsPath, "-d", "imagemonkey", "-U", "postgres")
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
	cmd := exec.Command("psql", "-f", triggersPath, "-d", "imagemonkey", "-U", "postgres")
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
	cmd := exec.Command("go", "run", "populate_labels.go", "common.go", "api_secrets.go", "--dryrun=false")
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
	cmd := exec.Command("psql", "-c", query, "-d", "imagemonkey", "-U", "postgres")
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
	cmd := exec.Command("psql", "-c", query, "-d", "imagemonkey", "-U", "postgres")
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

	_, err = p.db.Exec(`INSERT INTO account_permission(account_id, can_remove_label) 
							SELECT a.id, true FROM account a WHERE a.name = $1`, name)
	if err != nil {
		return err
	}

	return nil
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

func (p *ImageMonkeyDatabase) GetNumberOfImagesWithLabel(label string) (int32, error) {
	var num int32
	err := p.db.QueryRow(`SELECT count(*) 
						   FROM image_validation v 
						   JOIN label l ON v.label_id = l.id
						   WHERE l.name = $1 AND l.parent_id is null`, label).Scan(&num)
	return num, err
}

func (p *ImageMonkeyDatabase) GetRandomLabelName() (string, error) {
	var label string
	err := p.db.QueryRow(`SELECT l.name
						   FROM label l
						   WHERE l.parent_id is null
						   ORDER BY random() LIMIT 1`).Scan(&label)
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

func (p *ImageMonkeyDatabase) SetValidationValid(validationId string, num int) error {
	_, err := p.db.Exec(`UPDATE image_validation 
							SET num_of_valid = $2 
							WHERE uuid = $1`, validationId, num)
	return err
}

func (p *ImageMonkeyDatabase) Close() {
	p.db.Close()
}
