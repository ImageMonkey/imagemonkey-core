package main 

import (
	"fmt"
	"log"
	"database/sql"
	"encoding/json"
	_"github.com/lib/pq"
)

type Annotation struct{
    Left float32 `json:"left"`
    Top float32 `json:"top"`
    Width float32 `json:"width"`
    Height float32 `json:"height"`
}

type Annotations struct{
    Annotations []Annotation `json:"annotations"`
    Id int64 `json:"id"`
}


var db *sql.DB

func migrate() error {
	tx, err := db.Begin()
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

	for _, elem := range allAnnotations {
		var byt []byte
		fmt.Printf("Migrating annotations for id = %d\n", elem.Id)

		byt, err := json.Marshal(elem.Annotations)
        if err != nil {
            log.Fatal("[Main] Couldn't unmarshal: ", err.Error())
            return err
        }
        
        _, err = tx.Exec(`INSERT INTO annotation_data(image_annotation_id, annotation, annotation_type_id)
        					SELECT $1, q.*, 1 FROM json_array_elements($2) q`, elem.Id, byt)
        if err != nil {
        	log.Fatal("[Main] Couldn't insert: ", err.Error())
            return err
        }
	}



	return tx.Commit()
}

func main() {
	fmt.Printf("Starting migration...\n")

	var err error
	//open database and make sure that we can ping it
	db, err = sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
	if err != nil {
		log.Fatal("[Main] Couldn't open database: ", err.Error())
		return
	}

	err = migrate()
	if err != nil {
		log.Fatal("[Main] Couldn't migrate data: ", err.Error())
		return
	}

	fmt.Printf("Sucessfully migrated data\n")
}