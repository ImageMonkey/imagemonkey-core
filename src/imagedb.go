package main

import (
    "github.com/lib/pq"
	"github.com/getsentry/raven-go"
	log "github.com/Sirupsen/logrus"
    "encoding/json"
    "database/sql"
    //"errors"
    //"database/sql/driver"
)

type Annotation struct{
    Left float32 `json:"left"`
    Top float32 `json:"top"`
    Width float32 `json:"width"`
    Height float32 `json:"height"`
}

type Annotations struct{
    Annotations []Annotation `json:"annotations"`
    Label string `json:"label"`
    Sublabel string `json:"sublabel"`
}

type Image struct {
    Id string `json:"uuid"`
    Label string `json:"label"`
    Sublabel string `json:"sublabel"`
    Provider string `json:"provider"`
    Probability float32 `json:"probability"`
    NumOfValid int32 `json:"num_yes"`
    NumOfInvalid int32 `json:"num_no"`
    Annotations []Annotation `json:"annotations"`
    AllLabels []LabelMeEntry `json:"all_labels"`
}

type ImageValidation struct {
    Uuid string `json:"uuid"`
    Valid string `json:"valid"`
}

type ImageValidationBatch struct {
    Validations []ImageValidation `json:"validations"`
    Label string `json:"label"`
    Sublabel string `json:"sublabel"`
}

type GraphNode struct {
	Group int `json:"group"`
	Text string `json:"text"`
	Size int `json:"size"`
}

type ValidationStat struct {
    Label string `json:"label"`
    Count int `json:"count"`
    ErrorRate float32 `json:"error_rate"`
    TotalValidations int `json:"total_validations"`
}

type DonationsPerCountryStat struct {
    CountryCode string `json:"country_code"`
    Count int64 `json:"num"`
}

type ValidationsPerCountryStat struct {
    CountryCode string `json:"country_code"`
    Count int64 `json:"num"`
}

type AnnotationsPerCountryStat struct {
    CountryCode string `json:"country_code"`
    Count int64 `json:"num"`
}

type Statistics struct {
    Validations []ValidationStat `json:"validations"`
    DonationsPerCountry []DonationsPerCountryStat `json:"donations_per_country"`
    ValidationsPerCountry []ValidationsPerCountryStat `json:"validations_per_country"`
    AnnotationsPerCountry []AnnotationsPerCountryStat `json:"annotations_per_country"`
    NumOfUnlabeledDonations int64 `json:"num_of_unlabeled_donations"`
}

type LabelSearchItem struct {
    Label string `json:"label"`
    ParentLabel string `json:"parent_label"`
}

type LabelSearchResult struct {
    Labels []LabelSearchItem `json:"items"`
}

