package main 

import (
	"fmt"
	"log"
	"database/sql"
	"encoding/json"
	_"github.com/lib/pq"
)

type QuizExampleEntry struct {
	Filename string `json:"filename"`
	Attribution string `json:"attribution"`
}

type QuizAnswerEntry struct {
	Name string `json:"name"`
	Examples []QuizExampleEntry `json:"examples"`
}

type QuizQuestionEntry struct {
	Question string `json:"question"`
	Answers []QuizAnswerEntry `json:"answers"`
	ControlType string `json:"control_type"`
	AllowUnknown bool `json:"allow_unknown"`
	AllowOther bool `json:"allow_other"`
	BrowseByExample bool `json:"browse_by_example"`
}


var db *sql.DB

/*
delete from label_example;
delete from quiz_answer where id > 152;
delete from quiz_question where id > 77;
delete from label where id > 315;
*/

func addDogBreeds() error {
	/*question := "What am I?"
	allowOther := true
	allowUnknown := true
	browseByExample := true
	controlType := "dropdown"*/
	mainLabel := "dog"


	tx, err := db.Begin()
    if err != nil {
    	log.Fatal("[Main] Couldn't begin transaction: ", err.Error())
        return err
    }

    var quizQuestionEntry QuizQuestionEntry
	var jsonBlob = []byte(`{
					"question": "What am I?",
					"answers": [
						{
							"name": "Labrador Retriever",
							"examples": [
								{
									"filename": "LabradorRetriever.jpg"								
								}
							]
						},
						{
							"name": "English Cocker Spaniel",
							"examples": [
								{
									"filename": "EnglishCockerSpaniel.jpg"								
								}
							]
						},
						{
							"name": "English Springer Spaniel",
							"examples": [
								{
									"filename": "EnglishSpringerSpaniel.jpg"								
								}
							]
						},
						{
							"name": "German Shepherd",
							"examples": [
								{
									"filename": "GermanShepherd.jpg"
								}
							]
						},
						{
							"name": "Staffordshire Bull Terrier",
							"examples": [
								{
									"filename": "StaffordshireBullTerrier.jpg"
								}
							]
						},
						{
							"name": "Golden Retriever",
							"examples": [
								{
									"filename": "GoldenRetriever.jpg"
								}
							]
						},
						{
							"name": "Boxer",
							"examples": [
								{
									"filename": "Boxer.jpg"
								}
							]
						},
						{
							"name": "Beagle",
							"examples": [
								{
									"filename": "Beagle.jpg"
								}
							]
						},
						{
							"name": "Dachshund",
							"examples": [
								{
									"filename": "Dachshund.jpg"
								}
							]
						},
						{
							"name": "Poodle",
							"examples": [
								{
									"filename": "Poodle.jpg"
								}
							]
						},
						{
							"name": "Rottweiler",
							"examples": [
								{
									"filename": "Rottweiler.jpg"
								}
							]
						},
						{
							"name": "Siberian Husky",
							"examples": [
								{
									"filename": "SiberianHusky.jpg"
								}
							]
						},
						{
							"name": "Bulldog",
							"examples": [
								{
									"filename": "Bulldog.jpg"
								}
							]
						},
						{
							"name": "Mops",
							"examples": [
								{
									"filename": "Mops.jpg"
								}
							]
						},
						{
							"name": "Dalmatiner",
							"examples": [
								{
									"filename": "Dalmatiner.jpg"
								}
							]
						},
						{
							"name": "Great Dane",
							"examples": [
								{
									"filename": "GreatDane.jpg"
								}
							]
						}

						

						
					],
					"allow_unknown": true,
					"allow_other": true,
					"browse_by_example": true,
					"control_type": "dropdown"
					}`)
	

	err = json.Unmarshal(jsonBlob, &quizQuestionEntry)
	if err != nil {
		log.Fatal("Couldn't unmarshal data", err.Error())
	}

	
	var insertedQuizQuestionId int64
	err = tx.QueryRow(`insert into quiz_question (question, recommended_control, allow_unknown, allow_other, browse_by_example) values($1, $2, $3, $4, $5) RETURNING id`, 
							quizQuestionEntry.Question, quizQuestionEntry.ControlType, quizQuestionEntry.AllowUnknown, quizQuestionEntry.AllowOther, 
							quizQuestionEntry.BrowseByExample).Scan(&insertedQuizQuestionId)
	if err != nil {
		tx.Rollback()
		log.Fatal("[Main] Couldn't insert quiz_question: ", err.Error())
		return err
	}

	_, err = tx.Exec(`update quiz_question set refines_label_id = (select id from label l where l.name = $1 and l.parent_id is null) where id = $2`,
							mainLabel, insertedQuizQuestionId)
	if err != nil {
		tx.Rollback()
		log.Fatal("[Main] Couldn't update quiz_question: ", err.Error())
		return err
	}


	var insertedLabelId int64
	for _, answer := range quizQuestionEntry.Answers {
		err = tx.QueryRow(`insert into label (name, parent_id)
  								select $1, id FROM label WHERE name = $2 RETURNING id`, answer.Name, mainLabel).Scan(&insertedLabelId)
		if err != nil {
			tx.Rollback()
			log.Fatal("[Main] Couldn't insert labels: ", err.Error())
			return err
		}


		_, err = tx.Exec(`insert into quiz_answer (label_id, quiz_question_id) VALUES($1, $2)`, insertedLabelId, insertedQuizQuestionId)
		if err != nil {
			tx.Rollback()
			log.Fatal("[Main] Couldn't insert quiz_answer: ", err.Error())
			return err
		}

		for _, example := range answer.Examples {
			_, err = tx.Exec(`insert into label_example (label_id, filename) VALUES($1, $2)`, insertedLabelId, example.Filename)
			if err != nil {
				tx.Rollback()
				log.Fatal("[Main] Couldn't insert quiz_answer: ", err.Error())
				return err
			}
		}

  
	}


	return tx.Commit()
}

func main() {
	fmt.Printf("Starting adding...\n")

	var err error
	//open database and make sure that we can ping it
	db, err = sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
	if err != nil {
		log.Fatal("[Main] Couldn't open database: ", err.Error())
		return
	}

	err = addDogBreeds()
	if err != nil {
		log.Fatal("[Main] Couldn't add data: ", err.Error())
		return
	}

	fmt.Printf("Sucessfully added data\n")
}