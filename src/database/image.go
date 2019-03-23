package imagemonkeydb

import (
    "github.com/getsentry/raven-go"
    log "github.com/Sirupsen/logrus"
    "../datastructures"
    "database/sql"
    commons "../commons"
    parser "../parser/v2"
    "fmt"
    "errors"
    "encoding/json"
    "github.com/lib/pq"
    "time"
)

type ImageDonationErrorType int
const (
  ImageDonationSuccess ImageDonationErrorType = 1 << iota
  ImageDonationImageCollectionDoesntExistError
  ImageDonationInternalError
)

func sublabelsToStringlist(sublabels []datastructures.Sublabel) []string {
    var s []string
    for _, sublabel := range sublabels {
        s = append(s, sublabel.Name)
    }

    return s
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

//returns a list of n - random images (n = limit) that were uploaded with the given label. 
func (p *ImageMonkeyDatabase) GetRandomGroupedImages(label string, limit int) ([]datastructures.ValidationImage, error) {
    var images []datastructures.ValidationImage

    tx, err := p.db.Begin()
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
        randomNumber = commons.Random(0, end)
    }

    //fetch images
    rows, err := p.db.Query(`SELECT i.key, l.name, v.num_of_valid, v.num_of_invalid, v.uuid FROM image i 
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


func (p *ImageMonkeyDatabase) UnlockImage(imageId string) error {
    _,err := p.db.Exec("UPDATE image SET unlocked = true WHERE key = $1", imageId)
    if err != nil {
        log.Debug("[Unlock Image] Couldn't unlock image: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    return nil
}

func (p *ImageMonkeyDatabase) PutImageInQuarantine(imageId string) error {
    _,err := p.db.Exec(`INSERT INTO image_quarantine(image_id)
                        SELECT id FROM image WHERE key = $1
                        ON CONFLICT(image_id) DO NOTHING`, imageId)
    if err != nil {
        log.Debug("[Put Image in Quarantine] Couldn't put image in quarantine: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    return nil
}

func (p *ImageMonkeyDatabase) IsImageUnlocked(uuid string) (bool, error) {
    var unlocked bool
    unlocked = false
    rows, err := p.db.Query("SELECT unlocked FROM image WHERE key = $1", uuid)
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

func (p *ImageMonkeyDatabase) ImageExists(hash uint64) (bool, error) {
    //PostgreSQL can't handle unsigned 64bit, so we are casting the hash to a signed 64bit value when comparing against the stored hash (so values above maxuint64/2 are negative). 
    rows, err := p.db.Query("SELECT COUNT(hash) FROM image where hash = $1", int64(hash))
    if err != nil {
        log.Debug("[Checking if photo exists] Couldn't get hash: ", err.Error())
        raven.CaptureError(err, nil)
        return false, err
    }
    defer rows.Close()

    var numOfOccurences int
    if rows.Next() {
        err = rows.Scan(&numOfOccurences)
        if err != nil {
            log.Debug("[Checking if photo exists] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return false, err
        }
    }

    if numOfOccurences > 0 {
        return true, nil
    } else{
        return false, nil
    }
}

func (p *ImageMonkeyDatabase) ImageExistsForUser(imageId string, username string) (bool, error) {
    var queryValues []interface{}
    queryValues = append(queryValues, imageId)

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
        queryValues = append(queryValues, username)
        
    }

    q := fmt.Sprintf(`SELECT COUNT(i.id) FROM image i 
                      WHERE i.key = $1 AND (i.unlocked = true %s)`, includeOwnImageDonations)
    var num int = 0
    err := p.db.QueryRow(q, queryValues...).Scan(&num)
    if err != nil {
        log.Error("[Image exists for user] Couldn't determine whether image exists: ", err.Error())
        raven.CaptureError(err, nil)
        return false, err
    }

    if num > 0 {
        return true, nil
    }
    return false, nil
}

func (p *ImageMonkeyDatabase) AddDonatedPhoto(apiUser datastructures.APIUser, imageInfo datastructures.ImageInfo, autoUnlock bool, 
                                              labels []datastructures.LabelMeEntry, imageCollectionName string, labelMap map[string]datastructures.LabelMapEntry, 
                                              metalabels *commons.MetaLabels) ImageDonationErrorType {
	tx, err := p.db.Begin()
    if err != nil {
    	log.Debug("[Adding donated photo] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return ImageDonationInternalError
    }

    imageProvider := imageInfo.Source.Provider
    if imageProvider == "imagehunt" {
        imageProvider = "donation"
    }


    //PostgreSQL can't store unsigned 64bit, so we are casting the hash to a signed 64bit value when storing the hash (so values above maxuint64/2 are negative). 
    //this should be ok, as we do not need to order those values, but just need to check if a hash exists. So it should be fine
	var imageId int64 
	err = tx.QueryRow("INSERT INTO image(key, unlocked, image_provider_id, hash, width, height) SELECT $1, $2, p.id, $3, $5, $6 FROM image_provider p WHERE p.name = $4 RETURNING id", 
					  imageInfo.Name, autoUnlock, int64(imageInfo.Hash), imageProvider, imageInfo.Width, imageInfo.Height).Scan(&imageId)
	if err != nil {
		log.Debug("[Adding donated photo] Couldn't insert image: ", err.Error())
		raven.CaptureError(err, nil)
		tx.Rollback()
		return ImageDonationInternalError
	}

    var insertedValidationIds []int64
    if labels[0].Label != "" { //only create a image validation entry, if a label is provided

        //per default we start with 0 validations, except if we are importing an image from a trusted
        //source. in that case, already set "numOfValid" to 1.
        numOfValid := 0
        if imageInfo.Source.Trusted {
            numOfValid = 1
        }


        insertedValidationIds, err = _addLabelsAndLabelSuggestionsToImageInTransaction(tx, apiUser, labelMap, metalabels, imageInfo.Name, labels, numOfValid, 0)
        if err != nil {
            return ImageDonationInternalError //tx already rolled back in case of error, so we can just return here
        }
    }


    if imageProvider != "donation" {
        imageSourceId, err := _addImageSource(imageId, imageInfo.Source, tx)
        if err != nil {
            return ImageDonationInternalError //tx already rolled back in case of error, so we can just return here
        }

        err = _addImageValidationSources(imageSourceId, insertedValidationIds, tx)
        if err != nil {
            return ImageDonationInternalError //tx already rolled back in case of error, so we can just return here
        }
    }

    //in case a username is provided, link image to user account
    if apiUser.Name != "" {
        _, err := tx.Exec(`INSERT INTO user_image(image_id, account_id)
                            SELECT $1, id FROM account WHERE name = $2`, imageId, apiUser.Name)
        if err != nil {
            tx.Rollback()
            log.Debug("[Add user image entry] Couldn't add entry: ", err.Error())
            raven.CaptureError(err, nil)
            return ImageDonationInternalError
        }
    }

    if imageCollectionName != "" && apiUser.Name != "" {
        _, err := tx.Exec(`INSERT INTO image_collection_image(user_image_collection_id, image_id)
                           SELECT (SELECT u.id 
                                 FROM user_image_collection u 
                                 JOIN account a ON u.account_id = a.id
                                 WHERE u.name = $1 AND a.name = $2), $3`,
                           imageCollectionName, apiUser.Name, imageId)
        if err != nil {
            if err, ok := err.(*pq.Error); ok {
                log.Info(err.Code)
                if err.Code == "23502" {
                    return ImageDonationImageCollectionDoesntExistError
                }
            }
            tx.Rollback()
            log.Error("[Add donated Image To Collection] Couldn't add image to collection: ", err.Error())
            raven.CaptureError(err, nil)
            return ImageDonationInternalError
        }
    }

    if imageInfo.Source.Provider == "imagehunt" {
        if len(insertedValidationIds) != 1 {
            tx.Rollback()
            err = errors.New("Couldn't create imagehunt entry due to missing or invalid label")
            log.Error("[Create ImageHunt entry for donated image]", err.Error())
            raven.CaptureError(err, nil)
            return ImageDonationInternalError 
        }

        _, err := tx.Exec(`INSERT INTO imagehunt_task(image_validation_id, created)
                            VALUES($1, $2)`, insertedValidationIds[0], time.Now().Unix())
        if err != nil {
            tx.Rollback()
            log.Error("[Create ImageHunt entry for donated image] Couldn't create entry: ", err.Error())
            raven.CaptureError(err, nil)
            return ImageDonationInternalError
        }
    }

    err = tx.Commit()
    if err != nil {
        log.Error("[Add donated Image] Couldn't commit transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return ImageDonationInternalError
    }

    return ImageDonationSuccess
}


func (p *ImageMonkeyDatabase) IsOwnDonation(imageId string, username string) (bool, error) {
    isOwnDonation := false
    rows, err := p.db.Query(`SELECT count(*)
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

func (p *ImageMonkeyDatabase) ReportImage(imageId string, reason string) error{
	insertedId := 0
	err := p.db.QueryRow("INSERT INTO image_report(image_id, reason) SELECT i.id, $2 FROM image i WHERE i.key = $1 RETURNING id", 
					  imageId, reason).Scan(&insertedId)
	if err != nil {
		log.Debug("[Report image] Couldn't add report: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	return nil
}

func (p *ImageMonkeyDatabase) GetAllUnverifiedImages(imageProvider string, shuffle bool, limit int) (datastructures.LockedImages, error){
    var lockedImages datastructures.LockedImages
    var queryValues []interface{}

    orderRandomly := ""
    if shuffle {
        orderRandomly = "ORDER BY RANDOM()"
    }

    limitBy := ""
    if limit != -1 {
        limitBy = fmt.Sprintf("LIMIT $%d", len(queryValues) + 1)
        queryValues = append(queryValues, limit)
    }

    q1 := "WHERE q.image_id NOT IN (SELECT image_id FROM image_quarantine)"
    if imageProvider != "" {
        q1 = fmt.Sprintf("WHERE (p.name = $%d) AND q.image_id NOT IN (SELECT image_id FROM image_quarantine)", len(queryValues) + 1)
        queryValues = append(queryValues, imageProvider)
    }

    q := fmt.Sprintf(`SELECT q.image_key, q.image_width, q.image_height, string_agg(q.label_name::text, ',') as labels, 
                      p.name as image_provider
                      FROM 
                      (
                        SELECT i.key as image_key, i.width as image_width, i.height as image_height, 
                        l.name  as label_name, i.image_provider_id as image_provider_id, i.id as image_id
                        FROM image i  
                        LEFT JOIN image_validation v ON v.image_id = i.id
                        JOIN label l ON v.label_id = l.id
                        WHERE i.unlocked = false

                        UNION
                        
                        SELECT i.key as image_key, i.width as image_width, i.height as image_height,
                        g.name  as label_name, i.image_provider_id as image_provider_id, i.id as image_id
                        FROM image i
                        LEFT JOIN image_label_suggestion s ON s.image_id = i.id
                        JOIN label_suggestion g ON g.id = s.label_suggestion_id
                        WHERE i.unlocked = false
                     ) q
                    JOIN image_provider p ON p.id = q.image_provider_id
                    %s
                    GROUP BY image_key, image_width, image_height, p.name
                    %s
                    %s`, q1, orderRandomly, limitBy)


    totalImagesQuery := fmt.Sprintf(`SELECT count(*) 
                                     FROM 
                                     ( SELECT i.id as image_id, i.image_provider_id as image_provider_id,
                                       i.unlocked as unlocked
                                       FROM image i
                                     ) q
                                     JOIN image_provider p ON p.id = q.image_provider_id 
                                     %s AND q.unlocked = false`, q1)

    var err error
    var rows *sql.Rows
    
    tx, err := p.db.Begin()
    if err != nil {
        log.Debug("[Fetch unverified images] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return lockedImages, err
    }

    rows, err = tx.Query(q, queryValues...)

    if err != nil {
        log.Debug("[Fetch unverified images] Couldn't fetch unverified images: ", err.Error())
        raven.CaptureError(err, nil)
        return lockedImages, err
    }

    defer rows.Close()

    for rows.Next() {
        var image datastructures.LockedImage
        err = rows.Scan(&image.Id, &image.Width, &image.Height, &image.Labels, &image.Provider)
        if err != nil {
            log.Debug("[Fetch unverified images] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return lockedImages, err
        }

        lockedImages.Images = append(lockedImages.Images, image)
    }


    err = tx.QueryRow(totalImagesQuery).Scan(&lockedImages.Total)
    if err != nil {
        log.Debug("[Fetch unverified images] Couldn't get number of images: ", err.Error())
        raven.CaptureError(err, nil)
        return lockedImages, err
    }

    err = tx.Commit()
    if err != nil {
        log.Debug("[Fetch unverified images] Couldn't commit transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return lockedImages, err
    }

    return lockedImages, nil
}

func (p *ImageMonkeyDatabase) DeleteImage(uuid string) error {
    var imageId int64

    tx, err := p.db.Begin()
    if err != nil {
        log.Debug("[Delete image] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }


    _, err = tx.Exec(`DELETE FROM user_image
                      WHERE image_id IN (
                        SELECT id FROM image WHERE key = $1 
                      )`, uuid)
    if err != nil {
        tx.Rollback()
        log.Debug("[Delete image] Couldn't delete user_image entry: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }


    _, err = tx.Exec(`DELETE FROM image_validation
                      WHERE image_id IN (
                        SELECT id FROM image WHERE key = $1 
                      )`, uuid)
    if err != nil {
        tx.Rollback()
        log.Debug("[Delete image] Couldn't delete image_validation entry: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }


    imageId = -1
    err = tx.QueryRow(`DELETE FROM image i WHERE key = $1
                       RETURNING i.id`, uuid).Scan(&imageId)
    if err != nil {
        tx.Rollback()
        log.Debug("[Delete image] Couldn't delete image entry: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    if imageId == -1 {
        tx.Rollback()
        err = errors.New("nothing deleted")
        log.Debug("[Delete image] Couldn't delete image entry: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    _, err = tx.Exec(`DELETE FROM image_label_suggestion s 
                       WHERE image_id = $1`, imageId)

    if err != nil {
        tx.Rollback()
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

func (p *ImageMonkeyDatabase) Export(parseResult parser.ParseResult, annotationsOnly bool) ([]datastructures.ExportedImage, error){
    joinType := "FULL OUTER JOIN"
    if annotationsOnly {
        joinType = "JOIN"
    }

    
    q1 := ""
    q2 := ""
    q3 := ""
    identifier := ""
    if parseResult.IsUuidQuery {
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
                     GROUP BY i.key, q3.validations, i.width, i.height`, identifier, q1, parseResult.Query, identifier, q2, parseResult.Query, joinType, identifier, q3, parseResult.Query)
    rows, err := p.db.Query(q, parseResult.QueryValues...)
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