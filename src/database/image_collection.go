package imagemonkeydb

import (
	"database/sql"
	commons "github.com/bbernhard/imagemonkey-core/commons"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	"github.com/getsentry/raven-go"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"fmt"
	"time"
)

func (p *ImageMonkeyDatabase) GetImageCollections(apiUser datastructures.APIUser, apiBaseUrl string) ([]datastructures.ImageCollection, error) {
	imageCollections := []datastructures.ImageCollection{}

	rows, err := p.db.Query(`SELECT u.name, u.description, COALESCE(q.num, 0), COALESCE(q1.image_key, ''),
							 COALESCE(q1.image_width, 0), COALESCE(q1.image_height, 0), COALESCE(q1.image_unlocked, false)
							 FROM user_image_collection u
							 JOIN account a ON a.id = u.account_id
							 LEFT JOIN (
							 	SELECT c.user_image_collection_id as user_image_collection_id, count(*) as num 
							 	FROM image_collection_image c
							 	JOIN image i ON i.id = c.image_id
							 	GROUP BY c.user_image_collection_id
							 ) q ON q.user_image_collection_id = u.id
							 LEFT JOIN (
							 	SELECT q1.user_image_collection_id, i.key as image_key, i.width as image_width, 
							 	i.height as image_height, i.unlocked as image_unlocked
								FROM (
										SELECT u.id as user_image_collection_id, MIN(i.id) as image_id
							 			FROM image i
							 			JOIN image_collection_image im ON im.image_id = i.id
							 			JOIN user_image_collection u ON u.id = im.user_image_collection_id
							 			GROUP BY u.id
								) q1
								JOIN image i ON q1.image_id = i.id
							 ) q1 ON q1.user_image_collection_id = u.id 
							 WHERE a.name = $1`, apiUser.Name)
	if err != nil {
		log.Error("[Image Collections] Couldn't get image collections: ", err.Error())
		raven.CaptureError(err, nil)
		return imageCollections, err
	}

	defer rows.Close()

	for rows.Next() {
		var imageCollection datastructures.ImageCollection
		err = rows.Scan(&imageCollection.Name, &imageCollection.Description, &imageCollection.Count,
			&imageCollection.SampleImage.Id, &imageCollection.SampleImage.Width, &imageCollection.SampleImage.Height,
			&imageCollection.SampleImage.Unlocked)
		if err != nil {
			log.Error("[Image Collections] Couldn't scan image collection: ", err.Error())
			raven.CaptureError(err, nil)
			return imageCollections, err
		}

		if imageCollection.SampleImage.Id == "" {
			imageCollection.SampleImage.Id = ""
			imageCollection.SampleImage.Unlocked = true
			imageCollection.SampleImage.Width = 960
			imageCollection.SampleImage.Height = 796
			imageCollection.SampleImage.Url = "/img/default.png"
		} else {
			imageCollection.SampleImage.Url = commons.GetImageUrlFromImageId(apiBaseUrl, imageCollection.SampleImage.Id, imageCollection.SampleImage.Unlocked)
		}

		imageCollections = append(imageCollections, imageCollection)
	}

	return imageCollections, nil
}

func (p *ImageMonkeyDatabase) _addImageCollectionInTransaction(tx *sql.Tx, username string, name string, description string) error {
	_, err := tx.Exec(`INSERT INTO user_image_collection(account_id, name, description)
							SELECT id, $2, $3 FROM account WHERE name = $1`, username, name, description)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				tx.Rollback()
				return &DuplicateImageCollectionError{Description: "An Image Collection with that name already exists"}
			}
		}
		tx.Rollback()
		return err
	}

	return nil
}

func (p *ImageMonkeyDatabase) AddImageCollection(apiUser datastructures.APIUser, name string, description string) error {
	tx, err := p.db.Begin()
	if err != nil {
		log.Error("[Image Collections] Couldn't begin transaction: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	err = p._addImageCollectionInTransaction(tx, apiUser.Name, name, description)
	if err != nil { //transaction already rolled back
		log.Error("[Image Collections] Couldn't add image collection: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Error("[Image Collections] Couldn't commit transaction: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	return nil
}

func (p *ImageMonkeyDatabase) _addImageToImageCollectionInTransaction(tx *sql.Tx, username string,
	imageCollectionName string, imageId string, failIfAlreadyAssigned bool) error {

	currentUnixTimestamp := int64(time.Now().Unix())

	queryValues := []interface{}{imageCollectionName, username, imageId}

	q1 := ""
	if !failIfAlreadyAssigned {
		q1 = "ON CONFLICT(user_image_collection_id, image_id) DO UPDATE SET last_modified = $4"
		queryValues = append(queryValues, currentUnixTimestamp)
	}

	q := fmt.Sprintf(`INSERT INTO image_collection_image(user_image_collection_id, image_id)
						 SELECT (SELECT u.id 
						   		 FROM user_image_collection u 
						   		 JOIN account a ON u.account_id = a.id
						   		 WHERE u.name = $1 AND a.name = $2), (SELECT id FROM image WHERE key = $3)
					  %s`, q1)
	_, err := tx.Exec(q, queryValues...)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23502" {
				tx.Rollback()
				return &InvalidImageCollectionInputError{Description: "Invalid Image Collection Input"}
			} else if err.Code == "23505" {
				tx.Rollback()
				return &DuplicateImageCollectionError{Description: "Image already assigned to Image Collection"}
			}
		}
		tx.Rollback()
		return err
	}

	return nil
}

func (p *ImageMonkeyDatabase) AddImageToImageCollection(apiUser datastructures.APIUser,
	imageCollectionName string, imageId string, failIfAlreadyAssigned bool) error {

	tx, err := p.db.Begin()
	if err != nil {
		log.Error("[Add Image To Collection] Couldn't begin transaction: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	err = p._addImageToImageCollectionInTransaction(tx, apiUser.Name, imageCollectionName, imageId, failIfAlreadyAssigned)
	if err != nil { //transaction already rolled back
		log.Error("[Add Image to Collection] Couldn't add image to image collection: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Error("[Add Image to Collection] Couldn't commit transaction: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	return nil
}
