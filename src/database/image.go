package imagemonkeydb

import (
	"encoding/json"
	"errors"
	"fmt"
	commons "github.com/bbernhard/imagemonkey-core/commons"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	parser "github.com/bbernhard/imagemonkey-core/parser/v2"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
	"time"
	"context"
	"github.com/jackc/pgx/v4"
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

func _addImageSource(imageId int64, imageSource datastructures.ImageSource, tx pgx.Tx) (int64, error) {
	var insertedId int64
	err := tx.QueryRow(context.TODO(), "INSERT INTO image_source(image_id, url) VALUES($1, $2) RETURNING id", imageId, imageSource.Url).Scan(&insertedId)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Debug("[Add image source] Couldn't add image source: ", err.Error())
		raven.CaptureError(err, nil)
		return insertedId, err
	}

	return insertedId, nil
}

//returns a list of n - random images (n = limit) that were uploaded with the given label.
func (p *ImageMonkeyDatabase) GetRandomGroupedImages(label string, limit int) ([]datastructures.ValidationImage, error) {
	var images []datastructures.ValidationImage

	tx, err := p.db.Begin(context.TODO())
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
	err = tx.QueryRow(context.TODO(),
                       `SELECT count(*) FROM image i 
                        JOIN image_provider p ON i.image_provider_id = p.id 
                        JOIN image_validation v ON v.image_id = i.id
                        JOIN label l ON v.label_id = l.id
                        WHERE i.unlocked = true AND p.name = 'donation' AND l.name = $1 AND l.parent_id is null`, label).Scan(&numOfRows)
	if err != nil {
		tx.Rollback(context.TODO())
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
	rows, err := p.db.Query(context.TODO(),
                          `SELECT i.key, l.name, v.num_of_valid, v.num_of_invalid, v.uuid FROM image i 
                           JOIN image_provider p ON i.image_provider_id = p.id 
                           JOIN image_validation v ON v.image_id = i.id
                           JOIN label l ON v.label_id = l.id
                           WHERE i.unlocked = true AND p.name = 'donation' AND l.name = $1 AND l.parent_id is null
                           OFFSET $2 LIMIT $3`, label, randomNumber, limit)

	if err != nil {
		tx.Rollback(context.TODO())
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
			tx.Rollback(context.TODO())
			log.Debug("[Fetch random grouped image] Couldn't scan row: ", err.Error())
			raven.CaptureError(err, nil)
			return images, err
		}

		images = append(images, image)
	}

	return images, tx.Commit(context.TODO())
}

