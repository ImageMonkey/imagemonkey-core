package imagemonkeydb

import (
    "../datastructures"
    "github.com/getsentry/raven-go"
    log "github.com/Sirupsen/logrus"
    "database/sql"
    "fmt"
    "encoding/json"
    "github.com/lib/pq"
    commons "../commons"
    parser "../parser"
    "errors"
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
                                    SELECT jsonb_agg(jsonb_build_object('text', dsc.description, 'state', dsc.state::text, 'language', l.fullname)) as image_descriptions,
                                    i.id as image_id
                                    FROM image_description dsc
                                    JOIN language l ON l.id = dsc.language_id
                                    JOIN image i ON i.id = dsc.image_id
                                    WHERE dsc.state != 'locked' --only show non locked image descriptions
                                    GROUP BY i.id
                                ) q2 ON q2.image_id = q1.image_id
                                ORDER BY parent_label ASC NULLS FIRST -- return base labels first
                                                                      -- otherwise, the below logic won't work correctly
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


func (p *ImageMonkeyDatabase) AddLabelsToImage(apiUser datastructures.APIUser, labelMap map[string]datastructures.LabelMapEntry, 
                                                metalabels *commons.MetaLabels, imageId string, labels []datastructures.LabelMeEntry) error {
    tx, err := p.db.Begin()
    if err != nil {
        log.Debug("[Adding image labels] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    var knownLabels []datastructures.LabelMeEntry
    for _, item := range labels {
        if !commons.IsLabelValid(labelMap, metalabels, item.Label, item.Sublabels) { //if its a label that is not known to us
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
        _, err = AddLabelsToImageInTransaction(apiUser.ClientFingerprint, imageId, knownLabels, 0, 0, tx)
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

func AddLabelsToImageInTransaction(clientFingerprint string, imageId string, labels []datastructures.LabelMeEntry, 
        numOfValid int, numOfNotAnnotatable int, tx *sql.Tx) ([]int64, error) {
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
                                  ON CONFLICT DO NOTHING
                                  RETURNING id`,
                                  imageId, numOfValid, 0, clientFingerprint, item.Label, pq.Array(sublabelsToStringlist(item.Sublabels)), numOfNotAnnotatable)
            if err != nil {
                if err != sql.ErrNoRows { //handle no rows gracefully (can happen if label already exists)
                    pqErr := err.(*pq.Error)
                    if pqErr.Code.Name() != "unique_violation" {
                        tx.Rollback()
                        log.Debug("[Adding image labels] Couldn't insert image validation entries for sublabels: ", err.Error())
                        raven.CaptureError(err, nil)
                        return insertedIds, err
                    }
                }
            } else {
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
        }

        //add base label
        var insertedLabelId int64
        err = tx.QueryRow(`INSERT INTO image_validation(image_id, num_of_valid, num_of_invalid, fingerprint_of_last_modification, 
                                                            label_id, uuid, num_of_not_annotatable) 
                              SELECT $1, $2, $3, $4, id, uuid_generate_v4(), $6 from label l WHERE id NOT IN 
                              (
                                SELECT label_id from image_validation v where image_id = $1
                              ) AND l.name = $5 AND l.parent_id IS NULL
                              ON CONFLICT DO NOTHING
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

func (p *ImageMonkeyDatabase) GetAllImageLabels() ([]string, error) {
    var labels []string

    rows, err := p.db.Query(`SELECT l.name FROM label l`)
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

func (p *ImageMonkeyDatabase) GetImagesLabels(apiUser datastructures.APIUser, parseResult parser.ParseResult, 
        apiBaseUrl string, shuffle bool) ([]datastructures.ImageLabel, error) {
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
                                               )`, len(parseResult.QueryValues) + 1)
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
                        ),
                        img_descriptions AS (
                            SELECT i.id as image_id, 
                            jsonb_agg(jsonb_build_object('text', dsc.description, 'state', dsc.state::text, 'language', l.fullname)) as descriptions
                            FROM image i
                            JOIN image_description dsc ON dsc.image_id = i.id
                            JOIN language l ON l.id = dsc.language_id
                            WHERE dsc.state != 'locked' --do not show when locked
                            GROUP BY i.id
                        )


                        SELECT image_key, image_width, image_height, image_unlocked,
                        json_agg(json_build_object('name', q4.label, 'num_yes', q4.num_of_valid, 'num_no', q4.num_of_invalid, 'sublabels', q4.sublabels)),
                        coalesce(imgdsc.descriptions, '[]'::jsonb)
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
                        LEFT JOIN img_descriptions imgdsc ON imgdsc.image_id = q4.image_id
                        GROUP BY image_key, image_width, image_height, image_unlocked, imgdsc.descriptions
                        %s`, 
                      includeOwnImageDonations, includeOwnImageDonations, parseResult.Query, shuffleStr)

    var rows *sql.Rows
    if apiUser.Name != "" {
        parseResult.QueryValues = append(parseResult.QueryValues, apiUser.Name)
    }

    rows, err := p.db.Query(q, parseResult.QueryValues...)
    if err != nil {
        log.Debug("[Get Image Labels] Couldn't get image labels: ", err.Error())
        raven.CaptureError(err, nil)
        return imageLabels, err
    }

    defer rows.Close()


    for rows.Next() {
        var imageLabel datastructures.ImageLabel
        var labels []byte
        var imageDescriptionBytes []byte
        err = rows.Scan(&imageLabel.Image.Id, &imageLabel.Image.Width, &imageLabel.Image.Height, 
                            &imageLabel.Image.Unlocked, &labels, &imageDescriptionBytes)
        if err != nil {
            log.Debug("[Get Image Labels] Couldn't scan rows: ", err.Error())
            raven.CaptureError(err, nil)
            return imageLabels, err
        }

        err := json.Unmarshal(labels, &imageLabel.Labels)
        if err != nil {
            log.Debug("[Get Image Labels] Couldn't unmarshal image labels: ", err.Error())
            raven.CaptureError(err, nil)
            return nil, err
        }

        err = json.Unmarshal(imageDescriptionBytes, &imageLabel.Image.Descriptions)
        if err != nil {
            log.Debug("[Get Image Labels] Couldn't unmarshal image descriptions: ", err.Error())
            raven.CaptureError(err, nil)
            return nil, err
        }

        imageLabel.Image.Url = commons.GetImageUrlFromImageId(apiBaseUrl, imageLabel.Image.Id, imageLabel.Image.Unlocked)

        imageLabels = append(imageLabels, imageLabel)
    }

    return imageLabels, nil
}