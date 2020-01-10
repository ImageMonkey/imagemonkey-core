package imagemonkeydb

import (
	"context"
	commons "github.com/bbernhard/imagemonkey-core/commons"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
	"time"
)

func (p *ImageMonkeyDatabase) GetImageHuntTasks(apiUser datastructures.APIUser, apiBaseUrl string) ([]datastructures.ImageHuntTask, error) {
	imageHuntTasks := []datastructures.ImageHuntTask{}

	rows, err := p.db.Query(context.TODO(),
		`SELECT q.image_width, q.image_height, q.image_unlocked, q.image_key, q.label_accessor, q.label
							 FROM
							 (
								SELECT i.width as image_width, i.height as image_height, i.unlocked as image_unlocked, 
									   i.key as image_key, a.accessor as label_accessor, l.name as label
									FROM label_accessor a
									JOIN image_validation v ON v.label_id = a.label_id
									JOIN imagehunt_task h ON h.image_validation_id = v.id
									JOIN label l ON l.id = a.label_id
									JOIN image i ON i.id = v.image_id
									JOIN user_image u ON u.image_id = v.image_id
									JOIN account ac ON ac.id = u.account_id
									WHERE ac.name = $1 AND
									(l.label_type = 'normal' OR l.label_type = 'meta')
									AND l.parent_id is null

								 UNION ALL 

								 SELECT 0 as image_width, 0 as image_height, false as image_unlocked, 
								 		'' as image_key, a.accessor as label_accessor, l.name as label
									FROM label_accessor a
									JOIN label l ON l.id = a.label_id
									WHERE (l.label_type = 'normal' OR l.label_type = 'meta')
									AND l.id NOT IN (
										SELECT v.label_id 
											FROM image_validation v 
											JOIN imagehunt_task h ON h.image_validation_id = v.id 
											JOIN user_image u ON u.image_id = v.image_id
											JOIN account a ON a.id = u.account_id
											WHERE a.name = $1
									)
									AND l.parent_id is null
							 ) q
							 ORDER BY CASE WHEN image_key = '' THEN 0 ELSE 1 END ASC
							`, apiUser.Name)
	if err != nil {
		log.Error("[Get ImageHunt Tasks] Couldn't get tasks: ", err.Error())
		raven.CaptureError(err, nil)
		return imageHuntTasks, err
	}

	defer rows.Close()

	for rows.Next() {
		var imageHuntTask datastructures.ImageHuntTask
		var imageHuntTaskImage datastructures.ImageHuntTaskImage

		err = rows.Scan(&imageHuntTaskImage.Width, &imageHuntTaskImage.Height, &imageHuntTaskImage.Unlocked,
			&imageHuntTaskImage.Id, &imageHuntTask.Label.Accessor, &imageHuntTask.Label.Name)
		if err != nil {
			log.Error("[Get ImageHunt Tasks] Couldn't scan tasks: ", err.Error())
			raven.CaptureError(err, nil)
			return imageHuntTasks, err
		}

		if imageHuntTaskImage.Id != "" {
			imageHuntTaskImage.Url = commons.GetImageUrlFromImageId(apiBaseUrl, imageHuntTaskImage.Id, imageHuntTaskImage.Unlocked)
			imageHuntTask.Image = &imageHuntTaskImage
		} else {
			imageHuntTask.Image = nil
		}

		imageHuntTasks = append(imageHuntTasks, imageHuntTask)
	}

	return imageHuntTasks, nil
}

func isValidationValid(numOfValid int, numOfInvalid int) bool {
	isValid := false
	numOfTotalValidations := numOfValid + numOfInvalid
	if numOfTotalValidations == 0 {
		isValid = true
	} else {
		ratio := float64(numOfValid) / float64(numOfTotalValidations)
		if ratio > 0.5 {
			isValid = true
		}
	}
	return isValid
}

func (p *ImageMonkeyDatabase) GetImageHuntStats(apiUser datastructures.APIUser, apiBaseUrl string,
	numOfAvailableLabels int, utcOffset int64) (datastructures.ImageHuntStats, error) {
	var imageHuntStats datastructures.ImageHuntStats

	rows, err := p.db.Query(context.TODO(),
		`SELECT count(*) 
							FROM imagehunt_task h
							JOIN image_validation v ON v.id = h.image_validation_id
							JOIN user_image u ON u.image_id = v.image_id
							JOIN account a ON a.id = u.account_id
							WHERE a.name = $1`, apiUser.Name)

	if err != nil {
		log.Error("[Get ImageHunt Stats] Couldn't get stats: ", err.Error())
		raven.CaptureError(err, nil)
		return imageHuntStats, err
	}

	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&imageHuntStats.Stars)
		if err != nil {
			log.Error("[Get ImageHunt Stats] Couldn't scan row: ", err.Error())
			raven.CaptureError(err, nil)
			return imageHuntStats, err
		}
	}

	rows.Close()

	rows, err = p.db.Query(context.TODO(),
		`SELECT h.created, v.num_of_valid, v.num_of_invalid
								FROM image_validation v 
								JOIN imagehunt_task h ON h.image_validation_id = v.id 
								JOIN user_image u ON u.image_id = v.image_id
								JOIN account a ON a.id = u.account_id
								WHERE a.name = $1
								ORDER BY h.created ASC`, apiUser.Name)

	if err != nil {
		log.Error("[Get ImageHunt Stats] Couldn't get detailed stats: ", err.Error())
		raven.CaptureError(err, nil)
		return imageHuntStats, err
	}

	defer rows.Close()

	achievementsGenerator := commons.NewAchievementsGenerator()

	for rows.Next() {
		var created int64
		var numOfValid int
		var numOfInvalid int
		err = rows.Scan(&created, &numOfValid, &numOfInvalid)
		if err != nil {
			log.Error("[Get ImageHunt Stats] Couldn't scan detailed row: ", err.Error())
			raven.CaptureError(err, nil)
			return imageHuntStats, err
		}

		t := time.Unix(created, 0)             //unix timestamp -> time
		t.Add(time.Duration(utcOffset*10 ^ 9)) //add utc offset (in ns)
		isValid := isValidationValid(numOfValid, numOfInvalid)

		if isValid {
			achievementsGenerator.Add(t)

		}

	}

	imageHuntStats.Achievements, err = achievementsGenerator.GetAchievements(apiBaseUrl)
	if err != nil {
		log.Error("[Get ImageHunt Stats] Couldn't get achievements: ", err.Error())
		raven.CaptureError(err, nil)
		return imageHuntStats, err
	}

	return imageHuntStats, nil
}
