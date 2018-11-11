package main

import (
	log "github.com/Sirupsen/logrus"
	"flag"
	"database/sql"
	"errors"
	"github.com/lib/pq"
)

var db *sql.DB

func doLabelTransition(tx *sql.Tx, labelSuggestion string, imageCollectionName string, imageCollectionDescription string, username string) error {
	rows, err := tx.Query(`SELECT i.id, s.id
			  				FROM image i
			  				JOIN image_label_suggestion s ON s.image_id = i.id
			  				JOIN label_suggestion l ON l.id = s.label_suggestion_id
			  				WHERE l.name = $1`, labelSuggestion)
	if err != nil {
		return err
	}

	defer rows.Close()

	var imageIds []int
	var imageLabelSuggestionIds []int
	for rows.Next() {
		var imageId int
		var imageLabelSuggestionId int

		err = rows.Scan(&imageId, &imageLabelSuggestionId)
		if err != nil {
			return err
		}

		imageIds = append(imageIds, imageId)
		imageLabelSuggestionIds = append(imageLabelSuggestionIds, imageLabelSuggestionId)
	}
	rows.Close()

	//get user collection if exists
	userCollectionRows, err := tx.Query(`SELECT u.id 
						  				 FROM user_image_collection u 
						  				 JOIN account a ON a.id = u.account_id
						  				 WHERE a.name = $1 AND u.name = $2`, username, imageCollectionName)
	if err != nil {
		return err
	}

	defer userCollectionRows.Close()

	var userCollectionId int
	userCollectionId = -1
	if userCollectionRows.Next() {
		err = userCollectionRows.Scan(&userCollectionId)
		if err != nil {
			return err
		}
	} else {
		err = tx.QueryRow(`INSERT INTO user_image_collection(account_id, name, description)
								SELECT id, $2, $3 FROM account a WHERE a.name = $1 RETURNING id`, 
								username, imageCollectionName, imageCollectionDescription).Scan(&userCollectionId)
		if err != nil {
			return err
		}
	}

	userCollectionRows.Close()

	if userCollectionId == -1 {
		return errors.New("User collection id is empty!")
	}


	_, err = tx.Exec(`INSERT INTO image_collection_image(user_image_collection_id, image_id)
						SELECT $1, unnest($2::integer[])`, userCollectionId, pq.Array(imageIds))
	if err != nil {
		return err
	}

	log.Info("[Main] Adding ", len(imageIds), " images to image collection ", imageCollectionName)

	_, err = tx.Exec(`DELETE FROM image_label_suggestion WHERE id = ANY($1)`, pq.Array(imageLabelSuggestionIds))
	if err != nil {
		return err
	}

	log.Info("[Main] Deleting ", len(imageLabelSuggestionIds), " image label suggestions")

	return nil
}

func main() {
	log.SetLevel(log.DebugLevel)

	labelSuggestion := flag.String("labelsuggestion", "", "The name of the label that will be transitioned.")
	username := flag.String("username", "", "The user ")
	imageCollectionName := flag.String("imagecollection", "", "The name of the image collection the images will be attached to.")
	dryRun := flag.Bool("dryrun", true, "Specifies whether this is a dryrun or not")

	flag.Parse()

	if *labelSuggestion == "" {
		log.Fatal("Please provide a label suggestion")
	}

	if *username == "" {
		log.Fatal("Please provide a username")
	}

	if *imageCollectionName == "" {
		log.Fatal("Please provide a image collection name")
	}

	log.Info("[Main] Attaching labels to image collection")
	var err error
	db, err = sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
	if err != nil {
		log.Error(err)
		return
	}

	err = db.Ping()
	if err != nil {
		log.Error("[Main] Couldn't ping database: ", err.Error())
		return
	}
	defer db.Close()

	tx, err := db.Begin()
    if err != nil {
    	log.Error("[Main] Couldn't begin trensaction: ", err.Error())
        return
    }

    imageCollectionDescription := ""
    err = doLabelTransition(tx, *labelSuggestion, *imageCollectionName, imageCollectionDescription, *username)
    if err != nil {
    	log.Error("[Main] Couldn't perform transition: ", err.Error())
    	tx.Rollback()
    	return
    }

    if *dryRun {
    	err = tx.Rollback()
		if err != nil {
			log.Error("[Main] Couldn't rollback transaction: ", err.Error())
			return
		}

		log.Info("[Main] This was just a dry run - rolling back everything")
	} else {
		err = tx.Commit()
		if err != nil {
			log.Error("[Main] Couldn't commit transaction: ", err.Error())
			return
		}
	}


}