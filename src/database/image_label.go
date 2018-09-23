package imagemonkeydb

import (
    "../datastructures"
    "github.com/getsentry/raven-go"
    log "github.com/Sirupsen/logrus"
    "database/sql"
    "fmt"
    "encoding/json"
)

func (p *ImageMonkeyDatabase) GetImageToLabel(imageId string, username string) (datastructures.ImageToLabel, error) {
    var image datastructures.ImageToLabel
    var labelMeEntries []datastructures.LabelMeEntry
    image.Provider = "donation"

    tx, err := p.db.Begin()
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
                          q.image_width, q.image_height, COALESCE(q2.image_descriptions, '[]')::jsonb
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
                                ) q ON q.id = q1.image_id
                                LEFT JOIN (
                                    SELECT jsonb_agg(jsonb_build_object('description', dsc.description)) as image_descriptions,
                                    i.id as image_id
                                    FROM image_description dsc
                                    JOIN image i ON i.id = dsc.image_id
                                    GROUP BY i.id, dsc.id
                                ) q2 ON q2.image_id = q1.image_id
                                `, q1)

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
        var imageDescriptions []byte
        temp := make(map[string]datastructures.LabelMeEntry) 
        for rows.Next() {
            err = rows.Scan(&image.Id, &label, &parentLabel, &labelUnlocked, &labelAnnotatable, &labelUuid, 
                            &validationUuid, &numOfValid, &numOfInvalid, &image.Unlocked, &image.Width, &image.Height,
                            &imageDescriptions)
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

            err := json.Unmarshal(imageDescriptions, &image.ImageDescriptions)
            if err != nil {
                log.Debug("[Get Image to Label] Couldn't unmarshal image descriptions: ", err.Error())
                raven.CaptureError(err, nil)
                return image, err
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