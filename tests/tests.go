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

func (p *ImageMonkeyDatabase) GetNumberOfImages() (int32, error) {
	var numOfImages int32
	err := p.db.QueryRow("SELECT count(*) FROM image").Scan(&numOfImages)
	if err != nil {
		return 0, err
	}

	return numOfImages, err
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

func testImageValidation(uuid string, param string) error{
	url := BASE_URL + API_VERSION + "/validation/" + uuid + "/validate/" + param
	resp, err := resty.R().
			Post(url)
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

		if err := testImageValidation(randomValidationId, param); err != nil {
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

func main(){
	db = NewImageMonkeyDatabase()

	err := db.Open()
	if err != nil {
		log.Fatal("[Main] Couldn't open database: ", err.Error())
	}

	defer db.Close()

	fmt.Printf("Testing Image Donation...\n")
	err = testMultipleDonate()
	if err == nil {
		fmt.Printf("Successfully donated images\n")
	} else {
		fmt.Printf("Donating images FAILED: %s\n", err.Error())
		return
	}

	err = testRandomImageValidation(100)
	if err == nil {
		fmt.Printf("[SUCCESS] Random Image validation test succeeded\n")
		return
	} else {
		fmt.Printf("[FAIL] Random Image validation test failed %s\n", err.Error())
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