package main

import (
    "github.com/lib/pq"
	"github.com/getsentry/raven-go"
	log "github.com/Sirupsen/logrus"
    "encoding/json"
    //"errors"
    //"database/sql/driver"
)

type Annotation struct{
    Left float32 `json:"left"`
    Top float32 `json:"top"`
    Width float32 `json:"width"`
    Height float32 `json:"height"`
}

type Image struct {
    Id string `json:"uuid"`
    Label string `json:"label"`
    Provider string `json:"provider"`
    Probability float32 `json:"probability"`
    NumOfValid int32 `json:"num_yes"`
    NumOfInvalid int32 `json:"num_no"`
    Annotations []Annotation `json:"annotations"`
}

type ImageValidation struct {
    Uuid string `json:"uuid"`
    Valid string `json:"valid"`
}

type GraphNode struct {
	Group int `json:"group"`
	Text string `json:"text"`
	Size int `json:"size"`
}

func addDonatedPhoto(filename string, hash uint64, label string) error{
	tx, err := db.Begin()
    if err != nil {
    	log.Debug("[Adding donated photo] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    //PostgreSQL can't store unsigned 64bit, so we are casting the hash to a signed 64bit value when storing the hash (so values above maxuint64/2 are negative). 
    //this should be ok, as we do not need to order those values, but just need to check if a hash exists. So it should be fine
	imageId := 0
	err = tx.QueryRow("INSERT INTO image(key, unlocked, image_provider_id, hash) SELECT $1, $2, p.id, $3 FROM image_provider p WHERE p.name = $4 RETURNING id", 
					  filename, false, int64(hash), "donation").Scan(&imageId)
	if(err != nil){
		log.Debug("[Adding donated photo] Couldn't insert image: ", err.Error())
		raven.CaptureError(err, nil)
		tx.Rollback()
		return err
	}

	labelId := 0
	err = tx.QueryRow("INSERT INTO image_validation(image_id, num_of_valid, num_of_invalid, label_id) SELECT $1, $2, $3, l.id FROM label l WHERE l.name = $4 RETURNING id", 
					  imageId, 0, 0, label).Scan(&labelId)
	if(err != nil){
		tx.Rollback()
		log.Debug("[Adding donated photo] Couldn't insert image validation entry: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	return tx.Commit()
}

func imageExists(hash uint64) (bool, error){
    //PostgreSQL can't handle unsigned 64bit, so we are casting the hash to a signed 64bit value when comparing against the stored hash (so values above maxuint64/2 are negative). 
    rows, err := db.Query("SELECT COUNT(hash) FROM image where hash = $1", int64(hash))
    if(err != nil){
        log.Debug("[Checking if photo exists] Couldn't get hash: ", err.Error())
        raven.CaptureError(err, nil)
        return false, err
    }
    defer rows.Close()

    var numOfOccurences int
    if(rows.Next()){
        err = rows.Scan(&numOfOccurences)
        if(err != nil){
            log.Debug("[Checking if photo exists] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return false, err
        }
    }

    if(numOfOccurences > 0){
        return true, nil
    } else{
        return false, nil
    }
}

func validateDonatedPhoto(imageId string, valid bool) error{
	if(valid){
		_,err := db.Exec(`UPDATE image_validation AS v 
						  SET num_of_valid = num_of_valid + 1
						  FROM image AS i
						  WHERE v.image_id = i.id AND key = $1`, imageId)
		if(err != nil){
			log.Debug("[Validating donated photo] Couldn't increase num_of_valid: ", err.Error())
			raven.CaptureError(err, nil)
			return err
		}
	} else{
		_,err := db.Exec(`UPDATE image_validation AS v 
						  SET num_of_invalid = num_of_invalid + 1
						  FROM image AS i
						  WHERE v.image_id = i.id AND key = $1`, imageId)
		if(err != nil){
			log.Debug("[Validating donated photo] Couldn't increase num_of_invalid: ", err.Error())
			raven.CaptureError(err, nil)
			return err
		}
	}

	return nil
}

func validateImages(imageValidationBatch []ImageValidation) error {
    var validEntries []string
    var invalidEntries []string
    for i := range imageValidationBatch {
        if imageValidationBatch[i].Valid == "yes" {
            validEntries = append(validEntries, imageValidationBatch[i].Uuid)
        } else if imageValidationBatch[i].Valid == "no" {
            invalidEntries = append(invalidEntries, imageValidationBatch[i].Uuid)
        }
    }


    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Batch Validating donated photos] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    if len(invalidEntries) > 0 {
        _,err := tx.Exec(`UPDATE image_validation AS v 
                              SET num_of_invalid = num_of_invalid + 1
                              FROM image AS i
                              WHERE v.image_id = i.id AND key = ANY($1)`, pq.Array(invalidEntries))
        if err != nil {
            tx.Rollback()
            log.Debug("[Batch Validating donated photos] Couldn't increase num_of_invalid: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }
    }

    if len(validEntries) > 0 {
        _,err := tx.Exec(`UPDATE image_validation AS v 
                              SET num_of_valid = num_of_valid + 1
                              FROM image AS i
                              WHERE v.image_id = i.id AND key = ANY($1)`, pq.Array(validEntries))
        if err != nil {
            tx.Rollback()
            log.Debug("[Batch Validating donated photos] Couldn't increase num_of_valid: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }
    }

    return tx.Commit()
}

func export(labels []string) ([]Image, error){
    rows, err := db.Query(`SELECT i.key, l.name, CASE WHEN v.num_of_valid + v.num_of_invalid = 0 THEN 0 ELSE (CAST (v.num_of_valid AS float)/(v.num_of_valid + v.num_of_invalid)) END, 
    					   v.num_of_valid, v.num_of_invalid, a.annotations
    					   FROM image_validation v 
                           JOIN image i ON v.image_id = i.id 
                           JOIN label l ON v.label_id = l.id 
                           JOIN image_provider p ON i.image_provider_id = p.id 
                           LEFT JOIN image_annotation a ON a.image_id = i.id
                           WHERE i.unlocked = true and p.name = 'donation' AND l.name = ANY($1)`, pq.Array(labels))
    if err != nil {
        log.Debug("[Export] Couldn't export data: ", err.Error())
        raven.CaptureError(err, nil)
        return nil, err
    }
    defer rows.Close()

    imageEntries := []Image{}
    for rows.Next() {
    	var image Image
        var annotations []byte
    	image.Provider = "donation"

        err = rows.Scan(&image.Id, &image.Label, &image.Probability, &image.NumOfValid, &image.NumOfInvalid, &annotations)
    	if err != nil {
            log.Debug("[Export] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return nil, err
        }

        if len(annotations) > 0 {
            err := json.Unmarshal(annotations, &image.Annotations)
            if err != nil {
                log.Debug("[Export] Couldn't unmarshal: ", err.Error())
                raven.CaptureError(err, nil)
                return nil, err
            }
        }

        imageEntries = append(imageEntries, image)
    }
    return imageEntries, err
}

func explore() []GraphNode{
	graphNodeEntries := []GraphNode{}

    tx, err := db.Begin()
    if err != nil {
    	log.Debug("[Explore] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return graphNodeEntries
    }

    var minSize int32
    var maxSize int32
    maxSize = 0
    minSize = 0
    err = tx.QueryRow(`SELECT MIN(count), MAX(count) FROM 
    						(SELECT COUNT(*) FROM image_validation v 
    						 JOIN label l ON v.label_id = l.id 
    						 GROUP BY l.name) t`).Scan(&minSize, &maxSize)
    if(err != nil){
        log.Debug("[Explore] Couldn't explore min/max: ", err.Error())
        raven.CaptureError(err, nil)
        tx.Rollback()
        return graphNodeEntries
    }

    scaleFactor := float64((float64(maxSize) - float64(minSize))/float64(30))
    if(scaleFactor == 0){
    	scaleFactor = 30
    } else {
    	scaleFactor = 1/scaleFactor
    }

    rows, err := tx.Query(`SELECT l.name, count(l.name) 
    					   FROM image_validation v 
    					   JOIN label l ON v.label_id = l.id 
    					   GROUP BY l.name ORDER BY count(l.name) DESC`)
    if(err != nil){
        log.Debug("[Explore] Couldn't explore data: ", err.Error())
        raven.CaptureError(err, nil)
        tx.Rollback()
        return graphNodeEntries
    }
    defer rows.Close()

    groupNr := 1
    for(rows.Next()){
    	var graphNode GraphNode
    	graphNode.Group = groupNr
    	err = rows.Scan(&graphNode.Text, &graphNode.Size)
    	if(err != nil) {
            log.Debug("[Explore] Couldn't scan data row: ", err.Error())
            raven.CaptureError(err, nil)
            tx.Rollback()
            return graphNodeEntries
        }
        graphNode.Size = int(float64(graphNode.Size) * scaleFactor)
        graphNodeEntries = append(graphNodeEntries, graphNode)
        groupNr += 1
    }

    tx.Commit()

    return graphNodeEntries
}


func getRandomImage() Image{
	var image Image

	image.Id = ""
	image.Label = ""
	image.Provider = "donation"

	rows, err := db.Query(`SELECT i.key, l.name FROM image i 
						   JOIN image_provider p ON i.image_provider_id = p.id 
						   JOIN image_validation v ON v.image_id = i.id
						   JOIN label l ON v.label_id = l.id
						   WHERE ((i.unlocked = true) AND (p.name = 'donation') 
                           AND (v.num_of_valid = 0) AND (v.num_of_invalid = 0)) LIMIT 1`)
	if(err != nil){
		log.Debug("[Fetch random image] Couldn't fetch random image: ", err.Error())
		raven.CaptureError(err, nil)
		return image
	}
    defer rows.Close()
	
	if(!rows.Next()){
        otherRows, err := db.Query(`SELECT i.key, l.name FROM image i 
                           JOIN image_provider p ON i.image_provider_id = p.id 
                           JOIN image_validation v ON v.image_id = i.id
                           JOIN label l ON v.label_id = l.id
                           WHERE i.unlocked = true AND p.name = 'donation' 
                           OFFSET floor(random() * (SELECT count(*) FROM image i JOIN image_provider p ON i.image_provider_id = p.id 
                           WHERE i.unlocked = true AND p.name = 'donation')) LIMIT 1`)
        if(!otherRows.Next()){
    		log.Debug("[Fetch random image] Missing result set")
    		raven.CaptureMessage("[Fetch random image] Missing result set", nil)
    		return image
        }
        defer otherRows.Close()

        err = otherRows.Scan(&image.Id, &image.Label)
        if(err != nil){
            log.Debug("[Fetch random image] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return image
        }
	} else{
        err = rows.Scan(&image.Id, &image.Label)
        if(err != nil){
            log.Debug("[Fetch random image] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return image
        }
    }

	return image
}

func reportImage(imageId string, reason string) error{
	insertedId := 0
	err := db.QueryRow("INSERT INTO image_report(image_id, reason) SELECT i.id, $2 FROM image i WHERE i.key = $1 RETURNING id", 
					  imageId, reason).Scan(&insertedId)
	if(err != nil){
		log.Debug("[Report image] Couldn't add report: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	return nil
}

//returns a list of n - random images (n = limit) that were uploaded with the given label. 
func getRandomGroupedImages(label string, limit int) ([]Image, error) {
    var images []Image

    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Random grouped images] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return images, err
    }

    //get number of images for a given label. we need that to calculate a random number between
    //0 and (numOfImages - limit). If (numOfImages - limit) < 0 then offset = 0.

    //TODO: the following SQL query is a potential candidate for improvement, as it probably gets slow if there
    //are ten thousands of rows in the DB.
    var numOfRows int
    err = tx.QueryRow(`SELECT count(*) FROM image i 
                        JOIN image_provider p ON i.image_provider_id = p.id 
                        JOIN image_validation v ON v.image_id = i.id
                        JOIN label l ON v.label_id = l.id
                        WHERE i.unlocked = true AND p.name = 'donation' AND l.name = $1`, label).Scan(&numOfRows)
    if err != nil {
        tx.Rollback()
        log.Debug("[Random grouped images] Couldn't get num of rows: ", err.Error())
        raven.CaptureError(err, nil)
        return images, err
    }

    randomNumber := 0
    end := numOfRows - limit
    if end < 0 {
        end = 0
    } 

    if end != 0 {
        randomNumber = random(0, end)
    }

    //fetch images
    rows, err := db.Query(`SELECT i.key, l.name FROM image i 
                           JOIN image_provider p ON i.image_provider_id = p.id 
                           JOIN image_validation v ON v.image_id = i.id
                           JOIN label l ON v.label_id = l.id
                           WHERE i.unlocked = true AND p.name = 'donation' AND l.name = $1
                           OFFSET $2 LIMIT $3`, label, randomNumber, limit)

    if err != nil {
        tx.Rollback()
        log.Debug("[Random grouped images] Couldn't get images: ", err.Error())
        raven.CaptureError(err, nil)
        return images, err
    }

    defer rows.Close()

    for rows.Next() {
        var image Image
        image.Provider = "donation"
        err = rows.Scan(&image.Id, &image.Label)
        if err != nil {
            tx.Rollback()
            log.Debug("[Fetch random grouped image] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return images, err
        }

        images = append(images, image)
    }

    return images, tx.Commit()
}

func addAnnotations(imageId string, annotations []Annotation) error{
    //currently there is a uniqueness constraint on the image_id column to ensure that we only have
    //one image annotation per image. That means that the below query can fail with a unique constraint error. 
    //we might want to change that in the future to support multiple annotations per image (if there is a use case for it),
    //but for now it should be fine.
    byt, err := json.Marshal(annotations)
    if(err != nil){
        log.Debug("[Add Annotation] Couldn't create byte array: ", err.Error())
        return err
    }

    insertedId := 0
    err = db.QueryRow("INSERT INTO image_annotation(image_id, annotations, num_of_valid, num_of_invalid) SELECT i.id, $2, $3, $4 FROM image i WHERE i.key = $1 RETURNING id", 
                      imageId, byt, 0, 0).Scan(&insertedId)
    if(err != nil){
        log.Debug("[Add Annotation] Couldn't add annotations: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }
    return nil
}

func getRandomUnannotatedImage() Image{
    var image Image
    //select all images that aren't already annotated and have a label correctness probability of >= 0.8 
    rows, err := db.Query(`SELECT i.key, l.name FROM image i 
                               JOIN image_provider p ON i.image_provider_id = p.id 
                               JOIN image_validation v ON v.image_id = i.id
                               JOIN label l ON v.label_id = l.id
                               WHERE i.unlocked = true AND p.name = 'donation' AND 
                               CASE WHEN v.num_of_valid + v.num_of_invalid = 0 THEN 0 ELSE (CAST (v.num_of_valid AS float)/(v.num_of_valid + v.num_of_invalid)) END >= 0.8
                               AND i.id NOT IN
                               (
                                    SELECT image_id FROM image_annotation 
                               )
                               OFFSET floor
                               ( random() * 
                                   (
                                        SELECT count(*) FROM image i
                                        JOIN image_provider p ON i.image_provider_id = p.id
                                        JOIN image_validation v ON v.image_id = i.id
                                        WHERE i.unlocked = true AND p.name = 'donation' AND 
                                        CASE WHEN v.num_of_valid + v.num_of_invalid = 0 THEN 0 ELSE (CAST (v.num_of_valid AS float)/(v.num_of_valid + v.num_of_invalid)) END >= 0.8
                                        AND i.id NOT IN
                                        (
                                            SELECT image_id FROM image_annotation 
                                        )
                                   ) 
                               )LIMIT 1`)
    if(err != nil) {
        log.Debug("[Get Random Un-annotated Image] Couldn't fetch result: ", err.Error())
        raven.CaptureError(err, nil)
        return image
    }

    defer rows.Close()

    if(rows.Next()){
        image.Provider = "donation"

        err = rows.Scan(&image.Id, &image.Label)
        if(err != nil){
            log.Debug("[Get Random Un-annotated Image] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return image
        }
    }

    return image
}

func getRandomAnnotatedImage() Image{
    var image Image

    rows, err := db.Query(`SELECT i.key, l.name, a.annotations FROM image i 
                               JOIN image_provider p ON i.image_provider_id = p.id 
                               JOIN image_validation v ON v.image_id = i.id
                               JOIN image_annotation a ON a.image_id = v.image_id
                               JOIN label l ON v.label_id = l.id
                               WHERE i.unlocked = true AND p.name = 'donation' 
                               OFFSET floor(random() * 
                               (
                                SELECT count(*) FROM image i 
                                JOIN image_provider p ON i.image_provider_id = p.id 
                                JOIN image_annotation a ON a.image_id = i.id
                                WHERE i.unlocked = true AND p.name = 'donation')
                               ) LIMIT 1`)
    if(err != nil){
        log.Debug("[Get Random Annotated Image] Couldn't get annotated image: ", err.Error())
        raven.CaptureError(err, nil)
        return image
    }

    defer rows.Close()

    if(rows.Next()){
        var annotations []byte
        image.Provider = "donation"

        err = rows.Scan(&image.Id, &image.Label, &annotations)
        if(err != nil) {
            log.Debug("[Get Random Annotated Image] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return image
        }

        err := json.Unmarshal(annotations, &image.Annotations)
        if(err != nil) {
            log.Debug("[Get Random Annotated Image] Couldn't unmarshal: ", err.Error())
            raven.CaptureError(err, nil)
            return image
        }
    }

    return image
}

func validateAnnotatedImage(imageId string, valid bool) error{
    if(valid){
        _,err := db.Exec(`UPDATE image_annotation AS a 
                          SET num_of_valid = num_of_valid + 1
                          FROM image AS i
                          WHERE a.image_id = i.id AND key = $1`, imageId)
        if(err != nil){
            log.Debug("[Validating annotated photo] Couldn't increase num_of_valid: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }
    } else{
        _,err := db.Exec(`UPDATE image_annotation AS a 
                          SET num_of_invalid = num_of_invalid + 1
                          FROM image AS i
                          WHERE a.image_id = i.id AND key = $1`, imageId)
        if(err != nil){
            log.Debug("[Validating annotated photo] Couldn't increase num_of_invalid: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }
    }

    return nil
}

func getNumOfDonatedImages() (int64, error){
    var num int64
    err := db.QueryRow("SELECT count(*) FROM image").Scan(&num)
    if(err != nil){
        log.Debug("[Fetch images] Couldn't get num of available images: ", err.Error())
        raven.CaptureError(err, nil)
        return 0, err
    }

    return num, nil
}

func getNumOfAnnotatedImages() (int64, error){
    var num int64
    err := db.QueryRow("SELECT count(*) FROM image_annotation").Scan(&num)
    if(err != nil){
        log.Debug("[Fetch images] Couldn't get num of annotated images: ", err.Error())
        raven.CaptureError(err, nil)
        return 0, err
    }

    return num, nil
}


func getNumOfValidatedImages() (int64, error){
    var num int64
    err := db.QueryRow("SELECT count(*) FROM image_validation").Scan(&num)
    if(err != nil){
        log.Debug("[Fetch images] Couldn't get num of validated images: ", err.Error())
        raven.CaptureError(err, nil)
        return 0, err
    }

    return num, nil
}

func getAllUnverifiedImages() ([]Image, error){
    var images []Image
    rows, err := db.Query(`SELECT i.key, l.name FROM image i 
                            JOIN image_provider p ON i.image_provider_id = p.id 
                            JOIN image_validation v ON v.image_id = i.id
                            JOIN label l ON v.label_id = l.id
                            WHERE ((i.unlocked = false) AND (p.name = 'donation'))`)

    if(err != nil){
        log.Debug("[Fetch unverified images] Couldn't fetch unverified images: ", err.Error())
        raven.CaptureError(err, nil)
        return images, err
    }

    defer rows.Close()

    for rows.Next() {
        var image Image
        image.Provider = "donation"
        err = rows.Scan(&image.Id, &image.Label)
        if err != nil {
            log.Debug("[Fetch unverified images] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return images, err
        }

        images = append(images, image)
    }

    return images, nil
}

func unlockImage(imageId string) error {
    _,err := db.Exec("UPDATE image SET unlocked = true WHERE key = $1", imageId)
    if err != nil {
        log.Debug("[Unlock Image] Couldn't unlock image: ", err.Error())
        return err
    }

    return nil
}