package main

import (
	"fmt"
	"gopkg.in/resty.v1"
	"errors"
	log "github.com/Sirupsen/logrus"
	"database/sql"
	_"github.com/lib/pq"
	"io/ioutil"
	"math/rand"
	"time"
	"os/exec"
	"bytes"
	"encoding/json"
	"reflect"
)


const BASE_URL string = "http://127.0.0.1:8081/"
const API_VERSION string = "v1"

type AnnotatedImage struct {
    Image struct {
        Id string `json:"uuid"`
        Provider string `json:"provider"`
        Width int32 `json:"width"`
        Height int32 `json:"height"`
    } `json:"image"`

    Validation struct {
        Label string `json:"label"`
        Sublabel string `json:"sublabel"`
    } `json:"validation"`
    

    Id string `json:"uuid"`
    NumOfValid int32 `json:"num_yes"`
    NumOfInvalid int32 `json:"num_no"`
    Annotations []json.RawMessage `json:"annotations"`
    NumRevisions int32 `json:"num_revisions"`
    Revision int32 `json:"revision"`
}

type AnnotationRow struct {
	Image struct {
		Id string `json:"uuid"`
	} `json:"image"`

	Validation struct {
		Id string `json:"uuid"`
		Label string `json:"label"`
		Sublabel string `json:"sublabel"`
	} `json:"validation"`
}

func random(min, max int) int {
    rand.Seed(time.Now().Unix())
    return rand.Intn(max - min) + min
}

func randomBool() bool {
    return rand.Float32() < 0.5
}


