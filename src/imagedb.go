package main

import (
	"database/sql"
	"github.com/lib/pq"
	"github.com/getsentry/raven-go"
	log "github.com/Sirupsen/logrus"
)

type Image struct {
    Id string `json:"uuid"`
    Label string `json:"label"`
    Provider string `json:"provider"`
    Probability float32 `json:"probability"`
    NumOfValid int32 `json:"num_yes"`
    NumOfInvalid int32 `json:"num_no"`
}

type GraphNode struct {
	Group int `json:"group"`
	Text string `json:"text"`
	Size int `json:"size"`
}

func addDonatedPhoto(filename string, label string) error{
	db, err := sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
	if err != nil {
		log.Debug("[Adding donated photo] Couldn't open database: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	imageId := 0
	err = db.QueryRow("INSERT INTO image(key, unlocked, image_provider_id) SELECT $1, $2, p.id FROM image_provider p WHERE p.name = $3 RETURNING id", 
					  filename, false, "donation").Scan(&imageId)
	if(err != nil){
		log.Debug("[Adding donated photo] Couldn't insert image: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	labelId := 0
	err = db.QueryRow("INSERT INTO image_validation(image_id, num_of_valid, num_of_invalid, label_id) SELECT $1, $2, $3, l.id FROM label l WHERE l.name = $4 RETURNING id", 
					  imageId, 0, 0, label).Scan(&labelId)
	if(err != nil){
		log.Debug("[Adding donated photo] Couldn't insert image validation entry: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	return nil
}

func validateDonatedPhoto(imageId string, valid bool) error{
	db, err := sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
	if err != nil {
		log.Debug("[Validating donated photo] Couldn't open database: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

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
		_,err = db.Exec(`UPDATE image_validation AS v 
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

func export(labels []string) ([]Image, error){
    db, err := sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
    if err != nil {
        log.Debug("[Export] Couldn't open database: ", err.Error())
        raven.CaptureError(err, nil)
        return nil, err
    }

    rows, err := db.Query(`SELECT i.key, l.name, CASE WHEN v.num_of_valid + v.num_of_invalid = 0 THEN 0 ELSE (CAST (v.num_of_valid AS float)/(v.num_of_valid + v.num_of_invalid)) END, 
    					   v.num_of_valid, v.num_of_invalid
    					   FROM image_validation v 
                           JOIN image i ON v.image_id = i.id 
                           JOIN label l ON v.label_id = l.id 
                           JOIN image_provider p ON i.image_provider_id = p.id 
                           WHERE i.unlocked = true and p.name = 'donation' AND l.name = ANY($1)`, pq.Array(labels))
    if(err != nil){
        log.Debug("[Export] Couldn't export data: ", err.Error())
        raven.CaptureError(err, nil)
        return nil, err
    }

    imageEntries := []Image{}
    for(rows.Next()){
    	var image Image
    	image.Provider = "donation"

        err = rows.Scan(&image.Id, &image.Label, &image.Probability, &image.NumOfValid, &image.NumOfInvalid)
    	if(err != nil) {
            log.Debug("[Export] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return nil, err
        }

        imageEntries = append(imageEntries, image)
    }

    return imageEntries, nil
}

func explore() []GraphNode{
	graphNodeEntries := []GraphNode{}
	db, err := sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
    if err != nil {
        log.Debug("[Explore] Couldn't open database: ", err.Error())
        raven.CaptureError(err, nil)
        return graphNodeEntries
    }

    rows, err := db.Query(`SELECT MIN(count), MAX(count) FROM 
    						(SELECT COUNT(*) FROM image_validation v 
    						 JOIN label l ON v.label_id = l.id 
    						 GROUP BY l.name) t`)
    if(err != nil){
        log.Debug("[Explore] Couldn't explore min/max: ", err.Error())
        raven.CaptureError(err, nil)
        return graphNodeEntries
    }

    minSize := 0
    maxSize := 0
    if(rows.Next()){
    	err = rows.Scan(&minSize, &maxSize)
    	if(err != nil){
        	log.Debug("[Explore] Couldn't scan min/max row: ", err.Error())
        	raven.CaptureError(err, nil)
        	return graphNodeEntries
    	}
    }

    scaleFactor := float64((float64(maxSize) - float64(minSize))/float64(30))
    if(scaleFactor == 0){
    	scaleFactor = 30
    } else {
    	scaleFactor = 1/scaleFactor
    }

    rows, err = db.Query(`SELECT l.name, count(l.name) 
    					   FROM image_validation v 
    					   JOIN label l ON v.label_id = l.id 
    					   GROUP BY l.name ORDER BY count(l.name) DESC`)
    if(err != nil){
        log.Debug("[Explore] Couldn't explore data: ", err.Error())
        raven.CaptureError(err, nil)
        return graphNodeEntries
    }

    groupNr := 1
    for(rows.Next()){
    	var graphNode GraphNode
    	graphNode.Group = groupNr
    	err = rows.Scan(&graphNode.Text, &graphNode.Size)
    	if(err != nil) {
            log.Debug("[Explore] Couldn't scan data row: ", err.Error())
            raven.CaptureError(err, nil)
            return graphNodeEntries
        }
        graphNode.Size = int(float64(graphNode.Size) * scaleFactor)
        graphNodeEntries = append(graphNodeEntries, graphNode)
        groupNr += 1
    }
    return graphNodeEntries
}


func getRandomImage() Image{
	var image Image

	image.Id = ""
	image.Label = ""
	image.Provider = "donation"

	//open database
	db, err := sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
	if err != nil {
		log.Debug("[Fetch random image] Couldn't open database: ", err.Error())
		raven.CaptureError(err, nil)
		return image
	}

	rows, err := db.Query(`SELECT i.key, l.name FROM image i 
						   JOIN image_provider p ON i.image_provider_id = p.id 
						   JOIN image_validation v ON v.image_id = i.id
						   JOIN label l ON v.label_id = l.id
						   WHERE i.unlocked = true AND p.name = 'donation' 
						   OFFSET floor(random() * (SELECT count(*) FROM image i JOIN image_provider p ON i.image_provider_id = p.id WHERE i.unlocked = true AND p.name = 'donation')) LIMIT 1`)
	if(err != nil){
		log.Debug("[Fetch random image] Couldn't fetch random image: ", err.Error())
		raven.CaptureError(err, nil)
		return image
	}
	
	if(!rows.Next()){
		log.Debug("[Fetch random image] Missing result set")
		raven.CaptureMessage("[Fetch random image] Missing result set", nil)
		return image
	}

	err = rows.Scan(&image.Id, &image.Label)
	if(err != nil){
		log.Debug("[Fetch random image] Couldn't scan row: ", err.Error())
		raven.CaptureError(err, nil)
		return image
	}

	return image
}

func reportImage(imageId string, reason string) error{
	db, err := sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
	if err != nil {
		log.Debug("[Report image] Couldn't open database: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	insertedId := 0
	err = db.QueryRow("INSERT INTO image_report(image_id, reason) SELECT i.id, $2 FROM image i WHERE i.key = $1 RETURNING id", 
					  imageId, reason).Scan(&insertedId)
	if(err != nil){
		log.Debug("[Report image] Couldn't add report: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	return nil
}