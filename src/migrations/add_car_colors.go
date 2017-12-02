package main 

import (
	"fmt"
	"log"
	"database/sql"
	//"encoding/json"
	_"github.com/lib/pq"
)



var db *sql.DB

func addCarColors() error {
	question := "The color of the car is..."
	allowOther := true
	allowUnknown := true
	controlType := "color tags"
	mainLabel := "car"


	tx, err := db.Begin()
    if err != nil {
    	log.Fatal("[Main] Couldn't begin transaction: ", err.Error())
        return err
    }

	carColors := []string{"white", "silver", "black", "grey", "blue", "red", "brown", "green"}

	
	var insertedQuizQuestionId int64
	err = tx.QueryRow(`insert into quiz_question (question, recommended_control, allow_unknown, allow_other) values($1, $2, $3, $4) RETURNING id`, 
							question, controlType, allowUnknown, allowOther).Scan(&insertedQuizQuestionId)
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
	for _,carColor := range carColors {
		err = tx.QueryRow(`insert into label (name, parent_id)
  								select $1, id FROM label WHERE name = $2 RETURNING id`, carColor, mainLabel).Scan(&insertedLabelId)
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

	err = addCarColors()
	if err != nil {
		log.Fatal("[Main] Couldn't add data: ", err.Error())
		return
	}

	fmt.Printf("Sucessfully added data\n")
}