func addDonatedPhoto(clientFingerprint string, filename string, hash uint64, labels []LabelMeEntry ) error{
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

    if labels[0].Label != "" { //only create a image validation entry, if a label is provided
        err = _addLabelsToImage(clientFingerprint, filename, labels, tx)
        if err != nil {
            return err //tx already rolled back in case of error, so we can just return here
        }


    	/*labelId := 0
    	err = tx.QueryRow(`INSERT INTO image_validation(image_id, num_of_valid, num_of_invalid, fingerprint_of_last_modification, label_id) 
                            SELECT $1, $2, $3, $5, l.id FROM label l WHERE l.name = $4 RETURNING id`, 
    					  imageId, 0, 0, label, clientFingerprint).Scan(&labelId)
    	if(err != nil){
    		tx.Rollback()
    		log.Debug("[Adding donated photo] Couldn't insert image validation entry: ", err.Error())
    		raven.CaptureError(err, nil)
    		return err
    	}


        if addSublabels {
            TODO: add sublabels (image_validionn_entry)
        }*/

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

func validateDonatedPhoto(clientFingerprint string, imageId string, labelValidationEntry LabelValidationEntry, valid bool) error {
	if valid {
        log.Debug("HERE " +labelValidationEntry.Label)
        var err error
        if labelValidationEntry.Sublabel == "" {
    		_, err = db.Exec(`UPDATE image_validation AS v 
    						   SET num_of_valid = num_of_valid + 1, fingerprint_of_last_modification = $1
    						   FROM image AS i 
    						   WHERE v.image_id = i.id AND key = $2 AND v.label_id = (SELECT id FROM label WHERE name = $3 AND parent_id is null)`, 
                               clientFingerprint, imageId, labelValidationEntry.Label)
        } else {
            _, err = db.Exec(`UPDATE image_validation AS v 
                              SET num_of_valid = num_of_valid + 1, fingerprint_of_last_modification = $1
                              FROM image AS i 
                              WHERE v.image_id = i.id AND key = $2 AND v.label_id = (
                                SELECT l.id FROM label l 
                                JOIN label pl ON l.parent_id = pl.id
                                WHERE l.name = $3 AND pl.name = $4
                              )`, 
                              clientFingerprint, imageId, labelValidationEntry.Sublabel, labelValidationEntry.Label)
        }

		if err != nil {
			log.Debug("[Validating donated photo] Couldn't increase num_of_valid: ", err.Error())
			raven.CaptureError(err, nil)
			return err
		}
	} else {
        var err error
        if labelValidationEntry.Sublabel == "" {
    		_, err = db.Exec(`UPDATE image_validation AS v 
    						   SET num_of_invalid = num_of_invalid + 1, fingerprint_of_last_modification = $1
    						   FROM image AS i
    						   WHERE v.image_id = i.id AND key = $2 AND v.label_id = (SELECT id FROM label WHERE name = $3 AND parent_id is null)`, 
                               clientFingerprint, imageId, labelValidationEntry.Label)
        } else {
            _, err = db.Exec(`UPDATE image_validation AS v 
                               SET num_of_invalid = num_of_invalid + 1, fingerprint_of_last_modification = $1
                               FROM image AS i
                               WHERE v.image_id = i.id AND key = $2 AND v.label_id = (
                                SELECT l.id FROM label l 
                                JOIN label pl ON l.parent_id = pl.id
                                WHERE l.name = $3 AND pl.name = $4
                               )`, 
                               clientFingerprint, imageId, labelValidationEntry.Sublabel, labelValidationEntry.Label)
        }

		if err != nil {
			log.Debug("[Validating donated photo] Couldn't increase num_of_invalid: ", err.Error())
			raven.CaptureError(err, nil)
			return err
		}
	}

	return nil
}

func validateImages(clientFingerprint string, imageValidationBatch ImageValidationBatch) error {
    var validEntries []string
    var invalidEntries []string

    validations := imageValidationBatch.Validations

    for i := range validations {
        if validations[i].Valid == "yes" {
            validEntries = append(validEntries, validations[i].Uuid)
        } else if validations[i].Valid == "no" {
            invalidEntries = append(invalidEntries, validations[i].Uuid)
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
                              SET num_of_invalid = num_of_invalid + 1, fingerprint_of_last_modification = $1
                              FROM image AS i
                              WHERE v.image_id = i.id AND key = ANY($2) AND v.label_id = (SELECT id FROM label WHERE name = $3)`, 
                              clientFingerprint, pq.Array(invalidEntries),imageValidationBatch.Label)
        if err != nil {
            tx.Rollback()
            log.Debug("[Batch Validating donated photos] Couldn't increase num_of_invalid: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }
    }

    if len(validEntries) > 0 {
        _,err := tx.Exec(`UPDATE image_validation AS v 
                              SET num_of_valid = num_of_valid + 1, fingerprint_of_last_modification = $1
                              FROM image AS i
                              WHERE v.image_id = i.id AND key = ANY($2) AND v.label_id = (SELECT id FROM label WHERE name = $3)`, 
                              clientFingerprint, pq.Array(validEntries), imageValidationBatch.Label)
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
                           LEFT JOIN image_annotation a ON a.image_id = i.id AND a.label_id = l.id
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

func explore(words []string) (Statistics, error) {
    statistics := Statistics{}

    //use temporary map for faster lookup
    temp := make(map[string]ValidationStat)

    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Explore] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }
    
    rows, err := tx.Query(`SELECT l.name, count(l.name), 
                           CASE WHEN SUM(v.num_of_valid + v.num_of_invalid) = 0 THEN 0 ELSE (CAST (SUM(v.num_of_invalid) AS float)/(SUM(v.num_of_valid) + SUM(v.num_of_invalid))) END as error_rate, 
                           SUM(v.num_of_valid + v.num_of_invalid) as total_validations
                           FROM image_validation v 
                           JOIN label l ON v.label_id = l.id 
                           GROUP BY l.name ORDER BY count(l.name) DESC`)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore data: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }
    defer rows.Close()

    for rows.Next() {
        var validationStat ValidationStat
        err = rows.Scan(&validationStat.Label, &validationStat.Count, &validationStat.ErrorRate, &validationStat.TotalValidations)
        if err != nil {
            tx.Rollback()
            log.Debug("[Explore] Couldn't scan data row: ", err.Error())
            raven.CaptureError(err, nil)
            return statistics, err
        }

        temp[validationStat.Label] = validationStat
    }

    //add labels where we don't have a donation yet
    for _, value := range words {
        _, contains := temp[value]
        if !contains {
            var validationStat ValidationStat
            validationStat.Label = value
            validationStat.Count = 0
            temp[value] = validationStat
        }
    }

    for _, value := range temp {
        statistics.Validations = append(statistics.Validations, value)
    }

    //get donations grouped by country
    donationsPerCountryRows, err := tx.Query(`SELECT country_code, count FROM donations_per_country ORDER BY count DESC`)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore data: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }
    defer donationsPerCountryRows.Close()

    for donationsPerCountryRows.Next() {
        var donationsPerCountryStat DonationsPerCountryStat
        err = donationsPerCountryRows.Scan(&donationsPerCountryStat.CountryCode, &donationsPerCountryStat.Count)
        if err != nil {
            tx.Rollback()
            log.Debug("[Explore] Couldn't scan data row: ", err.Error())
            raven.CaptureError(err, nil)
            return statistics, err
        }

        statistics.DonationsPerCountry = append(statistics.DonationsPerCountry, donationsPerCountryStat)
    }


    //get validations grouped by country
    validationsPerCountryRows, err := tx.Query(`SELECT country_code, count FROM validations_per_country ORDER BY count DESC`)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore data: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }
    defer validationsPerCountryRows.Close()

    for validationsPerCountryRows.Next() {
        var validationsPerCountryStat ValidationsPerCountryStat
        err = validationsPerCountryRows.Scan(&validationsPerCountryStat.CountryCode, &validationsPerCountryStat.Count)
        if err != nil {
            tx.Rollback()
            log.Debug("[Explore] Couldn't scan data row: ", err.Error())
            raven.CaptureError(err, nil)
            return statistics, err
        }

        statistics.ValidationsPerCountry = append(statistics.ValidationsPerCountry, validationsPerCountryStat)
    }

    //get annotations grouped by country
    annotationsPerCountryRows, err := tx.Query(`SELECT country_code, count FROM annotations_per_country ORDER BY count DESC`)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore data: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }
    defer annotationsPerCountryRows.Close()

    for annotationsPerCountryRows.Next() {
        var annotationsPerCountryStat AnnotationsPerCountryStat
        err = annotationsPerCountryRows.Scan(&annotationsPerCountryStat.CountryCode, &annotationsPerCountryStat.Count)
        if err != nil {
            tx.Rollback()
            log.Debug("[Explore] Couldn't scan data row: ", err.Error())
            raven.CaptureError(err, nil)
            return statistics, err
        }

        statistics.AnnotationsPerCountry = append(statistics.AnnotationsPerCountry, annotationsPerCountryStat)
    }

    //get all unlabeled donations
    err = tx.QueryRow(`SELECT count(i.id) from image i WHERE i.id NOT IN (SELECT image_id FROM image_validation)`).Scan(&statistics.NumOfUnlabeledDonations)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't scan data row: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }



    return statistics, tx.Commit()
}


func getRandomImage() Image{
	var image Image

	image.Id = ""
	image.Label = ""
	image.Provider = "donation"

	rows, err := db.Query(`SELECT i.key, l.name, COALESCE(pl.name, '') FROM image i 
						   JOIN image_provider p ON i.image_provider_id = p.id 
						   JOIN image_validation v ON v.image_id = i.id
						   JOIN label l ON v.label_id = l.id
                           LEFT JOIN label pl ON l.parent_id = pl.id
						   WHERE ((i.unlocked = true) AND (p.name = 'donation') 
                           AND (v.num_of_valid = 0) AND (v.num_of_invalid = 0)) LIMIT 1`)
	if(err != nil){
		log.Debug("[Fetch random image] Couldn't fetch random image: ", err.Error())
		raven.CaptureError(err, nil)
		return image
	}
    defer rows.Close()
	
    var label1 string
    var label2 string
	if(!rows.Next()){
        otherRows, err := db.Query(`SELECT i.key, l.name, COALESCE(pl.name, '') FROM image i 
                                    JOIN image_provider p ON i.image_provider_id = p.id 
                                    JOIN image_validation v ON v.image_id = i.id
                                    JOIN label l ON v.label_id = l.id
                                    LEFT JOIN label pl ON l.parent_id = pl.id
                                    WHERE i.unlocked = true AND p.name = 'donation' 
                                    OFFSET floor(random() * 
                                        ( SELECT count(*) FROM image i 
                                          JOIN image_provider p ON i.image_provider_id = p.id 
                                          JOIN image_validation v ON v.image_id = i.id 
                                          WHERE i.unlocked = true AND p.name = 'donation'
                                        )
                                    ) LIMIT 1`)
        if(!otherRows.Next()){
    		log.Debug("[Fetch random image] Missing result set")
    		raven.CaptureMessage("[Fetch random image] Missing result set", nil)
    		return image
        }
        defer otherRows.Close()

        err = otherRows.Scan(&image.Id, &label1, &label2)
        if(err != nil){
            log.Debug("[Fetch random image] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return image
        }
	} else{
        err = rows.Scan(&image.Id, &label1, &label2)
        if(err != nil){
            log.Debug("[Fetch random image] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return image
        }
    }

    if label2 == "" {
        image.Label = label1
        image.Sublabel = ""
    } else {
        image.Label = label2
        image.Sublabel = label1
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
                        WHERE i.unlocked = true AND p.name = 'donation' AND l.name = $1 AND l.parent_id is null`, label).Scan(&numOfRows)
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
                           WHERE i.unlocked = true AND p.name = 'donation' AND l.name = $1 AND l.parent_id is null
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

func addAnnotations(clientFingerprint string, imageId string, annotations Annotations) error{
    //currently there is a uniqueness constraint on the image_id column to ensure that we only have
    //one image annotation per image. That means that the below query can fail with a unique constraint error. 
    //we might want to change that in the future to support multiple annotations per image (if there is a use case for it),
    //but for now it should be fine.
    byt, err := json.Marshal(annotations.Annotations)
    if err != nil {
        log.Debug("[Add Annotation] Couldn't create byte array: ", err.Error())
        return err
    }

    insertedId := 0

    if annotations.Sublabel == "" {
        err = db.QueryRow(`INSERT INTO image_annotation(label_id, annotations, num_of_valid, num_of_invalid, fingerprint_of_last_modification, image_id) 
                            SELECT (SELECT l.id FROM label l WHERE l.name = $6 AND l.parent_id is null), $2, $3, $4, $5, (SELECT i.id FROM image i WHERE i.key = $1) RETURNING id`, 
                          imageId, byt, 0, 0, clientFingerprint, annotations.Label).Scan(&insertedId)
    } else {
        err = db.QueryRow(`INSERT INTO image_annotation(label_id, annotations, num_of_valid, num_of_invalid, fingerprint_of_last_modification, image_id) 
                            SELECT (SELECT l.id FROM label l JOIN label pl ON l.parent_id = pl.id WHERE l.name = $6 AND pl.name = $7), $2, $3, $4, $5, (SELECT i.id FROM image i WHERE i.key = $1) RETURNING id`, 
                          imageId, byt, 0, 0, clientFingerprint, annotations.Sublabel, annotations.Label).Scan(&insertedId)
    }


    if err != nil {
        log.Debug("[Add Annotation] Couldn't add annotations: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }
    return nil
}

func getRandomUnannotatedImage() Image{
    var image Image
    //select all images that aren't already annotated and have a label correctness probability of >= 0.8 
    rows, err := db.Query(`SELECT i.key, l.name, COALESCE(pl.name, '') FROM image i 
                               JOIN image_provider p ON i.image_provider_id = p.id 
                               JOIN image_validation v ON v.image_id = i.id
                               JOIN label l ON v.label_id = l.id
                               LEFT JOIN label pl ON l.parent_id = pl.id
                               WHERE i.unlocked = true AND p.name = 'donation' AND 
                               CASE WHEN v.num_of_valid + v.num_of_invalid = 0 THEN 0 ELSE (CAST (v.num_of_valid AS float)/(v.num_of_valid + v.num_of_invalid)) END >= 0.8
                               AND NOT EXISTS
                                (
                                    SELECT 1 FROM image_annotation a WHERE a.label_id = v.label_id AND a.image_id = v.image_id
                                )
                               OFFSET floor
                               ( random() * 
                                   (
                                        SELECT count(*) FROM image i
                                        JOIN image_provider p ON i.image_provider_id = p.id
                                        JOIN image_validation v ON v.image_id = i.id
                                        WHERE i.unlocked = true AND p.name = 'donation' AND 
                                        CASE WHEN v.num_of_valid + v.num_of_invalid = 0 THEN 0 ELSE (CAST (v.num_of_valid AS float)/(v.num_of_valid + v.num_of_invalid)) END >= 0.8
                                        AND NOT EXISTS
                                        (
                                            SELECT 1 FROM image_annotation a WHERE a.label_id = v.label_id AND a.image_id = v.image_id
                                        )
                                   ) 
                               )LIMIT 1`)
    if(err != nil) {
        log.Debug("[Get Random Un-annotated Image] Couldn't fetch result: ", err.Error())
        raven.CaptureError(err, nil)
        return image
    }

    defer rows.Close()

    var label1 string
    var label2 string
    if(rows.Next()){
        image.Provider = "donation"

        err = rows.Scan(&image.Id, &label1, &label2)
        if(err != nil){
            log.Debug("[Get Random Un-annotated Image] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return image
        }

        if label2 == "" {
            image.Label = label1
            image.Sublabel = ""
        } else {
            image.Label = label2
            image.Sublabel = label1
        }
    }

    return image
}

func getRandomAnnotatedImage() Image{
    var image Image

    rows, err := db.Query(`SELECT i.key, l.name, COALESCE(pl.name, ''), a.annotations FROM image i 
                               JOIN image_provider p ON i.image_provider_id = p.id 
                               JOIN image_annotation a ON a.image_id = i.id
                               JOIN label l ON a.label_id = l.id
                               LEFT JOIN label pl ON l.parent_id = pl.id
                               WHERE i.unlocked = true AND p.name = 'donation' 
                               OFFSET floor(random() * 
                               (
                                SELECT count(*) FROM image i 
                                JOIN image_provider p ON i.image_provider_id = p.id 
                                JOIN image_annotation a ON a.image_id = i.id
                                JOIN label l ON a.label_id = l.id
                                WHERE i.unlocked = true AND p.name = 'donation')
                               ) LIMIT 1`)
    if(err != nil){
        log.Debug("[Get Random Annotated Image] Couldn't get annotated image: ", err.Error())
        raven.CaptureError(err, nil)
        return image
    }

    defer rows.Close()

    var label1 string
    var label2 string
    if(rows.Next()){
        var annotations []byte
        image.Provider = "donation"

        err = rows.Scan(&image.Id, &label1, &label2, &annotations)
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

        if label2 == "" {
            image.Label = label1
            image.Sublabel = ""
        } else {
            image.Label = label2
            image.Sublabel = label1
        }
    }

    return image
}

func validateAnnotatedImage(clientFingerprint string, imageId string, labelValidationEntry LabelValidationEntry, valid bool) error {
    if valid {
        var err error
        if labelValidationEntry.Sublabel == "" {
            _, err = db.Exec(`UPDATE image_annotation AS a 
                              SET num_of_valid = num_of_valid + 1, fingerprint_of_last_modification = $1
                              FROM image AS i
                              WHERE a.image_id = i.id AND key = $2 AND a.label_id = (SELECT id FROM label WHERE name = $3 AND parent_id is null)`, 
                              clientFingerprint, imageId, labelValidationEntry.Label)
        } else {
            _, err = db.Exec(`UPDATE image_annotation AS a 
                              SET num_of_valid = num_of_valid + 1, fingerprint_of_last_modification = $1
                              FROM image AS i
                              WHERE a.image_id = i.id AND key = $2 AND a.label_id = (
                                SELECT l.id FROM label l 
                                JOIN label pl ON l.parent_id = pl.id
                                WHERE l.name = $3 AND pl.name = $4
                              )`, 
                              clientFingerprint, imageId, labelValidationEntry.Sublabel, labelValidationEntry.Label)
        }


        if err != nil {
            log.Debug("[Validating annotated photo] Couldn't increase num_of_valid: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }
    } else {
        var err error
        if labelValidationEntry.Sublabel == "" {
            _,err = db.Exec(`UPDATE image_annotation AS a 
                              SET num_of_invalid = num_of_invalid + 1, fingerprint_of_last_modification = $1
                              FROM image AS i
                              WHERE a.image_id = i.id AND key = $2 AND a.label_id = (
                                SELECT id FROM label WHERE name = $3 AND parent_id is null
                              )`, 
                              clientFingerprint, imageId, labelValidationEntry.Label)
        } else {
            _,err = db.Exec(`UPDATE image_annotation AS a 
                              SET num_of_invalid = num_of_invalid + 1, fingerprint_of_last_modification = $1
                              FROM image AS i
                              WHERE a.image_id = i.id AND key = $2 AND a.label_id = (
                                SELECT l.id FROM label l 
                                JOIN label pl ON l.parent_id = pl.id
                                WHERE l.name = $3 AND pl.name = $4
                              )`, 
                              clientFingerprint, imageId, labelValidationEntry.Sublabel, labelValidationEntry.Label)
        }


        if err != nil {
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

func updateContributionsPerCountry(contributionType string, countryCode string) error {
    if contributionType == "donation" {
        _, err := db.Exec(`INSERT INTO donations_per_country (country_code, count)
                            VALUES ($1, $2) ON CONFLICT (country_code)
                            DO UPDATE SET count = donations_per_country.count + 1`, countryCode, 1)
        if err != nil {
            log.Debug("[Update Contributions per Country] Couldn't insert into/update donations_per_country: ", err.Error())
            return err
        }
    } else if contributionType == "validation" {
        _, err := db.Exec(`INSERT INTO validations_per_country (country_code, count)
                            VALUES ($1, $2) ON CONFLICT (country_code)
                            DO UPDATE SET count = validations_per_country.count + 1`, countryCode, 1)
        if err != nil {
            log.Debug("[Update Contributions per Country] Couldn't insert into/update validations_per_country: ", err.Error())
            return err
        }
    } else if contributionType == "annotation" {
        _, err := db.Exec(`INSERT INTO annotations_per_country (country_code, count)
                            VALUES ($1, $2) ON CONFLICT (country_code)
                            DO UPDATE SET count = annotations_per_country.count + 1`, countryCode, 1)
        if err != nil {
            log.Debug("[Update Contributions per Country] Couldn't insert into/update annotations_per_country: ", err.Error())
            return err
        }
    }

    return nil
}

func getImageToLabel() (Image, error) {
    var image Image
    var labelMeEntries []LabelMeEntry
    image.Provider = "donation"

    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Get Image to Label] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return image, err
    }

    unlabeledRows, err := tx.Query(`SELECT i.key from image i WHERE i.unlocked = true AND i.id NOT IN (SELECT image_id FROM image_validation) LIMIT 1`)
    if err != nil {
        tx.Rollback()
        raven.CaptureError(err, nil)
        log.Debug("[Get Image to Label] Couldn't get unlabeled image: ", err.Error())
        return image, err
    }

    defer unlabeledRows.Close()

    if !unlabeledRows.Next() {
        rows, err := tx.Query(`SELECT q.key, l.name as label, COALESCE(pl.name, '') as parentLabel
                               FROM image_validation v 
                               JOIN (SELECT i.id as id, i.key as key FROM image i OFFSET floor(random() * (SELECT count(*) FROM image i WHERE unlocked = true)) LIMIT 1) q ON q.id = v.image_id
                               JOIN label l on v.label_id = l.id 
                               LEFT JOIN label pl on l.parent_id = pl.id`)


        if err != nil {
            tx.Rollback()
            raven.CaptureError(err, nil)
            log.Debug("[Get Image to Label] Couldn't get image: ", err.Error())
            return image, err
        }

        defer rows.Close()

        //store in temporary map for faster access
        var label string
        var parentLabel string
        var baseLabel string
        temp := make(map[string]LabelMeEntry) 
        for rows.Next() {
            err = rows.Scan(&image.Id, &label, &parentLabel)
            if err != nil {
                tx.Rollback()
                raven.CaptureError(err, nil)
                log.Debug("[Get Image to Label] Couldn't scan labeled row: ", err.Error())
                return image, err
            }

            baseLabel = parentLabel
            if parentLabel == "" {
                baseLabel = label
            }

            if val, ok := temp[baseLabel]; ok {
                if parentLabel != "" {
                    val.Sublabels = append(val.Sublabels, label)
                }
                temp[baseLabel] = val
            } else {
                var labelMeEntry LabelMeEntry
                labelMeEntry.Label = baseLabel
                if parentLabel != "" {
                    labelMeEntry.Sublabels = append(labelMeEntry.Sublabels, label)
                }
                temp[baseLabel] = labelMeEntry 
            }
        }

        rows.Close()

        //map -> list
        for  _, value := range temp {
            labelMeEntries = append(labelMeEntries, value)
        }

    } else {
        err = unlabeledRows.Scan(&image.Id)
        if err != nil {
            tx.Rollback()
            raven.CaptureError(err, nil)
            log.Debug("[Get Image to Label] Couldn't scan row: ", err.Error())
            return image, err
        }
        unlabeledRows.Close()
    }

    image.AllLabels = labelMeEntries

    err = tx.Commit()
    if err != nil {
        raven.CaptureError(err, nil)
        log.Debug("[Get Image to Label] Couldn't commit changes: ", err.Error())
        return image, err
    }


    return image, nil
}

func addLabelsToImage(clientFingerprint string, imageId string, labels []LabelMeEntry) error {
    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Adding image labels] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    err = _addLabelsToImage(clientFingerprint, imageId, labels, tx)
    if err != nil { 
        return err //tx already rolled back in case of error, so we can just return here 
    }

    err = tx.Commit()
    if err != nil {
        log.Debug("[Adding image labels] Couldn't commit changes: ", err.Error())
        raven.CaptureError(err, nil)
        return err 
    }
    return err
}


func _addLabelsToImage(clientFingerprint string, imageId string, labels []LabelMeEntry, tx *sql.Tx) error {
    for _, item := range labels {
        rows, err := tx.Query(`SELECT i.id FROM image i WHERE i.key = $1`, imageId)
        if err != nil {
            tx.Rollback()
            log.Debug("[Adding image labels] Couldn't get image ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }

        defer rows.Close()

        var imageId int64
        if rows.Next() {
            err = rows.Scan(&imageId)
            if err != nil {
                tx.Rollback()
                log.Debug("[Adding image labels] Couldn't scan image image entry: ", err.Error())
                raven.CaptureError(err, nil)
                return err
            }
        }

        rows.Close()

        //add sublabels
        if len(item.Sublabels) > 0 {
            _, err = tx.Exec(`INSERT INTO image_validation(image_id, num_of_valid, num_of_invalid, fingerprint_of_last_modification, label_id) 
                            SELECT $1, $2, $3, $4, l.id FROM label l LEFT JOIN label cl ON cl.id = l.parent_id WHERE (cl.name = $5 AND l.name = ANY($6))`,
                            imageId, 0, 0, clientFingerprint, item.Label, pq.Array(item.Sublabels))
            if err != nil {
                tx.Rollback()
                log.Debug("[Adding image labels] Couldn't insert image validation entries for sublabels: ", err.Error())
                raven.CaptureError(err, nil)
                return err
            }
        }

        //add base label
        _, err = tx.Exec(`INSERT INTO image_validation(image_id, num_of_valid, num_of_invalid, fingerprint_of_last_modification, label_id) 
                            SELECT $1, $2, $3, $4, id from label l WHERE id NOT IN (
                                SELECT label_id from image_validation v where image_id = $1
                            ) AND l.name = $5 AND l.parent_id IS NULL`,
                            imageId, 0, 0, clientFingerprint, item.Label)
        if err != nil {
            pqErr := err.(*pq.Error)
            if pqErr.Code.Name() != "unique_violation" {
                tx.Rollback()
                log.Debug("[Adding image labels] Couldn't insert image validation entry for label: ", err.Error())
                raven.CaptureError(err, nil)
                return err
            }
        }
    }

    return nil
}

func getAllImageLabels() ([]string, error) {
    var labels []string

    rows, err := db.Query(`SELECT l.name FROM label l`)
    if err != nil {
        log.Debug("[Getting all image labels] Couldn't get image labels: ", err.Error())
        raven.CaptureError(err, nil)
        return labels, err
    }

    defer rows.Close()

    for rows.Next() {
        var label string
        err = rows.Scan(&label)
        if err != nil {
            log.Debug("[Getting all image labels] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return labels, err 
        }

        labels = append(labels, label)
    }

    return labels, nil
}

/*func getLabelSuggestions(labelsStr string) ([]string, error) {
    var labelSuggestions []string
    similarity := 0.3

    rows, err := db.Query(`SELECT q.labels FROM
                            (
                                SELECT string_agg(l.name::text, ',') AS labels FROM image_validation v JOIN label l ON v.label_id = l.id GROUP BY image_id
                            ) q 
                           WHERE similarity(q.labels, $1) > $2`, labelsStr, similarity)
    if err != nil {
        log.Debug("[Fetching label suggestions] Couldn't get label suggestions: ", err.Error())
        raven.CaptureError(err, nil)
        return labelSuggestions, err 
    }

    defer rows.Close()

    var labelSuggestionStr string
    temp := make(map[string]bool)
    for rows.Next() {
        err = rows.Scan(&labelSuggestionStr)
        if err != nil {
            log.Debug("[Fetching label suggestions] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return labelSuggestions, err 
        }

        labels := strings.Split(labelSuggestionStr, ",")
        for _, label := range labels {
            temp[label] = true; //store labels in map for faster lookup (and to make sure that we don't get any duplicates)
        }
    }

    //map -> string
    for key, _ := range temp { 
        labelSuggestions = append(labelSuggestions, key)
    }

    return labelSuggestions, nil
}*/

func getMostPopularLabels(limit int32) ([]string, error) {
    var labels []string

    rows, err := db.Query(`SELECT l.name FROM image_validation v 
                            JOIN label l ON v.label_id = l.id 
                            WHERE l.parent_id is NULL
                            GROUP BY l.id
                            ORDER BY count(l.id) DESC LIMIT $1`, limit)
    if err != nil {
        log.Debug("[Most Popular Labels] Couldn't fetch results: ", err.Error())
        raven.CaptureError(err, nil)
        return labels, err
    }

    defer rows.Close()

    for rows.Next() {
        var label string
        err = rows.Scan(&label)
        if err != nil {
           log.Debug("[Most Popular Labels] Couldn't scan row: ", err.Error())
           raven.CaptureError(err, nil)
           return labels, err 
        }

        labels = append(labels, label)
    }

    return labels, nil
}

/*func autocompleteLabel(text string) (LabelSearchResult, error) {
    var labelSearchResult LabelSearchResult

    rows, err := db.Query(`SELECT l.name as label, COALESCE(pl.name, '') as parent_label
                           FROM label l
                           LEFT JOIN label pl ON pl.id = l.parent_id
                           WHERE l.name % $1
                           ORDER BY similarity(l.name, $1) DESC`, text)

    if err != nil {
        log.Debug("[Autocomplete Label] Couldn't fetch results: ", err.Error())
        raven.CaptureError(err, nil)
        return labelSearchResult, err
    }

    defer rows.Close()

    for rows.Next() {
        var labelSearchItem LabelSearchItem
        err = rows.Scan(&labelSearchItem.Label, &labelSearchItem.ParentLabel)
        if err != nil {
            log.Debug("[Autocomplete Label] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return labelSearchResult, err
        }
        labelSearchResult.Labels = append(labelSearchResult.Labels, labelSearchItem)
    }
    return labelSearchResult, nil*/

    /*var labelSearchResult LabelSearchResult

    rows, err := db.Query(`SELECT COALESCE(l.name, '') as parent_label, q.name as label from label l 
                           RIGHT JOIN ( 
                                SELECT name, parent_id
                                FROM label l
                                WHERE name % $1
                                ORDER BY similarity(name, $1) DESC
                           ) q
                           ON q.parent_id = l.id`, text)

    if err != nil {
        log.Debug("[Autocomplete Label] Couldn't fetch results: ", err.Error())
        raven.CaptureError(err, nil)
        return labelSearchResult, err
    }

    defer rows.Close()

    for rows.Next() {
        var labelSearchItem LabelSearchItem
        err = rows.Scan(&labelSearchItem.ParentLabel, &labelSearchItem.Label)
        if err != nil {
            log.Debug("[Autocomplete Label] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return labelSearchResult, err
        }
        labelSearchResult.Labels = append(labelSearchResult.Labels, labelSearchItem)
    }
    return labelSearchResult, nil*/

    /*rows, err := db.Query(`SELECT name
                           FROM label l
                           WHERE name % $1
                           ORDER BY similarity(name, $1) DESC`, text)
    if err != nil {
        log.Debug("[Autocomplete Label] Couldn't fetch results: ", err.Error())
        raven.CaptureError(err, nil)
        return labelSearchResult, err
    }

    defer rows.Close()

    for rows.Next() {
        var labelSearchItem LabelSearchItem
        err = rows.Scan(&labelSearchItem.Label)
        if err != nil {
            log.Debug("[Autocomplete Label] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return labelSearchResult, err
        }
        labelSearchResult.Labels = append(labelSearchResult.Labels, labelSearchItem)
    }
    return labelSearchResult, nil*/
//}