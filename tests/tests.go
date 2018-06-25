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
)


const BASE_URL string = "http://127.0.0.1:8081/"
const API_VERSION string = "v1"

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

	/*_, err = localDb.Exec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s", pq.QuoteIdentifier("uuid-ossp")))
	if err != nil {
		return err
	}*/

	err = installUuidExtension()
	if err != nil {
		return err
	}

	/*_, err = localDb.Exec("GRANT ALL PRIVILEGES ON DATABASE imagemonkey to monkey")
	if err != nil {
		return err
	}

	_, err = localDb.Exec("GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO monkey")
	if err != nil {
		return err
	}

	_, err = localDb.Exec("GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO monkey")
	*/

	
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
	if err != nil {
		return numOfYes, numOfNo, err
	}

	return numOfYes, numOfNo, nil
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