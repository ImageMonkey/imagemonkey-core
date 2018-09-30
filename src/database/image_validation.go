package imagemonkeydb

import (
    "github.com/getsentry/raven-go"
    log "github.com/Sirupsen/logrus"
    "../datastructures"
    "database/sql"
    "github.com/lib/pq"
    "errors"
    "fmt"
)

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

func (p *ImageMonkeyDatabase) ValidateImages(apiUser datastructures.APIUser, 
		imageValidationBatch datastructures.ImageValidationBatch, moderatorAction bool) error {
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


    tx, err := p.db.Begin()
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


func (p *ImageMonkeyDatabase) GetImageToValidate(imageId string, labelId string, username string) (datastructures.ValidationImage, error) {
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

    rows, err = p.db.Query(q, queryParams...)
	

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
                otherRows, err = p.db.Query(q1, labelId, username)
            } else {
                otherRows, err = p.db.Query(q1, labelId)
            }
        } else {
            if username != "" {
                otherRows, err = p.db.Query(q1, username)
            } else {
                otherRows, err = p.db.Query(q1)
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

func (p *ImageMonkeyDatabase) BlacklistForAnnotation(validationId string, apiUser datastructures.APIUser) error {
    _, err := p.db.Exec(`INSERT INTO user_annotation_blacklist(image_validation_id, account_id)
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

func (p *ImageMonkeyDatabase) MarkValidationAsNotAnnotatable(validationId string) error {
    _, err := p.db.Exec(`UPDATE image_validation SET num_of_not_annotatable = num_of_not_annotatable + 1 
                       WHERE uuid = $1`, validationId)
    if err != nil {
        log.Debug("[Mark Validation as not annotatable] Couldn't mark validation as not-annotatable: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    return nil
}

func (p *ImageMonkeyDatabase) GetUnannotatedValidations(apiUser datastructures.APIUser, imageId string) ([]datastructures.UnannotatedValidation, error) {
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
        rows, err = p.db.Query(q, imageId)
    } else {
        rows, err = p.db.Query(q, imageId, apiUser.Name)
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