package imagemonkeydb

import (
	"context"
	"encoding/json"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	languages "github.com/bbernhard/imagemonkey-core/languages"
	"github.com/getsentry/raven-go"
	"github.com/gofrs/uuid"
	log "github.com/sirupsen/logrus"
)

type UnlockImageDescriptionErrorType int

const (
	UnlockImageDescriptionSuccess UnlockImageDescriptionErrorType = 1 << iota
	UnlockImageDescriptionInternalError
	UnlockImageDescriptionInvalidId
)

type LockImageDescriptionErrorType int

const (
	LockImageDescriptionSuccess LockImageDescriptionErrorType = 1 << iota
	LockImageDescriptionInternalError
	LockImageDescriptionInvalidId
)

type AddImageDescriptionErrorType int

const (
	AddImageDescriptionSuccess AddImageDescriptionErrorType = 1 << iota
	AddImageDescriptionInternalError
	AddImageDescriptionInvalidImageDescription
	AddImageDescriptionInvalidLanguage
)

func (p *ImageMonkeyDatabase) AddImageDescriptions(imageId string, descriptions []datastructures.ImageDescription) AddImageDescriptionErrorType {
	var imageDescriptions []string
	var langs []string
	for _, val := range descriptions {
		if val.Description == "" {
			return AddImageDescriptionInvalidImageDescription
		}
		imageDescriptions = append(imageDescriptions, val.Description)

		if val.Language == "" || !languages.IsValid(val.Language) {
			return AddImageDescriptionInvalidLanguage
		}
		langs = append(langs, val.Language)
	}

	tx, err := p.db.Begin(context.TODO())
	if err != nil {
		log.Error("[Adding image description] Couldn't begin transaction: ", err.Error())
		raven.CaptureError(err, nil)
		return AddImageDescriptionInternalError
	}

	rows, err := tx.Query(context.TODO(),
		`SELECT l.id FROM
						  (
							SELECT lang, nr FROM unnest($1::text[])
							WITH ORDINALITY As T (lang, nr) --ensure that result is correctly ordered after unnest
						  ) q
						  JOIN language l ON l.name = q.lang
						  ORDER BY nr`, langs)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Error("[Adding image description] Couldn't get languages: ", err.Error())
		raven.CaptureError(err, nil)
		return AddImageDescriptionInternalError
	}

	defer rows.Close()

	var languageIds []int64
	for rows.Next() {
		var languageId int64
		err = rows.Scan(&languageId)
		if err != nil {
			tx.Rollback(context.TODO())
			log.Error("[Adding image description] Couldn't scan language: ", err.Error())
			raven.CaptureError(err, nil)
			return AddImageDescriptionInternalError
		}

		languageIds = append(languageIds, languageId)
	}

	if len(languageIds) != len(imageDescriptions) {
		tx.Rollback(context.TODO())
		log.Error("[Adding image description] language ids mismatch: ", err.Error())
		raven.CaptureError(err, nil)
		return AddImageDescriptionInternalError
	}

	_, err = tx.Exec(context.TODO(),
		`INSERT INTO image_description(image_id, description, num_of_valid, num_of_invalid, state, processed_by, uuid, language_id)
							SELECT (SELECT i.id FROM image i WHERE i.key = $1), 
									unnest($2::text[]), 0, 0, 'unknown', null, uuid_generate_v4(), unnest($3::integer[])
						  ON CONFLICT(image_id, description) DO UPDATE SET num_of_valid = image_description.num_of_valid + 1`,
		imageId, imageDescriptions, languageIds)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Error("[Adding image description] Couldn't add image description: ", err.Error())
		raven.CaptureError(err, nil)
		return AddImageDescriptionInternalError
	}

	err = tx.Commit(context.TODO())
	if err != nil {
		tx.Rollback(context.TODO())
		log.Error("[Adding image description] Couldn't commit transaction: ", err.Error())
		raven.CaptureError(err, nil)
		return AddImageDescriptionInternalError
	}

	return AddImageDescriptionSuccess
}

