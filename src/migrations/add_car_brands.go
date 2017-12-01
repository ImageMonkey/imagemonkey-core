package main 

import (
	"fmt"
	"log"
	"database/sql"
	//"encoding/json"
	_"github.com/lib/pq"
)



var db *sql.DB

func addCarBrands() error {
	question := "It's a..."
	allowOther := true
	allowUnknown := true
	controlType := "dropdown"
	mainLabel := "car"


	tx, err := db.Begin()
    if err != nil {
    	log.Fatal("[Main] Couldn't begin transaction: ", err.Error())
        return err
    }

	carBrands := []string{"Seat", "Renault", "Peugot", "BMW", "Ford", "Opel", "Alfa Romeo", "Chevrolet", "Porsche", "Honda", "Subaru", "Mazda", "Mitsubishi", "Lexus", "Toyota", "Volkswagen", "Suzuki", "Mercedes-Benz", "Saab", "Audi", "Kia", "Land Rover", "Doge", "Chrysler", "Hummer", "Hyundai", "Jaguar", "Jeep", "Nissan", "Volvo", "Daewoo", "Fiat", "MINI", "Smart"}

	
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
	for _,brand := range carBrands {
		err = tx.QueryRow(`insert into label (name, parent_id)
  								select $1, id FROM label WHERE name = $2 RETURNING id`, brand, mainLabel).Scan(&insertedLabelId)
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
	/*tx, err := db.Begin()
    if err != nil {
    	log.Fatal("[Main] Couldn't begin transaction: ", err.Error())
        return err
    }


	rows, err := tx.Query(`SELECT id, annotations FROM image_annotation`)
	if err != nil {
		log.Fatal("[Main] Couldn't select annotations: ", err.Error())
		return err
	}

	defer rows.Close()

	var allAnnotations []Annotations
	for rows.Next() {
		var byt []byte
		var annotations Annotations
		err = rows.Scan(&annotations.Id, &byt)

		err := json.Unmarshal(byt, &annotations.Annotations)
        if err != nil {
            log.Fatal("[Main] Couldn't unmarshal: ", err.Error())
            return err
        }

        allAnnotations = append(allAnnotations, annotations)
	}

	rows.Close()

	return tx.Commit()*/
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

	err = addCarBrands()
	if err != nil {
		log.Fatal("[Main] Couldn't add data: ", err.Error())
		return
	}

	fmt.Printf("Sucessfully added data\n")
}