func equalJson(s1, s2 string) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal([]byte(s1), &o1)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 1 :: %s", err.Error())
	}
	err = json.Unmarshal([]byte(s2), &o2)
	if err != nil {
		return false, fmt.Errorf("Error mashalling string 2 :: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
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


type ImageMonkeyDatabase struct {
    db *sql.DB
}

var db *ImageMonkeyDatabase

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

	_, err = localDb.Exec("CREATE EXTENSION IF NOT EXISTS temporal_tables")
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
	err := p.db.QueryRow(`SELECT a.uuid FROM image_annotation a LIMIT 1`).Scan(&annotationId)
	return annotationId, err
}


func (p *ImageMonkeyDatabase) Close() {
	p.db.Close()
}



type ReportResult struct {
    Reason string `json:"reason"`
}

type LoginResult struct {
	Token string `json:"token"`
}


type ValidateResult struct {
    Id string `json:"uuid"`
    Url string `json:"url"`
    Label string `json:"label"`
    Provider string `json:"provider"`
    Probability float32 `json:"probability"`
    NumOfValid int32 `json:"num_yes"`
    NumOfInvalid int32 `json:"num_no"`
}

func testReport(uuid string, reason string) error{
	url := BASE_URL + API_VERSION + "/report/" + uuid
	_, err := resty.R().
    	SetHeader("Content-Type", "application/json").
     	SetBody(`{"reason":"whoops"}`).
     	SetResult(&ReportResult{}).
     	Post(url)
     return err
}

func testSignUp(username string, password string, email string) error {
	numBefore, err := db.GetNumberOfUsers()
	if err != nil {
		return err
	}


	url := BASE_URL + API_VERSION + "/signup"
	resp, err := resty.R().
    	SetHeader("Content-Type", "application/json").
     	SetBody(map[string]interface{}{"username": username, "password": password, "email": email}).
     	Post(url)

    if resp.StatusCode() != 201 {
		return errors.New("Couldn't signup")
    }

    numAfter, err := db.GetNumberOfUsers(); 
    if err != nil {
		return err
	}

	if numAfter != numBefore + 1 {
		return errors.New("number of account entries in database is off by one!")
	} 

    return err
}

func testLogin(username string, password string) (string, error) {
	url := BASE_URL + API_VERSION + "/login"
	resp, err := resty.R().
     	SetBasicAuth(username, password).
     	SetResult(&LoginResult{}).
     	Post(url)

    if resp.StatusCode() != 200 {
		return "", errors.New("Couldn't login")
    }

    return resp.Result().(*LoginResult).Token, err
}

func testDonate(path string, label string) error {
	numBefore, err := db.GetNumberOfImages()
	if err != nil {
		return err
	}

	url := BASE_URL + API_VERSION + "/donate"
	resp, err := resty.R().
      SetFile("image", path).
      SetFormData(map[string]string{
        "label": label,
      }).
      Post(url)

    if resp.StatusCode() != 200 {
		return errors.New("Couldn't push donation")
    }

    numAfter, err := db.GetNumberOfImages(); 
    if err != nil {
		return err
	}

	if numAfter != numBefore + 1 {
		return errors.New("number of image entries in database is off by one!")
	} 

	//after image donation, unlock all images
    err = db.UnlockAllImages()

    return err
}

func testMultipleDonate() error {
	dirname := "./images/apples/"
	files, err := ioutil.ReadDir(dirname)
    if err != nil {
        log.Fatal(err)
    }

    for _, f := range files {
        if err := testDonate(dirname + f.Name(), "apple"); err != nil {
        	return err
        }
    }

    return nil
}

func testValidate() error{
	url := BASE_URL + API_VERSION + "/validate"
	_, err := resty.R().
			SetResult(&ValidateResult{}).
			Get(url)
	return err
}

func testImageValidation(uuid string, param string, moderated bool, token string) error{
	url := BASE_URL + API_VERSION + "/validation/" + uuid + "/validate/" + param

	var resp *resty.Response
	var err error
	if moderated {
		resp, err = resty.R().
				SetHeader("X-Moderation", "true").
				SetAuthToken(token).
				Post(url)
	} else {
		resp, err = resty.R().
				Post(url)
	}

	if resp.StatusCode() != 200 {
		return errors.New("Couldn't validate image with uuid " + uuid)
	}


	
	return err
}

func testRandomImageValidation(num int) error {
	for i := 0; i < num; i++ {
		param := ""
		randomBool := randomBool()
		if randomBool {
			param = "yes"
		} else {
			param = "no"
		}

		randomValidationId, err := db.GetRandomValidationId()
		if err != nil {
			return err
		}

		beforeChangeNumValid, beforeChangeNumInvalid, err := db.GetValidationCount(randomValidationId)
		if err != nil {
			return err
		}

		if err := testImageValidation(randomValidationId, param, false, ""); err != nil {
			return err
		}

		afterChangeNumValid, afterChangeNumInvalid, err := db.GetValidationCount(randomValidationId)
		if err != nil {
			return err
		}

		if param == "yes" {
			if afterChangeNumValid != (beforeChangeNumValid + 1) {
				return errors.New("image validation valid count is off by one!")
			}
		} else {
			if afterChangeNumInvalid != (beforeChangeNumInvalid + 1) {
				return errors.New("image validation invalid count is off by one!")
			}
		}
	}

	return nil
}

func testRandomModeratedImageValidation(num int, token string) error {
	for i := 0; i < num; i++ {


		randomValidationId, err := db.GetRandomValidationId()
		if err != nil {
			return err
		}

		_, beforeChangeNumInvalid, err := db.GetValidationCount(randomValidationId)
		if err != nil {
			return err
		}

		if err := testImageValidation(randomValidationId, "no", true, token); err != nil {
			return err
		}

		_, afterChangeNumInvalid, err := db.GetValidationCount(randomValidationId)
		if err != nil {
			return err
		}


		if afterChangeNumInvalid != (beforeChangeNumInvalid + 5) {
			return errors.New("image validation invalid count is off!")
		}
	}

	return nil
}

func testRandomAnnotate(num int, annotations string) error {
	type Annotation struct {
		Annotations []json.RawMessage `json:"annotations"`
		Label string `json:"label"`
		Sublabel string `json:"sublabel"`
	}

	for i := 0; i < num; i++ {
		annotationRow, err := db.GetRandomImageForAnnotation()
		if err != nil {
			return err
		}

		annotationEntry := Annotation{Label: annotationRow.Validation.Label, 
										Sublabel: annotationRow.Validation.Sublabel}

		err = json.Unmarshal([]byte(annotations), &annotationEntry.Annotations)
		if err != nil {
			return err
		}

		url := BASE_URL + API_VERSION + "/annotate/" + annotationRow.Image.Id

		 resp, err := resty.R().
					SetHeader("Content-Type", "application/json").
					SetBody(annotationEntry).
					Post(url)
		if err != nil {
			return err
		}

		if resp.StatusCode() != 201 {
			return errors.New("Couldn't annotate image " + annotationRow.Image.Id)
		}

		//export annotations again
		url = resp.Header().Get("Location")
		resp, err = resty.R().
					SetHeader("Content-Type", "application/json").
					SetResult(&AnnotatedImage{}).
					Get(url)
		if err != nil {
			return err
		}

		if resp.StatusCode() != 200 {
			return errors.New("Couldn't verify annotate image " + annotationRow.Image.Id)
		}

		j, err := json.Marshal(&resp.Result().(*AnnotatedImage).Annotations)
	    if err != nil {
	        return err
	    }

		equal, err := equalJson(string(j), annotations)
		if !equal {
			return errors.New("Exported annotations do not match!")
		}

		if err != nil {
			return err
		}

		/*j, err := json.Marshal(&resp.Result().(*AnnotatedImage).Annotations)
	    if err != nil {
	        return err
	    }

	    if string(j) != annotations {
	    	return errors.New("Exported annotations do not match!")
	    }*/
	}

	return nil
}

func testRandomAnnotationRework(num int, annotations string) error {
	type Annotation struct {
		Annotations []json.RawMessage `json:"annotations"`
	}

	for i := 0; i < num; i++ {
		annotationId, err := db.GetRandomAnnotationId()
		if err != nil {
			return err
		}

		oldAnnotationRevision, err := db.GetAnnotationRevision(annotationId)
		if err != nil {
			return err
		}

		oldAnnotationDataIds, err := db.GetAnnotationDataIds(annotationId)
		if err != nil {
			return err
		}

		annotationEntry := Annotation{}

		err = json.Unmarshal([]byte(annotations), &annotationEntry.Annotations)
		if err != nil {
			return err
		}

		url := BASE_URL + API_VERSION + "/annotation/" + annotationId
		resp, err := resty.R().
					SetHeader("Content-Type", "application/json").
					SetBody(annotationEntry).
					Put(url)
		if err != nil {
			return err
		}

		if resp.StatusCode() != 201 {
			return errors.New("Couldn't rework annotation entry")
		}

		newAnnotationRevision, err := db.GetAnnotationRevision(annotationId)
		if err != nil {
			return err
		}

		newAnnotationDataIds, err := db.GetOldAnnotationDataIds(annotationId, oldAnnotationRevision)
		if err != nil {
			return err
		}

		if newAnnotationRevision != (oldAnnotationRevision + 1) {
			return errors.New("annotation revision does not match expected value!")
		}

		equal := reflect.DeepEqual(oldAnnotationDataIds, newAnnotationDataIds)

		if !equal {
			return errors.New("annotation data ids changed, although they shouldn't!")
		}

	}

	return nil
}

func main(){
	db = NewImageMonkeyDatabase()

	fmt.Printf("Initializing database...\n")
	err := db.Initialize()
	if err != nil {
		panic(err)
	}

	err = db.Open()
	if err != nil {
		log.Fatal("[Main] Couldn't open database: ", err.Error())
	}

	defer db.Close()

	fmt.Printf("Testing signup...\n")
	err = testSignUp("testuser", "testpassword", "testuser@imagemonkey.io")
	if err == nil {
		fmt.Printf("[SUCCESS] Successfully signed up\n")
	} else {
		fmt.Printf("[FAIL] Couldn't signup: %s\n", err.Error())
		return
	}

	fmt.Printf("Testing login...\n")
	_,err = testLogin("testuser", "testpassword")
	if err == nil {
		fmt.Printf("[SUCCESS] Successfully logged in\n")
	} else {
		fmt.Printf("[FAIL] Login failed: %s\n", err.Error())
		return
	}

	fmt.Printf("Testing Image Donation...\n")
	err = testMultipleDonate()
	if err == nil {
		fmt.Printf("[SUCCESS] Successfully donated images\n")
	} else {
		fmt.Printf("[FAIL] Couldn't donate images: %s\n", err.Error())
		return
	}

	err = testRandomImageValidation(100)
	if err == nil {
		fmt.Printf("[SUCCESS] Random Image validation test succeeded\n")
	} else {
		fmt.Printf("[FAIL] Random Image validation test failed %s\n", err.Error())
		return
	}

	fmt.Printf("Add user 'moderator'...\n")
	err = testSignUp("moderator", "moderator", "moderator@imagemonkey.io")
	if err == nil {
		fmt.Printf("[SUCCESS] user 'moderator' successfully added\n")
	} else {
		fmt.Printf("[FAIL] Couldn't add user 'moderator': %s\n", err.Error())
		return
	}

	fmt.Printf("Logging in with user 'moderator'...\n")
	moderatorToken, err := testLogin("moderator", "moderator")
	if err == nil {
		fmt.Printf("[SUCCESS] Successfully logged in user 'moderator'\n")
	} else {
		fmt.Printf("[FAIL] Couldn't login user 'moderator': %s\n", err.Error())
		return
	}

	fmt.Printf("Giving user 'moderator' moderator permissions...\n")
	err = db.GiveUserModeratorRights("moderator")
	if err == nil {
		fmt.Printf("[SUCCESS] Successfully gave user 'moderator' moderator permissions\n")
	} else {
		fmt.Printf("[FAIL] Couldn't give user 'moderator' moderator permissions: %s\n", err.Error())
		return
	}

	err = testRandomModeratedImageValidation(100, moderatorToken)
	if err == nil {
		fmt.Printf("[SUCCESS] Random moderated Image validation test succeeded\n")
	} else {
		fmt.Printf("[FAIL] Random moderated Image validation test failed %s\n", err.Error())
		return
	}



	err = testRandomAnnotate(2, `[{"top":100,"left":200,"type":"rect","angle":0,"width":40,"height":60,"stroke":{"color":"red","width":1}}]`)
	if err == nil {
		fmt.Printf("[SUCCESS] Random annotate test succeeded\n")
	} else {
		fmt.Printf("[FAIL] Random annotate test failed %s\n", err.Error())
		return
	}


	err = testRandomAnnotationRework(2, `[{"top":200,"left":300,"type":"rect","angle":10,"width":50,"height":30,"stroke":{"color":"blue","width":3}}]`)
	if err == nil {
		fmt.Printf("[SUCCESS] Random annotate rework test succeeded\n")
	} else {
		fmt.Printf("[FAIL] Random annotate rework test failed %s\n", err.Error())
		return
	}









	/*uuid := "48750a63-df1f-48d1-99ee-6c60e535a271"
	fmt.Printf("Testing Image Reporting...\n")
	if(testReport(uuid, "b") == nil){
		fmt.Printf("Successfully reported image with uuid %s\n", uuid)
	} else {
		fmt.Printf("Reporting image with uuid %s FAILED\n", uuid)
		return
	}

	fmt.Printf("Testing Image Donation...\n")
	err = testMultipleDonate()
	if err == nil {
		fmt.Printf("Successfully donated images\n")
	} else {
		fmt.Printf("Donating images FAILED: %s\n", err.Error())
		return
	}

	fmt.Printf("Testing Image Validation...\n")
	if(testValidate() == nil){
		fmt.Printf("Successfully received an image to validate\n")
	} else {
		fmt.Printf("Couldn't get image to validate\n")
		return
	}

	fmt.Printf("Testing Image Validation (UUID)...\n")
	if(testValidateWithUuid(uuid) == nil){
		fmt.Printf("Successfully received an image to validate\n")
	} else {
		fmt.Printf("Couldn't get image to validate\n")
		return
	}*/

}