package imagemonkeydb

import (
    "github.com/getsentry/raven-go"
    log "github.com/Sirupsen/logrus"
    "../datastructures"
    "../commons"
    "github.com/lib/pq"
)

type AddImageCollectionErrorType int

const (
  AddImageCollectionSuccess AddImageCollectionErrorType = 1 << iota
  AddImageCollectionInternalError
  AddImageCollectionAlreadyExistsError
)


type AddImageToImageCollectionErrorType int

const (
  AddImageToImageCollectionSuccess AddImageToImageCollectionErrorType = 1 << iota
  AddImageToImageCollectionInternalError
  AddImageToImageCollectionInvalidInputError
  AddImageToImageCollectionDuplicateEntryError
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
							 	SELECT u.id as user_image_collection_id, MIN(i.key) as image_key, 
							 	i.width as image_width, i.height as image_height, i.unlocked as image_unlocked
							 	FROM image i
							 	JOIN image_collection_image im ON im.image_id = i.id
							 	JOIN user_image_collection u ON u.id = im.user_image_collection_id
							 	GROUP BY u.id, i.width, i.height, i.unlocked
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

func (p *ImageMonkeyDatabase) AddImageCollection(apiUser datastructures.APIUser, name string, description string) AddImageCollectionErrorType {
	_, err := p.db.Exec(`INSERT INTO user_image_collection(account_id, name, description)
							SELECT id, $2, $3 FROM account WHERE name = $1`, apiUser.Name, name, description)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23505" {
				return AddImageCollectionAlreadyExistsError
			}
		}
		log.Error("[Image Collections] Couldn't add image collection: ", err.Error())
        raven.CaptureError(err, nil)
        return AddImageCollectionInternalError
	}

	return AddImageCollectionSuccess
}

func (p *ImageMonkeyDatabase) AddImageToImageCollection(apiUser datastructures.APIUser, 
	imageCollectionName string, imageId string) AddImageToImageCollectionErrorType {

	_, err := p.db.Exec(`INSERT INTO image_collection_image(user_image_collection_id, image_id)
						 SELECT (SELECT u.id 
						   		 FROM user_image_collection u 
						   		 JOIN account a ON u.account_id = a.id
						   		 WHERE u.name = $1 AND a.name = $2), (SELECT id FROM image WHERE key = $3)`, 
						   		 imageCollectionName, apiUser.Name, imageId)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == "23502" {
				return AddImageToImageCollectionInvalidInputError
			} else if err.Code == "23505" {
				return AddImageToImageCollectionDuplicateEntryError	
			}
		}
		log.Error("[Add Image To Collection] Couldn't add image to collection: ", err.Error())
        raven.CaptureError(err, nil)
        return AddImageToImageCollectionInternalError
	}

	return AddImageToImageCollectionSuccess
}