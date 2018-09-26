package imagemonkeydb

import (
	"../datastructures"
	"github.com/getsentry/raven-go"
	log "github.com/Sirupsen/logrus"
	"encoding/json"
	"github.com/lib/pq"
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


func (p *ImageMonkeyDatabase) AddImageDescriptions(imageId string, descriptions []datastructures.ImageDescription) error {
	var lst []string
	for _, val := range descriptions {
		lst = append(lst, val.Description)
	}

	_, err := p.db.Query(`INSERT INTO image_description(image_id, description, num_of_valid, num_of_invalid, state, processed_by, uuid)
							SELECT (SELECT i.id FROM image i WHERE i.key = $1), unnest($2::text[]), 0, 0, 'unknown', null, uuid_generate_v4()
						  ON CONFLICT(image_id, description) DO UPDATE SET num_of_valid = image_description.num_of_valid + 1`, 
							imageId, pq.Array(lst))
	if err != nil {
        log.Error("[Adding image description] Couldn't add image description: ", err.Error())
        raven.CaptureError(err, nil)
        return err
	}

	return nil
}

func (p *ImageMonkeyDatabase) UnlockImageDescription(imageId string, descriptionId string) (UnlockImageDescriptionErrorType) {
	rows, err := p.db.Query(`UPDATE image_description AS d
							SET state = 'unlocked' 
							FROM image AS i
							WHERE i.id = d.image_id AND i.key = $1 AND uuid = $2
							RETURNING d.id`, 
							imageId, descriptionId)
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


func (p *ImageMonkeyDatabase) LockImageDescription(imageId string, descriptionId string) (LockImageDescriptionErrorType) {
	rows, err := p.db.Query(`UPDATE image_description AS d
							SET state = 'locked' 
							FROM image AS i
							WHERE i.id = d.image_id AND i.key = $1 AND uuid = $2
							RETURNING d.id`, 
							imageId, descriptionId)
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

	rows, err := p.db.Query(`SELECT i.key, json_agg(json_build_object('text', dsc.description, 'uuid', dsc.uuid))
							 FROM image_description dsc
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