func (p *ImageMonkeyDatabase) UnlockImageDescription(apiUser datastructures.APIUser, imageId string, descriptionId string) UnlockImageDescriptionErrorType {
	_, err := uuid.FromString(descriptionId)
	if err != nil {
		return UnlockImageDescriptionInvalidId
	}

	rows, err := p.db.Query(context.TODO(),
		`UPDATE image_description AS d
							SET state = 'unlocked', processed_by = (SELECT id FROM account WHERE name = $3)
							FROM image AS i
							WHERE i.id = d.image_id AND i.key = $1 AND uuid = $2
							RETURNING d.id`,
		imageId, descriptionId, apiUser.Name)
	if err != nil {
		log.Error("[Unlocking image description] Couldn't unlock image description: ", err.Error())
		raven.CaptureError(err, nil)
		return UnlockImageDescriptionInternalError
	}

	defer rows.Close()

	if rows.Next() {
		return UnlockImageDescriptionSuccess
	}

	return UnlockImageDescriptionInvalidId
}

func (p *ImageMonkeyDatabase) LockImageDescription(apiUser datastructures.APIUser, imageId string, descriptionId string) LockImageDescriptionErrorType {
	_, err := uuid.FromString(descriptionId)
	if err != nil {
		return LockImageDescriptionInvalidId
	}

	rows, err := p.db.Query(context.TODO(),
		`UPDATE image_description AS d
							SET state = 'locked', processed_by = (SELECT id FROM account WHERE name = $3)
							FROM image AS i
							WHERE i.id = d.image_id AND i.key = $1 AND uuid = $2
							RETURNING d.id`,
		imageId, descriptionId, apiUser.Name)
	if err != nil {
		log.Error("[Unlocking image description] Couldn't lock image description: ", err.Error())
		raven.CaptureError(err, nil)
		return LockImageDescriptionInternalError
	}

	defer rows.Close()

	if rows.Next() {
		return LockImageDescriptionSuccess
	}

	return LockImageDescriptionInvalidId
}

func (p *ImageMonkeyDatabase) GetUnprocessedImageDescriptions() ([]datastructures.DescriptionsPerImage, error) {
	var res []datastructures.DescriptionsPerImage

	rows, err := p.db.Query(context.TODO(),
		`SELECT i.key, json_agg(json_build_object('text', dsc.description, 'uuid', dsc.uuid, 'language', l.fullname))
							 FROM image_description dsc
							 JOIN language l ON l.id = dsc.language_id
							 JOIN image i ON i.id = dsc.image_id
							 WHERE dsc.state = 'unknown'
							 GROUP BY i.key`)

	if err != nil {
		log.Error("[Get Locked image descriptions] Couldn't get locked image descriptions: ", err.Error())
		raven.CaptureError(err, nil)
		return res, err
	}

	defer rows.Close()

	for rows.Next() {
		var descriptionsPerImage datastructures.DescriptionsPerImage
		var descriptions []byte

		err = rows.Scan(&descriptionsPerImage.Image.Id, &descriptions)
		if err != nil {
			log.Error("[Get Locked image descriptions] Couldn't get locked image descriptions: ", err.Error())
			raven.CaptureError(err, nil)
			return res, err
		}

		err := json.Unmarshal(descriptions, &descriptionsPerImage.Image.Descriptions)
		if err != nil {
			log.Error("[Get Locked image descriptions] Couldn't unmarshal descriptions: ", err.Error())
			raven.CaptureError(err, nil)
			return res, err
		}

		res = append(res, descriptionsPerImage)
	}
	return res, nil
}

func (p *ImageMonkeyDatabase) GetNumOfUnprocessedDescriptions() (int, error) {
	var num int
	num = 0

	err := p.db.QueryRow(context.TODO(), `SELECT count(*) FROM image_description WHERE state = 'unknown'`).Scan(&num)
	if err != nil {
		return num, err
	}

	return num, nil
}
