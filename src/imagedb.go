package main

import (
    "github.com/lib/pq"
	"github.com/getsentry/raven-go"
	log "github.com/Sirupsen/logrus"
    "encoding/json"
    "database/sql"
    "fmt"
)

type RectangleAnnotation struct {
    Id int64 `json:"id"`
    Left float32 `json:"left"`
    Top float32 `json:"top"`
    Width float32 `json:"width"`
    Height float32 `json:"height"`
    Angle float32 `json:"angle"`
}

type EllipsisAnnotation struct {
    Id int64 `json:"id"`
    Left float32 `json:"left"`
    Top float32 `json:"top"`
    Rx float32 `json:"rx"`
    Ry float32 `json:"ry"`
    Angle float32 `json:"angle"`
}


type PolygonPoint struct {
    X float32 `json:"x"`
    Y float32 `json:"y"`
}

type PolygonAnnotation struct {
    Id int64 `json:"id"`
    Left float32 `json:"left"`
    Top float32 `json:"top"`
    Points []PolygonPoint `json:"points"`
    Angle float32 `json:"angle"`
}

type Annotations struct {
    Annotations []json.RawMessage `json:"annotations"`
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
    Annotations []json.RawMessage `json:"annotations"`
    AllLabels []LabelMeEntry `json:"all_labels"`
}

