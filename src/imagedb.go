package main

import (
    "github.com/lib/pq"
	"github.com/getsentry/raven-go"
	log "github.com/Sirupsen/logrus"
    "encoding/json"
    "database/sql"
    "fmt"
    "errors"
    "time"
    "github.com/dgrijalva/jwt-go"
    "./datastructures"
)

func sublabelsToStringlist(sublabels []datastructures.Sublabel) []string {
    var s []string
    for _, sublabel := range sublabels {
        s = append(s, sublabel.Name)
    }

    return s
}


func addDonatedPhoto(username string, imageInfo datastructures.ImageInfo, autoUnlock bool, clientFingerprint string, 
        labels []datastructures.LabelMeEntry) error{
	tx, err := db.Begin()
    if err != nil {
    	log.Debug("[Adding donated photo] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    //PostgreSQL can't store unsigned 64bit, so we are casting the hash to a signed 64bit value when storing the hash (so values above maxuint64/2 are negative). 
    //this should be ok, as we do not need to order those values, but just need to check if a hash exists. So it should be fine
	var imageId int64 
	err = tx.QueryRow("INSERT INTO image(key, unlocked, image_provider_id, hash, width, height) SELECT $1, $2, p.id, $3, $5, $6 FROM image_provider p WHERE p.name = $4 RETURNING id", 
					  imageInfo.Name, autoUnlock, int64(imageInfo.Hash), imageInfo.Source.Provider, imageInfo.Width, imageInfo.Height).Scan(&imageId)
	if err != nil {
		log.Debug("[Adding donated photo] Couldn't insert image: ", err.Error())
		raven.CaptureError(err, nil)
		tx.Rollback()
		return err
	}

    var insertedValidationIds []int64
    if labels[0].Label != "" { //only create a image validation entry, if a label is provided

        //per default we start with 0 validations, except if we are importing an image from a trusted
        //source. in that case, already set "numOfValid" to 1.
        numOfValid := 0
        if imageInfo.Source.Trusted {
            numOfValid = 1
        }

        insertedValidationIds, err = _addLabelsToImage(clientFingerprint, imageInfo.Name, labels, numOfValid, 0, tx)
        if err != nil {
            return err //tx already rolled back in case of error, so we can just return here
        }
    }


    if imageInfo.Source.Provider != "donation" {
        imageSourceId, err := _addImageSource(imageId, imageInfo.Source, tx)
        if err != nil {
            return err //tx already rolled back in case of error, so we can just return here
        }

        err = _addImageValidationSources(imageSourceId, insertedValidationIds, tx)
        if err != nil {
            return err //tx already rolled back in case of error, so we can just return here
        }
    }

    //in case a username is provided, link image to user account
    if username != "" {
        _, err := tx.Exec(`INSERT INTO user_image(image_id, account_id)
                            SELECT $1, id FROM account WHERE name = $2`, imageId, username)
        if err != nil {
            tx.Rollback()
            log.Debug("[Add user image entry] Couldn't add entry: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }
    }

	return tx.Commit()
}

func _addImageValidationSources(imageSourceId int64, imageValidationIds []int64, tx *sql.Tx) error {
    for _, id := range imageValidationIds {
        _, err := tx.Exec("INSERT INTO image_validation_source(image_source_id, image_validation_id) VALUES($1, $2)", imageSourceId, id)
        if err != nil {
            tx.Rollback()
            log.Debug("[Add image validation source] Couldn't add image validation source: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }
    }

    return nil
}

func _addImageSource(imageId int64, imageSource datastructures.ImageSource, tx *sql.Tx) (int64, error) {
    var insertedId int64
    err := tx.QueryRow("INSERT INTO image_source(image_id, url) VALUES($1, $2) RETURNING id", imageId, imageSource.Url).Scan(&insertedId)
    if err != nil {
        tx.Rollback()
        log.Debug("[Add image source] Couldn't add image source: ", err.Error())
        raven.CaptureError(err, nil)
        return insertedId, err
    }

    return insertedId, nil
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

func validateImages(apiUser datastructures.APIUser, imageValidationBatch datastructures.ImageValidationBatch, moderatorAction bool) error {
    var validEntries []string
    var invalidEntries []string
    var updatedRowIds []int64

    stepSize := 1
    if moderatorAction {
        stepSize = 5
    }

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
        rows, err := tx.Query(`UPDATE image_validation AS v 
                               SET num_of_invalid = num_of_invalid + $3, fingerprint_of_last_modification = $1
                               WHERE uuid = ANY($2) RETURNING id`, 
                               apiUser.ClientFingerprint, pq.Array(invalidEntries), stepSize)
        if err != nil {
            tx.Rollback()
            log.Debug("[Batch Validating donated photos] Couldn't increase num_of_invalid: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }

        defer rows.Close()

        for rows.Next() {
            var updatedRowId int64
            err = rows.Scan(&updatedRowId)
            if err != nil {
                tx.Rollback()
                log.Debug("[Batch Validating donated photos] Couldn't scan row: ", err.Error())
                raven.CaptureError(err, nil)
                return err
            }

            updatedRowIds = append(updatedRowIds, updatedRowId)
        }
    }

    if len(validEntries) > 0 {
        rows1, err := tx.Query(`UPDATE image_validation AS v 
                              SET num_of_valid = num_of_valid + $3, fingerprint_of_last_modification = $1
                              WHERE uuid = ANY($2) RETURNING id`, 
                              apiUser.ClientFingerprint, pq.Array(validEntries), stepSize)
        if err != nil {
            tx.Rollback()
            log.Debug("[Batch Validating donated photos] Couldn't increase num_of_valid: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }

        defer rows1.Close()

        for rows1.Next() {
            var updatedRowId int64
            err = rows1.Scan(&updatedRowId)
            if err != nil {
                tx.Rollback()
                log.Debug("[Batch Validating donated photos] Couldn't scan row: ", err.Error())
                raven.CaptureError(err, nil)
                return err
            }

            updatedRowIds = append(updatedRowIds, updatedRowId)
        }
    }


    if apiUser.Name != "" {
        if len(updatedRowIds) == 0 {
            tx.Rollback()
            err := errors.New("nothing updated")
            log.Debug("[Batch Validating donated photos] ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }


        _, err = tx.Exec(`INSERT INTO user_image_validation(image_validation_id, account_id, timestamp)
                            SELECT unnest($1::integer[]), a.id, CURRENT_TIMESTAMP FROM account a WHERE a.name = $2`, pq.Array(updatedRowIds), apiUser.Name)
        if err != nil {
            tx.Rollback()
            log.Debug("[Batch Validating donated photos] Couldn't add user image validation entry: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }
    }

    err = tx.Commit()
    if err != nil {
        log.Debug("[Batch Validating donated photos] Couldn't commit transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }


    return nil
}

func export(parseResult ParseResult, annotationsOnly bool) ([]datastructures.ExportedImage, error){
    joinType := "FULL OUTER JOIN"
    if annotationsOnly {
        joinType = "JOIN"
    }

    
    q1 := ""
    q2 := ""
    q3 := ""
    identifier := ""
    if parseResult.isUuidQuery {
        q1 = "JOIN label l ON l.id = r.label_id"
        q2 = "JOIN label l ON l.id = n.label_id"
        q3 = "JOIN label l ON l.id = v.label_id"
        identifier = "l.name"
    } else {
        q1 = "JOIN label_accessor a ON r.label_id = a.label_id"
        q2 = "JOIN label_accessor a ON n.label_id = a.label_id"
        q3 = "JOIN label_accessor a ON a.label_id = v.label_id"
        identifier = "a.accessor"
    }


    q := fmt.Sprintf(`SELECT i.key, CASE WHEN json_agg(q3.annotations)::jsonb = '[null]'::jsonb THEN '[]' ELSE json_agg(q3.annotations)::jsonb END as annotations, 
                      q3.validations, i.width, i.height
                      FROM image i 
                      JOIN
                      (
                          SELECT COALESCE(q.image_id, q1.image_id) as image_id, q.annotations, q1.validations FROM 
                          (
                            SELECT an.image_id as image_id, (d.annotation || ('{"label":"' || %s || '"}')::jsonb || ('{"type":"' || t.name || '"}')::jsonb)::jsonb as annotations 
                            FROM image_annotation_refinement r 
                            JOIN annotation_data d ON r.annotation_data_id = d.id
                            JOIN annotation_type t ON d.annotation_type_id = t.id
                            JOIN image_annotation an ON d.image_annotation_id = an.id
                            %s
                            WHERE ((%s) AND an.auto_generated = false)

                            UNION

                            SELECT n.image_id as image_id, (d.annotation || ('{"label":"' || %s || '"}')::jsonb || ('{"type":"' || t.name || '"}')::jsonb)::jsonb as annotations 
                            FROM image_annotation n
                            JOIN annotation_data d ON d.image_annotation_id = n.id
                            JOIN annotation_type t ON d.annotation_type_id = t.id
                            %s
                            WHERE ((%s) AND n.auto_generated = false)
                          ) q
                          
                          %s (
                            SELECT i.id as image_id, json_agg(json_build_object('label', %s, 'num_yes', num_of_valid, 'num_no', num_of_invalid))::jsonb as validations
                            FROM image i 
                            JOIN image_validation v ON i.id = v.image_id
                            %s
                            WHERE (%s)
                            GROUP BY i.id
                          ) q1 
                          ON q1.image_id = q.image_id
                      )q3
                              
                     ON i.id = q3.image_id
                      
                     WHERE i.unlocked = true
                     GROUP BY i.key, q3.validations, i.width, i.height`, identifier, q1, parseResult.query, identifier, q2, parseResult.query, joinType, identifier, q3, parseResult.query)
    rows, err := db.Query(q, parseResult.queryValues...)
    if err != nil {
        log.Debug("[Export] Couldn't export data: ", err.Error())
        raven.CaptureError(err, nil)
        return nil, err
    }
    defer rows.Close()

    imageEntries := []datastructures.ExportedImage{}
    for rows.Next() {
        var image datastructures.ExportedImage
        var annotations []byte
        var validations []byte
        image.Provider = "donation"

        err = rows.Scan(&image.Id, &annotations, &validations, &image.Width, &image.Height)
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

func explore(words []string) (datastructures.Statistics, error) {
    statistics := datastructures.Statistics{}

    //use temporary map for faster lookup
    temp := make(map[string]datastructures.ValidationStat)

    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Explore] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }
    
    rows, err := tx.Query(`SELECT CASE WHEN pl.name is null THEN l.name ELSE l.name || '/' || pl.name END, count(l.name), 
                           CASE 
                            WHEN SUM(v.num_of_valid + v.num_of_invalid) = 0 THEN 0 
                            ELSE (CAST (SUM(v.num_of_invalid) AS float)/(SUM(v.num_of_valid) + SUM(v.num_of_invalid))) 
                           END as error_rate, 
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
        var validationStat datastructures.ValidationStat
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
            var validationStat datastructures.ValidationStat
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
        var donationsPerCountryStat datastructures.DonationsPerCountryStat
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
        var validationsPerCountryStat datastructures.ValidationsPerCountryStat
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
        var annotationsPerCountryStat datastructures.AnnotationsPerCountryStat
        err = annotationsPerCountryRows.Scan(&annotationsPerCountryStat.CountryCode, &annotationsPerCountryStat.Count)
        if err != nil {
            tx.Rollback()
            log.Debug("[Explore] Couldn't scan data row: ", err.Error())
            raven.CaptureError(err, nil)
            return statistics, err
        }

        statistics.AnnotationsPerCountry = append(statistics.AnnotationsPerCountry, annotationsPerCountryStat)
    }


    //get annotation refinements grouped by country
    annotationRefinementsPerCountryRows, err := tx.Query(`SELECT country_code, count FROM annotation_refinements_per_country ORDER BY count DESC`)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore data: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }
    defer annotationRefinementsPerCountryRows.Close()

    for annotationRefinementsPerCountryRows.Next() {
        var annotationRefinementsPerCountryStat datastructures.AnnotationRefinementsPerCountryStat
        err = annotationRefinementsPerCountryRows.Scan(&annotationRefinementsPerCountryStat.CountryCode, &annotationRefinementsPerCountryStat.Count)
        if err != nil {
            tx.Rollback()
            log.Debug("[Explore] Couldn't scan data row: ", err.Error())
            raven.CaptureError(err, nil)
            return statistics, err
        }

        statistics.AnnotationRefinementsPerCountry = append(statistics.AnnotationRefinementsPerCountry, annotationRefinementsPerCountryStat)
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

    statistics.NumOfDonations, err = _getTotalDonations(tx)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't get total donations: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    statistics.NumOfAnnotations, err = _getTotalAnnotations(tx, false)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't get total annotations: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    statistics.NumOfValidations, err = _getTotalValidations(tx)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't get total validations: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    statistics.NumOfAnnotationRefinements, err = _getTotalAnnotationRefinements(tx)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't get total annotation refinements: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    statistics.NumOfLabelSuggestions, err = _getTotalLabelSuggestions(tx)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't get total label suggestions: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    statistics.NumOfLabels, err = _getTotalLabels(tx)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't get total labels: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    return statistics, tx.Commit()
}

func _getTotalDonations(tx *sql.Tx) (int64, error) {
    var numOfTotalDonations int64
    numOfTotalDonations = 0

    rows, err := tx.Query(`SELECT count(*) FROM image i`)
    if err != nil {
        return numOfTotalDonations, nil
    }

    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&numOfTotalDonations)
        if err != nil {
            return numOfTotalDonations, err
        }
    }

    return numOfTotalDonations, nil
}

func _getTotalLabels(tx *sql.Tx) (int64, error) {
    var numOfTotalLabels int64
    numOfTotalLabels = 0

    rows, err := tx.Query(`SELECT count(*) FROM label l`)
    if err != nil {
        return numOfTotalLabels, nil
    }

    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&numOfTotalLabels)
        if err != nil {
            return numOfTotalLabels, err
        }
    }

    return numOfTotalLabels, nil
}

func _getTotalLabelSuggestions(tx *sql.Tx) (int64, error) {
    var numOfTotalLabelSuggestions int64
    numOfTotalLabelSuggestions = 0

    rows, err := tx.Query(`SELECT count(*) FROM label_suggestion l`)
    if err != nil {
        return numOfTotalLabelSuggestions, nil
    }

    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&numOfTotalLabelSuggestions)
        if err != nil {
            return numOfTotalLabelSuggestions, err
        }
    }

    return numOfTotalLabelSuggestions, nil
}

func _getTotalAnnotations(tx *sql.Tx, includeAutoGeneratedAnnotations bool) (int64, error) {
    var numOfAnnotations int64
    numOfAnnotations = 0

    q1 := ""
    if !includeAutoGeneratedAnnotations {
        q1 = "WHERE a.auto_generated = false"
    }

    q := fmt.Sprintf(`SELECT count(*) FROM image_annotation a %s`, q1)

    rows, err := tx.Query(q)
    if err != nil {
        return numOfAnnotations, nil
    }

    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&numOfAnnotations)
        if err != nil {
            return numOfAnnotations, err
        }
    }

    return numOfAnnotations, nil
}

func _getTotalValidations(tx *sql.Tx) (int64, error) {
    var numOfValidations int64
    numOfValidations = 0

    rows, err := tx.Query(`SELECT count(*) FROM image_validation v`)
    if err != nil {
        return numOfValidations, nil
    }

    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&numOfValidations)
        if err != nil {
            return numOfValidations, err
        }
    }

    return numOfValidations, nil
}

func _getTotalAnnotationRefinements(tx *sql.Tx) (int64, error) {
    var numOfAnnotationRefinements int64
    numOfAnnotationRefinements = 0

    rows, err := tx.Query(`SELECT count(*) FROM image_annotation_refinement r`)
    if err != nil {
        return numOfAnnotationRefinements, nil
    }

    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&numOfAnnotationRefinements)
        if err != nil {
            return numOfAnnotationRefinements, err
        }
    }

    return numOfAnnotationRefinements, nil
}

func _exploreAnnotationsPerApp(tx *sql.Tx) ([]datastructures.AnnotationsPerAppStat, error) {
    var annotationsPerApp []datastructures.AnnotationsPerAppStat

    //get annotations grouped by app
    annotationsPerAppRows, err := tx.Query(`SELECT app_identifier, count FROM annotations_per_app ORDER BY count DESC`)
    if err != nil {
        return annotationsPerApp, err
    }
    defer annotationsPerAppRows.Close()

    for annotationsPerAppRows.Next() {
        var annotationsPerAppStat datastructures.AnnotationsPerAppStat
        err = annotationsPerAppRows.Scan(&annotationsPerAppStat.AppIdentifier, &annotationsPerAppStat.Count)
        if err != nil {
            return annotationsPerApp, err
        }

        annotationsPerApp = append(annotationsPerApp, annotationsPerAppStat)
    }

    return annotationsPerApp, nil
}

func _exploreDonationsPerApp(tx *sql.Tx) ([]datastructures.DonationsPerAppStat, error) {
    var donationsPerApp []datastructures.DonationsPerAppStat

    //get donations grouped by app
    donationsPerAppRows, err := tx.Query(`SELECT app_identifier, count FROM donations_per_app ORDER BY count DESC`)
    if err != nil {
        return donationsPerApp, err
    }
    defer donationsPerAppRows.Close()

    for donationsPerAppRows.Next() {
        var donationsPerAppStat datastructures.DonationsPerAppStat
        err = donationsPerAppRows.Scan(&donationsPerAppStat.AppIdentifier, &donationsPerAppStat.Count)
        if err != nil {
            return donationsPerApp, err
        }

        donationsPerApp = append(donationsPerApp, donationsPerAppStat)
    }

    return donationsPerApp, nil
}

func _exploreValidationsPerApp(tx *sql.Tx) ([]datastructures.ValidationsPerAppStat, error) {
    var validationsPerApp []datastructures.ValidationsPerAppStat

    //get validations grouped by app
    validationsPerAppRows, err := tx.Query(`SELECT app_identifier, count FROM validations_per_app ORDER BY count DESC`)
    if err != nil {
        return validationsPerApp, err
    }
    defer validationsPerAppRows.Close()

    for validationsPerAppRows.Next() {
        var validationsPerAppStat datastructures.ValidationsPerAppStat
        err = validationsPerAppRows.Scan(&validationsPerAppStat.AppIdentifier, &validationsPerAppStat.Count)
        if err != nil {
            return validationsPerApp, err
        }

        validationsPerApp = append(validationsPerApp, validationsPerAppStat)
    }

    return validationsPerApp, nil
}


func getImageToValidate(imageId string, labelId string, username string) (datastructures.ValidationImage, error) {
	var image datastructures.ValidationImage

	image.Id = ""
	image.Label = ""
	image.Provider = "donation"

    nextParam := 1
    labelIdStr := ""
    if labelId != "" {
        if imageId == "" {
            labelIdStr = " AND l.uuid = $1"
            nextParam = 2
        } else {
            labelIdStr = " AND l.uuid = $2"
            nextParam = 3
        }
    } else {
        if imageId != "" {
            nextParam = 2
        }
    }

    includeOwnImageDonations := ""
    if username != "" {
        includeOwnImageDonations = fmt.Sprintf(`OR (
                                                EXISTS 
                                                    (
                                                        SELECT 1 
                                                        FROM user_image u
                                                        JOIN account a ON a.id = u.account_id
                                                        WHERE u.image_id = i.id AND a.name = $%d
                                                    )
                                                AND NOT EXISTS 
                                                    (
                                                        SELECT 1 
                                                        FROM image_quarantine q 
                                                        WHERE q.image_id = i.id 
                                                    )
                                               )`, nextParam)
    }

    //either select a specific image with a given image id or try to select 
    //an image that's not already validated (as they have preference). 
    imageIdStr := "(v.num_of_valid = 0) AND (v.num_of_invalid = 0)"
    if imageId != "" {
        imageIdStr = "i.key = $1"
    }

    q := fmt.Sprintf(`SELECT i.key, l.name, COALESCE(pl.name, ''), v.num_of_valid, v.num_of_invalid, v.uuid, i.unlocked
                        FROM image i 
                        JOIN image_provider p ON i.image_provider_id = p.id 
                        JOIN image_validation v ON v.image_id = i.id
                        JOIN label l ON v.label_id = l.id
                        LEFT JOIN label pl ON l.parent_id = pl.id
                        WHERE ((i.unlocked = true %s) AND (p.name = 'donation') 
                        AND %s%s) LIMIT 1`,includeOwnImageDonations, imageIdStr, labelIdStr)

	var rows *sql.Rows
    var err error
    var queryParams []interface{}
    if imageId != "" {
        queryParams = append(queryParams, imageId) 
    }

    if labelId != "" {
        queryParams = append(queryParams, labelId) 
    }

    if username != "" {
        queryParams = append(queryParams, username) 
    }

    rows, err = db.Query(q, queryParams...)
	

    if err != nil {
		log.Debug("[Fetch image] Couldn't fetch random image: ", err.Error())
		raven.CaptureError(err, nil)
		return image, err
	}
    defer rows.Close()
	
    var label1 string
    var label2 string
	if !rows.Next() {
        //if we provided a image id, but we get no result, its an error. So return here
        if imageId != "" {
            return image, nil
        }


        var otherRows *sql.Rows

        q1 := fmt.Sprintf(`SELECT i.key, l.name, COALESCE(pl.name, ''), v.num_of_valid, v.num_of_invalid, v.uuid, i.unlocked
                            FROM image i 
                            JOIN image_provider p ON i.image_provider_id = p.id 
                            JOIN image_validation v ON v.image_id = i.id
                            JOIN label l ON v.label_id = l.id
                            LEFT JOIN label pl ON l.parent_id = pl.id
                            WHERE (i.unlocked = true %s) AND p.name = 'donation'%s 
                            OFFSET floor(random() * 
                                ( SELECT count(*) FROM image i 
                                  JOIN image_provider p ON i.image_provider_id = p.id 
                                  JOIN image_validation v ON v.image_id = i.id 
                                  JOIN label l ON v.label_id = l.id
                                  WHERE (i.unlocked = true %s) AND p.name = 'donation'%s
                                )
                            ) LIMIT 1`, includeOwnImageDonations, labelIdStr, includeOwnImageDonations, labelIdStr)

        if labelId != "" {
            if username != "" {
                otherRows, err = db.Query(q1, labelId, username)
            } else {
                otherRows, err = db.Query(q1, labelId)
            }
        } else {
            if username != "" {
                otherRows, err = db.Query(q1, username)
            } else {
                otherRows, err = db.Query(q1)
            }
        }

        if err != nil {
            log.Debug("[Fetch random image] Couldn't fetch random image: ", err.Error())
            raven.CaptureError(err, nil)
            return image, err
        }

        defer otherRows.Close()
        
        if otherRows.Next() {
            err = otherRows.Scan(&image.Id, &label1, &label2, &image.Validation.NumOfValid, 
                                    &image.Validation.NumOfInvalid, &image.Validation.Id, &image.Unlocked)
            if err != nil {
                log.Debug("[Fetch random image] Couldn't scan row: ", err.Error())
                raven.CaptureError(err, nil)
                return image, err
            }
        }
	} else{
        err = rows.Scan(&image.Id, &label1, &label2, &image.Validation.NumOfValid, 
                            &image.Validation.NumOfInvalid, &image.Validation.Id, &image.Unlocked)
        if err != nil {
            log.Debug("[Fetch random image] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return image, err
        }
    }

    if label2 == "" {
        image.Label = label1
        image.Sublabel = ""
    } else {
        image.Label = label2
        image.Sublabel = label1
    }

	return image, nil
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
func getRandomGroupedImages(label string, limit int) ([]datastructures.ValidationImage, error) {
    var images []datastructures.ValidationImage

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
    rows, err := db.Query(`SELECT i.key, l.name, v.num_of_valid, v.num_of_invalid, v.uuid FROM image i 
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
        var image datastructures.ValidationImage
        image.Provider = "donation"
        err = rows.Scan(&image.Id, &image.Label, &image.Validation.NumOfValid, &image.Validation.NumOfInvalid, &image.Validation.Id)
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

func updateAnnotation(apiUser datastructures.APIUser, annotationId string, annotations datastructures.Annotations) error {
    byt, err := json.Marshal(annotations.Annotations)
    if err != nil {
        log.Debug("[Add Annotation] Couldn't create byte array: ", err.Error())
        return err
    }

    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Update Annotation] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    var imageAnnotationRevisionId int64
    //add entry to image_annotation_revision table
    err = tx.QueryRow(`INSERT INTO image_annotation_revision(image_annotation_id, revision)
                         SELECT a.id, a.revision FROM image_annotation a
                         WHERE a.uuid = $1 RETURNING id`, annotationId).Scan(&imageAnnotationRevisionId)
    if err != nil {
        tx.Rollback()
        log.Debug("[Update Annotation] Couldn't insert to annotation revision: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    _, err = tx.Exec(`UPDATE annotation_data
                     SET image_annotation_id = NULL, image_annotation_revision_id = $2
                     FROM image_annotation a WHERE a.uuid = $1 
                     AND a.id = image_annotation_id`, 
                     annotationId, imageAnnotationRevisionId)
    if err != nil {
        tx.Rollback()
        log.Debug("[Update Annotation] Couldn't update annotation data: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    var imageAnnotationId int64
    err = tx.QueryRow(`UPDATE image_annotation a SET num_of_valid = 0, num_of_invalid = 0, revision = revision + 1
                       WHERE uuid = $1 
                       RETURNING id`, annotationId).Scan(&imageAnnotationId)
    if err != nil {
        tx.Rollback()
        log.Debug("[Update Annotation] Couldn't update annotation: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    //insertes annotation data; 'type' gets removed before inserting data
    _, err = tx.Exec(`INSERT INTO annotation_data(image_annotation_id, uuid, annotation, annotation_type_id)
                            SELECT $1, uuid_generate_v4(), ((q.*)::jsonb - 'type'), (SELECT id FROM annotation_type where name = ((q.*)->>'type')::text) FROM json_array_elements($2) q`, imageAnnotationId, byt)
    if err != nil {
        tx.Rollback()
        log.Debug("[Update Annotation] Couldn't add new annotation data: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    err = tx.Commit()
    if err != nil {
        log.Debug("[Update Annotation] Couldn't commit transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    return nil
}

func addAnnotations(apiUser datastructures.APIUser, imageId string, annotations datastructures.Annotations, autoGenerated bool) (string, error) {
    //currently there is a uniqueness constraint on the image_id column to ensure that we only have
    //one image annotation per image. That means that the below query can fail with a unique constraint error. 
    //at the moment the uniqueness constraint errors are handled gracefully - that means we return nil.
    //we might want to change that in the future to support multiple annotations per image (if there is a use case for it),
    //but for now it should be fine.
    var annotationId string
    annotationId = ""

    byt, err := json.Marshal(annotations.Annotations)
    if err != nil {
        log.Debug("[Add Annotation] Couldn't create byte array: ", err.Error())
        return annotationId, err
    }


    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Add Annotation] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return annotationId, err
    }

    insertedId := 0
    if annotations.Sublabel == "" {
        err = tx.QueryRow(`INSERT INTO image_annotation(label_id, num_of_valid, num_of_invalid, fingerprint_of_last_modification, image_id, uuid, auto_generated, revision) 
                            SELECT (SELECT l.id FROM label l WHERE l.name = $5 AND l.parent_id is null), $2, $3, $4, (SELECT i.id FROM image i WHERE i.key = $1), 
                            uuid_generate_v4(), $6, $7 RETURNING id, uuid`, 
                          imageId, 0, 0, apiUser.ClientFingerprint, annotations.Label, autoGenerated, 1).Scan(&insertedId, &annotationId)
    } else {
        err = tx.QueryRow(`INSERT INTO image_annotation(label_id, num_of_valid, num_of_invalid, fingerprint_of_last_modification, image_id, uuid, auto_generated, revision) 
                            SELECT (SELECT l.id FROM label l JOIN label pl ON l.parent_id = pl.id WHERE l.name = $5 AND pl.name = $6), $2, $3, $4, 
                            (SELECT i.id FROM image i WHERE i.key = $1), uuid_generate_v4(), $7, $8 RETURNING id, uuid`, 
                          imageId, 0, 0, apiUser.ClientFingerprint, annotations.Sublabel, annotations.Label, autoGenerated, 1).Scan(&insertedId, &annotationId)
    }


    if err != nil {
        if pqErr, ok := err.(*pq.Error); ok {
            if pqErr.Code.Name() == "unique_violation" {
                tx.Commit()
                return annotationId, err
            }
        }

        tx.Rollback()
        log.Debug("[Add Annotation] Couldn't add image annotation: ", err.Error())
        raven.CaptureError(err, nil)
        return annotationId, err
    }

    //insertes annotation data; 'type' gets removed before inserting data
    _, err = tx.Exec(`INSERT INTO annotation_data(image_annotation_id, uuid, annotation, annotation_type_id)
                            SELECT $1, uuid_generate_v4(), ((q.*)::jsonb - 'type'), (SELECT id FROM annotation_type where name = ((q.*)->>'type')::text) FROM json_array_elements($2) q`, insertedId, byt)
    if err != nil {
        tx.Rollback()
        log.Debug("[Add Annotation] Couldn't add annotations: ", err.Error())
        raven.CaptureError(err, nil)
        return annotationId, err
    }

    if apiUser.Name != "" {
        var id int64

        id = 0
        err = tx.QueryRow(`INSERT INTO user_image_annotation(image_annotation_id, account_id, timestamp)
                                SELECT $1, a.id, CURRENT_TIMESTAMP FROM account a WHERE a.name = $2 RETURNING id`, insertedId, apiUser.Name).Scan(&id)
        if err != nil {
            tx.Rollback()
            log.Debug("[Add User Annotation] Couldn't add user annotation entry: ", err.Error())
            raven.CaptureError(err, nil)
            return annotationId, err
        }

        if id == 0 {
            tx.Rollback()
            log.Debug("[Add User Annotation] Nothing inserted")
            return annotationId, errors.New("nothing inserted")
        }
    }


    err = tx.Commit()
    if err != nil {
        log.Debug("[Add Annotation] Couldn't commit transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return annotationId, err
    }

    return annotationId, nil
}

func _getImageForAnnotationFromValidationId(username string, validationId string, addAutoAnnotations bool) (datastructures.UnannotatedImage, error) {
    var unannotatedImage datastructures.UnannotatedImage

    includeOwnImageDonations := ""
    if username != "" {
        includeOwnImageDonations = `OR (
                                        EXISTS 
                                        (
                                            SELECT 1 
                                            FROM user_image u
                                            JOIN account a ON a.id = u.account_id
                                            WHERE u.image_id = i.id AND a.name = $2
                                        )
                                        AND NOT EXISTS 
                                        (
                                            SELECT 1 
                                            FROM image_quarantine q 
                                            WHERE q.image_id = i.id 
                                        )
                                       )`
        
    }

    q := fmt.Sprintf(`SELECT i.key, l.name, COALESCE(pl.name, '') as parent_label, i.width, i.height, v.uuid, 
                           json_agg(q1.annotation || ('{"type":"' || q1.name || '"}')::jsonb)::jsonb as auto_annotations,
                           i.unlocked
                            FROM image i 
                            JOIN image_provider p ON i.image_provider_id = p.id 
                            JOIN image_validation v ON v.image_id = i.id
                            JOIN label l ON v.label_id = l.id
                            LEFT JOIN label pl ON l.parent_id = pl.id

                            LEFT JOIN 
                            (
                                SELECT a.label_id as label_id, a.image_id as image_id, d.annotation, t.name
                                FROM image_annotation a 
                                JOIN annotation_data d ON d.image_annotation_id = a.id
                                JOIN annotation_type t on d.annotation_type_id = t.id
                                WHERE a.auto_generated = true
                            ) q1 ON l.id = q1.label_id AND i.id = q1.image_id 
                            WHERE (i.unlocked = true %s) AND p.name = 'donation' AND v.uuid::text = $1
                            GROUP BY i.key, l.name, pl.name, width, height, v.uuid, i.unlocked`, includeOwnImageDonations)

    //we do not check, whether there already exists a annotation for the given validation id. 
    //there is anyway only one annotation per validation allowed, so if someone tries to push another annotation, the corresponding POST request 
    //would fail 
    var rows *sql.Rows
    var err error

    if username == "" {
        rows, err = db.Query(q, validationId)
    } else {
        rows, err = db.Query(q, validationId, username)
    }

    if err != nil {
        log.Debug("[Get specific Image for Annotation] Couldn't get annotation ", err.Error())
        raven.CaptureError(err, nil)
        return unannotatedImage, err
    }

    defer rows.Close()

    var label1 string
    var label2 string
    var autoAnnotationBytes []byte
    if rows.Next() {
        unannotatedImage.Provider = "donation"

        err = rows.Scan(&unannotatedImage.Id, &label1, &label2, &unannotatedImage.Width, &unannotatedImage.Height, 
                            &unannotatedImage.Validation.Id, &autoAnnotationBytes, &unannotatedImage.Unlocked)
        if err != nil {
            log.Debug("[Get specific Image for Annotation] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return unannotatedImage, err
        }

        if addAutoAnnotations {
            if len(autoAnnotationBytes) > 0 {
                err = json.Unmarshal(autoAnnotationBytes, &unannotatedImage.AutoAnnotations)
                if err != nil {
                    log.Debug("[Get specific Image for Annotation] Couldn't unmarshal auto annotations: ", err.Error())
                    raven.CaptureError(err, nil)
                    return unannotatedImage, err
                }
            }
        }

        if label2 == "" {
            unannotatedImage.Label = label1
            unannotatedImage.Sublabel = ""
        } else {
            unannotatedImage.Label = label2
            unannotatedImage.Sublabel = label1
        }
    }

    return unannotatedImage, nil
}

func getImageForAnnotation(username string, addAutoAnnotations bool, validationId string, labelId string) (datastructures.UnannotatedImage, error) {
    //if a validation id is provided, use a different code path. 
    //selecting a single image given a validation id is totally different from selecting a random image
    //so it makes sense to use a different code path here. 
    if validationId != "" {
        return _getImageForAnnotationFromValidationId(username, validationId, addAutoAnnotations)
    }


    var unannotatedImage datastructures.UnannotatedImage

    //specify the max. number of not-annotatables before we skip the annotation task
    maxNumNotAnnotatable := 3

    q1 := ""
    posNum := 1
    if labelId != "" {
        q1 = "AND l.uuid = $1"
        posNum = 2
    }

    q3 := fmt.Sprintf("AND v.num_of_not_annotatable < $%d", posNum)
    posNum += 1

    includeOwnImageDonations := ""
    q2 := ""
    if username != "" {
        q2 = fmt.Sprintf(`AND NOT EXISTS
                           (
                                SELECT 1 FROM user_annotation_blacklist bl 
                                JOIN account acc ON acc.id = bl.account_id
                                WHERE bl.image_validation_id = v.id AND acc.name = $%d
                           )`, posNum)

        includeOwnImageDonations = fmt.Sprintf(`OR (
                                                    EXISTS 
                                                        (
                                                            SELECT 1 
                                                            FROM user_image u
                                                            JOIN account a ON a.id = u.account_id
                                                            WHERE u.image_id = i.id AND a.name = $%d
                                                        )
                                                    AND NOT EXISTS 
                                                        (
                                                            SELECT 1 
                                                            FROM image_quarantine q 
                                                            WHERE q.image_id = i.id 
                                                        )
                                                   )`, posNum) 
        
    }


    q := fmt.Sprintf(`SELECT q.image_key, q.label, q.parent_label, q.image_width, q.image_height, q.validation_uuid, 
                        CASE WHEN json_agg(q1.annotation)::jsonb = '[null]'::jsonb THEN '[]' ELSE json_agg(q1.annotation || ('{"type":"' || q1.annotation_type || '"}')::jsonb)::jsonb END as auto_annotations,
                        q.image_unlocked
                        FROM
                        (SELECT l.id as label_id, i.id as image_id, i.key as image_key, l.name as label, COALESCE(pl.name, '') as parent_label, 
                            width as image_width, height as image_height, v.uuid as validation_uuid, i.unlocked as image_unlocked
                            FROM image i 
                            JOIN image_provider p ON i.image_provider_id = p.id 
                            JOIN image_validation v ON v.image_id = i.id
                            JOIN label l ON v.label_id = l.id
                            LEFT JOIN label pl ON l.parent_id = pl.id
                            WHERE (i.unlocked = true %s) AND p.name = 'donation' AND 
                            CASE WHEN v.num_of_valid + v.num_of_invalid = 0 THEN 0 ELSE (CAST (v.num_of_valid AS float)/(v.num_of_valid + v.num_of_invalid)) END >= 0.8
                            %s
                            AND NOT EXISTS
                            (
                                SELECT 1 FROM image_annotation a 
                                WHERE a.label_id = v.label_id AND a.image_id = v.image_id AND a.auto_generated = false
                            )
                            %s
                            %s
                            OFFSET floor
                            ( random() * 
                                (
                                    SELECT count(*) FROM image i
                                    JOIN image_provider p ON i.image_provider_id = p.id
                                    JOIN image_validation v ON v.image_id = i.id
                                    JOIN label l ON v.label_id = l.id
                                    WHERE (i.unlocked = true %s) AND p.name = 'donation' AND 
                                    CASE WHEN v.num_of_valid + v.num_of_invalid = 0 THEN 0 ELSE (CAST (v.num_of_valid AS float)/(v.num_of_valid + v.num_of_invalid)) END >= 0.8
                                    %s
                                    AND NOT EXISTS
                                    (
                                        SELECT 1 FROM image_annotation a 
                                        WHERE a.label_id = v.label_id AND a.image_id = v.image_id AND a.auto_generated = false
                                    )
                                    %s
                                    %s
                                ) 
                            )LIMIT 1
                        ) q
                        LEFT JOIN 
                        (
                            SELECT a.label_id as label_id, a.image_id as image_id, d.annotation as annotation, t.name as annotation_type
                            FROM image_annotation a 
                            JOIN annotation_data d ON d.image_annotation_id = a.id
                            JOIN annotation_type t on d.annotation_type_id = t.id
                            WHERE a.auto_generated = true 
                        ) q1 ON q.label_id = q1.label_id AND q.image_id = q1.image_id
                        GROUP BY q.image_key, q.label, q.parent_label, 
                        q.image_width, q.image_height, q.validation_uuid, q.image_unlocked`, 
                        includeOwnImageDonations, q1, q2, q3, includeOwnImageDonations, q1, q2, q3)

    //select all images that aren't already annotated and have a label correctness probability of >= 0.8 
    var rows *sql.Rows
    var err error
    if labelId == "" {
        if username != "" {
            rows, err = db.Query(q, maxNumNotAnnotatable, username)
        } else {
            rows, err = db.Query(q, maxNumNotAnnotatable)
        } 
    } else {
        if username != "" {
            rows, err = db.Query(q, labelId, maxNumNotAnnotatable, username)
        } else {
            rows, err = db.Query(q, labelId, maxNumNotAnnotatable)
        }
    }

    if err != nil {
        log.Debug("[Get Random Un-annotated Image] Couldn't fetch result: ", err.Error())
        raven.CaptureError(err, nil)
        return unannotatedImage, err
    }

    defer rows.Close()

    var label1 string
    var label2 string
    var autoAnnotationBytes []byte
    if rows.Next() {
        unannotatedImage.Provider = "donation"

        err = rows.Scan(&unannotatedImage.Id, &label1, &label2, &unannotatedImage.Width, &unannotatedImage.Height, 
            &unannotatedImage.Validation.Id, &autoAnnotationBytes, &unannotatedImage.Unlocked)
        if err != nil {
            log.Debug("[Get Random Un-annotated Image] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return unannotatedImage, err
        }

        if addAutoAnnotations {
            if len(autoAnnotationBytes) > 0 {
                err = json.Unmarshal(autoAnnotationBytes, &unannotatedImage.AutoAnnotations)
                if err != nil {
                    log.Debug("[Get Random Un-annotated Image] Couldn't unmarshal auto annotations: ", err.Error())
                    raven.CaptureError(err, nil)
                    return unannotatedImage, err
                }
            }
        }

        if label2 == "" {
            unannotatedImage.Label = label1
            unannotatedImage.Sublabel = ""
        } else {
            unannotatedImage.Label = label2
            unannotatedImage.Sublabel = label1
        }
    }

    return unannotatedImage, nil
}

func getAnnotatedImage(apiUser datastructures.APIUser, annotationId string, autoGenerated bool, revision int32) (datastructures.AnnotatedImage, error) {
    var annotatedImage datastructures.AnnotatedImage

    includeOwnImageDonations := ""
    includeOwnImageDonationsStr := `OR (
                                        EXISTS 
                                            (
                                                SELECT 1 
                                                FROM user_image u
                                                JOIN account a ON a.id = u.account_id
                                                WHERE u.image_id = i.id AND a.name = $%d
                                            )
                                        AND NOT EXISTS 
                                            (
                                                SELECT 1 
                                                FROM image_quarantine q 
                                                WHERE q.image_id = i.id 
                                            )
                                      )` 

    q := ""
    if revision != -1 && annotationId != "" {
        
        if apiUser.Name != "" {
            includeOwnImageDonations = fmt.Sprintf(includeOwnImageDonationsStr, 3)
        }

        q = fmt.Sprintf(`SELECT q1.key, l.name, COALESCE(pl.name, ''), q1.annotation_uuid, 
                                json_agg(q.annotation || ('{"type":"' || q.annotation_type || '"}')::jsonb)::jsonb as annotations, 
                                 q1.num_of_valid, q1.num_of_invalid, q1.width, q1.height, q1.image_unlocked
                                   FROM (
                                     SELECT i.key as key, i.id as image_id, q2.label_id as label_id, 
                                     q2.id as entry_id, q2.annotation_uuid as annotation_uuid, q2.num_of_valid as num_of_valid, 
                                     q2.num_of_invalid as num_of_invalid, i.width as width, i.height as height, q2.is_revision,
                                     i.unlocked as image_unlocked
                                     FROM image i
                                     JOIN image_provider p ON i.image_provider_id = p.id
                                     JOIN (
                                        SELECT DISTINCT a.image_id as image_id, a.label_id as label_id, a.uuid as annotation_uuid,
                                        a.num_of_valid as num_of_valid, a.num_of_invalid as num_of_invalid,
                                        CASE WHEN r.revision = $1 THEN r.id ELSE a.id END as id, 
                                        CASE WHEN r.revision = $1 THEN true ELSE false END as is_revision
                                        FROM image_annotation a
                                        LEFT JOIN image_annotation_revision r ON r.image_annotation_id = a.id
                                        where a.uuid::text = $2 
                                        AND a.auto_generated = false and (r.revision = $1 or a.revision = $1)
                                     ) q2 ON q2.image_id = i.id
                                     WHERE (i.unlocked = true %s) AND p.name = 'donation'
                                     
                                     
                                   ) q1

                                   JOIN
                                   (
                                     SELECT d.annotation as annotation, t.name as annotation_type,
                                     d.image_annotation_id as image_annotation_id, d.image_annotation_revision_id as image_annotation_revision_id
                                     FROM annotation_data d 
                                     JOIN annotation_type t on d.annotation_type_id = t.id
                                   ) q ON 
                                     CASE 
                                        WHEN q1.is_revision THEN q.image_annotation_revision_id = q1.entry_id
                                        ELSE q.image_annotation_id = q1.entry_id 
                                     END


                                   JOIN label l ON q1.label_id = l.id
                                   LEFT JOIN label pl ON l.parent_id = pl.id
                                   GROUP BY q1.key, q1.annotation_uuid, l.name, pl.name, 
                                   q1.num_of_valid, q1.num_of_invalid, q1.width, q1.height, q1.image_unlocked`, includeOwnImageDonations)


    } else {
        q1 := ""
        if annotationId != "" {
            q1 = "AND a.uuid::text = $2"

            if apiUser.Name != "" {
                includeOwnImageDonations = fmt.Sprintf(includeOwnImageDonationsStr, 3)
            }

        } else {
            if apiUser.Name != "" {
                includeOwnImageDonations = fmt.Sprintf(includeOwnImageDonationsStr, 2)
            }
            
            q1 = fmt.Sprintf(`OFFSET floor(
                                            random() * 
                                            ( 
                                                SELECT count(*) FROM image i 
                                                JOIN image_provider p ON i.image_provider_id = p.id 
                                                JOIN image_annotation a ON a.image_id = i.id
                                                WHERE (i.unlocked = true %s) AND p.name = 'donation' AND a.auto_generated = $1
                                            )
                                          )
                              LIMIT 1`, includeOwnImageDonations)
        }

        q = fmt.Sprintf(`SELECT q1.key, l.name, COALESCE(pl.name, ''), q1.annotation_uuid, 
                                 json_agg(q.annotation || ('{"type":"' || q.annotation_type || '"}')::jsonb)::jsonb as annotations, 
                                 q1.num_of_valid, q1.num_of_invalid, q1.width, q1.height, q1.image_unlocked
                                   FROM (
                                     SELECT i.key as key, i.id as image_id, a.label_id as label_id, 
                                     a.id as image_annotation_id, a.uuid as annotation_uuid, a.num_of_valid as num_of_valid, 
                                     a.num_of_invalid as num_of_invalid, i.width as width, i.height as height, i.unlocked as image_unlocked
                                     FROM image i
                                     JOIN image_provider p ON i.image_provider_id = p.id
                                     JOIN image_annotation a ON a.image_id = i.id
                                     WHERE (i.unlocked = true %s) AND p.name = 'donation' AND a.auto_generated = $1
                                     %s
                                     
                                     
                                   ) q1

                                   JOIN
                                   (
                                     SELECT d.image_annotation_id as image_annotation_id, d.annotation as annotation,
                                     t.name as annotation_type
                                     FROM annotation_data d 
                                     JOIN annotation_type t on d.annotation_type_id = t.id
                                   ) q ON q.image_annotation_id = q1.image_annotation_id


                                   JOIN label l ON q1.label_id = l.id
                                   LEFT JOIN label pl ON l.parent_id = pl.id
                                   GROUP BY q1.key, q1.annotation_uuid, l.name, pl.name, 
                                   q1.num_of_valid, q1.num_of_invalid, q1.width, q1.height, q1.image_unlocked`, includeOwnImageDonations, q1)
    }

    var err error

    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Get Annotated Image] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return annotatedImage, err
    }
    
    var rows *sql.Rows

    if revision != -1 && annotationId != "" {
        if apiUser.Name == "" {
            rows, err = tx.Query(q, revision, annotationId)
        } else {
            rows, err = tx.Query(q, revision, annotationId, apiUser.Name)
        }
    } else {
        if annotationId == "" {
            if apiUser.Name == "" {
                rows, err = db.Query(q, autoGenerated)
            } else {
                rows, err = db.Query(q, autoGenerated, apiUser.Name)
            }
        } else {
            if apiUser.Name == "" {
                rows, err = db.Query(q, autoGenerated, annotationId)
            } else {
                rows, err = db.Query(q, autoGenerated, annotationId, apiUser.Name)
            }
        }
    }

    if err != nil {
        tx.Rollback()
        log.Debug("[Get Annotated Image] Couldn't get annotated image: ", err.Error())
        raven.CaptureError(err, nil)
        return annotatedImage, err
    }

    defer rows.Close()

    var label1 string
    var label2 string
    if rows.Next() {
        var annotations []byte
        annotatedImage.Image.Provider = "donation"

        err = rows.Scan(&annotatedImage.Image.Id, &label1, &label2, &annotatedImage.Id, 
                        &annotations, &annotatedImage.NumOfValid, &annotatedImage.NumOfInvalid, 
                        &annotatedImage.Image.Width, &annotatedImage.Image.Height, &annotatedImage.Image.Unlocked)
        if err != nil {
            tx.Rollback()
            log.Debug("[Get Annotated Image] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return annotatedImage, err
        }

        err := json.Unmarshal(annotations, &annotatedImage.Annotations)
        if err != nil {
            tx.Rollback()
            log.Debug("[Get Annotated Image] Couldn't unmarshal: ", err.Error())
            raven.CaptureError(err, nil)
            return annotatedImage, err
        }

        if label2 == "" {
            annotatedImage.Validation.Label = label1
            annotatedImage.Validation.Sublabel = ""
        } else {
            annotatedImage.Validation.Label = label2
            annotatedImage.Validation.Sublabel = label1
        }
    }

    if annotationId != "" {
        rows.Close()
        err := tx.QueryRow(`SELECT (SUM(CASE WHEN r.id is null THEN 0 ELSE 1 END) + 1)::integer as num 
                            FROM image_annotation a 
                            LEFT JOIN image_annotation_revision r ON r.image_annotation_id = a.id 
                            WHERE a.uuid::text = $1`, annotationId).Scan(&annotatedImage.NumRevisions)
        if err != nil {
            tx.Rollback()
            log.Debug("[Get Annotated Image] Couldn't get number of annotation revisions: ", err.Error())
            raven.CaptureError(err, nil)
            return annotatedImage, err
        }

        annotatedImage.Revision = revision
    }

    err = tx.Commit()
    if err != nil {
        log.Debug("[Get Annotated Image] Couldn't commit transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return annotatedImage, err
    }

    return annotatedImage, nil
}

func validateAnnotatedImage(clientFingerprint string, annotationId string, labelValidationEntry datastructures.LabelValidationEntry, valid bool) error {
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

func getAllUnverifiedImages(imageProvider string, shuffle bool) ([]datastructures.Image, error){
    var images []datastructures.Image

    orderRandomly := ""
    if shuffle {
        orderRandomly = "ORDER BY RANDOM()"
    }

    q1 := "WHERE q.image_id NOT IN (SELECT image_id FROM image_quarantine)"
    params := false
    if imageProvider != "" {
        params = true
        q1 = "WHERE (p.name = $1) AND q.image_id NOT IN (SELECT image_id FROM image_quarantine)"
    }

    q := fmt.Sprintf(`SELECT q.image_key, string_agg(q.label_name::text, ',') as labels, 
                      MAX(p.name) as image_provider
                      FROM 
                      (
                        SELECT i.key as image_key, l.name  as label_name, 
                        i.image_provider_id as image_provider_id, i.id as image_id
                        FROM image i  
                        LEFT JOIN image_validation v ON v.image_id = i.id
                        JOIN label l ON v.label_id = l.id
                        WHERE i.unlocked = false

                        UNION
                        
                        SELECT i.key as image_key, g.name  as label_name, 
                        i.image_provider_id as image_provider_id, i.id as image_id
                        FROM image i
                        LEFT JOIN image_label_suggestion s ON s.image_id = i.id
                        JOIN label_suggestion g ON g.id = s.label_suggestion_id
                        WHERE i.unlocked = false
                     ) q
                    JOIN image_provider p ON p.id = q.image_provider_id
                    %s
                    GROUP BY image_key
                    %s`, q1, orderRandomly)

    var err error
    var rows *sql.Rows
    if params {
        rows, err = db.Query(q, imageProvider)
    } else {
        rows, err = db.Query(q)
    }

    if err != nil {
        log.Debug("[Fetch unverified images] Couldn't fetch unverified images: ", err.Error())
        raven.CaptureError(err, nil)
        return images, err
    }

    defer rows.Close()

    for rows.Next() {
        var image datastructures.Image
        err = rows.Scan(&image.Id, &image.Label, &image.Provider)
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
        raven.CaptureError(err, nil)
        return err
    }

    return nil
}

func putImageInQuarantine(imageId string) error {
    _,err := db.Exec(`INSERT INTO image_quarantine(image_id)
                        SELECT id FROM image WHERE key = $1
                        ON CONFLICT(image_id) DO NOTHING`, imageId)
    if err != nil {
        log.Debug("[Put Image in Quarantine] Couldn't put image in quarantine: ", err.Error())
        raven.CaptureError(err, nil)
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
    } else if contributionType == "annotation-refinement" {
        _, err := db.Exec(`INSERT INTO annotation_refinements_per_country (country_code, count)
                            VALUES ($1, $2) ON CONFLICT (country_code)
                            DO UPDATE SET count = annotation_refinements_per_country.count + 1`, countryCode, 1)
        if err != nil {
            log.Debug("[Update Contributions per Country] Couldn't insert into/update annotation_refinements_per_country: ", err.Error())
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



func getImageToLabel(imageId string, username string) (datastructures.Image, error) {
    var image datastructures.Image
    var labelMeEntries []datastructures.LabelMeEntry
    image.Provider = "donation"

    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Get Image to Label] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return image, err
    }


    includeOwnImageDonations := ""
    if username != "" {
        includeOwnImageDonations = `OR (
                                        EXISTS 
                                            (
                                                SELECT 1 
                                                FROM user_image u
                                                JOIN account a ON a.id = u.account_id
                                                WHERE u.image_id = i.id AND a.name = $1
                                            )
                                        AND NOT EXISTS 
                                            (
                                                SELECT 1 
                                                FROM image_quarantine q 
                                                WHERE q.image_id = i.id 
                                            )
                                       )` 
    }


    var unlabeledRows *sql.Rows 
    if imageId == "" {
        q := fmt.Sprintf(`SELECT i.key, i.unlocked, i.width, i.height
                            FROM image i 
                            WHERE (i.unlocked = true %s)

                            AND i.id NOT IN (
                                SELECT image_id FROM image_validation
                            ) AND i.id NOT IN (
                                SELECT image_id FROM image_label_suggestion
                            ) LIMIT 1`, includeOwnImageDonations)

        if username == "" {
            unlabeledRows, err = tx.Query(q)
        } else {
            unlabeledRows, err = tx.Query(q, username)
        }

        if err != nil {
            tx.Rollback()
            raven.CaptureError(err, nil)
            log.Debug("[Get Image to Label] Couldn't get unlabeled image: ", err.Error())
            return image, err
        }

        defer unlabeledRows.Close()
    }

    if imageId != "" || !unlabeledRows.Next() {
        q1 := ""
        if imageId == "" {
            //either get a random image or image with specific id
            q1 = fmt.Sprintf(`SELECT i.id as id, i.key as key, 
                              i.unlocked as image_unlocked, i.width as image_width, i.height as image_height
                               FROM image i WHERE (i.unlocked = true %s)
                               OFFSET floor(random() * (
                                                        SELECT count(*) FROM image i WHERE (unlocked = true %s)
                                                       )
                                           ) LIMIT 1`, includeOwnImageDonations, includeOwnImageDonations)
        } else {
            paramPos := 1
            if username != "" {
                paramPos = 2
            }

            q1 = fmt.Sprintf(`SELECT i.id as id, i.key as key, 
                              i.unlocked as image_unlocked, i.width as image_width, i.height as image_height
                              FROM image i 
                              WHERE (i.unlocked = true %s) AND i.key = $%d`, includeOwnImageDonations, paramPos)
        } 

        q := fmt.Sprintf(`SELECT q.key, COALESCE(label, ''), COALESCE(parent_label, '') as parent_label, 
                          COALESCE(q1.unlocked, false) as label_unlocked, COALESCE(q1.annotatable, false) as annotatable, 
                          COALESCE(q1.label_uuid, '') as label_uuid, COALESCE(q1.validation_uuid, '') as validation_uuid, 
                          COALESCE(q1.num_of_valid, 0) as num_of_valid, COALESCE(q1.num_of_invalid, 0) as num_of_invalid, q.image_unlocked,
                          q.image_width, q.image_height
                               FROM 
                                (
                                    SELECT v.image_id as image_id, l.name as label, 
                                    COALESCE(pl.name, '') as parent_label, true as unlocked, true as annotatable,
                                    l.uuid::text as label_uuid, v.uuid::text as validation_uuid, v.num_of_valid as num_of_valid,
                                    v.num_of_invalid as num_of_invalid
                                    FROM image_validation v 
                                    JOIN label l on v.label_id = l.id 
                                    LEFT JOIN label pl on l.parent_id = pl.id

                                    UNION ALL

                                    SELECT ils.image_id as image_id, s.name as label, 
                                    '' as parent_label, false as unlocked, ils.annotatable as annotatable,
                                    '' as label_uuid, '' as validation_uuid, 0 as num_of_valid, 0 as num_of_invalid
                                    FROM image_label_suggestion ils
                                    JOIN label_suggestion s on ils.label_suggestion_id = s.id
                                ) q1
                                RIGHT JOIN (
                                    %s
                                ) q
                                ON q.id = q1.image_id`, q1)

        var rows *sql.Rows
        if imageId == "" {
            if username == ""  {
                rows, err = tx.Query(q)
            } else {
                rows, err = tx.Query(q, username)
            }
        } else {
            if username == "" {
                rows, err = tx.Query(q, imageId)
            } else {
                rows, err = tx.Query(q, username, imageId)
            }
        }

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
        var labelUnlocked bool
        var labelAnnotatable bool
        var labelUuid string
        var validationUuid string
        var numOfValid int32
        var numOfInvalid int32
        temp := make(map[string]datastructures.LabelMeEntry) 
        for rows.Next() {
            err = rows.Scan(&image.Id, &label, &parentLabel, &labelUnlocked, &labelAnnotatable, &labelUuid, 
                            &validationUuid, &numOfValid, &numOfInvalid, &image.Unlocked, &image.Width, &image.Height)
            if err != nil {
                tx.Rollback()
                raven.CaptureError(err, nil)
                log.Debug("[Get Image to Label] Couldn't scan labeled row: ", err.Error())
                return image, err
            }

            //can happen if we are selecting an image by id and that image has no labels yet
            if label == "" {
                continue
            }

            baseLabel = parentLabel
            if parentLabel == "" {
                baseLabel = label
            }

            if val, ok := temp[baseLabel]; ok {
                if parentLabel != "" {
                    var validation *datastructures.LabelMeValidation
                    validation = nil
                    if validationUuid != "" {
                        validation = &datastructures.LabelMeValidation{Uuid: validationUuid, NumOfValid: numOfValid, NumOfInvalid: numOfInvalid}
                    }

                    val.Sublabels = append(val.Sublabels, datastructures.Sublabel {Name: label, Unlocked: labelUnlocked, 
                                                                    Annotatable: labelAnnotatable, Uuid: labelUuid,
                                                                    Validation: validation})
                }
                temp[baseLabel] = val
            } else {
                var labelMeEntry datastructures.LabelMeEntry
                labelMeEntry.Label = baseLabel
                labelMeEntry.Unlocked = labelUnlocked
                labelMeEntry.Annotatable = labelAnnotatable
                labelMeEntry.Uuid = labelUuid
                labelMeEntry.Validation = &datastructures.LabelMeValidation{Uuid: validationUuid, NumOfValid: numOfValid, NumOfInvalid: numOfInvalid}
                if parentLabel != "" {
                    var validation *datastructures.LabelMeValidation
                    validation = nil
                    if validationUuid != "" {
                        validation = &datastructures.LabelMeValidation{Uuid: validationUuid, NumOfValid: numOfValid, NumOfInvalid: numOfInvalid}
                    }


                    labelMeEntry.Sublabels = append(labelMeEntry.Sublabels, datastructures.Sublabel {Name: label, Unlocked: labelUnlocked, 
                                                                                      Annotatable: labelAnnotatable, Uuid: labelUuid,
                                                                                      Validation: validation})
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
        err = unlabeledRows.Scan(&image.Id, &image.Unlocked, &image.Width, &image.Height)
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

func addLabelsToImage(apiUser datastructures.APIUser, labelMap map[string]datastructures.LabelMapEntry, 
                        imageId string, labels []datastructures.LabelMeEntry) error {
    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Adding image labels] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    var knownLabels []datastructures.LabelMeEntry
    for _, item := range labels {
        if !isLabelValid(labelMap, item.Label, item.Sublabels) { //if its a label that is not known to us
            if apiUser.Name != "" { //and request is coming from a authenticated user, add it to the label suggestions
                err := _addLabelSuggestionToImage(apiUser, item.Label, imageId, item.Annotatable, tx)
                if err != nil {
                    return err //tx already rolled back in case of error, so we can just return here 
                }
            } else {
                tx.Rollback()
                log.Debug("you need to be authenticated")
                return errors.New("you need to be authenticated to perform this action") 
            }
        } else {
            knownLabels = append(knownLabels, item)
        }
    }

    if len(knownLabels) > 0 {
        _, err = _addLabelsToImage(apiUser.ClientFingerprint, imageId, knownLabels, 0, 0, tx)
        if err != nil { 
            return err //tx already rolled back in case of error, so we can just return here 
        }
    }

    
    err = tx.Commit()
    if err != nil {
        log.Debug("[Adding image labels] Couldn't commit changes: ", err.Error())
        raven.CaptureError(err, nil)
        return err 
    }
    return err
}

func _addLabelSuggestionToImage(apiUser datastructures.APIUser, label string, imageId string, annotatable bool, tx *sql.Tx) error {
    var labelSuggestionId int64

    labelSuggestionId = -1
    rows, err := tx.Query("SELECT id FROM label_suggestion WHERE name = $1", label)
    if err != nil {
        tx.Rollback()
        log.Debug("[Adding suggestion label] Couldn't get label: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    if !rows.Next() { //label does not exist yet, insert it
        rows.Close()

        err := tx.QueryRow(`INSERT INTO label_suggestion(name, proposed_by) 
                            SELECT $1, id FROM account a WHERE a.name = $2 
                            ON CONFLICT (name) DO NOTHING RETURNING id`, label, apiUser.Name).Scan(&labelSuggestionId)
        if err != nil {
            tx.Rollback()
            log.Debug("[Adding suggestion label] Couldn't add label: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }
    } else {
        err = rows.Scan(&labelSuggestionId)
        rows.Close()
        if err != nil {
            tx.Rollback()
            log.Debug("[Adding suggestion label] Couldn't scan label: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }
    }

    _, err = tx.Exec(`INSERT INTO image_label_suggestion (fingerprint_of_last_modification, image_id, label_suggestion_id, annotatable) 
                        SELECT $1, id, $3, $4 FROM image WHERE key = $2
                        ON CONFLICT(image_id, label_suggestion_id) DO NOTHING`, apiUser.ClientFingerprint, imageId, labelSuggestionId, annotatable)
    if err != nil {
        tx.Rollback()
        log.Debug("[Adding image label suggestion] Couldn't add image label suggestion: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    return nil
}

func _addLabelsToImage(clientFingerprint string, imageId string, labels []datastructures.LabelMeEntry, numOfValid int, numOfNotAnnotatable int, tx *sql.Tx) ([]int64, error) {
    var insertedIds []int64
    for _, item := range labels {
        rows, err := tx.Query(`SELECT i.id FROM image i WHERE i.key = $1`, imageId)
        if err != nil {
            tx.Rollback()
            log.Debug("[Adding image labels] Couldn't get image ", err.Error())
            raven.CaptureError(err, nil)
            return insertedIds, err
        }

        defer rows.Close()

        var imageId int64
        if rows.Next() {
            err = rows.Scan(&imageId)
            if err != nil {
                tx.Rollback()
                log.Debug("[Adding image labels] Couldn't scan image image entry: ", err.Error())
                raven.CaptureError(err, nil)
                return insertedIds, err
            }
        }

        rows.Close()

        //add sublabels
        if len(item.Sublabels) > 0 {
            rows, err = tx.Query(`INSERT INTO image_validation(image_id, num_of_valid, num_of_invalid, fingerprint_of_last_modification, label_id, uuid, num_of_not_annotatable) 
                                  SELECT $1, $2, $3, $4, l.id, uuid_generate_v4(), $7 FROM label l LEFT JOIN label cl ON cl.id = l.parent_id WHERE (cl.name = $5 AND l.name = ANY($6))
                                  RETURNING id`,
                                  imageId, numOfValid, 0, clientFingerprint, item.Label, pq.Array(sublabelsToStringlist(item.Sublabels)), numOfNotAnnotatable)
            if err != nil {
                tx.Rollback()
                log.Debug("[Adding image labels] Couldn't insert image validation entries for sublabels: ", err.Error())
                raven.CaptureError(err, nil)
                return insertedIds, err
            }

            for rows.Next() {
                var insertedSublabelId int64
                err = rows.Scan(&insertedSublabelId)
                if err != nil {
                    rows.Close()
                    tx.Rollback()
                    log.Debug("[Adding image labels] Couldn't scan sublabels: ", err.Error())
                    raven.CaptureError(err, nil)
                    return insertedIds, err
                }
                insertedIds = append(insertedIds, insertedSublabelId)
            }
            rows.Close()
        }

        //add base label
        var insertedLabelId int64
        err = tx.QueryRow(`INSERT INTO image_validation(image_id, num_of_valid, num_of_invalid, fingerprint_of_last_modification, 
                                                            label_id, uuid, num_of_not_annotatable) 
                              SELECT $1, $2, $3, $4, id, uuid_generate_v4(), $6 from label l WHERE id NOT IN 
                              (
                                SELECT label_id from image_validation v where image_id = $1
                              ) AND l.name = $5 AND l.parent_id IS NULL
                              RETURNING id`,
                              imageId, numOfValid, 0, clientFingerprint, item.Label, numOfNotAnnotatable).Scan(&insertedLabelId)
        if err != nil {
            if err != sql.ErrNoRows { //handle no rows gracefully (can happen if label already exists)
                pqErr := err.(*pq.Error)
                if pqErr.Code.Name() != "unique_violation" {
                    tx.Rollback()
                    log.Debug("[Adding image labels] Couldn't insert image validation entry for label: ", err.Error())
                    raven.CaptureError(err, nil)
                    return insertedIds, err
                }
            }
        } else {
            insertedIds = append(insertedIds, insertedLabelId)
        }
    }

    return insertedIds, nil
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

func getUnannotatedValidations(apiUser datastructures.APIUser, imageId string) ([]datastructures.UnannotatedValidation, error) {
    var unannotatedValidations []datastructures.UnannotatedValidation

    includeOwnImageDonations := ""
    if apiUser.Name != "" {
        includeOwnImageDonations = `OR (
                                        EXISTS 
                                        (
                                            SELECT 1 
                                            FROM user_image u
                                            JOIN account a ON a.id = u.account_id
                                            WHERE u.image_id = i.id AND a.name = $2
                                        )
                                        AND NOT EXISTS 
                                        (
                                            SELECT 1 
                                            FROM image_quarantine q 
                                            WHERE q.image_id = i.id 
                                        )
                                    )`
    }

    q := fmt.Sprintf(`SELECT v.uuid::text, l.name, COALESCE(pl.name, '') 
                             FROM image_validation v 
                             JOIN label l ON v.label_id = l.id 
                             JOIN image i ON v.image_id = i.id
                             LEFT JOIN label pl on l.parent_id = pl.id
                             WHERE i.key = $1 AND (i.unlocked = true %s) AND NOT exists (
                                SELECT 1 FROM image_annotation a WHERE
                                a.image_id = i.id AND a.label_id = l.id
                             )`, includeOwnImageDonations)
    var rows *sql.Rows
    var err error

    if apiUser.Name == "" {
        rows, err = db.Query(q, imageId)
    } else {
        rows, err = db.Query(q, imageId, apiUser.Name)
    }
    
    if err != nil {
        log.Debug("[Get unannotated validation ids] Couldn't get validation ids: ", err.Error())
        raven.CaptureError(err, nil)
        return unannotatedValidations, err
    }

    defer rows.Close()

    for rows.Next() {
        var unannotatedValidation datastructures.UnannotatedValidation
        err = rows.Scan(&unannotatedValidation.Validation.Id, &unannotatedValidation.Validation.Label, 
                            &unannotatedValidation.Validation.Sublabel)
        if err != nil {
            log.Debug("[Get unannotated validation ids] Couldn't scan rows: ", err.Error())
            raven.CaptureError(err, nil)
            return unannotatedValidations, err
        }

        unannotatedValidations = append(unannotatedValidations, unannotatedValidation)
    }

    return unannotatedValidations, nil
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

func getRandomAnnotationForQuizRefinement() (datastructures.AnnotationRefinement, error) {
    var bytes []byte
    var annotationBytes []byte
    var refinement datastructures.AnnotationRefinement
    var annotations []json.RawMessage
    rows, err := db.Query(`SELECT i.key, s.quiz_question_id, s.quiz_question, s.quiz_answers, s1.annotations, s.recommended_control::text, 
                            s1.uuid, s.allow_unknown, s.allow_other, s.browse_by_example, s.multiselect
                            FROM ( 
                                    SELECT qq.question as quiz_question, qq.recommended_control as recommended_control,
                                    json_agg(json_build_object('uuid', l.uuid, 'label', l.name, 'examples', COALESCE(s2.examples, '[]'))) as quiz_answers, 
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
                                    SELECT a.uuid, a.label_id, a.image_id, json_agg(d.annotation || ('{"uuid":"' || d.uuid || '"}')::jsonb || ('{"type":"'||t.name||'"}')::jsonb)::jsonb as annotations 
                                    FROM image_annotation a
                                    JOIN image i ON i.id = a.image_id
                                    JOIN annotation_data d ON d.image_annotation_id = a.id
                                    JOIN annotation_type t ON d.annotation_type_id = t.id
                                    WHERE CASE WHEN a.num_of_valid + a.num_of_invalid = 0 THEN 0 ELSE (CAST (a.num_of_valid AS float)/(a.num_of_valid + a.num_of_invalid)) END >= 0.8
                                    AND i.unlocked = true
                                    GROUP BY a.label_id, a.image_id, a.uuid
                                 ) as s1
                            ON s1.label_id =  s.refines_label_id 
                            JOIN image i ON i.id = s1.image_id
                            OFFSET floor(random() * 
                                ( SELECT count(*) FROM image_annotation a 
                                  JOIN quiz_question q ON q.refines_label_id = a.label_id
                                  WHERE CASE WHEN a.num_of_valid + a.num_of_invalid = 0 THEN 0 ELSE (CAST (a.num_of_valid AS float)/(a.num_of_valid + a.num_of_invalid)) END >= 0.8
                                )
                            ) LIMIT 1`)

    if err != nil {
        log.Debug("[Random Quiz question] Couldn't get random image quiz: ", err.Error())
        raven.CaptureError(err, nil)
        return refinement, err 
    }

    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&refinement.Image.Uuid, &refinement.Question.Uuid, 
                            &refinement.Question.Question, &bytes, &annotationBytes, &refinement.Question.RecommendedControl, 
                            &refinement.Annotation.Uuid, &refinement.Metainfo.AllowUnknown, &refinement.Metainfo.AllowOther, 
                            &refinement.Metainfo.BrowseByExample, &refinement.Metainfo.MultiSelect)

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
    }

    return refinement, nil
}

func addOrUpdateRefinements(annotationUuid string, annotationDataId string, annotationRefinementEntries []datastructures.AnnotationRefinementEntry, 
                                clientFingerprint string) error {
    var err error

    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Add or Update annotation refinement] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    for _, item := range annotationRefinementEntries {

        _, err = tx.Exec(`INSERT INTO image_annotation_refinement(annotation_data_id, label_id, num_of_valid, fingerprint_of_last_modification)
                            SELECT d.id, (SELECT l.id FROM label l WHERE l.uuid = $2), $3, $4 
                            FROM image_annotation a JOIN annotation_data d ON d.image_annotation_id = a.id WHERE a.uuid = $5 AND d.uuid = $1
                          ON CONFLICT (annotation_data_id, label_id)
                          DO UPDATE SET fingerprint_of_last_modification = $4, num_of_valid = image_annotation_refinement.num_of_valid + 1
                          WHERE image_annotation_refinement.annotation_data_id = (SELECT d.id FROM annotation_data d WHERE d.uuid = $1) 
                          AND image_annotation_refinement.label_id = (SELECT l.id FROM label l WHERE l.uuid = $2)`, 
                               annotationDataId, item.LabelId, 1, clientFingerprint, annotationUuid)
        
        if err != nil {
            tx.Rollback()
            log.Debug("[Add or Update annotation refinement] Couldn't update: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }
    }

    err = tx.Commit()
    if err != nil {
        log.Debug("[Add or Update annotation refinement] Couldn't commit transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    return nil
}


func batchAnnotationRefinement(annotationRefinementEntries []datastructures.BatchAnnotationRefinementEntry, apiUser datastructures.APIUser) error {
    var err error

    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Add or Update annotation refinement] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    for _, item := range annotationRefinementEntries {
        _, err = tx.Exec(`INSERT INTO image_annotation_refinement(annotation_data_id, label_id, num_of_valid, fingerprint_of_last_modification)
                            SELECT d.id, (SELECT l.id FROM label l WHERE l.uuid = $2), $3, $4 
                            FROM annotation_data d WHERE d.uuid = $1
                          ON CONFLICT (annotation_data_id, label_id)
                          DO UPDATE SET fingerprint_of_last_modification = $4, num_of_valid = image_annotation_refinement.num_of_valid + 1
                          WHERE image_annotation_refinement.annotation_data_id = (SELECT d.id FROM annotation_data d WHERE d.uuid = $1)`, 
                               item.AnnotationDataId, item.LabelId, 1, apiUser.ClientFingerprint)
        
        if err != nil {
            tx.Rollback()
            log.Debug("[Batch annotation refinement] Couldn't update: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }
    }

    err = tx.Commit()
    if err != nil {
        log.Debug("[Batch annotation refinement] Couldn't commit transaction: ", err.Error())
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

func getLabelCategories() ([]string, error) {
    var labels []string
    rows, err := db.Query(`SELECT pl.name 
                            FROM label l 
                            JOIN label pl on pl.id = l.parent_id
                            WHERE pl.label_type = 'refinement_category'`)
    if err != nil {
        log.Debug("[Get label categories] Couldn't get category: ", err.Error())
        raven.CaptureError(err, nil)
        return labels, err
    }
    defer rows.Close()

    var label string
    for rows.Next() {
        err = rows.Scan(&label)
        if err != nil {
           log.Debug("[Get label categories] Couldn't scan row: ", err.Error())
           raven.CaptureError(err, nil)
           return labels, err 
        }

        labels = append(labels, label)
    }

    return labels, nil
}

func getLabelAccessors() ([]string, error) {
    var labels []string
    rows, err := db.Query(`SELECT accessor FROM label_accessor`)
    if err != nil {
        log.Debug("[Get label accessors] Couldn't get accessor: ", err.Error())
        raven.CaptureError(err, nil)
        return labels, err
    }
    defer rows.Close()

    var label string
    for rows.Next() {
        err = rows.Scan(&label)
        if err != nil {
           log.Debug("[Get label accessors] Couldn't scan row: ", err.Error())
           raven.CaptureError(err, nil)
           return labels, err 
        }

        labels = append(labels, label)
    }

    return labels, nil
}

func deleteImage(uuid string) error {
    var deletedId int64

    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Delete image] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }


    deletedId = -1
    err = tx.QueryRow(`DELETE FROM image_validation
                       WHERE image_id IN (
                        SELECT id FROM image WHERE key = $1 
                       )
                       RETURNING id`, uuid).Scan(&deletedId)
    if err != nil {
        tx.Rollback()
        log.Debug("[Delete image] Couldn't delete image_validation entry: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    if deletedId == -1 {
        tx.Rollback()
        err = errors.New("nothing deleted")
        log.Debug("[Delete image] Couldn't delete image_validation entry: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    deletedId = -1
    err = tx.QueryRow(`DELETE FROM image i WHERE key = $1
                       RETURNING i.id`, uuid).Scan(&deletedId)
    if err != nil {
        tx.Rollback()
        log.Debug("[Delete image] Couldn't delete image entry: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }
    imageId := deletedId
    if deletedId == -1 {
        tx.Rollback()
        err = errors.New("nothing deleted")
        log.Debug("[Delete image] Couldn't delete image entry: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }
    

    deletedId = -1 
    err = tx.QueryRow(`DELETE FROM image_label_suggestion s 
                       WHERE image_id = $1 RETURNING s.id`, imageId).Scan(&deletedId)

    if deletedId == -1 {
        tx.Rollback()
        err = errors.New("nothing deleted")
        log.Debug("[Delete image] Couldn't delete image_label_suggestion entry: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }
    

    err = tx.Commit()
    if err != nil {
        log.Debug("[Delete image] Couldn't commit transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    return nil
}

func getImagesForAutoAnnotation(labels []string) ([]datastructures.AutoAnnotationImage, error) {
    var autoAnnotationImages []datastructures.AutoAnnotationImage
    rows, err := db.Query(`SELECT i.key, i.width, i.height, json_agg(l.name)  FROM image i 
                           JOIN image_validation v ON v.image_id = i.id
                           JOIN label l on v.label_id = l.id
                           WHERE i.id NOT IN (
                              SELECT image_id FROM image_annotation WHERE auto_generated = true
                           ) AND l.parent_id is null AND i.unlocked = true AND l.name = ANY($1)
                           GROUP BY i.key, i.width, i.height`, 
                           pq.Array(labels))
    if err != nil {
        log.Debug("[Get images for auto annotation] Couldn't get: ", err.Error())
        raven.CaptureError(err, nil)
        return autoAnnotationImages, err
    }

    defer rows.Close()

    for rows.Next() {
        var autoAnnotationImage datastructures.AutoAnnotationImage
        var data []byte
        err = rows.Scan(&autoAnnotationImage.Image.Id, &autoAnnotationImage.Image.Width, &autoAnnotationImage.Image.Height, &data)
        if err != nil {
           log.Debug("[Get images for auto annotation] Couldn't scan row: ", err.Error())
           raven.CaptureError(err, nil)
           return autoAnnotationImages, err 
        }

        err = json.Unmarshal(data, &autoAnnotationImage.Labels)
        if err != nil {
            log.Debug("[Get images for auto annotation] Couldn't unmarshal: ", err.Error())
            raven.CaptureError(err, nil)
            return autoAnnotationImages, err
        }

        autoAnnotationImages = append(autoAnnotationImages, autoAnnotationImage)
    }

    
    

    return autoAnnotationImages, nil
}

func userExists(username string) (bool, error) {
    var numOfExistingUsers int32
    err := db.QueryRow("SELECT count(*) FROM account u WHERE u.name = $1", username).Scan(&numOfExistingUsers)
    if err != nil {
        log.Debug("[User exists] Couldn't get num of existing users: ", err.Error())
        raven.CaptureError(err, nil)
        return false, err
    }

    if numOfExistingUsers > 0 {
        return true, nil
    }
    return false, nil
}

func getUserInfo(username string) (datastructures.UserInfo, error) {
    var userInfo datastructures.UserInfo
    var removeLabelPermission bool
    removeLabelPermission = false

    userInfo.Name = ""
    userInfo.Created = 0
    userInfo.ProfilePicture = ""
    userInfo.IsModerator = false

    err := db.QueryRow(`SELECT a.name, COALESCE(a.profile_picture, ''), a.created, a.is_moderator,
                        COALESCE(p.can_remove_label, false) as remove_label_permission
                        FROM account a 
                        LEFT JOIN account_permission p ON p.account_id = a.id 
                        WHERE a.name = $1`, username).Scan(&userInfo.Name, &userInfo.ProfilePicture, &userInfo.Created, 
                                                            &userInfo.IsModerator, &removeLabelPermission)
    if err != nil {
        log.Debug("[User Info] Couldn't get user info: ", err.Error())
        raven.CaptureError(err, nil)
        return userInfo, err
    }

    if userInfo.IsModerator {
        permissions := &datastructures.UserPermissions {CanRemoveLabel: removeLabelPermission}
        userInfo.Permissions = permissions
    }

    return userInfo, nil
}

func emailExists(email string) (bool, error) {
    var numOfExistingUsers int32
    err := db.QueryRow("SELECT count(*) FROM account u WHERE u.email = $1", email).Scan(&numOfExistingUsers)
    if err != nil {
        log.Debug("[Email exists] Couldn't get num of existing users: ", err.Error())
        raven.CaptureError(err, nil)
        return false, err
    }

    if numOfExistingUsers > 0 {
        return true, nil
    }
    return false, nil
}

func getHashedPasswordForUser(username string) (string, error) {
    var hashedPassword string
    err := db.QueryRow("SELECT hashed_password FROM account u WHERE u.name = $1", username).Scan(&hashedPassword)
    if err != nil {
        log.Debug("[Hashed Password] Couldn't get hashed password for user: ", err.Error())
        raven.CaptureError(err, nil)
        return "", err
    }

    return hashedPassword, nil
}


func addAccessToken(username string, accessToken string, expirationTime int64) error {
    var insertedId int64

    insertedId = 0
    err := db.QueryRow(`INSERT INTO access_token(user_id, token, expiration_time)
                        SELECT id, $2, $3 FROM account a WHERE a.name = $1 RETURNING id`, username, accessToken, expirationTime).Scan(&insertedId)
    if err != nil {
        log.Debug("[Add Access Token] Couldn't add access token: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    if insertedId == 0 {
        log.Debug("[Add Access Token] Nothing inserted")
        return errors.New("Nothing inserted")
    }

    return nil
}

func removeAccessToken(accessToken string) error {
    _, err := db.Exec(`DELETE FROM access_token WHERE token = $1`, accessToken)
    if err != nil {
        log.Debug("[Remove Access Token] Couldn't remove access token: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    return nil
}

func accessTokenExists(accessToken string) bool {
    var numOfAccessTokens int32

    numOfAccessTokens = 0
    err := db.QueryRow("SELECT count(*) FROM access_token WHERE token = $1", accessToken).Scan(&numOfAccessTokens)
    if err != nil {
        log.Debug("[Add Access Token] Couldn't add access token: ", err.Error())
        raven.CaptureError(err, nil)
        return false
    }

    if numOfAccessTokens == 0 {
        return false
    }

    return true
}

func createUser(username string, hashedPassword []byte, email string) error {
    var insertedId int64

    created := int64(time.Now().Unix())

    insertedId = 0
    err := db.QueryRow(`INSERT INTO account(name, hashed_password, email, created, is_moderator) 
                        VALUES($1, $2, $3, $4, $5) RETURNING id`, 
                        username, hashedPassword, email, created, false).Scan(&insertedId)
    if err != nil {
        log.Debug("[Creating User] Couldn't create user: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    if insertedId == 0 {
        return errors.New("nothing inserted")
    }

    return nil
}

func getUserStatistics(username string) (datastructures.UserStatistics, error) {
    var userStatistics datastructures.UserStatistics

    tx, err := db.Begin()
    if err != nil {
        log.Debug("[User Statistics] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return userStatistics, err
    }

    userStatistics.Total.Annotations = 0
    err = tx.QueryRow("SELECT count(*) FROM image_annotation").Scan(&userStatistics.Total.Annotations)
    if err != nil {
        tx.Rollback()
        log.Debug("[User Statistics] Couldn't get total annotations: ", err.Error())
        raven.CaptureError(err, nil)
        return userStatistics, err
    }


    userStatistics.User.Annotations = 0
    err = tx.QueryRow(`SELECT count(*) FROM user_image_annotation u
                       JOIN account a on u.account_id = a.id WHERE a.name = $1`, username).Scan(&userStatistics.User.Annotations)
    if err != nil {
        tx.Rollback()
        log.Debug("[User Statistics] Couldn't get user annotations: ", err.Error())
        raven.CaptureError(err, nil)
        return userStatistics, err
    }


    userStatistics.Total.Validations = 0
    err = tx.QueryRow("SELECT count(*) FROM image_validation").Scan(&userStatistics.Total.Validations)
    if err != nil {
        tx.Rollback()
        log.Debug("[User Statistics] Couldn't get total validations: ", err.Error())
        raven.CaptureError(err, nil)
        return userStatistics, err
    }

    userStatistics.User.Validations = 0
    err = tx.QueryRow(`SELECT count(*) FROM user_image_validation u
                       JOIN account a on u.account_id = a.id WHERE a.name = $1`, username).Scan(&userStatistics.User.Validations)
    if err != nil {
        tx.Rollback()
        log.Debug("[User Statistics] Couldn't get user validations: ", err.Error())
        raven.CaptureError(err, nil)
        return userStatistics, err
    }


    err = tx.Commit()
    if err != nil {
        log.Debug("[User Statistics] Couldn't commit transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return userStatistics, err
    }


    return userStatistics, nil
}

func changeProfilePicture(username string, uuid string) (string, error) {
    var existingProfilePicture string

    existingProfilePicture = ""

    tx, err := db.Begin()
    if err != nil {
        log.Debug("[Change Profile Picture] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return existingProfilePicture, err
    }

    err = tx.QueryRow(`SELECT COALESCE(a.profile_picture, '') FROM account a WHERE a.name = $1`, username).Scan(&existingProfilePicture)
    if err != nil {
        log.Debug("[Change Profile Picture] Couldn't get existing profile picture: ", err.Error())
        raven.CaptureError(err, nil)
        return existingProfilePicture, err
    }

    _, err = tx.Exec(`UPDATE account SET profile_picture = $1 WHERE name = $2`, uuid, username)
    if err != nil {
        log.Debug("[Change Profile Picture] Couldn't change profile picture: ", err.Error())
        raven.CaptureError(err, nil)
        return existingProfilePicture, err
    }


    err = tx.Commit()
    if err != nil {
        log.Debug("[Change Profile Picture] Couldn't commit transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return existingProfilePicture, err
    }

    return existingProfilePicture, nil
}


func getAnnotationRefinementStatistics(period string) ([]datastructures.DataPoint, error) {
    var annotationRefinementStatistics []datastructures.DataPoint

    if period != "last-month" {
        return annotationRefinementStatistics, errors.New("Only last-month statistics are supported at the moment")
    }

    rows, err := db.Query(`WITH dates AS (
                            SELECT *
                            FROM generate_series((CURRENT_DATE - interval '1 month'), CURRENT_DATE, '1 day') date
                           ),
                           num_of_annotation_refinements AS (
                            SELECT sys_period FROM image_annotation_refinement_history h
                            UNION ALL 
                            SELECT sys_period FROM image_annotation_refinement h1
                           )
                          SELECT to_char(date(date), 'YYYY-MM-DD'),
                           ( SELECT count(*) FROM num_of_annotation_refinements s
                             WHERE date(lower(s.sys_period)) = date(date) 
                           ) as num
                           FROM dates
                           GROUP BY date
                           ORDER BY date`)
    if err != nil {
        log.Debug("[Get Annotation Refinement Statistics] Couldn't get statistics: ", err.Error())
        raven.CaptureError(err, nil)
        return annotationRefinementStatistics, err
    }

    defer rows.Close()

    for rows.Next() {
        var datapoint datastructures.DataPoint
        err = rows.Scan(&datapoint.Date, &datapoint.Value)
        if err != nil {
            log.Debug("[Get Annotation Refinement Statistics] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return annotationRefinementStatistics, err
        }

        annotationRefinementStatistics = append(annotationRefinementStatistics, datapoint)
    }

    return annotationRefinementStatistics, nil
}



func getAnnotationStatistics(period string) ([]datastructures.DataPoint, error) {
    var annotationStatistics []datastructures.DataPoint

    if period != "last-month" {
        return annotationStatistics, errors.New("Only last-month statistics are supported at the moment")
    }

    rows, err := db.Query(`WITH dates AS (
                            SELECT *
                            FROM generate_series((CURRENT_DATE - interval '1 month'), CURRENT_DATE, '1 day') date
                           ),
                           num_of_annotations AS (
                            SELECT sys_period FROM image_annotation_history h
                            UNION ALL 
                            SELECT sys_period FROM image_annotation h1
                           )
                          SELECT to_char(date(date), 'YYYY-MM-DD'),
                           ( SELECT count(*) FROM num_of_annotations s
                             WHERE date(lower(s.sys_period)) = date(date) 
                           ) as num
                           FROM dates
                           GROUP BY date
                           ORDER BY date`)
    if err != nil {
        log.Debug("[Get Statistics] Couldn't get statistics: ", err.Error())
        raven.CaptureError(err, nil)
        return annotationStatistics, err
    }

    defer rows.Close()

    for rows.Next() {
        var datapoint datastructures.DataPoint
        err = rows.Scan(&datapoint.Date, &datapoint.Value)
        if err != nil {
            log.Debug("[Get Statistics] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return annotationStatistics, err
        }

        annotationStatistics = append(annotationStatistics, datapoint)
    }

    return annotationStatistics, nil
}

func getAnnotatedStatistics(apiUser datastructures.APIUser) ([]datastructures.AnnotatedStat, error) {
    var annotatedStats []datastructures.AnnotatedStat
    var queryValues []interface{}

    includeOwnImageDonations := ""
    if apiUser.Name != "" {
        includeOwnImageDonations = `OR (
                                            EXISTS 
                                            (
                                                SELECT 1 
                                                FROM user_image u
                                                JOIN account a ON a.id = u.account_id
                                                WHERE u.image_id = i.id AND a.name = $1
                                            )
                                            AND NOT EXISTS 
                                            (
                                                SELECT 1 
                                                FROM image_quarantine q 
                                                WHERE q.image_id = i.id 
                                            )
                                        )`
        queryValues = append(queryValues, apiUser.Name)
    }



    q := fmt.Sprintf(`WITH num_validations AS (
                        SELECT v.label_id, COUNT(v.label_id) as num
                        FROM image_validation v
                        JOIN image i ON v.image_id = i.id
                        WHERE (i.unlocked = true %s)
                        GROUP BY v.label_id
                     ),
                     num_annotations AS (
                        SELECT a.label_id, COUNT(a.label_id) as num
                        FROM image_annotation a
                        JOIN image i ON a.image_id = i.id
                        WHERE (i.unlocked = true %s) AND a.auto_generated = false
                        GROUP BY a.label_id
                     )
                     SELECT l.uuid, acc.accessor, COALESCE(v.num, 0) as num_total, COALESCE(a.num, 0) as num_completed
                     FROM num_validations v
                     JOIN label_accessor acc ON acc.label_id = v.label_id
                     JOIN label l ON acc.label_id = l.id
                     LEFT JOIN num_annotations a ON a.label_id = acc.label_id
                     ORDER BY 
                        CASE 
                            WHEN v.num = 0 THEN 0
                            ELSE a.num/v.num
                        END DESC`, 
                     includeOwnImageDonations, includeOwnImageDonations)

    rows, err := db.Query(q, queryValues...)
    if err != nil {
        log.Debug("[Get Annotated Statistics] Couldn't get annotated statistics: ", err.Error())
        raven.CaptureError(err, nil)
        return annotatedStats, err
    }

    defer rows.Close()

    for rows.Next() {
        var annotatedStat datastructures.AnnotatedStat
        err = rows.Scan(&annotatedStat.Label.Id, &annotatedStat.Label.Name, &annotatedStat.Num.Total, &annotatedStat.Num.Completed)
        if err != nil {
            log.Debug("[Get Annotated Statistics] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return annotatedStats, err
        }

        annotatedStats = append(annotatedStats, annotatedStat)
    }

    return annotatedStats, nil
}


func getValidationStatistics(period string) ([]datastructures.DataPoint, error) {
    var validationStatistics []datastructures.DataPoint

    if period != "last-month" {
        return validationStatistics, errors.New("Only last-month statistics are supported at the moment")
    }

    rows, err := db.Query(`WITH dates AS (
                            SELECT *
                            FROM generate_series((CURRENT_DATE - interval '1 month'), CURRENT_DATE, '1 day') date
                           ),
                           num_of_validations AS (
                            SELECT sys_period FROM image_validation_history h
                            UNION ALL
                            SELECT sys_period FROM image_validation h1
                           )
                          SELECT to_char(date(date), 'YYYY-MM-DD'),
                           ( SELECT count(*) FROM num_of_validations s
                             WHERE date(lower(s.sys_period)) = date(date) 
                           ) as num
                           FROM dates
                           GROUP BY date
                           ORDER BY date`)
    if err != nil {
        log.Debug("[Get Statistics] Couldn't get statistics: ", err.Error())
        raven.CaptureError(err, nil)
        return validationStatistics, err
    }

    defer rows.Close()

    for rows.Next() {
        var datapoint datastructures.DataPoint
        err = rows.Scan(&datapoint.Date, &datapoint.Value)
        if err != nil {
            log.Debug("[Get Statistics] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return validationStatistics, err
        }

        validationStatistics = append(validationStatistics, datapoint)
    }

    return validationStatistics, nil
}

func getActivity(period string) ([]datastructures.Activity, error) {
    var activity []datastructures.Activity

    if period != "last-month" {
        return activity, errors.New("Only last-month statistics are supported at the moment")
    }

    rows, err := db.Query(`SELECT l.name, i.key, q.type, date(q.dt), i.width, i.height, 
                           (d.annotation || ('{"type":"' || t.name || '"}')::jsonb)::jsonb as annotation, q.activity_name 
                           FROM
                            (
                                (
                                    (
                                        SELECT label_id, image_id, 'created' as type, lower(a.sys_period) as dt, 
                                        a.id as annotation_id, 'annotation' as activity_name
                                        FROM image_annotation a 
                                        WHERE id NOT IN ( SELECT id FROM image_annotation_history h
                                                          WHERE h.label_id = a.label_id and a.image_id = h.image_id
                                                        )
                                        AND 
                                        (
                                                date(lower(a.sys_period)) <= CURRENT_DATE 
                                                AND 
                                                date(lower(a.sys_period)) >= (CURRENT_DATE - interval '1 month')
                                        )
                                    )

                                    UNION

                                    (
                                        SELECT label_id, image_id, 'created' as type, lower(h.sys_period) as dt, 
                                        h.id as annotation_id, 'annotation' as activity_name
                                        FROM image_annotation_history h
                                        WHERE h.num_of_valid = 0 AND h.num_of_invalid = 0
                                        AND 
                                        (
                                            date(upper(h.sys_period)) <= CURRENT_DATE
                                            AND 
                                            date(upper(h.sys_period)) >= (CURRENT_DATE - interval '1 month')
                                        )
                                    )

                                    UNION ALL

                                    (
                                        SELECT a.label_id, a.image_id, 'verified' as type, upper(h.sys_period) as dt, 
                                        h.id as annotation_id, 'annotation' as activity_name
                                        FROM image_annotation_history h
                                        JOIN image_annotation a 
                                        ON a.id = h.id
                                        AND 
                                        (
                                            date(upper(h.sys_period)) <= CURRENT_DATE 
                                            AND 
                                            date(upper(h.sys_period)) >= (CURRENT_DATE - interval '1 month')
                                        )
                                    )
                                )


                                UNION ALL
                                (
                                    (
                                        SELECT label_id, image_id, 'created' as type, lower(v.sys_period) as dt, 
                                        null::bigint as annotation_id, 'validation' as activity_name
                                        FROM image_validation v 
                                        WHERE id NOT IN ( SELECT id FROM image_validation_history h
                                                          WHERE h.label_id = v.label_id and v.image_id = h.image_id
                                                        )
                                        AND 
                                        (
                                                date(lower(v.sys_period)) <= CURRENT_DATE 
                                                AND 
                                                date(lower(v.sys_period)) >= (CURRENT_DATE - interval '1 month')
                                        )
                                    )

                                    UNION

                                    (
                                        SELECT label_id, image_id, 'created' as type, lower(h.sys_period) as dt, 
                                        null::bigint as annotation_id, 'validation' as activity_name
                                        FROM image_validation_history h
                                        WHERE h.num_of_valid = 0 AND h.num_of_invalid = 0
                                        AND 
                                        (
                                            date(upper(h.sys_period)) <= CURRENT_DATE
                                            AND 
                                            date(upper(h.sys_period)) >= (CURRENT_DATE - interval '1 month')
                                        )
                                    )

                                    UNION ALL

                                    (
                                        SELECT v.label_id, v.image_id, 'verified' as type, upper(h.sys_period) as dt, 
                                        null::bigint as annotation_id, 'validation' as activity_name
                                        FROM image_validation_history h
                                        JOIN image_validation v 
                                        ON v.id = h.id
                                        AND 
                                        (
                                            date(upper(h.sys_period)) <= CURRENT_DATE 
                                            AND 
                                            date(upper(h.sys_period)) >= (CURRENT_DATE - interval '1 month')
                                        )
                                    )
                                )
                            ) q
                            JOIN label l ON q.label_id = l.id
                            JOIN image i ON q.image_id = i.id
                            LEFT JOIN annotation_data d ON q.annotation_id = d.image_annotation_id
                            LEFT JOIN annotation_type t ON d.annotation_type_id = t.id
                            WHERE i.unlocked = true
                            order by dt desc`)
    if err != nil {
        log.Debug("[Get Activity] Couldn't get activity: ", err.Error())
        raven.CaptureError(err, nil)
        return activity, err
    }

    defer rows.Close()

    for rows.Next() {
        var a datastructures.Activity
        var annotation []byte
        err = rows.Scan(&a.Image.Label, &a.Image.Id, &a.Type, &a.Date, &a.Image.Width, &a.Image.Height, &annotation, &a.Name)
        if err != nil {
            log.Debug("[Get Activity] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return activity, err
        }

        if len(annotation) > 0 {
            err := json.Unmarshal(annotation, &a.Image.Annotation)
            if err != nil {
                log.Debug("[Get Activity] Couldn't unmarshal annotations: ", err.Error())
                raven.CaptureError(err, nil)
                return activity, err
            }
        }

        activity = append(activity, a)
    }

    return activity, nil
}

func getLabelSuggestions() ([]string, error) {
    var labelSuggestions []string

    rows, err := db.Query("SELECT name FROM label_suggestion")
    if err != nil {
        log.Debug("[Get Label Suggestions] Couldn't get label suggestions: ", err.Error())
        raven.CaptureError(err, nil)
        return labelSuggestions, err
    }

    defer rows.Close()

    for rows.Next() {
        var labelSuggestion string
        err := rows.Scan(&labelSuggestion)
        if err != nil {
            log.Debug("[Get Label Suggestions] Couldn't scan label suggestions: ", err.Error())
            raven.CaptureError(err, nil)
            return labelSuggestions, err
        }

        labelSuggestions = append(labelSuggestions, labelSuggestion)
    }

    return labelSuggestions, nil
}

func blacklistForAnnotation(validationId string, apiUser datastructures.APIUser) error {
    _, err := db.Exec(`INSERT INTO user_annotation_blacklist(image_validation_id, account_id)
                        SELECT v.id, (SELECT a.id FROM account a WHERE a.name = $1) as account_id 
                               FROM image_validation v WHERE v.uuid = $2
                        ON CONFLICT DO NOTHING`, apiUser.Name, validationId)
    if err != nil {
        log.Debug("[Blacklist Annotation] Couldn't blacklist annotation: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }
    return nil
}

func markValidationAsNotAnnotatable(validationId string) error {
    _, err := db.Exec(`UPDATE image_validation SET num_of_not_annotatable = num_of_not_annotatable + 1 
                       WHERE uuid = $1`, validationId)
    if err != nil {
        log.Debug("[Mark Validation as not annotatable] Couldn't mark validation as not-annotatable: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    return nil
}

func isImageUnlocked(uuid string) (bool, error) {
    var unlocked bool
    unlocked = false
    rows, err := db.Query("SELECT unlocked FROM image WHERE key = $1", uuid)
    if err != nil {
        log.Debug("[Is Image Unlocked] Couldn't get row: ", err.Error())
        raven.CaptureError(err, nil)
        return false, err
    }

    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&unlocked)
        if err != nil {
            log.Debug("[Is Image Unlocked] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return false, err
        }
    }

    return unlocked, nil
}

func getApiTokens(username string) ([]datastructures.APIToken, error) {
    var apiTokens []datastructures.APIToken
    rows, err := db.Query(`SELECT token, issued_at, description, revoked 
                           FROM api_token a
                           JOIN account a1 ON a1.id = a.account_id
                           WHERE a1.name = $1`, username)
    if err != nil {
        log.Debug("[Get API Tokens] Couldn't get rows: ", err.Error())
        raven.CaptureError(err, nil)
        return apiTokens, err
    }

    defer rows.Close() 

    for rows.Next() {
        var apiToken datastructures.APIToken
        err = rows.Scan(&apiToken.Token, &apiToken.IssuedAtUnixTimestamp, &apiToken.Description, &apiToken.Revoked)
        if err != nil {
            log.Debug("[Get API Tokens] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return apiTokens, err
        }

        apiTokens = append(apiTokens, apiToken)
    }

    return apiTokens, nil
}

func isApiTokenRevoked(token string) (bool, error) {
    var revoked bool
    err := db.QueryRow("SELECT revoked FROM api_token WHERE token = $1", token).Scan(&revoked)
    if err != nil {
        log.Debug("[Is API Token revoked] Couldn't scan row: ", err.Error())
        raven.CaptureError(err, nil)
        return false, err
    }

    return revoked, nil
}

func generateApiToken(username string, description string) (datastructures.APIToken, error) {
    type MyCustomClaims struct {
        Username string `json:"username"`
        Created int64 `json:"created"`
        jwt.StandardClaims
    }

    var apiToken datastructures.APIToken

    issuedAt := time.Now()
    expiresAt := issuedAt.Add(time.Hour * 24 * 365 * 10) //10 years

    claims := MyCustomClaims {
                  username,
                  issuedAt.Unix(),
                  jwt.StandardClaims{
                        ExpiresAt: expiresAt.Unix(),
                        Issuer: "imagemonkey-api",
                  },
              }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    tokenString, err := token.SignedString([]byte(JWT_SECRET))
    if err != nil {
        return apiToken, err
    }


    _, err = db.Exec(`INSERT INTO api_token(account_id, issued_at, description, revoked, token, expires_at)
                        SELECT id, $2, $3, $4, $5, $6 FROM account WHERE name = $1`, 
                        username, issuedAt.Unix(), description, false, tokenString, expiresAt.Unix())
    if err != nil {
        log.Debug("[Generate API Token] Couldn't insert entry: ", err.Error())
        raven.CaptureError(err, nil)
        return apiToken, err
    }

    apiToken.Description = description
    apiToken.Token = tokenString
    apiToken.IssuedAtUnixTimestamp = issuedAt.Unix()

    return apiToken, nil
}

func revokeApiToken(username string, apiToken string) (bool, error) {
    var modifiedId int64
    err := db.QueryRow(`UPDATE api_token AS a 
                       SET revoked = true
                       FROM account AS acc 
                       WHERE acc.id = a.account_id AND acc.name = $1 AND a.token = $2
                       RETURNING a.id`, username, apiToken).Scan(&modifiedId)
    if err != nil {
        log.Debug("[Revoke API Token] Couldn't revoke token: ", err.Error())
        raven.CaptureError(err, nil)
        return false, err
    }

    return true, nil
}

func getAvailableAnnotationTasks(apiUser datastructures.APIUser, parseResult ParseResult, orderRandomly bool, apiBaseUrl string) ([]datastructures.AnnotationTask, error) {
    var annotationTasks []datastructures.AnnotationTask


    includeOwnImageDonations := ""
    if apiUser.Name != "" {
        includeOwnImageDonations = fmt.Sprintf(`OR (
                                                EXISTS 
                                                    (
                                                        SELECT 1 
                                                        FROM user_image u
                                                        JOIN account a ON a.id = u.account_id
                                                        WHERE u.image_id = i.id AND a.name = $%d
                                                    )
                                                AND NOT EXISTS 
                                                    (
                                                        SELECT 1 
                                                        FROM image_quarantine q 
                                                        WHERE q.image_id = i.id 
                                                    )
                                               )`, len(parseResult.queryValues) + 1)
    }

    q1 := ""
    if orderRandomly {
        q1 = " ORDER BY RANDOM()"
    }

    q2 := ""
    if apiUser.Name != "" {
        q2 = fmt.Sprintf(` AND NOT EXISTS
                           (
                                SELECT 1 FROM user_annotation_blacklist bl 
                                JOIN account acc ON acc.id = bl.account_id
                                WHERE bl.image_validation_id = v.id AND acc.name = $%d
                           )`, len(parseResult.queryValues) + 1)
    }

    q := fmt.Sprintf(`SELECT qqq.image_key, qqq.image_width, qqq.image_height, qqq.validation_uuid, qqq.image_unlocked
                      FROM
                      (
                        SELECT qq.image_key, qq.image_width, qq.image_height, 
                        unnest(string_to_array(qq.validation_uuids, ',')) as validation_uuid, qq.image_unlocked
                        FROM
                        (    
                              SELECT q.image_key, q.image_width, q.image_height, q.validation_uuids, q.image_unlocked
                              FROM
                              (   SELECT i.key as image_key, i.width as image_width, i.height as image_height, 
                                  array_to_string(array_agg(CASE WHEN (%s) THEN v.uuid ELSE NULL END), ',') as validation_uuids, 
                                  array_agg(a.accessor)::text[] as accessors, COALESCE(c.annotated_percentage, 0) as annotated_percentage, 
                                  i.unlocked as image_unlocked
                                  FROM image i 
                                  JOIN image_validation v ON v.image_id = i.id 
                                  JOIN label l ON l.id = v.label_id
                                  JOIN label_accessor a ON l.id = a.label_id
                                  LEFT JOIN image_annotation_coverage c ON c.image_id = i.id
                                  WHERE (i.unlocked = true %s)

                                  GROUP BY i.key, i.width, i.height, i.unlocked, c.annotated_percentage
                              ) q WHERE %s
                        )qq
                      ) qqq
                      JOIN image_validation v ON v.uuid::text = qqq.validation_uuid
                      WHERE NOT EXISTS (
                        SELECT 1 FROM image_annotation a 
                        WHERE a.label_id = v.label_id AND a.image_id = v.image_id
                      )%s%s`, parseResult.subquery, includeOwnImageDonations, parseResult.query, q2, q1)

    //first item in query value is the label we want to annotate
    //parseResult.queryValues = append([]interface{}{parseResult.queryValues[0]}, parseResult.queryValues...)

    var rows *sql.Rows
    var err error
    if apiUser.Name == "" {
        rows, err = db.Query(q, parseResult.queryValues...)
    } else {
        parseResult.queryValues = append(parseResult.queryValues, apiUser.Name)
        rows, err = db.Query(q, parseResult.queryValues...)
    }
    if err != nil {
        log.Debug("[Annotation Tasks] Couldn't get available annotation tasks: ", err.Error())
        raven.CaptureError(err, nil)
        return annotationTasks, err
    }

    defer rows.Close()

    for rows.Next() {
        var annotationTask datastructures.AnnotationTask
        err = rows.Scan(&annotationTask.Image.Id, &annotationTask.Image.Width, &annotationTask.Image.Height, 
                            &annotationTask.Id, &annotationTask.Image.Unlocked)
        if err != nil {
            log.Debug("[Annotation Tasks] Couldn't get available annotation tasks: ", err.Error())
            raven.CaptureError(err, nil)
            return annotationTasks, err
        }

        if annotationTask.Id == "" {
            continue
        }

        annotationTask.Image.Url = getImageUrlFromImageId(apiBaseUrl, annotationTask.Image.Id, annotationTask.Image.Unlocked)

        annotationTasks = append(annotationTasks, annotationTask)
    }

    return annotationTasks, nil
}

/*func getNumberOfImageAnnotationRevisions(annotationId string) (int32, error) {
    var numOfRevisions int32
    numOfRevisions = 0
    err := db.QueryRow(`SELECT (SUM(CASE WHEN r.id is null THEN 0 ELSE 1 END) + 1) as num 
                        FROM image_annotation a 
                        JOIN image_annotation_revision r ON r.image_annotation_id = a.id 
                        WHERE a.uuid::text = $1`, annotationId).Scan(&numOfRevisions)
    if err != nil {
        log.Debug("[Number Of Annotation Revisions] Couldn't get number of annotation revisions: ", err.Error())
        raven.CaptureError(err, nil)
        return numOfRevisions, err
    }

    return numOfRevisions, nil
}*/


func getAnnotations(apiUser datastructures.APIUser, parseResult ParseResult, 
                        imageId string, apiBaseUrl string) ([]datastructures.AnnotatedImage, error) {
    var annotatedImages []datastructures.AnnotatedImage
    var queryValues []interface{}


    q1 := ""
    if imageId == "" {
        q1 = parseResult.query
        queryValues = parseResult.queryValues
    } else {
        q1 = "q.key = $1"
        queryValues = append(queryValues, imageId)
    }

    includeOwnImageDonations := ""
    if apiUser.Name != "" {
        includeOwnImageDonations = fmt.Sprintf(`OR (
                                                EXISTS 
                                                    (
                                                        SELECT 1 
                                                        FROM user_image u
                                                        JOIN account a ON a.id = u.account_id
                                                        WHERE u.image_id = i.id AND a.name = $%d
                                                    )
                                                AND NOT EXISTS 
                                                    (
                                                        SELECT 1 
                                                        FROM image_quarantine q 
                                                        WHERE q.image_id = i.id 
                                                    )
                                               )`, len(queryValues) + 1)
        queryValues = append(queryValues, apiUser.Name)
    }

    q := fmt.Sprintf(`SELECT q1.key, l.name, COALESCE(pl.name, ''), q1.annotation_uuid, 
                             json_agg(q.annotation || ('{"type":"' || q.annotation_type || '"}')::jsonb)::jsonb as annotations, 
                             q1.num_of_valid, q1.num_of_invalid, q1.image_width, q1.image_height, q1.image_unlocked
                                   FROM (
                                     SELECT key, image_id, q.label_id, entry_id, annotation_uuid, num_of_valid,
                                     num_of_invalid, image_width, image_height, image_unlocked, annotated_percentage
                                     FROM
                                     (
                                         SELECT i.key as key, i.id as image_id, an.label_id as label_id, 
                                         an.id as entry_id, an.uuid as annotation_uuid, an.num_of_valid as num_of_valid, 
                                         an.num_of_invalid as num_of_invalid, i.width as image_width, i.height as image_height,
                                         i.unlocked as image_unlocked, qq.annotated_percentage
                                         FROM image i
                                         JOIN image_annotation an ON i.id = an.image_id
                                         JOIN image_provider p ON i.image_provider_id = p.id
                                         LEFT JOIN image_annotation_coverage qq ON qq.image_id = i.id
                                         WHERE (i.unlocked = true %s) AND p.name = 'donation'
                                         AND an.auto_generated = false
                                     ) q
                                     JOIN label_accessor a ON a.label_id = q.label_id
                                     WHERE %s
                                   ) q1

                                   JOIN
                                   (
                                     SELECT d.image_annotation_id as annotation_id, d.annotation as annotation, t.name as annotation_type 
                                     FROM annotation_data d 
                                     JOIN annotation_type t on d.annotation_type_id = t.id
                                   ) q ON q.annotation_id = q1.entry_id


                                   JOIN label l ON q1.label_id = l.id
                                   LEFT JOIN label pl ON l.parent_id = pl.id
                                   GROUP BY q1.key, q.annotation_id, q1.annotation_uuid, l.name, pl.name, 
                                   q1.num_of_valid, q1.num_of_invalid, q1.image_width, q1.image_height, q1.image_unlocked`, 
                                   includeOwnImageDonations, q1)

    rows, err := db.Query(q, queryValues...)
    if err != nil {
        log.Debug("[Get Annotated Images] Couldn't get annotations: ", err.Error())
        raven.CaptureError(err, nil)
        return annotatedImages, err
    }

    defer rows.Close()

    var label1 string
    var label2 string
    var annotations []byte
    for rows.Next() {
        var annotatedImage datastructures.AnnotatedImage
        annotatedImage.Image.Provider = "donation"

        err = rows.Scan(&annotatedImage.Image.Id, &label1, &label2, &annotatedImage.Id, 
                        &annotations, &annotatedImage.NumOfValid, &annotatedImage.NumOfInvalid, 
                        &annotatedImage.Image.Width, &annotatedImage.Image.Height, &annotatedImage.Image.Unlocked)
        if err != nil {
            log.Debug("[Get Annotated Images] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return annotatedImages, err
        }

        err := json.Unmarshal(annotations, &annotatedImage.Annotations)
        if err != nil {
            log.Debug("[Get Annotated Images] Couldn't unmarshal: ", err.Error())
            raven.CaptureError(err, nil)
            return annotatedImages, err
        }

        if label2 == "" {
            annotatedImage.Validation.Label = label1
            annotatedImage.Validation.Sublabel = ""
        } else {
            annotatedImage.Validation.Label = label2
            annotatedImage.Validation.Sublabel = label1
        }

        annotatedImage.Image.Url = getImageUrlFromImageId(apiBaseUrl, annotatedImage.Image.Id, annotatedImage.Image.Unlocked)

        annotatedImages = append(annotatedImages, annotatedImage)

    }
    return annotatedImages, nil
}

func isOwnDonation(imageId string, username string) (bool, error) {
    isOwnDonation := false
    rows, err := db.Query(`SELECT count(*)
                            FROM image i 
                            WHERE i.key = $1 AND EXISTS 
                                                        (
                                                            SELECT 1 
                                                            FROM user_image u
                                                            JOIN account a ON a.id = u.account_id
                                                            WHERE u.image_id = i.id AND a.name = $2
                                                        )
                                                        AND NOT EXISTS 
                                                        (
                                                            SELECT 1 
                                                            FROM image_quarantine q 
                                                            WHERE q.image_id = i.id 
                                                        )`, imageId, username)
    if err != nil {
        log.Debug("[Is Own Donation] Couldn't retrieve information: ", err.Error())
        raven.CaptureError(err, nil)
        return isOwnDonation, err
    }

    defer rows.Close()

    for rows.Next() {
        var num int32
        err = rows.Scan(&num)
        if err != nil {
            log.Debug("[Is Own Donation] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return isOwnDonation, err
        }

        if num > 0 {
            isOwnDonation = true
        }
    }

    return isOwnDonation, nil
}

func getImagesLabels(apiUser datastructures.APIUser, parseResult ParseResult, apiBaseUrl string, shuffle bool) ([]datastructures.ImageLabel, error) {
    var imageLabels []datastructures.ImageLabel

    shuffleStr := ""
    if shuffle {
        shuffleStr = "ORDER BY RANDOM()"
    }

    includeOwnImageDonations := ""
    if apiUser.Name != "" {
        includeOwnImageDonations = fmt.Sprintf(`OR (
                                                EXISTS 
                                                    (
                                                        SELECT 1 
                                                        FROM user_image u
                                                        JOIN account a ON a.id = u.account_id
                                                        WHERE u.image_id = i.id AND a.name = $%d
                                                    )
                                                AND NOT EXISTS 
                                                    (
                                                        SELECT 1 
                                                        FROM image_quarantine q 
                                                        WHERE q.image_id = i.id 
                                                    )
                                               )`, len(parseResult.queryValues) + 1)
    }

    q := fmt.Sprintf(`WITH 
                        image_productive_labels AS (
                                           SELECT i.id as image_id, a.accessor as accessor, a.label_id as label_id
                                                                FROM image_validation v 
                                                                JOIN label_accessor a ON v.label_id = a.label_id
                                                                JOIN image i ON v.image_id = i.id
                                                                WHERE (i.unlocked = true %s)
                        ),image_trending_labels AS (

                                                            SELECT i.id as image_id, s.name as label
                                                                FROM image_label_suggestion ils
                                                                JOIN label_suggestion s on ils.label_suggestion_id = s.id
                                                                JOIN image i ON ils.image_id = i.id
                                                                WHERE (i.unlocked = true %s)
                        ),
                        image_ids AS (
                            SELECT image_id, annotated_percentage, image_width, image_height, image_key, image_unlocked
                            FROM
                            (
                                SELECT image_id, accessors, annotated_percentage, i.width as image_width, i.height as image_height, 
                                i.unlocked as image_unlocked, i.key as image_key
                                FROM
                                (
                                    SELECT q1.image_id, array_agg(label)::text[] as accessors, 
                                    COALESCE(c.annotated_percentage, 0) as annotated_percentage
                                    FROM 
                                    (
                                        SELECT image_id, accessor as label
                                        FROM image_productive_labels p 

                                        UNION ALL

                                        SELECT image_id, label as label
                                        FROM image_trending_labels t
                                    ) q1
                                    LEFT JOIN image_annotation_coverage c ON c.image_id = q1.image_id
                                    GROUP BY q1.image_id, c.annotated_percentage
                                ) q2
                                JOIN image i ON i.id = q2.image_id
                            ) q
                            WHERE %s
                        )


                        SELECT image_key, image_width, image_height, image_unlocked,
                        json_agg(json_build_object('name', q4.label, 'num_yes', q4.num_of_valid, 'num_no', q4.num_of_invalid, 'sublabels', q4.sublabels))
                        FROM
                        (
                            SELECT q3.image_id, q3.label, q3.num_of_valid, q3.num_of_invalid,
                            coalesce(json_agg(json_build_object('name', q3.sublabel, 'num_yes', q3.num_of_valid, 'num_no', q3.num_of_invalid)) 
                                                FILTER (WHERE q3.sublabel is not null), '[]'::json) as sublabels,
                            image_key, image_width, image_height, image_unlocked
                            FROM
                            (
                                SELECT ii.image_id, CASE WHEN pl.name is not null then pl.name else l.name end as label, 
                                       COALESCE(CASE WHEN l.parent_id is not null then l.name else null end, null) as sublabel,
                                       v.num_of_valid as num_of_valid, v.num_of_invalid as num_of_invalid, 
                                       ii.image_key as image_key, ii.image_width as image_width, ii.image_height as image_height,
                                       ii.image_unlocked as image_unlocked
                                FROM
                                image_ids ii
                                JOIN image_productive_labels p on p.image_id = ii.image_id
                                JOIN label l on l.id = p.label_id
                                LEFT JOIN label pl on pl.id = l.parent_id
                                JOIN image_validation v ON ii.image_id = v.image_id AND v.label_id = l.id

                                UNION ALL

                                SELECT ii.image_id, s.name as label, null as sublabel, 
                                0 as num_of_valid, 0 as num_of_invalid,
                                ii.image_key as image_key, ii.image_width as image_width, ii.image_height as image_height,
                                ii.image_unlocked as image_unlocked
                                FROM image_ids ii
                                JOIN image_label_suggestion ils on ii.image_id = ils.image_id
                                JOIN label_suggestion s on ils.label_suggestion_id = s.id
                            ) q3
                            GROUP BY image_id, image_key, image_width, image_height, image_unlocked, label, num_of_valid, num_of_invalid
                        ) q4
                        GROUP BY image_key, image_width, image_height, image_unlocked
                        %s`, 
                      includeOwnImageDonations, includeOwnImageDonations, parseResult.query, shuffleStr)

    var rows *sql.Rows
    if apiUser.Name != "" {
        parseResult.queryValues = append(parseResult.queryValues, apiUser.Name)
    }

    rows, err := db.Query(q, parseResult.queryValues...)
    if err != nil {
        log.Debug("[Get Image Labels] Couldn't get image labels: ", err.Error())
        raven.CaptureError(err, nil)
        return imageLabels, err
    }

    defer rows.Close()


    for rows.Next() {
        var imageLabel datastructures.ImageLabel
        var labels []byte
        err = rows.Scan(&imageLabel.Image.Id, &imageLabel.Image.Width, &imageLabel.Image.Height, &imageLabel.Image.Unlocked, &labels)
        if err != nil {
            log.Debug("[Get Image Labels] Couldn't scan rows: ", err.Error())
            raven.CaptureError(err, nil)
            return imageLabels, err
        }

        err := json.Unmarshal(labels, &imageLabel.Labels)
        if err != nil {
            log.Debug("[Export] Couldn't unmarshal image labels: ", err.Error())
            raven.CaptureError(err, nil)
            return nil, err
        }

        imageLabel.Image.Url = getImageUrlFromImageId(apiBaseUrl, imageLabel.Image.Id, imageLabel.Image.Unlocked)

        imageLabels = append(imageLabels, imageLabel)
    }

    return imageLabels, nil
}

func getAnnotationsForRefinement(parseResult ParseResult, apiBaseUrl string, 
        annotationDataId string) ([]datastructures.AnnotationRefinementTask, error) {
    var annotationRefinementTasks []datastructures.AnnotationRefinementTask

    q1 := ""
    if annotationDataId != "" {
        q1 = fmt.Sprintf("WHERE d.uuid::text = $%d", len(parseResult.queryValues) + 1)
    }

    q2 := ""
    if len(parseResult.queryValues) > 0 {
        q2 = fmt.Sprintf("WHERE %s", parseResult.query)
    }

    q := fmt.Sprintf(`WITH 
                        productive_image_annotation_data_entries AS (
                            SELECT q.annotation_data_id, array_agg(q.label)::text[] as accessors
                            FROM (
                                    SELECT d.id as annotation_data_id, an.image_id as image_id, a.accessor as label 
                                    FROM image_annotation an
                                    JOIN annotation_data d on d.image_annotation_id = an.id
                                    JOIN label_accessor a on a.label_id = an.label_id
                                    WHERE an.auto_generated = false

                                    UNION ALL
                                    
                                    SELECT d.id as annotation_data_id, an.image_id as image_id, a.accessor as accessor
                                    FROM image_annotation an
                                    JOIN annotation_data d on d.image_annotation_id = an.id
                                    JOIN image_annotation_refinement r on r.annotation_data_id = d.id
                                    JOIN label_accessor a on a.label_id = r.label_id
                                    JOIN label l on l.id = r.label_id
                                    LEFT JOIN label pl ON l.parent_id = pl.id
                                    WHERE pl.label_type = 'refinement_category' AND an.auto_generated = false

                                    UNION ALL

                                    SELECT d.id as annotation_data_id, an.image_id as image_id, pl.name as accessor
                                    FROM image_annotation an
                                    JOIN annotation_data d on d.image_annotation_id = an.id
                                    JOIN image_annotation_refinement r on r.annotation_data_id = d.id
                                    JOIN label l on l.id = r.label_id
                                    LEFT JOIN label pl ON l.parent_id = pl.id
                                    WHERE pl.label_type = 'refinement_category' AND an.auto_generated = false
                                ) q
                                JOIN image i ON q.image_id = i.id
                                WHERE i.unlocked = true
                                GROUP BY q.annotation_data_id
                        ), 
                        filtered_image_annotation_data_entries AS (
                            SELECT annotation_data_id 
                            FROM productive_image_annotation_data_entries q
                            %s
                        )
                        SELECT i.key, i.unlocked, i.width, i.height, a.uuid,
                        (d.annotation || ('{"uuid":"' || d.uuid || '"}')::jsonb || ('{"type":"' || t.name || '"}')::jsonb)::jsonb as annotation,
                        COALESCE(json_agg(json_build_object('name', l.name, 'uuid', l.uuid)) FILTER (WHERE l.id is not null), '[]'::json) as labels
                        FROM filtered_image_annotation_data_entries f
                        JOIN annotation_data d ON d.id = f.annotation_data_id
                        JOIN image_annotation a ON a.id = d.image_annotation_id
                        JOIN image i ON i.id = a.image_id
                        JOIN annotation_type t ON d.annotation_type_id = t.id
                        LEFT JOIN image_annotation_refinement r ON r.annotation_data_id = d.id
                        LEFT JOIN label l ON l.id = r.label_id
                        %s
                        GROUP BY i.key, i.unlocked, i.width, i.height, a.uuid, d.annotation, d.uuid, t.name`, q2, q1)

    if annotationDataId != "" {
        parseResult.queryValues = append(parseResult.queryValues, annotationDataId)
    }

    rows, err := db.Query(q, parseResult.queryValues...)
    if err != nil {
        log.Debug("[Get Annotations For Refinement] Couldn't get annotations for refinement: ", err.Error())
        raven.CaptureError(err, nil)
        return annotationRefinementTasks, err
    }

    defer rows.Close()

    for rows.Next() {
        var annotationBytes []byte
        var labelAccessorsBytes []byte
        var annotationRefinementTask datastructures.AnnotationRefinementTask
        rows.Scan(&annotationRefinementTask.Image.Id, &annotationRefinementTask.Image.Unlocked, 
                    &annotationRefinementTask.Image.Width, &annotationRefinementTask.Image.Height, 
                    &annotationRefinementTask.Annotation.Id, &annotationBytes, &labelAccessorsBytes)

        err = json.Unmarshal(annotationBytes, &annotationRefinementTask.Annotation.Data)
        if err != nil {
            log.Debug("[Get Annotations For Refinement] Couldn't unmarshal annotation: ", err.Error())
            raven.CaptureError(err, nil)
            return annotationRefinementTasks, err
        }

        err = json.Unmarshal(labelAccessorsBytes, &annotationRefinementTask.Refinements)
        if err != nil {
            log.Debug("[Get Annotations For Refinement] Couldn't unmarshal labels: ", err.Error())
            raven.CaptureError(err, nil)
            return annotationRefinementTasks, err
        }


        annotationRefinementTask.Image.Url = getImageUrlFromImageId(apiBaseUrl, annotationRefinementTask.Image.Id, 
                                                                        annotationRefinementTask.Image.Unlocked)

        annotationRefinementTasks = append(annotationRefinementTasks, annotationRefinementTask)
    }

    return annotationRefinementTasks, nil
}

func getAnnotationCoverage(imageId string) ([]datastructures.ImageAnnotationCoverage, error) {
    var imageAnnotationCoverages []datastructures.ImageAnnotationCoverage

    q1 := ""
    var queryValues []interface{}
    if imageId != "" {
        q1 = "WHERE i.key = $1"
        queryValues = append(queryValues, imageId)
    }

    q := fmt.Sprintf(`SELECT i.key, i.width, i.height, c.annotated_percentage
                      FROM image_annotation_coverage c
                      JOIN image i ON c.image_id = i.id
                      %s`, q1)

    rows, err := db.Query(q, queryValues...)
    if err != nil {
        log.Debug("[Get annotation coverage] Couldn't get annotation coverage: ", err.Error())
        raven.CaptureError(err, nil)
        return imageAnnotationCoverages, err
    }

    defer rows.Close()

    for rows.Next() {
        var imageAnnotationCoverage datastructures.ImageAnnotationCoverage

        err = rows.Scan(&imageAnnotationCoverage.Image.Id, &imageAnnotationCoverage.Image.Width, 
                            &imageAnnotationCoverage.Image.Height, &imageAnnotationCoverage.Coverage)
        if err != nil {
            log.Debug("[Get annotation coverage] Couldn't scan rows: ", err.Error())
            raven.CaptureError(err, nil)
            return imageAnnotationCoverages, err
        }

        imageAnnotationCoverages = append(imageAnnotationCoverages, imageAnnotationCoverage)
    }

    return imageAnnotationCoverages, nil
}