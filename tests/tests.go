package main

import (
	"fmt"
	"gopkg.in/resty.v0"
)


const BASE_URL string = "http://127.0.0.1:8081/"
const API_VERSION string = "v1"

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

func testDonate() error{
	url := BASE_URL + API_VERSION + "/donate"
	_, err := resty.R().
      SetFile("image", "./eggs.jpg").
      SetFormData(map[string]string{
        "label": "egg",
      }).
      Post(url)

    return err
}

func testValidate() error{
	url := BASE_URL + API_VERSION + "/validate"
	_, err := resty.R().
			SetResult(&ValidateResult{}).
			Get(url)
	return err
}

func testValidateWithUuid(uuid string) error{
	url := BASE_URL + API_VERSION + "/validate/" + uuid
	_, err := resty.R().
			SetResult(&ValidateResult{}).
			Get(url)
	return err
}

func main(){
	uuid := "48750a63-df1f-48d1-99ee-6c60e535a271"
	fmt.Printf("Testing Image Reporting...\n")
	if(testReport(uuid, "b") == nil){
		fmt.Printf("Successfully reported image with uuid %s\n", uuid)
	} else {
		fmt.Printf("Reporting image with uuid %s FAILED\n", uuid)
	}

	fmt.Printf("Testing Image Donation...\n")
	if(testDonate() == nil){
		fmt.Printf("Successfully donated image\n")
	} else {
		fmt.Printf("Donating image FAILED\n")
	}

	fmt.Printf("Testing Image Validation...\n")
	if(testValidate() == nil){
		fmt.Printf("Successfully received an image to validate\n")
	} else {
		fmt.Printf("Couldn't get image to validate\n")
	}

	fmt.Printf("Testing Image Validation (UUID)...\n")
	if(testValidateWithUuid(uuid) == nil){
		fmt.Printf("Successfully received an image to validate\n")
	} else {
		fmt.Printf("Couldn't get image to validate\n")
	}

}