type AnnotatedImage struct {
    ImageId string `json:"image_uuid"`
    AnnotationId string `json:"annotation_uuid"`
    Label string `json:"label"`
    Sublabel string `json:"sublabel"`
    Provider string `json:"provider"`
    Annotations []json.RawMessage `json:"annotations"`
    NumOfValid int32 `json:"num_yes"`
    NumOfInvalid int32 `json:"num_no"`
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

type DonationsPerAppStat struct {
    AppIdentifier string `json:"app_identifier"`
    Count int64 `json:"num"`
}

type ValidationsPerAppStat struct {
    AppIdentifier string `json:"app_identifier"`
    Count int64 `json:"num"`
}

type AnnotationsPerAppStat struct {
    AppIdentifier string `json:"app_identifier"`
    Count int64 `json:"num"`
}

type Statistics struct {
    Validations []ValidationStat `json:"validations"`
    DonationsPerCountry []DonationsPerCountryStat `json:"donations_per_country"`
    ValidationsPerCountry []ValidationsPerCountryStat `json:"validations_per_country"`
    AnnotationsPerCountry []AnnotationsPerCountryStat `json:"annotations_per_country"`
    DonationsPerApp []DonationsPerAppStat `json:"donations_per_app"`
    ValidationsPerApp []ValidationsPerAppStat `json:"validations_per_app"`
    AnnotationsPerApp []AnnotationsPerAppStat `json:"annotations_per_app"`
    NumOfUnlabeledDonations int64 `json:"num_of_unlabeled_donations"`
}

type LabelSearchItem struct {
    Label string `json:"label"`
    ParentLabel string `json:"parent_label"`
}

type LabelSearchResult struct {
    Labels []LabelSearchItem `json:"items"`
}

type AnnotationRefinementQuestion struct {
    Question string `json:"question"`
    Uuid int64 `json:"uuid"`
    RecommendedControl string `json:"recommended_control"`
}

type AnnotationRefinementAnswerExample struct {
    Filename string `json:"filename"`
    Attribution string `json:"attribution"`
}

type AnnotationRefinementAnswer struct {
    Label string `json:"label"`
    Id int64 `json:"id"`
    Examples []AnnotationRefinementAnswerExample `json:"examples"`
}

type AnnotationRefinement struct {
    Question AnnotationRefinementQuestion `json:"question"`
    //Answers []AnnotationRefinementAnswer `json:"answers"`
    Answers []json.RawMessage `json:"answers"`

    Metainfo struct {
        BrowseByExample bool `json:"browse_by_example"`
        AllowOther bool `json:"allow_other"`
        AllowUnknown bool `json:"allow_unknown"`
        MultiSelect bool `json:"multiselect"`
    } `json:"metainfo"`

    Image struct {
        Uuid string `json:"uuid"`
    } `json:"image"`

    Annotation struct{
        Uuid string `json:"uuid"`
        Annotation json.RawMessage `json:"annotation"`
    } `json:"annotation"`
}

type AnnotationRefinementEntry struct {
    LabelId int64 `json:"label_id"`
}

type ExportedImage struct {
    Id string `json:"uuid"`
    Provider string `json:"provider"`
    Annotations []json.RawMessage `json:"annotations"`
    Validations []json.RawMessage `json:"validations"`
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

func export(parseResult ParseResult) ([]ExportedImage, error){
    q := fmt.Sprintf(`SELECT i.key, json_agg(q3.annotations), q3.validations
                      FROM image i 
                      JOIN
                      (
                          SELECT COALESCE(q.image_id, q1.image_id) as image_id, q.annotations, q1.validations FROM 
                          (
                            SELECT an.image_id as image_id, (d.annotation || ('{"label":"' || a.accessor || '"}')::jsonb || ('{"type":"' || t.name || '"}')::jsonb)::jsonb as annotations 
                            FROM image_annotation_refinement r 
                            JOIN annotation_data d ON r.annotation_data_id = d.id
                            JOIN annotation_type t ON d.annotation_type_id = t.id
                            JOIN image_annotation an ON d.image_annotation_id = an.id
                            JOIN label_accessor a ON r.label_id = a.label_id
                            WHERE (%s)

                            UNION

                            SELECT n.image_id as image_id, (d.annotation || ('{"label":"' || a.accessor || '"}')::jsonb || ('{"type":"' || t.name || '"}')::jsonb)::jsonb as annotations 
                            FROM image_annotation n
                            JOIN annotation_data d ON d.image_annotation_id = n.id
                            JOIN annotation_type t ON d.annotation_type_id = t.id
                            JOIN label_accessor a ON n.label_id = a.label_id
                            WHERE (%s)
                          ) q
                          
                          FULL OUTER JOIN (
                            SELECT i.id as image_id, json_agg(json_build_object('label', accessor, 'num_yes', num_of_valid, 'num_no', num_of_invalid))::jsonb as validations
                            FROM image i 
                            JOIN image_validation v ON i.id = v.image_id
                            JOIN label_accessor a ON a.label_id = v.label_id
                            WHERE (%s)
                            GROUP BY i.id
                          ) q1 
                          ON q1.image_id = q.image_id
                      )q3
                              
                     ON i.id = q3.image_id
                      
                     WHERE i.unlocked = true
                     GROUP BY i.key, q3.validations`, parseResult.annotationQuery, parseResult.annotationQuery, parseResult.annotationQuery)
    rows, err := db.Query(q, parseResult.queryValues...)
    if err != nil {
        log.Debug("[Export] Couldn't export data: ", err.Error())
        raven.CaptureError(err, nil)
        return nil, err
    }
    defer rows.Close()

    imageEntries := []ExportedImage{}
    for rows.Next() {
        var image ExportedImage
        var annotations []byte
        var validations []byte
        image.Provider = "donation"

        err = rows.Scan(&image.Id, &annotations, &validations)
        if err != nil {
            log.Debug("[Export] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return nil, err
        }

        if len(annotations) > 0 {
            err := json.Unmarshal(annotations, &image.Annotations)
            if err != nil {
                log.Debug("[Export] Couldn't unmarshal annotations: ", err.Error())
                raven.CaptureError(err, nil)
                return nil, err
            }
        }

        if len(validations) > 0 {
            err := json.Unmarshal(validations, &image.Validations)
            if err != nil {
                log.Debug("[Export] Couldn't unmarshal validations: ", err.Error())
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
    
    rows, err := tx.Query(`SELECT CASE WHEN pl.name is null THEN l.name ELSE l.name || '/' || pl.name END, count(l.name), 
                           CASE WHEN SUM(v.num_of_valid + v.num_of_invalid) = 0 THEN 0 ELSE (CAST (SUM(v.num_of_invalid) AS float)/(SUM(v.num_of_valid) + SUM(v.num_of_invalid))) END as error_rate, 
                           SUM(v.num_of_valid + v.num_of_invalid) as total_validations
                           FROM image_validation v 
                           JOIN label l ON v.label_id = l.id 
                           LEFT JOIN label pl on l.parent_id = pl.id
                           GROUP BY l.name, pl.name ORDER BY count(l.name) DESC`)
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

    statistics.AnnotationsPerApp, err = _exploreAnnotationsPerApp(tx)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore annotations per app: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    statistics.DonationsPerApp, err = _exploreDonationsPerApp(tx)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore donations per app: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    statistics.ValidationsPerApp, err = _exploreValidationsPerApp(tx)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore validations per app: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    return statistics, tx.Commit()
}

func _exploreAnnotationsPerApp(tx *sql.Tx) ([]AnnotationsPerAppStat, error) {
    var annotationsPerApp []AnnotationsPerAppStat

    //get annotations grouped by app
    annotationsPerAppRows, err := tx.Query(`SELECT app_identifier, count FROM annotations_per_app ORDER BY count DESC`)
    if err != nil {
        return annotationsPerApp, err
    }
    defer annotationsPerAppRows.Close()

    for annotationsPerAppRows.Next() {
        var annotationsPerAppStat AnnotationsPerAppStat
        err = annotationsPerAppRows.Scan(&annotationsPerAppStat.AppIdentifier, &annotationsPerAppStat.Count)
        if err != nil {
            return annotationsPerApp, err
        }

        annotationsPerApp = append(annotationsPerApp, annotationsPerAppStat)
    }

    return annotationsPerApp, nil
}

func _exploreDonationsPerApp(tx *sql.Tx) ([]DonationsPerAppStat, error) {
    var donationsPerApp []DonationsPerAppStat

    //get donations grouped by app
    donationsPerAppRows, err := tx.Query(`SELECT app_identifier, count FROM donations_per_app ORDER BY count DESC`)
    if err != nil {
        return donationsPerApp, err
    }
    defer donationsPerAppRows.Close()

    for donationsPerAppRows.Next() {
        var donationsPerAppStat DonationsPerAppStat
        err = donationsPerAppRows.Scan(&donationsPerAppStat.AppIdentifier, &donationsPerAppStat.Count)
        if err != nil {
            return donationsPerApp, err
        }

        donationsPerApp = append(donationsPerApp, donationsPerAppStat)
    }

    return donationsPerApp, nil
}

func _exploreValidationsPerApp(tx *sql.Tx) ([]ValidationsPerAppStat, error) {
    var validationsPerApp []ValidationsPerAppStat

    //get validations grouped by app
    validationsPerAppRows, err := tx.Query(`SELECT app_identifier, count FROM validations_per_app ORDER BY count DESC`)
    if err != nil {
        return validationsPerApp, err
    }
    defer validationsPerAppRows.Close()

    for validationsPerAppRows.Next() {
        var validationsPerAppStat ValidationsPerAppStat
        err = validationsPerAppRows.Scan(&validationsPerAppStat.AppIdentifier, &validationsPerAppStat.Count)
        if err != nil {
            return validationsPerApp, err
        }

        validationsPerApp = append(validationsPerApp, validationsPerAppStat)
    }

    return validationsPerApp, nil
}


func getRandomImage() Image{
	var image Image

	image.Id = ""
	image.Label = ""
	image.Provider = "donation"

	rows, err := db.Query(`SELECT i.key, l.name, COALESCE(pl.name, ''), v.num_of_valid, v.num_of_invalid 
                           FROM image i 
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
        otherRows, err := db.Query(`SELECT i.key, l.name, COALESCE(pl.name, ''), v.num_of_valid, v.num_of_invalid
                                    FROM image i 
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

        err = otherRows.Scan(&image.Id, &label1, &label2, &image.NumOfValid, &image.NumOfInvalid)
        if(err != nil){
            log.Debug("[Fetch random image] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return image
        }
	} else{
        err = rows.Scan(&image.Id, &label1, &label2, &image.NumOfValid, &image.NumOfInvalid)
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


    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Add Annotation] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    insertedId := 0
    if annotations.Sublabel == "" {
        err = tx.QueryRow(`INSERT INTO image_annotation(label_id, num_of_valid, num_of_invalid, fingerprint_of_last_modification, image_id, uuid) 
                            SELECT (SELECT l.id FROM label l WHERE l.name = $5 AND l.parent_id is null), $2, $3, $4, (SELECT i.id FROM image i WHERE i.key = $1), 
                            uuid_generate_v4() RETURNING id`, 
                          imageId, 0, 0, clientFingerprint, annotations.Label).Scan(&insertedId)
    } else {
        err = tx.QueryRow(`INSERT INTO image_annotation(label_id, num_of_valid, num_of_invalid, fingerprint_of_last_modification, image_id, uuid) 
                            SELECT (SELECT l.id FROM label l JOIN label pl ON l.parent_id = pl.id WHERE l.name = $5 AND pl.name = $6), $2, $3, $4, 
                            (SELECT i.id FROM image i WHERE i.key = $1), uuid_generate_v4() RETURNING id`, 
                          imageId, 0, 0, clientFingerprint, annotations.Sublabel, annotations.Label).Scan(&insertedId)
    }


    if err != nil {
        tx.Rollback()
        log.Debug("[Add Annotation] Couldn't add image annotation: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    //insertes annotation data; 'type' gets removed before inserting data
    _, err = tx.Exec(`INSERT INTO annotation_data(image_annotation_id, annotation, annotation_type_id)
                            SELECT $1, ((q.*)::jsonb - 'type'), (SELECT id FROM annotation_type where name = ((q.*)->>'type')::text) FROM json_array_elements($2) q`, insertedId, byt)
    if err != nil {
        tx.Rollback()
        log.Debug("[Add Annotation] Couldn't add annotations: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }


    err = tx.Commit()
    if err != nil {
        log.Debug("[Add Annotation] Couldn't commit transaction: ", err.Error())
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

func getRandomAnnotatedImage() (AnnotatedImage, error) {
    var annotatedImage AnnotatedImage

    rows, err := db.Query(`SELECT i.key, l.name, COALESCE(pl.name, ''), a.uuid, q.annotations as annotations, a.num_of_valid, a.num_of_invalid 
                               FROM image i 
                               JOIN image_provider p ON i.image_provider_id = p.id 
                               JOIN image_annotation a ON a.image_id = i.id

                               JOIN
                               (
                                 SELECT json_agg(d.annotation || ('{"type":"' || t.name || '"}')::jsonb)::jsonb as annotations, 
                                 d.image_annotation_id as image_annotation_id
                                 FROM annotation_data d 
                                 JOIN annotation_type t on d.annotation_type_id = t.id
                                 GROUP BY d.image_annotation_id
                               ) q ON q.image_annotation_id = a.id


                               JOIN label l ON a.label_id = l.id
                               LEFT JOIN label pl ON l.parent_id = pl.id
                               WHERE i.unlocked = true AND p.name = 'donation'
                               GROUP BY i.key, a.uuid, l.name, pl.name, a.num_of_valid, a.num_of_invalid, q.annotations
                               OFFSET floor(random() * 
                               (
                                SELECT count(*) FROM image i 
                                JOIN image_provider p ON i.image_provider_id = p.id 
                                JOIN image_annotation a ON a.image_id = i.id
                                JOIN label l ON a.label_id = l.id
                                WHERE i.unlocked = true AND p.name = 'donation')
                               )LIMIT 1`)
    if err != nil {
        log.Debug("[Get Random Annotated Image] Couldn't get annotated image: ", err.Error())
        raven.CaptureError(err, nil)
        return annotatedImage, err
    }

    defer rows.Close()

    var label1 string
    var label2 string
    if rows.Next() {
        var annotations []byte
        annotatedImage.Provider = "donation"

        err = rows.Scan(&annotatedImage.ImageId, &label1, &label2, &annotatedImage.AnnotationId, &annotations, &annotatedImage.NumOfValid, &annotatedImage.NumOfInvalid)
        if err != nil {
            log.Debug("[Get Random Annotated Image] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return annotatedImage, err
        }

        err := json.Unmarshal(annotations, &annotatedImage.Annotations)
        if err != nil {
            log.Debug("[Get Random Annotated Image] Couldn't unmarshal: ", err.Error())
            raven.CaptureError(err, nil)
            return annotatedImage, err
        }

        if label2 == "" {
            annotatedImage.Label = label1
            annotatedImage.Sublabel = ""
        } else {
            annotatedImage.Label = label2
            annotatedImage.Sublabel = label1
        }
    }

    return annotatedImage, nil
}

func validateAnnotatedImage(clientFingerprint string, annotationId string, labelValidationEntry LabelValidationEntry, valid bool) error {
    if valid {
        var err error
        if labelValidationEntry.Sublabel == "" {
            _, err = db.Exec(`UPDATE image_annotation AS a 
                              SET num_of_valid = num_of_valid + 1, fingerprint_of_last_modification = $1
                              WHERE a.uuid = $2 AND a.label_id = (SELECT id FROM label WHERE name = $3 AND parent_id is null)`, 
                              clientFingerprint, annotationId, labelValidationEntry.Label)
        } else {
            _, err = db.Exec(`UPDATE image_annotation AS a 
                              SET num_of_valid = num_of_valid + 1, fingerprint_of_last_modification = $1
                              WHERE a.uuid = $2 AND a.label_id = (
                                SELECT l.id FROM label l 
                                JOIN label pl ON l.parent_id = pl.id
                                WHERE l.name = $3 AND pl.name = $4
                              )`, 
                              clientFingerprint, annotationId, labelValidationEntry.Sublabel, labelValidationEntry.Label)
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
                              WHERE a.uuid = $2 AND a.label_id = (
                                SELECT id FROM label WHERE name = $3 AND parent_id is null
                              )`, 
                              clientFingerprint, annotationId, labelValidationEntry.Label)
        } else {
            _,err = db.Exec(`UPDATE image_annotation AS a 
                              SET num_of_invalid = num_of_invalid + 1, fingerprint_of_last_modification = $1
                              WHERE a.uuid = $2 AND a.label_id = (
                                SELECT l.id FROM label l 
                                JOIN label pl ON l.parent_id = pl.id
                                WHERE l.name = $3 AND pl.name = $4
                              )`, 
                              clientFingerprint, annotationId, labelValidationEntry.Sublabel, labelValidationEntry.Label)
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

func updateContributionsPerApp(contributionType string, appIdentifier string) error {
    if contributionType == "donation" {
        _, err := db.Exec(`INSERT INTO donations_per_app (app_identifier, count)
                            VALUES ($1, $2) ON CONFLICT (app_identifier)
                            DO UPDATE SET count = donations_per_app.count + 1`, appIdentifier, 1)
        if err != nil {
            log.Debug("[Update Contributions per App] Couldn't insert into/update donations_per_app: ", err.Error())
            return err
        }
    } else if contributionType == "validation" {
        _, err := db.Exec(`INSERT INTO validations_per_app (app_identifier, count)
                            VALUES ($1, $2) ON CONFLICT (app_identifier)
                            DO UPDATE SET count = validations_per_app.count + 1`, appIdentifier, 1)
        if err != nil {
            log.Debug("[Update Contributions per App] Couldn't insert into/update validations_per_app: ", err.Error())
            return err
        }
    } else if contributionType == "annotation" {
        _, err := db.Exec(`INSERT INTO annotations_per_app (app_identifier, count)
                            VALUES ($1, $2) ON CONFLICT (app_identifier)
                            DO UPDATE SET count = annotations_per_app.count + 1`, appIdentifier, 1)
        if err != nil {
            log.Debug("[Update Contributions per App] Couldn't insert into/update annotations_per_app: ", err.Error())
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
                               JOIN (SELECT i.id as id, i.key as key FROM image i WHERE i.unlocked = true OFFSET floor(random() * (SELECT count(*) FROM image i WHERE unlocked = true)) LIMIT 1) q ON q.id = v.image_id
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

func getRandomAnnotationForRefinement() (AnnotationRefinement, error) {
    var bytes []byte
    var annotationBytes []byte
    var refinement AnnotationRefinement
    var annotations []json.RawMessage
    err := db.QueryRow(`SELECT i.key, s.quiz_question_id, s.quiz_question, s.quiz_answers, s1.annotations, s.recommended_control::text, s1.uuid, s.allow_unknown, 
                        s.allow_other, s.browse_by_example, s.multiselect
                        FROM ( 
                                SELECT qq.question as quiz_question, qq.recommended_control as recommended_control,
                                json_agg(json_build_object('id', l.id, 'label', l.name, 'examples', COALESCE(s2.examples, '[]'))) as quiz_answers, 
                                qq.refines_label_id as refines_label_id, qq.id as quiz_question_id, qq.allow_unknown as allow_unknown, qq.allow_other as allow_other, 
                                qq.browse_by_example as browse_by_example, qq.multiselect
    
                                FROM quiz_question qq 
                                JOIN quiz_answer q ON q.quiz_question_id = qq.id 
                                JOIN label l ON q.label_id = l.id
                                LEFT JOIN (
                                    SELECT e.label_id, json_agg(json_build_object('filename', e.filename, 'attribution', e.attribution))::jsonb as examples
                                    FROM label_example e GROUP BY label_id
                                ) s2 
                                ON s2.label_id = l.id 
                                GROUP BY qq.question, qq.refines_label_id, qq.id, qq.recommended_control
                             ) as s
                        JOIN (
                                SELECT a.uuid, a.label_id, a.image_id, json_agg(d.annotation || ('{"id":'||d.id||'}')::jsonb || ('{"type":"'||t.name||'"}')::jsonb)::jsonb as annotations 
                                FROM image_annotation a
                                JOIN annotation_data d ON d.image_annotation_id = a.id
                                JOIN annotation_type t ON d.annotation_type_id = t.id
                                WHERE CASE WHEN a.num_of_valid + a.num_of_invalid = 0 THEN 0 ELSE (CAST (a.num_of_valid AS float)/(a.num_of_valid + a.num_of_invalid)) END >= 0.8
                                GROUP BY a.label_id, a.image_id, a.uuid
                             ) as s1
                        ON s1.label_id =  s.refines_label_id 
                        JOIN image i ON i.id = s1.image_id
                        OFFSET floor(random() * 
                            ( SELECT count(*) FROM image_annotation a 
                              JOIN quiz_question q ON q.refines_label_id = a.label_id
                              WHERE CASE WHEN a.num_of_valid + a.num_of_invalid = 0 THEN 0 ELSE (CAST (a.num_of_valid AS float)/(a.num_of_valid + a.num_of_invalid)) END >= 0.8
                            )
                        ) LIMIT 1`).Scan(&refinement.Image.Uuid, &refinement.Question.Uuid, 
                            &refinement.Question.Question, &bytes, &annotationBytes, &refinement.Question.RecommendedControl, &refinement.Annotation.Uuid, &refinement.Metainfo.AllowUnknown, 
                            &refinement.Metainfo.AllowOther, &refinement.Metainfo.BrowseByExample, &refinement.Metainfo.MultiSelect)
    if err != nil {
        log.Debug("[Random Quiz question] Couldn't scan row: ", err.Error())
        raven.CaptureError(err, nil)
        return refinement, err 
    }

    err = json.Unmarshal(bytes, &refinement.Answers)
    if err != nil {
        log.Debug("[Random Quiz question] Couldn't unmarshal answers: ", err.Error())
        raven.CaptureError(err, nil)
        return refinement, err
    }

    err = json.Unmarshal(annotationBytes, &annotations)
    if err != nil {
        log.Debug("[Random Quiz question] Couldn't unmarshal annotations: ", err.Error())
        raven.CaptureError(err, nil)
        return refinement, err
    }

    if len(annotations) == 1 {
        refinement.Annotation.Annotation = annotations[0]
    } else if len(annotations) > 1 {
        randomVal := random(0, (len(annotations) - 1))
        refinement.Annotation.Annotation = annotations[randomVal]
    }

    return refinement, nil
}

func addOrUpdateRefinements(annotationUuid string, annotationDataId int64, annotationRefinementEntries []AnnotationRefinementEntry, clientFingerprint string) error {
    var err error

    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Add or Update Random Quiz question] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    for _, item := range annotationRefinementEntries {

        _, err = tx.Exec(`INSERT INTO image_annotation_refinement(annotation_data_id, label_id, num_of_valid, fingerprint_of_last_modification)
                            SELECT $1, $2, $3, $4 FROM image_annotation a JOIN annotation_data d ON d.image_annotation_id = a.id WHERE a.uuid = $5 AND d.id = $1
                          ON CONFLICT (annotation_data_id, label_id)
                          DO UPDATE SET fingerprint_of_last_modification = $4, num_of_valid = image_annotation_refinement.num_of_valid + 1
                          WHERE image_annotation_refinement.annotation_data_id = $1 AND image_annotation_refinement.label_id = $2`, 
                               annotationDataId, item.LabelId, 1, clientFingerprint, annotationUuid)
        
        if err != nil {
            tx.Rollback()
            log.Debug("[Add or Update Random Quiz question] Couldn't update: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }
    }

    err = tx.Commit()
    if err != nil {
        log.Debug("[Add or Update Random Quiz question] Couldn't commit transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    return nil
}

func addLabelSuggestion(suggestedLabel string) error {
     _, err := db.Exec(`INSERT INTO label_suggestion(name) VALUES($1)
                       ON CONFLICT (name) DO NOTHING`, suggestedLabel)
    if err != nil {
        log.Debug("[Add label suggestion] Couldn't insert: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    return nil
} 

func getLabelAccessors() ([]string, error) {
    var labels []string
    rows, err := db.Query(`SELECT accessor FROM label_accessor`)
    if err != nil {
        log.Debug("[Get label accessor] Couldn't insert: ", err.Error())
        raven.CaptureError(err, nil)
        return labels, err
    }
    defer rows.Close()

    var label string
    for rows.Next() {
        err = rows.Scan(&label)
        if err != nil {
           log.Debug("[Get label accessor] Couldn't scan row: ", err.Error())
           raven.CaptureError(err, nil)
           return labels, err 
        }

        labels = append(labels, label)
    }

    return labels, nil
}