func (p *ImageMonkeyDatabase) UnlockImage(imageId string) error {
	_, err := p.db.Exec(context.TODO(), "UPDATE image SET unlocked = true WHERE key = $1", imageId)
	if err != nil {
		log.Debug("[Unlock Image] Couldn't unlock image: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	return nil
}

func (p *ImageMonkeyDatabase) PutImageInQuarantine(imageId string) error {
	_, err := p.db.Exec(context.TODO(), `INSERT INTO image_quarantine(image_id)
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
	rows, err := p.db.Query(context.TODO(), "SELECT unlocked FROM image WHERE key = $1", uuid)
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
	rows, err := p.db.Query(context.TODO(), "SELECT COUNT(hash) FROM image where hash = $1", int64(hash))
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
	} else {
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
	err := p.db.QueryRow(context.TODO(), q, queryValues...).Scan(&num)
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
	metalabels *commons.MetaLabels) error {
	tx, err := p.db.Begin(context.TODO())
	if err != nil {
		log.Debug("[Adding donated photo] Couldn't begin transaction: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	imageProvider := imageInfo.Source.Provider
	if imageProvider == "imagehunt" {
		imageProvider = "donation"
	}

	//PostgreSQL can't store unsigned 64bit, so we are casting the hash to a signed 64bit value when storing the hash (so values above maxuint64/2 are negative).
	//this should be ok, as we do not need to order those values, but just need to check if a hash exists. So it should be fine
	var imageId int64
	err = tx.QueryRow(context.TODO(),
                       `INSERT INTO image(key, unlocked, image_provider_id, hash, width, height, sys_period) 
                        SELECT $1, $2, p.id, $3, $5, $6, '["now()",]'::tstzrange FROM image_provider p WHERE p.name = $4 
                        RETURNING id`,
		imageInfo.Name, autoUnlock, int64(imageInfo.Hash), imageProvider, imageInfo.Width, imageInfo.Height).Scan(&imageId)
	if err != nil {
		log.Debug("[Adding donated photo] Couldn't insert image: ", err.Error())
		raven.CaptureError(err, nil)
		tx.Rollback(context.TODO())
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

		insertedValidationIds, err = _addLabelsAndLabelSuggestionsToImageInTransaction(tx, apiUser, labelMap, metalabels, imageInfo.Name, labels, numOfValid, 0)
		if err != nil {
			return err //tx already rolled back in case of error, so we can just return here
		}
	}

	if imageProvider != "donation" {
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
	if apiUser.Name != "" {
		_, err := tx.Exec(context.TODO(),
                           `INSERT INTO user_image(image_id, account_id)
                            SELECT $1, id FROM account WHERE name = $2`, imageId, apiUser.Name)
		if err != nil {
			tx.Rollback(context.TODO())
			log.Debug("[Add user image entry] Couldn't add entry: ", err.Error())
			raven.CaptureError(err, nil)
			return err
		}
	}

	//if user is logged in
	if apiUser.Name != "" {
		//add image to default image collection
		err = p._addImageToImageCollectionInTransaction(tx, apiUser.Name, MyDonations, imageInfo.Name, true)
		if err != nil {
			//transaction already rolled back
			log.Error("[Add donated Image To Collection] Couldn't add image to default collection: ", err.Error())
			raven.CaptureError(err, nil)
			return err
		}

		//if image collection is provided
		if imageCollectionName != "" {
			err = p._addImageToImageCollectionInTransaction(tx, apiUser.Name, imageCollectionName, imageInfo.Name, true)
			if err != nil {
				//transaction already rolled back
				log.Error("[Add donated Image To Collection] Couldn't add image to collection: ", err.Error())
				raven.CaptureError(err, nil)
				return err
			}
		}
	}

	if imageInfo.Source.Provider == "imagehunt" {
		if len(insertedValidationIds) != 1 {
			tx.Rollback(context.TODO())
			err = errors.New("Couldn't create imagehunt entry due to missing or invalid label")
			log.Error("[Create ImageHunt entry for donated image]", err.Error())
			raven.CaptureError(err, nil)
			return err
		}

		_, err := tx.Exec(context.TODO(),
                           `INSERT INTO imagehunt_task(image_validation_id, created)
                            VALUES($1, $2)`, insertedValidationIds[0], time.Now().Unix())
		if err != nil {
			tx.Rollback(context.TODO())
			log.Error("[Create ImageHunt entry for donated image] Couldn't create entry: ", err.Error())
			raven.CaptureError(err, nil)
			return err
		}
	}

	err = tx.Commit(context.TODO())
	if err != nil {
		log.Error("[Add donated Image] Couldn't commit transaction: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	return nil
}

func (p *ImageMonkeyDatabase) IsOwnDonation(imageId string, username string) (bool, error) {
	isOwnDonation := false
	rows, err := p.db.Query(context.TODO(),
                           `SELECT count(*)
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

func (p *ImageMonkeyDatabase) ReportImage(imageId string, reason string) error {
	insertedId := 0
	err := p.db.QueryRow(context.TODO(), "INSERT INTO image_report(image_id, reason) SELECT i.id, $2 FROM image i WHERE i.key = $1 RETURNING id",
		imageId, reason).Scan(&insertedId)
	if err != nil {
		log.Debug("[Report image] Couldn't add report: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	return nil
}

func (p *ImageMonkeyDatabase) GetAllUnverifiedImages(imageProvider string, shuffle bool, limit int) (datastructures.LockedImages, error) {
	var lockedImages datastructures.LockedImages
	var queryValues []interface{}

	orderRandomly := ""
	if shuffle {
		orderRandomly = "ORDER BY RANDOM()"
	}

	limitBy := ""
	if limit != -1 {
		limitBy = fmt.Sprintf("LIMIT $%d", len(queryValues)+1)
		queryValues = append(queryValues, limit)
	}

	q1 := "WHERE q.image_id NOT IN (SELECT image_id FROM image_quarantine)"
	if imageProvider != "" {
		q1 = fmt.Sprintf("WHERE (p.name = $%d) AND q.image_id NOT IN (SELECT image_id FROM image_quarantine)", len(queryValues)+1)
		queryValues = append(queryValues, imageProvider)
	}

	q := fmt.Sprintf(`SELECT q.image_key, q.image_width, q.image_height, COALESCE(string_agg(q.label_name::text, ','), '') as labels, 
                      p.name as image_provider
                      FROM 
                      (
                        SELECT i.key as image_key, i.width as image_width, i.height as image_height, 
                        l.name  as label_name, i.image_provider_id as image_provider_id, i.id as image_id
                        FROM image i  
                        LEFT JOIN image_validation v ON v.image_id = i.id
                        LEFT JOIN label l ON v.label_id = l.id
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
	var rows pgx.Rows

	tx, err := p.db.Begin(context.TODO())
	if err != nil {
		log.Debug("[Fetch unverified images] Couldn't begin transaction: ", err.Error())
		raven.CaptureError(err, nil)
		return lockedImages, err
	}

	rows, err = tx.Query(context.TODO(), q, queryValues...)

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

	err = tx.QueryRow(context.TODO(), totalImagesQuery).Scan(&lockedImages.Total)
	if err != nil {
		log.Debug("[Fetch unverified images] Couldn't get number of images: ", err.Error())
		raven.CaptureError(err, nil)
		return lockedImages, err
	}

	err = tx.Commit(context.TODO())
	if err != nil {
		log.Debug("[Fetch unverified images] Couldn't commit transaction: ", err.Error())
		raven.CaptureError(err, nil)
		return lockedImages, err
	}

	return lockedImages, nil
}

func (p *ImageMonkeyDatabase) DeleteImage(uuid string) error {
	var imageId int64

	tx, err := p.db.Begin(context.TODO())
	if err != nil {
		log.Debug("[Delete image] Couldn't begin transaction: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	_, err = tx.Exec(context.TODO(),
                     `DELETE FROM user_image
                      WHERE image_id IN (
                        SELECT id FROM image WHERE key = $1 
                      )`, uuid)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Debug("[Delete image] Couldn't delete user_image entry: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	_, err = tx.Exec(context.TODO(),
                     `DELETE FROM image_validation
                      WHERE image_id IN (
                        SELECT id FROM image WHERE key = $1 
                      )`, uuid)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Debug("[Delete image] Couldn't delete image_validation entry: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	imageId = -1
	err = tx.QueryRow(context.TODO(),
                      `DELETE FROM image i WHERE key = $1
                       RETURNING i.id`, uuid).Scan(&imageId)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Debug("[Delete image] Couldn't delete image entry: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	if imageId == -1 {
		tx.Rollback(context.TODO())
		err = errors.New("nothing deleted")
		log.Debug("[Delete image] Couldn't delete image entry: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	_, err = tx.Exec(context.TODO(),
                      `DELETE FROM image_label_suggestion s 
                       WHERE image_id = $1`, imageId)

	if err != nil {
		tx.Rollback(context.TODO())
		log.Debug("[Delete image] Couldn't delete image_label_suggestion entry: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	err = tx.Commit(context.TODO())
	if err != nil {
		log.Debug("[Delete image] Couldn't commit transaction: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	return nil
}

func (p *ImageMonkeyDatabase) Export(parseResult parser.ParseResult, annotationsOnly bool) ([]datastructures.ExportedImage, error) {
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

	q := fmt.Sprintf(`WITH image_annotation_refinements as (
						SELECT an.image_id as image_id, (d.annotation || ('{"label":"' || %s || '"}')::jsonb || ('{"type":"' || t.name || '"}')::jsonb)::jsonb as annotations, %s as label
                            FROM image_annotation_refinement r 
                            JOIN annotation_data d ON r.annotation_data_id = d.id
                            JOIN annotation_type t ON d.annotation_type_id = t.id
                            JOIN image_annotation an ON d.image_annotation_id = an.id
                            %s
                            WHERE an.auto_generated = false
					  ), image_annotations as (
					 	SELECT n.image_id as image_id, (d.annotation || ('{"label":"' || %s || '"}')::jsonb || ('{"type":"' || t.name || '"}')::jsonb)::jsonb as annotations, %s as label
                            FROM image_annotation n
                            JOIN annotation_data d ON d.image_annotation_id = n.id
                            JOIN annotation_type t ON d.annotation_type_id = t.id
                            %s
                            WHERE n.auto_generated = false 
					  ), image_validations as (
					  	SELECT i.id as image_id, json_agg(json_build_object('label', %s, 'num_yes', num_of_valid, 'num_no', num_of_invalid))::jsonb as validations, array_agg(%s) as accessors
                            FROM image i 
                            JOIN image_validation v ON i.id = v.image_id
                            %s
                            GROUP BY i.id
					  ), unlocked_images as (
					  	SELECT i.id as image_id, i.key as image_uuid, i.width as image_width, i.height as image_height
						 FROM image i 
						 WHERE i.unlocked = true
					  ), filtered_annotations_and_refinements as (
					    SELECT image_id, annotations, accessors 
						FROM (
							SELECT r.image_id as image_id, r.annotations as annotations, array_agg(label) as accessors
							FROM image_annotation_refinements r
							GROUP BY r.image_id, r.annotations
							
							UNION
							
							SELECT a.image_id as image_id, a.annotations as annotations, array_agg(label) as accessors
							FROM image_annotations a
							GROUP BY a.image_id, a.annotations
						) q
						WHERE %s
					  ), filtered_image_validations as (
					  	SELECT image_id, validations, accessors 
						FROM image_validations q
						WHERE %s
					  )

					  SELECT i.image_uuid, CASE WHEN json_agg(q2.annotations)::jsonb = '[null]'::jsonb THEN '[]' ELSE json_agg(q2.annotations)::jsonb END as annotations
					  	FROM unlocked_images i
						JOIN
						(
						  SELECT COALESCE(a.image_id, v.image_id) as image_id, a.annotations, v.validations 
							FROM filtered_annotations_and_refinements a
							%s
							filtered_image_validations v
							ON a.image_id = v.image_id
						) q2
						ON q2.image_id = i.image_id
					    GROUP BY i.image_uuid, q2.validations, i.image_width, i.image_height`, identifier, identifier, q1, identifier, identifier, q2, identifier, identifier, q3, parseResult.Query, parseResult.Query, joinType)
	rows, err := p.db.Query(context.TODO(), q, parseResult.QueryValues...)
	if err != nil {
		log.Error("Couldn't export data: ", err.Error())
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

func (p *ImageMonkeyDatabase) GetImageDetails(imageId string) (datastructures.ImageDetails, error) {
	rows, err := p.db.Query(context.TODO(), `SELECT i.unlocked, i.width, i.height
												FROM image i WHERE i.key = $1`, imageId)
	var imageDetails datastructures.ImageDetails
	if err != nil {
		log.Error("[Image Info] Couldn't query image info: ", err.Error())
		raven.CaptureError(err, nil)
		return imageDetails, err
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&imageDetails.Unlocked, &imageDetails.Width, &imageDetails.Height)
		if err != nil {
			log.Error("[Image Info] Couldn't scan row: ", err.Error())
			raven.CaptureError(err, nil)
			return imageDetails, err
		}
	} else {
		return imageDetails, &NotFoundError{Description: "No image with that id found"}
	}
	return imageDetails, nil
}
