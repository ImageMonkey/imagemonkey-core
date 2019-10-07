package imagemonkeydb

import (
    "github.com/getsentry/raven-go"
    log "github.com/sirupsen/logrus"
    datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
    parser "github.com/bbernhard/imagemonkey-core/parser/v2"
    commons "github.com/bbernhard/imagemonkey-core/commons" 
	"encoding/json"
    "errors"
    "fmt"
    "database/sql"
    "github.com/lib/pq"
    "image"
    "github.com/gofrs/uuid"
    //"github.com/francoispqt/gojay"
    //"bytes"
)


func updateAnnotationInTransaction2(tx *sql.Tx, apiUser datastructures.APIUser, label string, sublabel string, 
                                                            annotationsContainer datastructures.AnnotationsContainer) error {
    var queryValues []interface{}
    query := ""
    if label != "" && sublabel != "" {
        if annotationsContainer.IsSuggestion {
			tx.Rollback()
			return errors.New("Unexpected sublabel set for annotation suggestion")
		} else {
			query = `SELECT a.uuid 
                        FROM image_annotation a 
                        JOIN label l ON l.id = a.label_id
                        JOIN label pl ON l.parent_id = pl.id 
                        WHERE l.name = $1 AND pl.name = $2` 
        	queryValues = append(queryValues, label, sublabel)
			//queryValues = append(queryValues, label)
        	//queryValues = append(queryValues, sublabel)
		}
    } else {
		if annotationsContainer.IsSuggestion {
        	query = `SELECT a.uuid
						FROM image_annotation_suggestion a
						JOIN label_suggestion l ON l.id = a.label_suggestion_id
						WHERE l.name = $1` 
		} else {
			query = `SELECT a.uuid 
                        FROM image_annotation a 
                        JOIN label l ON l.id = a.label_id
                        WHERE l.name = $1`
    	}
		queryValues = append(queryValues, label)
		//queryValues = append(queryValues, label)
	}

    rows, err := tx.Query(query, queryValues...)
    if err != nil {
        tx.Rollback()
        log.Error("[Update Annotation] Couldn't get annotation id: ", err.Error())
        return err
    }

    if rows.Next() {
        var annotationId string
        err = rows.Scan(&annotationId)
        if err != nil {
            tx.Rollback()
            log.Error("[Update Annotation] Couldn't scan annotation id: ", err.Error())
            return err
        }

        rows.Close()

        return updateAnnotationInTransaction(tx, apiUser, annotationId, annotationsContainer)
    }
    tx.Rollback()
    return errors.New("[Update Annotation] Couldn't get uuid for label")
}

func updateAnnotationInTransaction(tx *sql.Tx, apiUser datastructures.APIUser, annotationId string, 
                                                annotationsContainer datastructures.AnnotationsContainer) error {
    byt, err := json.Marshal(annotationsContainer.Annotations.Annotations)
    if err != nil {
        tx.Rollback()
        log.Error("[Update Annotation] Couldn't create byte array: ", err.Error())
        return err
    }

	if annotationsContainer.IsSuggestion {
		if apiUser.Name == "" {
			tx.Rollback()
			return &AuthenticationRequiredError{Description: "Couldn't process request - you need to be authenticated to perform this action"}
		}
	}

    var imageAnnotationRevisionId int64

    //add entry to image_annotation_revision table
	insertImageAnnotationRevisionQuery := ""
	if annotationsContainer.IsSuggestion {
		insertImageAnnotationRevisionQuery = `INSERT INTO image_annotation_suggestion_revision(image_annotation_suggestion_id, revision)
                         						SELECT a.id, a.revision FROM image_annotation_suggestion a
                         						WHERE a.uuid = $1 
												RETURNING id`
	} else {
		insertImageAnnotationRevisionQuery = `INSERT INTO image_annotation_revision(image_annotation_id, revision)
                         						SELECT a.id, a.revision FROM image_annotation a
                         						WHERE a.uuid = $1 
												RETURNING id`
	}

    err = tx.QueryRow(insertImageAnnotationRevisionQuery, annotationId).Scan(&imageAnnotationRevisionId)
    if err != nil {
        tx.Rollback()
        log.Error("[Update Annotation] Couldn't insert to annotation revision: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

	updateAnnotationDataQuery := ""
	if annotationsContainer.IsSuggestion {
		updateAnnotationDataQuery = `UPDATE annotation_suggestion_data
                    					SET image_annotation_suggestion_id = NULL, image_annotation_suggestion_revision_id = $2
                     					FROM image_annotation_suggestion a WHERE a.uuid = $1 
                     					AND a.id = image_annotation_suggestion_id`
	} else {
		updateAnnotationDataQuery = `UPDATE annotation_data
                    					SET image_annotation_id = NULL, image_annotation_revision_id = $2
                     					FROM image_annotation a WHERE a.uuid = $1 
                     					AND a.id = image_annotation_id`
	}

    _, err = tx.Exec(updateAnnotationDataQuery, annotationId, imageAnnotationRevisionId)
    if err != nil {
        tx.Rollback()
        log.Error("[Update Annotation] Couldn't update annotation data: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    var imageAnnotationId int64
	
	updateImageAnnotationQuery := ""
	if annotationsContainer.IsSuggestion {
		updateImageAnnotationQuery = `UPDATE image_annotation_suggestion a SET num_of_valid = 0, num_of_invalid = 0, revision = revision + 1
                       					WHERE uuid = $1 
                       					RETURNING id`
	} else {
		updateImageAnnotationQuery = `UPDATE image_annotation a SET num_of_valid = 0, num_of_invalid = 0, revision = revision + 1
                       					WHERE uuid = $1 
                       					RETURNING id`
	}

    err = tx.QueryRow(updateImageAnnotationQuery, annotationId).Scan(&imageAnnotationId)
    if err != nil {
        tx.Rollback()
        log.Error("[Update Annotation] Couldn't update annotation: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }


	 //insertes annotation data; 'type' and 'refinements' are removed removed before inserting data
	insertAnnotationDataQuery := ""
	if annotationsContainer.IsSuggestion {
		insertAnnotationDataQuery = `INSERT INTO annotation_suggestion_data(image_annotation_suggestion_id, uuid, annotation, annotation_type_id)
                            	   		SELECT $1, uuid_generate_v4(), ((q.*)::jsonb - 'type' - 'refinements'), 
                                     		(SELECT id FROM annotation_type where name = ((q.*)->>'type')::text) 
                                     	 	 FROM json_array_elements($2) q
                                   		RETURNING uuid`
	} else {
		insertAnnotationDataQuery = `INSERT INTO annotation_data(image_annotation_id, uuid, annotation, annotation_type_id)
                            	   		SELECT $1, uuid_generate_v4(), ((q.*)::jsonb - 'type' - 'refinements'), 
                                     		(SELECT id FROM annotation_type where name = ((q.*)->>'type')::text) 
                                     	 	 FROM json_array_elements($2) q
                                   		RETURNING uuid`
	}

    var rows *sql.Rows
    rows, err = tx.Query(insertAnnotationDataQuery, imageAnnotationId, byt)
    if err != nil {
        tx.Rollback()
        log.Error("[Update Annotation] Couldn't add annotations: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }
    defer rows.Close()
    annotationDataIds := make(map[int]string)
    i := 0
    for rows.Next() {
        var annotationDataId string
        err = rows.Scan(&annotationDataId)
        if err != nil {
            tx.Rollback()
            log.Error("[Update Annotation] Couldn't scan annotation data ids: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }
        annotationDataIds[i] = annotationDataId
        i += 1
    }
    rows.Close()

    if len(annotationsContainer.AllowedRefinements) != len(annotationDataIds) {
        tx.Rollback()
        err = errors.New("Num of annotation refinements do not match num of annotation data ids!")
        log.Error("[Update Annotation] Couldn't add annotations : ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    for i, refinements := range annotationsContainer.AllowedRefinements {
        if val, ok := annotationDataIds[i]; ok {
            err = addOrUpdateRefinementsInTransaction(tx, annotationId, val, refinements, apiUser.ClientFingerprint, annotationsContainer.IsSuggestion)
            if err != nil { //transaction already rolled back, so we can return here
                return err
            }
        }
    }
    return nil
}

func (p *ImageMonkeyDatabase) UpdateAnnotation(apiUser datastructures.APIUser, annotationId string, 
		                                          annotationsContainer datastructures.AnnotationsContainer) error {

    tx, err := p.db.Begin()
    if err != nil {
        log.Debug("[Update Annotation] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    err = updateAnnotationInTransaction(tx, apiUser, annotationId, annotationsContainer)
    if err != nil { //transaction already rolled back, so we can return here
		log.Error(err.Error())
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

func (p *ImageMonkeyDatabase) AddAnnotations(apiUser datastructures.APIUser, imageId string, 
			annotations []datastructures.AnnotationsContainer) ([]string, error) {

    annotationIds := []string{}

    tx, err := p.db.Begin()
    if err != nil {
        log.Error("[Add Annotation] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return annotationIds, err
    }

    //currently there is a uniqueness constraint on the image_id column to ensure that we only have
    //one image annotation per image. That means that the below query can fail with a unique constraint error. 
    //at the moment the uniqueness constraint errors are handled gracefully - that means we return nil.
    //we might want to change that in the future to support multiple annotations per image (if there is a use case for it),
    //but for now it should be fine.

    for _, annotation := range annotations {
		if annotation.IsSuggestion { //label is not known to us (i.e non-productive)
			if apiUser.Name == "" {
				tx.Rollback()
				return annotationIds, &AuthenticationRequiredError{Description: "you need to be authenticated to perform this action"}
			}
		}
		
		
		byt, err := json.Marshal(annotation.Annotations.Annotations)
        if err != nil {
            tx.Rollback()
            log.Error("[Add Annotation] Couldn't create byte array: ", err.Error())
            return annotationIds, err
        }
	
        var idRows *sql.Rows
		var insertImageAnnotationQueryValues []interface{}
		insertImageAnnotationQuery := ""
		if annotation.IsSuggestion {
			insertImageAnnotationQuery = `INSERT INTO image_annotation_suggestion(label_suggestion_id, num_of_valid, num_of_invalid, 
												fingerprint_of_last_modification, image_id, uuid, auto_generated, revision) 
												SELECT (SELECT l.id FROM label_suggestion l WHERE l.name = $5), 
													$2, $3, $4, 
													(SELECT i.id FROM image i WHERE i.key = $1), 
													uuid_generate_v4(), $6, $7 
										  ON CONFLICT DO NOTHING RETURNING id, uuid`
			
			insertImageAnnotationQueryValues = append(insertImageAnnotationQueryValues, imageId, 0, 0, apiUser.ClientFingerprint, 
															annotation.Annotations.Label, annotation.AutoGenerated, 1)
		
		} else {
			if annotation.Annotations.Sublabel == "" {
				insertImageAnnotationQuery = `INSERT INTO image_annotation(label_id, num_of_valid, num_of_invalid, 
												fingerprint_of_last_modification, image_id, uuid, auto_generated, revision) 
													SELECT (SELECT l.id FROM label l WHERE l.name = $5 AND l.parent_id is null), 
														$2, $3, $4, 
														(SELECT i.id FROM image i WHERE i.key = $1), 
														uuid_generate_v4(), $6, $7 
											  ON CONFLICT DO NOTHING RETURNING id, uuid`
			
				insertImageAnnotationQueryValues = append(insertImageAnnotationQueryValues, imageId, 0, 0, apiUser.ClientFingerprint, 
															annotation.Annotations.Label, annotation.AutoGenerated, 1)

			} else {
				insertImageAnnotationQuery = `INSERT INTO image_annotation(label_id, num_of_valid, num_of_invalid, 
												fingerprint_of_last_modification, image_id, uuid, auto_generated, revision) 
													SELECT (SELECT l.id FROM label l JOIN label pl ON l.parent_id = pl.id WHERE l.name = $5 AND pl.name = $6), 
														$2, $3, $4, 
														(SELECT i.id FROM image i WHERE i.key = $1), 
														uuid_generate_v4(), $7, $8 
											  ON CONFLICT DO NOTHING RETURNING id, uuid`
            
			insertImageAnnotationQueryValues = append(insertImageAnnotationQueryValues, imageId, 0, 0, apiUser.ClientFingerprint, 
														annotation.Annotations.Sublabel, annotation.Annotations.Label, annotation.AutoGenerated, 1)
			}
        }

		idRows, err = tx.Query(insertImageAnnotationQuery, insertImageAnnotationQueryValues...)
        if err != nil {
            tx.Rollback()
            log.Error("[Update Annotation] Couldn't add image annotation: ", err.Error())
            return annotationIds, err
        }

        defer idRows.Close()

        var annotationId string
        var insertedId int64

        if !idRows.Next() { //we get no result set in case there already exists an entry 
                            //in that case, just update the annotation
            idRows.Close()
            err = updateAnnotationInTransaction2(tx, apiUser, annotation.Annotations.Label, annotation.Annotations.Sublabel, annotation)
            if err != nil { //transaction already rolled back, so we can return here
                return annotationIds, err
            }
            continue 
        } else { //image annotation successfully added, get inserted ids
            err = idRows.Scan(&insertedId, &annotationId)
            if err != nil {
                tx.Rollback()
                log.Error("[Update Annotation] Couldn't scan image annotation row: ", err.Error())
                return annotationIds, err
            }
        }

        idRows.Close()

        //insertes annotation data; 'type' and 'refinements' are removed removed before inserting data
        var rows *sql.Rows
		insertAnnotationDataQuery := ""
		if annotation.IsSuggestion {
			insertAnnotationDataQuery = `INSERT INTO annotation_suggestion_data(image_annotation_suggestion_id, uuid, annotation, annotation_type_id)
                                		  SELECT $1, uuid_generate_v4(), 
                                        	((q.*)::jsonb - 'type' - 'refinements'), 
                                        	(SELECT id FROM annotation_type where name = ((q.*)->>'type')::text) 
                                            FROM json_array_elements($2) q
                                		  RETURNING uuid`
		} else {
			insertAnnotationDataQuery = `INSERT INTO annotation_data(image_annotation_id, uuid, annotation, annotation_type_id)
                                		  SELECT $1, uuid_generate_v4(), 
                                        	((q.*)::jsonb - 'type' - 'refinements'), 
                                        	(SELECT id FROM annotation_type where name = ((q.*)->>'type')::text) 
                                            FROM json_array_elements($2) q
                                		  RETURNING uuid` 
		}

        rows, err = tx.Query(insertAnnotationDataQuery, insertedId, byt)
        if err != nil {
            tx.Rollback()
            log.Error("[Add Annotation] Couldn't add annotations: ", err.Error())
            raven.CaptureError(err, nil)
            return annotationIds, err
        }
        defer rows.Close()
        annotationDataIds := make(map[int]string)
        i := 0
        for rows.Next() {
            var annotationDataId string
            err = rows.Scan(&annotationDataId)
            if err != nil {
                tx.Rollback()
                log.Error("[Add Annotation] Couldn't scan annotation data ids: ", err.Error())
                raven.CaptureError(err, nil)
                return annotationIds, err
            }
            annotationDataIds[i] = annotationDataId
            i += 1
        }
        rows.Close()

        if len(annotation.AllowedRefinements) != len(annotationDataIds) {
            tx.Rollback()
            err = errors.New("Num of annotation refinements do not match num of annotation data ids!")
            log.Error("[Add Annotation] Couldn't add annotations : ", err.Error())
            raven.CaptureError(err, nil)
            return annotationIds, err
        }

        for i, refinements := range annotation.AllowedRefinements {
            if val, ok := annotationDataIds[i]; ok {
                err = addOrUpdateRefinementsInTransaction(tx, annotationId, val, refinements, apiUser.ClientFingerprint, annotation.IsSuggestion)
                if err != nil { //transaction already rolled back, so we can return here
                    return annotationIds, err
                }
            }
        }

        if apiUser.Name != "" {
            var id int64

            id = 0
			userImageAnnotationQuery := ""
			if annotation.IsSuggestion {
				userImageAnnotationQuery = `INSERT INTO user_image_annotation_suggestion(image_annotation_suggestion_id, account_id, timestamp)
                                    			SELECT $1, a.id, CURRENT_TIMESTAMP 
												FROM account a WHERE a.name = $2 
												RETURNING id`
			} else {
				userImageAnnotationQuery = `INSERT INTO user_image_annotation(image_annotation_id, account_id, timestamp)
                                    			SELECT $1, a.id, CURRENT_TIMESTAMP 
												FROM account a WHERE a.name = $2 
												RETURNING id`
			}
            err = tx.QueryRow(userImageAnnotationQuery, insertedId, apiUser.Name).Scan(&id)
            if err != nil {
                tx.Rollback()
                log.Error("[Add User Annotation] Couldn't add user annotation entry: ", err.Error())
                raven.CaptureError(err, nil)
                return annotationIds, err
            }

            if id == 0 {
                tx.Rollback()
                log.Error("[Add User Annotation] Nothing inserted")
                return annotationIds, errors.New("nothing inserted")
            }
        }

        annotationIds = append(annotationIds, annotationId)
    }


    err = tx.Commit()
    if err != nil {
        log.Error("[Add Annotation] Couldn't commit transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return annotationIds, err
    }

    return annotationIds, nil
}

func (p *ImageMonkeyDatabase) _getImageForAnnotationFromValidationId(username string, validationId string, 
		addAutoAnnotations bool) (datastructures.UnannotatedImage, error) {
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

    q := fmt.Sprintf(`SELECT i.key, label_name, parent_label_name, q2.label_accessor, i.width, i.height, validation_uuid, 
                           json_agg(q1.annotation || ('{"type":"' || q1.name || '"}')::jsonb)::jsonb as auto_annotations,
                           i.unlocked
                            FROM image i 
                            JOIN image_provider p ON i.image_provider_id = p.id 
                            
							JOIN (
								SELECT v.image_id as image_id, v.uuid as validation_uuid, 
								l.name as label_name, COALESCE(pl.name, '') as parent_label_name,
								acc.accessor as label_accessor, l.id as label_id, false as is_suggestion
                            	FROM image_validation v
								JOIN label l ON v.label_id = l.id
                            	JOIN label_accessor acc ON acc.label_id = v.label_id
                            	LEFT JOIN label pl ON l.parent_id = pl.id

								UNION ALL

								SELECT s.image_id as image_id, s.uuid as validation_uuid,
								l.name as label_name, '' as parent_label_name, l.name as label_accessor,
								l.id as label_id, true as is_suggestion
								FROM image_label_suggestion s
								JOIN label_suggestion l ON l.id = s.label_suggestion_id

							) q2 ON q2.image_id = i.id


                            LEFT JOIN 
                            (
                                SELECT a.label_id as label_id, a.image_id as image_id, d.annotation, t.name,
								false as is_suggestion
                                FROM image_annotation a 
                                JOIN annotation_data d ON d.image_annotation_id = a.id
                                JOIN annotation_type t on d.annotation_type_id = t.id
                                WHERE a.auto_generated = true

								UNION ALL

								SELECT a.label_suggestion_id as label_id, a.image_id as image_id, d.annotation, t.name,
								true as is_suggestion
                                FROM image_annotation_suggestion a 
                                JOIN annotation_suggestion_data d ON d.image_annotation_suggestion_id = a.id
                                JOIN annotation_type t on d.annotation_type_id = t.id
                                WHERE a.auto_generated = true

                            ) q1 ON q2.label_id = q1.label_id AND i.id = q1.image_id AND q2.is_suggestion = q1.is_suggestion
                            WHERE (i.unlocked = true %s) AND p.name = 'donation' AND q2.validation_uuid::text = $1
                            GROUP BY i.key, q2.label_name, q2.parent_label_name, q2.label_accessor, 
							i.width, i.height, q2.validation_uuid, i.unlocked`, includeOwnImageDonations)

    //we do not check, whether there already exists a annotation for the given validation id. 
    //there is anyway only one annotation per validation allowed, so if someone tries to push another annotation, the corresponding POST request 
    //would fail 
    var rows *sql.Rows
    var err error

    if username == "" {
        rows, err = p.db.Query(q, validationId)
    } else {
        rows, err = p.db.Query(q, validationId, username)
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

        err = rows.Scan(&unannotatedImage.Id, &label1, &label2, &unannotatedImage.Label.Accessor, 
                            &unannotatedImage.Width, &unannotatedImage.Height, &unannotatedImage.Validation.Id, 
                            &autoAnnotationBytes, &unannotatedImage.Unlocked)
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
            unannotatedImage.Label.Label = label1
            unannotatedImage.Label.Sublabel = ""
        } else {
            unannotatedImage.Label.Label = label2
            unannotatedImage.Label.Sublabel = label1
        }
    }

    return unannotatedImage, nil
}

func (p *ImageMonkeyDatabase) GetImageForAnnotation(username string, addAutoAnnotations bool, 
			validationId string, labelId string) (datastructures.UnannotatedImage, error) {
    //if a validation id is provided, use a different code path. 
    //selecting a single image given a validation id is totally different from selecting a random image
    //so it makes sense to use a different code path here. 
    if validationId != "" {
        return p._getImageForAnnotationFromValidationId(username, validationId, addAutoAnnotations)
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


    q := fmt.Sprintf(`SELECT q.image_key, q.label, q.parent_label, q.accessor, q.image_width, q.image_height, q.validation_uuid, 
                        CASE WHEN json_agg(q1.annotation)::jsonb = '[null]'::jsonb THEN '[]' ELSE json_agg(q1.annotation || ('{"type":"' || q1.annotation_type || '"}')::jsonb)::jsonb END as auto_annotations,
                        q.image_unlocked
                        FROM
                        (SELECT l.id as label_id, i.id as image_id, i.key as image_key, l.name as label, COALESCE(pl.name, '') as parent_label, 
                            acc.accessor as accessor, width as image_width, height as image_height, v.uuid as validation_uuid, i.unlocked as image_unlocked
                            FROM image i 
                            JOIN image_provider p ON i.image_provider_id = p.id 
                            JOIN image_validation v ON v.image_id = i.id
                            JOIN label l ON v.label_id = l.id
                            JOIN label_accessor acc ON acc.label_id = v.label_id
                            LEFT JOIN label pl ON l.parent_id = pl.id
                            WHERE (i.unlocked = true %s) AND p.name = 'donation' AND l.label_type != 'meta' AND
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
                                    WHERE (i.unlocked = true %s) AND p.name = 'donation' AND l.label_type != 'meta' AND 
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
                        GROUP BY q.image_key, q.label, q.parent_label, q.accessor,
                        q.image_width, q.image_height, q.validation_uuid, q.image_unlocked`, 
                        includeOwnImageDonations, q1, q2, q3, includeOwnImageDonations, q1, q2, q3)

    //select all images that aren't already annotated and have a label correctness probability of >= 0.8 
    var rows *sql.Rows
    var err error
    if labelId == "" {
        if username != "" {
            rows, err = p.db.Query(q, maxNumNotAnnotatable, username)
        } else {
            rows, err = p.db.Query(q, maxNumNotAnnotatable)
        } 
    } else {
        if username != "" {
            rows, err = p.db.Query(q, labelId, maxNumNotAnnotatable, username)
        } else {
            rows, err = p.db.Query(q, labelId, maxNumNotAnnotatable)
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

        err = rows.Scan(&unannotatedImage.Id, &label1, &label2, &unannotatedImage.Label.Accessor, 
            &unannotatedImage.Width, &unannotatedImage.Height, &unannotatedImage.Validation.Id, 
            &autoAnnotationBytes, &unannotatedImage.Unlocked)
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
            unannotatedImage.Label.Label = label1
            unannotatedImage.Label.Sublabel = ""
        } else {
            unannotatedImage.Label.Label = label2
            unannotatedImage.Label.Sublabel = label1
        }
    }

    return unannotatedImage, nil
}

func (p *ImageMonkeyDatabase) GetAnnotatedImage(apiUser datastructures.APIUser, annotationId string, 
		autoGenerated bool, revision int32) (datastructures.AnnotatedImage, error) {
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

        q = fmt.Sprintf(`SELECT q2.image_key, q2.label_name, q2.parent_label_name, q2.annotation_uuid, json_agg(q2.annotation), 
                            q2.num_of_valid, q2.num_of_invalid, q2.image_width, q2.image_height, q2.image_unlocked, q2.is_suggestion
                            FROM
                            (
                                SELECT q1.key as image_key, q1.label_name, q1.parent_label_name, q1.annotation_uuid as annotation_uuid, 
                                    q.annotation || ('{"type":"' || q.annotation_type || '"}')::jsonb
                                     || jsonb_strip_nulls(jsonb_build_object('refinements', ((json_agg(jsonb_build_object('label_uuid', q.annotation_refinement_uuid)) 
                                        FILTER (WHERE q.annotation_refinement_uuid IS NOT NULL))))) as annotation, 
                                     q1.num_of_valid as num_of_valid, q1.num_of_invalid as num_of_invalid, q1.width as image_width, 
                                     q1.height as image_height, q1.image_unlocked as image_unlocked, q1.is_suggestion
                                       FROM (
                                         SELECT i.key as key, i.id as image_id, q2.label_id as label_id, 
                                         q2.id as entry_id, q2.annotation_uuid as annotation_uuid, q2.num_of_valid as num_of_valid, 
                                         q2.num_of_invalid as num_of_invalid, i.width as width, i.height as height, q2.is_revision,
                                         i.unlocked as image_unlocked, q2.is_suggestion, q2.label_name, q2.parent_label_name
                                         FROM image i
                                         JOIN image_provider p ON i.image_provider_id = p.id
                                         JOIN (
                                            SELECT DISTINCT a.image_id as image_id, a.label_id as label_id, a.uuid as annotation_uuid,
                                            a.num_of_valid as num_of_valid, a.num_of_invalid as num_of_invalid,
                                            CASE WHEN r.revision = $1 THEN r.id ELSE a.id END as id, 
                                            CASE WHEN r.revision = $1 THEN true ELSE false END as is_revision, false as is_suggestion,
											l.name as label_name, COALESCE(pl.name, '') as parent_label_name
                                            FROM image_annotation a
                                            JOIN label l ON a.label_id = l.id
                                       		LEFT JOIN label pl ON l.parent_id = pl.id
											LEFT JOIN image_annotation_revision r ON r.image_annotation_id = a.id
                                            where a.uuid::text = $2 
                                            AND a.auto_generated = false and (r.revision = $1 or a.revision = $1)

											UNION ALL

											SELECT DISTINCT a.image_id as image_id, a.label_suggestion_id as label_id, a.uuid as annotation_uuid,
                                            a.num_of_valid as num_of_valid, a.num_of_invalid as num_of_invalid, 	
                                            CASE WHEN r.revision = $1 THEN r.id ELSE a.id END as id, 
                                            CASE WHEN r.revision = $1 THEN true ELSE false END as is_revision, true as is_suggestion,
											l.name as label_name, '' as parent_label_name
                                            FROM image_annotation_suggestion a
											JOIN label_suggestion l ON l.id = a.label_suggestion_id
                                            LEFT JOIN image_annotation_suggestion_revision r ON r.image_annotation_suggestion_id = a.id
                                            where a.uuid::text = $2 
                                            AND a.auto_generated = false and (r.revision = $1 or a.revision = $1)

                                         ) q2 ON q2.image_id = i.id
                                         WHERE (i.unlocked = true %s) AND p.name = 'donation'
                                         
                                         
                                       ) q1

                                       JOIN
                                       (
                                         SELECT d.annotation as annotation, l.uuid as annotation_refinement_uuid, t.name as annotation_type,
                                         d.image_annotation_id as image_annotation_id, d.image_annotation_revision_id as image_annotation_revision_id,
										 false as is_suggestion
                                         FROM annotation_data d 
                                         JOIN annotation_type t on d.annotation_type_id = t.id
                                         LEFT JOIN image_annotation_refinement r ON r.annotation_data_id = d.id
                                         LEFT JOIN label l ON l.id = r.label_id

										 UNION ALL

										 SELECT d.annotation as annotation, l.uuid as annotation_refinement_uuid, t.name as annotation_type,
                                         d.image_annotation_suggestion_id as image_annotation_id, 
										 d.image_annotation_suggestion_revision_id as image_annotation_revision_id, true as is_suggestion
                                         FROM annotation_suggestion_data d 
                                         JOIN annotation_type t on d.annotation_type_id = t.id
                                         LEFT JOIN image_annotation_suggestion_refinement r ON r.annotation_suggestion_data_id = d.id
                                         LEFT JOIN label l ON l.id = r.label_id
                                       ) q ON 
                                         CASE 
                                            WHEN q1.is_revision THEN q.image_annotation_revision_id = q1.entry_id
                                            ELSE q.image_annotation_id = q1.entry_id 
                                         END AND q.is_suggestion = q1.is_suggestion

                                       GROUP BY q1.key, q.annotation, q.annotation_type, q1.annotation_uuid, label_name, parent_label_name, 
                                       q1.num_of_valid, q1.num_of_invalid, q1.width, q1.height, q1.image_unlocked, q1.is_suggestion
                            ) q2
                            GROUP BY q2.image_key, q2.label_name, q2.parent_label_name, q2.annotation_uuid, 
                            q2.num_of_valid, q2.num_of_invalid, q2.image_width, q2.image_height, q2.image_unlocked,
							q2.is_suggestion`, includeOwnImageDonations)
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

        q = fmt.Sprintf(`SELECT q2.image_key, q2.label_name, q2.parent_label_name, q2.annotation_uuid, json_agg(q2.annotation), q2.num_of_valid,
                            q2.num_of_invalid, q2.image_width, q2.image_height, q2.image_unlocked, q2.is_suggestion
                            FROM
                            (
                                SELECT q1.key as image_key, label_name, parent_label_name, q1.annotation_uuid as annotation_uuid, 
                                     q.annotation || ('{"type":"' || q.annotation_type || '"}')::jsonb
                                     || jsonb_strip_nulls(jsonb_build_object('refinements', ((json_agg(jsonb_build_object('label_uuid', q.annotation_refinement_uuid)) 
                                        FILTER (WHERE q.annotation_refinement_uuid IS NOT NULL))))) as annotation, 
                                     q1.num_of_valid as num_of_valid, q1.num_of_invalid as num_of_invalid, q1.width as image_width, 
                                     q1.height as image_height, q1.image_unlocked as image_unlocked, q1.is_suggestion
                                       FROM (
                                         SELECT i.key as key, i.id as image_id, label_name, parent_label_name,
                                         image_annotation_id, a.uuid as annotation_uuid, a.num_of_valid as num_of_valid, 
                                         a.num_of_invalid as num_of_invalid, i.width as width, i.height as height, i.unlocked as image_unlocked,
										 a.is_suggestion
                                         FROM image i
                                         JOIN image_provider p ON i.image_provider_id = p.id
                                         
										 JOIN (
										 	SELECT an.id as image_annotation_id, an.uuid as uuid, 
											an.image_id as image_id, an.auto_generated as auto_generated,
											l.name as label_name, COALESCE(pl.name, '') as parent_label_name,
											an.num_of_valid as num_of_valid, an.num_of_invalid as num_of_invalid,
											false as is_suggestion
											FROM image_annotation an
											JOIN label l ON an.label_id = l.id
                                       		LEFT JOIN label pl ON l.parent_id = pl.id
											UNION ALL

											SELECT ans.id as image_annotation_id, ans.uuid as uuid, 
											ans.image_id as image_id, ans.auto_generated as auto_generated,
											l.name as label_name, '' as parent_label_name,
											ans.num_of_valid as num_of_valid, ans.num_of_invalid as num_of_invalid,
											true as is_suggestion
											FROM image_annotation_suggestion ans
											JOIN label_suggestion l ON l.id = ans.label_suggestion_id
										 ) a ON a.image_id = i.id
										 
										 
										 WHERE (i.unlocked = true %s) AND p.name = 'donation' AND a.auto_generated = $1
                                         %s
                                         
                                         
                                       ) q1

                                       JOIN
                                       (
                                         SELECT d.image_annotation_id as image_annotation_id, l.uuid as annotation_refinement_uuid, 
                                         d.annotation as annotation, t.name as annotation_type, false as is_suggestion
                                         FROM annotation_data d 
                                         JOIN annotation_type t on d.annotation_type_id = t.id
                                         LEFT JOIN image_annotation_refinement r ON r.annotation_data_id = d.id
                                         LEFT JOIN label l ON l.id = r.label_id
										 
										 UNION ALL

										 SELECT d.image_annotation_suggestion_id as image_annotation_id, l.uuid as annotation_refinement_uuid, 
                                         d.annotation as annotation, t.name as annotation_type, true as is_suggestion
                                         FROM annotation_suggestion_data d 
                                         JOIN annotation_type t on d.annotation_type_id = t.id
										 LEFT JOIN image_annotation_suggestion_refinement r ON r.annotation_suggestion_data_id = d.id
                                         LEFT JOIN label l ON l.id = r.label_id
                                       ) q ON q.image_annotation_id = q1.image_annotation_id AND q.is_suggestion = q1.is_suggestion


                                      
                                       GROUP BY q1.key, q.annotation, q.annotation_type, q1.annotation_uuid, label_name, parent_label_name, 
                                       q1.num_of_valid, q1.num_of_invalid, q1.width, q1.height, q1.image_unlocked, q1.is_suggestion
                            ) q2
                            GROUP BY q2.image_key, q2.label_name, q2.parent_label_name, q2.annotation_uuid, q2.num_of_valid,
                            q2.num_of_invalid, q2.image_width, q2.image_height, q2.image_unlocked, q2.is_suggestion`, includeOwnImageDonations, q1)
	}

    var err error

    tx, err := p.db.Begin()
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
                rows, err = p.db.Query(q, autoGenerated)
            } else {
                rows, err = p.db.Query(q, autoGenerated, apiUser.Name)
            }
        } else {
            if apiUser.Name == "" {
                rows, err = p.db.Query(q, autoGenerated, annotationId)
            } else {
                rows, err = p.db.Query(q, autoGenerated, annotationId, apiUser.Name)
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
	var isSuggestion bool = false
    if rows.Next() {
        var annotations []byte
        annotatedImage.Image.Provider = "donation"

        err = rows.Scan(&annotatedImage.Image.Id, &label1, &label2, &annotatedImage.Id, 
                        &annotations, &annotatedImage.NumOfValid, &annotatedImage.NumOfInvalid, 
                        &annotatedImage.Image.Width, &annotatedImage.Image.Height, &annotatedImage.Image.Unlocked, &isSuggestion)
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
        if isSuggestion {
			err = tx.QueryRow(`SELECT (SUM(CASE WHEN r.id is null THEN 0 ELSE 1 END) + 1)::integer as num 
                               FROM image_annotation_suggestion a 
                               LEFT JOIN image_annotation_suggestion_revision r ON r.image_annotation_suggestion_id = a.id 
                               WHERE a.uuid::text = $1`, annotationId).Scan(&annotatedImage.NumRevisions)
		} else {
			err = tx.QueryRow(`SELECT (SUM(CASE WHEN r.id is null THEN 0 ELSE 1 END) + 1)::integer as num 
                               FROM image_annotation a 
                               LEFT JOIN image_annotation_revision r ON r.image_annotation_id = a.id 
                               WHERE a.uuid::text = $1`, annotationId).Scan(&annotatedImage.NumRevisions)
		}

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

func (p *ImageMonkeyDatabase) ValidateAnnotatedImage(clientFingerprint string, annotationId string, 
		labelValidationEntry datastructures.LabelValidationEntry, valid bool) error {
    if valid {
        var err error
        if labelValidationEntry.Sublabel == "" {
            _, err = p.db.Exec(`UPDATE image_annotation AS a 
                              SET num_of_valid = num_of_valid + 1, fingerprint_of_last_modification = $1
                              WHERE a.uuid = $2 AND a.label_id = (SELECT id FROM label WHERE name = $3 AND parent_id is null)`, 
                              clientFingerprint, annotationId, labelValidationEntry.Label)
        } else {
            _, err = p.db.Exec(`UPDATE image_annotation AS a 
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
            _,err = p.db.Exec(`UPDATE image_annotation AS a 
                              SET num_of_invalid = num_of_invalid + 1, fingerprint_of_last_modification = $1
                              WHERE a.uuid = $2 AND a.label_id = (
                                SELECT id FROM label WHERE name = $3 AND parent_id is null
                              )`, 
                              clientFingerprint, annotationId, labelValidationEntry.Label)
        } else {
            _,err = p.db.Exec(`UPDATE image_annotation AS a 
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

func (p *ImageMonkeyDatabase) GetAnnotationCoverage(imageId string) ([]datastructures.ImageAnnotationCoverage, error) {
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

    rows, err := p.db.Query(q, queryValues...)
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


func (p *ImageMonkeyDatabase) GetAnnotationsForRefinement(parseResult parser.ParseResult, apiBaseUrl string, 
        annotationDataId string) ([]datastructures.AnnotationRefinementTask, error) {
    var annotationRefinementTasks []datastructures.AnnotationRefinementTask

    q1 := ""
    if annotationDataId != "" {
        q1 = fmt.Sprintf("WHERE d.uuid::text = $%d", len(parseResult.QueryValues) + 1)
    }

    q2 := ""
    if len(parseResult.QueryValues) > 0 {
        q2 = fmt.Sprintf("WHERE %s", parseResult.Query)
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
        parseResult.QueryValues = append(parseResult.QueryValues, annotationDataId)
    }

    rows, err := p.db.Query(q, parseResult.QueryValues...)
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


        annotationRefinementTask.Image.Url = commons.GetImageUrlFromImageId(apiBaseUrl, annotationRefinementTask.Image.Id, 
                                                                        annotationRefinementTask.Image.Unlocked)

        annotationRefinementTasks = append(annotationRefinementTasks, annotationRefinementTask)
    }

    return annotationRefinementTasks, nil
}


func (p *ImageMonkeyDatabase) GetAnnotations(apiUser datastructures.APIUser, parseResult parser.ParseResult, 
                        imageId string, apiBaseUrl string) ([]datastructures.AnnotatedImage, error) {
    annotatedImages := []datastructures.AnnotatedImage{}
    var queryValues []interface{}


    q1 := ""
    if imageId == "" {
        q1 = "WHERE " + parseResult.Query
        queryValues = parseResult.QueryValues
    } else {
        q1 = "WHERE image_key = $1"
        queryValues = append(queryValues, imageId)
    }

	q2 := "acc.name is null"
    includeOwnImageDonations := ""
    if apiUser.Name != "" {
        q2 = fmt.Sprintf(`acc.name = $%d`, len(queryValues) + 1)
		
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

    q := fmt.Sprintf(`SELECT q2.image_key, q2.label_name, q2.parent_label_name, q2.annotation_uuid, json_agg(q2.annotation), 
                      q2.num_of_valid, q2.num_of_invalid, q2.image_width, q2.image_height, q2.image_unlocked
                      FROM
                      (
                        SELECT q1.image_key as image_key, label_name, parent_label_name, 
                            q1.annotation_uuid as annotation_uuid, 
                             q.annotation || ('{"type":"' || q.annotation_type || '"}')::jsonb
                                || jsonb_strip_nulls(jsonb_build_object('refinements', ((json_agg(jsonb_build_object('label_uuid', q.annotation_refinement_uuid)) 
                                FILTER (WHERE q.annotation_refinement_uuid IS NOT NULL))))) as annotation, 
                             q1.num_of_valid as num_of_valid, q1.num_of_invalid as num_of_invalid, 
                             q1.image_width as image_width, q1.image_height as image_height, q1.image_unlocked as image_unlocked, q1.is_suggestion
                                   FROM (
                                     SELECT image_key, image_id, entry_id, annotation_uuid, num_of_valid,
                                     num_of_invalid, image_width, image_height, image_unlocked, annotated_percentage, image_collection, label_name,
									 parent_label_name, is_suggestion
                                     FROM
                                     (
                                         SELECT i.key as image_key, i.id as image_id, 
                                         entry_id, annotation_uuid, num_of_valid, num_of_invalid, i.width as image_width, i.height as image_height,
                                         i.unlocked as image_unlocked, qq.annotated_percentage, coll.image_collection_name as image_collection,
										 label_name, parent_label_name, accessor, is_suggestion

                                         FROM image i
                                         JOIN image_provider p ON i.image_provider_id = p.id
                                         
										 JOIN (
										 	SELECT an.id as entry_id, an.num_of_invalid as num_of_invalid, an.num_of_valid as num_of_valid,
											an.uuid as annotation_uuid, l.name as label_name, COALESCE(pl.name, '') as parent_label_name, an.image_id as image_id,
											la.accessor as accessor, false as is_suggestion
										 	FROM image_annotation an 
                                         	JOIN label_accessor la ON la.label_id = an.label_id
                                         	JOIN label l ON an.label_id = l.id
                                         	LEFT JOIN label pl ON l.parent_id = pl.id
											WHERE an.auto_generated = false

											UNION ALL

											SELECT an.id as entry_id, an.num_of_invalid as num_of_invalid, an.num_of_valid as num_of_valid,
											an.uuid as annotation_uuid, l.name as label_name, '' as parent_label_name, an.image_id as image_id,
											l.name as accessor, true as is_suggestion
										 	FROM image_annotation_suggestion an
											JOIN label_suggestion l ON l.id = an.label_suggestion_id
											WHERE an.auto_generated = false
										 ) ann_labels ON ann_labels.image_id = i.id
                                         
										 LEFT JOIN image_annotation_coverage qq ON qq.image_id = i.id


										 LEFT JOIN 
										 (
											SELECT ui.name as image_collection_name, c.image_id as image_id
											FROM image_collection_image c
											JOIN user_image_collection ui ON c.user_image_collection_id = ui.id
											JOIN account acc ON acc.id = ui.account_id
											WHERE %s
										 ) coll ON coll.image_id = i.id
                                         WHERE (i.unlocked = true %s) AND p.name = 'donation'
                                     ) a 
                                     %s
                                   ) q1

                                   JOIN
                                   (
                                     
									 SELECT d.image_annotation_id as annotation_id, d.annotation as annotation, t.name as annotation_type,
                                     l.uuid as annotation_refinement_uuid, false as is_suggestion
                                     FROM annotation_data d 
                                     JOIN annotation_type t on d.annotation_type_id = t.id
                                     LEFT JOIN image_annotation_refinement r ON r.annotation_data_id = d.id
                                     LEFT JOIN label l ON l.id = r.label_id

									 UNION ALL

									 SELECT d.image_annotation_suggestion_id as annotation_id, d.annotation as annotation, t.name as annotation_type,
                                     l.uuid as annotation_refinement_uuid, true as is_suggestion
                                     FROM annotation_suggestion_data d 
                                     JOIN annotation_type t on d.annotation_type_id = t.id
									 LEFT JOIN image_annotation_suggestion_refinement r ON r.annotation_suggestion_data_id = d.id
                                     LEFT JOIN label l ON l.id = r.label_id

                                   ) q ON q.annotation_id = q1.entry_id AND q1.is_suggestion = q.is_suggestion 

                                   GROUP BY image_key, q.annotation_id, q.annotation, q.annotation_type, q1.annotation_uuid, label_name, parent_label_name, 
                                   q1.num_of_valid, q1.num_of_invalid, q1.image_width, q1.image_height, q1.image_unlocked, q1.is_suggestion
                      ) q2
                      GROUP BY q2.image_key, q2.label_name, q2.parent_label_name, q2.annotation_uuid, 
                            q2.num_of_valid, q2.num_of_invalid, q2.image_width, q2.image_height, q2.image_unlocked
                      `, q2, includeOwnImageDonations, q1)

    rows, err := p.db.Query(q, queryValues...)
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

        annotatedImage.Image.Url = commons.GetImageUrlFromImageId(apiBaseUrl, annotatedImage.Image.Id, annotatedImage.Image.Unlocked)

        annotatedImages = append(annotatedImages, annotatedImage)

    }
    return annotatedImages, nil
}

func (p *ImageMonkeyDatabase) GetAvailableAnnotationTasks(apiUser datastructures.APIUser, parseResult parser.ParseResult, 
		orderRandomly bool, apiBaseUrl string, includeImageSuggestions bool) ([]datastructures.AnnotationTask, error) {
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
                                               )`, len(parseResult.QueryValues) + 1)
    }

    orderBy := ""
    if orderRandomly {
        orderBy = " ORDER BY RANDOM()"
    }

    q2 := ""
    q3 := "acc.name is null"
	if apiUser.Name != "" {	
		q2 = fmt.Sprintf(` AND NOT EXISTS
                           (
                                SELECT 1 FROM user_annotation_blacklist bl 
                                JOIN account acc ON acc.id = bl.account_id
                                WHERE bl.image_validation_id = v.id AND acc.name = $%d
                           )`, len(parseResult.QueryValues) + 1)
    
		q3 = fmt.Sprintf(`acc.name = $%d`, len(parseResult.QueryValues) + 1)
	}

	q4 := ""
	if includeImageSuggestions {
		q4 = `UNION ALL

			  SELECT s.uuid as validation_uuid, l.name as label_accessor,
			  0 as annotated_percentage, s.image_id as image_id, false as is_productive 
			  FROM image_label_suggestion s
			  JOIN label_suggestion l ON l.id = s.label_suggestion_id
			  WHERE NOT EXISTS (
				SELECT 1 FROM image_annotation_suggestion a 
				WHERE a.label_suggestion_id = s.label_suggestion_id AND a.image_id = s.image_id
			  )
			 ` 	
	}

	//in case no subquery is provided, set 1=1 to "catch all". if we won't do that, the query
	//breaks due to a syntax error
	if parseResult.Subquery == "" {
		parseResult.Subquery = "1 = 1"
	}

    q := fmt.Sprintf(`SELECT qqq.image_key, qqq.image_width, qqq.image_height, qqq.validation_uuid, qqq.image_unlocked, 
                      accessor
                      FROM
                      (
                        SELECT qq.image_key, qq.image_width, qq.image_height, 
                        unnest(string_to_array(qq.validation_uuids, ',')) as validation_uuid, qq.image_unlocked, 
						unnest(qq.label_types) as label_types, unnest(qq.filtered_accessors) as accessor
                        FROM
                        (    
                              SELECT q.image_key, q.image_width, q.image_height, q.validation_uuids, 
							  q.image_unlocked, image_collection, q.label_types, q.filtered_accessors
                              FROM
                              (   SELECT i.key as image_key, i.width as image_width, i.height as image_height, 
                                  array_to_string(array_agg(CASE WHEN (%s) THEN a.validation_uuid ELSE NULL END), ',') as validation_uuids, 
                                  array_agg(a.accessor)::text[] as accessors, MAX(COALESCE(a.annotated_percentage, 0)) as annotated_percentage, 
                                  i.unlocked as image_unlocked, coll.image_collection as image_collection, 
								  array_agg(a.is_productive) FILTER(WHERE %s) as label_types,
								  array_agg(a.accessor) FILTER(WHERE %s) as filtered_accessors
                                  FROM image i 

								  JOIN (
                                  	SELECT v.uuid as validation_uuid, a.accessor as accessor, 
									c.annotated_percentage as annotated_percentage, v.image_id as image_id, true as is_productive
									FROM image_validation v
                                  	JOIN label l ON l.id = v.label_id
                                 	JOIN label_accessor a ON l.id = a.label_id
                                  	LEFT JOIN image_annotation_coverage c ON c.image_id = v.image_id
									WHERE NOT EXISTS (
									  SELECT 1 FROM image_annotation a 
									  WHERE a.label_id = v.label_id AND a.image_id = v.image_id
								    )%s


									%s
                                  ) a ON a.image_id = i.id
								  LEFT JOIN 
									(
										SELECT ui.name as image_collection, c.image_id as image_id
										FROM image_collection_image c
										JOIN user_image_collection ui ON c.user_image_collection_id = ui.id
										JOIN account acc ON acc.id = ui.account_id
										WHERE %s
									) coll ON coll.image_id = i.id
								  WHERE (i.unlocked = true %s)

                                  GROUP BY i.key, i.width, i.height, i.unlocked, coll.image_collection
                              ) q WHERE %s
                        )qq
                      ) qqq
                      %s`, parseResult.Subquery, parseResult.Subquery, parseResult.Subquery, 
					  		q2, q4, q3, includeOwnImageDonations, parseResult.Query, orderBy)

    //first item in query value is the label we want to annotate
    //parseResult.queryValues = append([]interface{}{parseResult.queryValues[0]}, parseResult.queryValues...)

    var rows *sql.Rows
    var err error
    if apiUser.Name == "" {
        rows, err = p.db.Query(q, parseResult.QueryValues...)
    } else {
        parseResult.QueryValues = append(parseResult.QueryValues, apiUser.Name)
        rows, err = p.db.Query(q, parseResult.QueryValues...)
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
                            &annotationTask.Id, &annotationTask.Image.Unlocked, &annotationTask.Label.Accessor)
        if err != nil {
            log.Debug("[Annotation Tasks] Couldn't get available annotation tasks: ", err.Error())
            raven.CaptureError(err, nil)
            return annotationTasks, err
        }

        if annotationTask.Id == "" {
            continue
        }

        annotationTask.Image.Url = commons.GetImageUrlFromImageId(apiBaseUrl, annotationTask.Image.Id, annotationTask.Image.Unlocked)

        annotationTasks = append(annotationTasks, annotationTask)
    }

    return annotationTasks, nil
}

func (p *ImageMonkeyDatabase) GetRandomAnnotationForQuizRefinement() (datastructures.AnnotationRefinement, error) {
    var bytes []byte
    var annotationBytes []byte
    var refinement datastructures.AnnotationRefinement
    var annotations []json.RawMessage
    rows, err := p.db.Query(`SELECT i.key, s.quiz_question_id, s.quiz_question, s.quiz_answers, s1.annotations, s.recommended_control::text, 
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
            randomVal := commons.Random(0, (len(annotations) - 1))
            refinement.Annotation.Annotation = annotations[randomVal]
        }
    }

    return refinement, nil
}


func addOrUpdateRefinementsInTransaction(tx *sql.Tx, annotationUuid string, annotationDataId string, 
            annotationRefinementEntries []datastructures.AnnotationRefinementEntry, clientFingerprint string, 
			isSuggestion bool) error {
    for _, item := range annotationRefinementEntries {

        _, err := uuid.FromString(item.LabelId)
        if err != nil {
            tx.Rollback()
            log.Error("[Add or Update annotation refinement] Couldn't add/update refinements - invalid label id")
            raven.CaptureError(err, nil)
            return &InvalidLabelIdError{Description: "invalid label id"}
        }

		if isSuggestion {
			_, err = tx.Exec(`INSERT INTO image_annotation_suggestion_refinement(annotation_suggestion_data_id, 
												label_id, num_of_valid, fingerprint_of_last_modification)
                            SELECT d.id, (SELECT l.id FROM label l WHERE l.uuid = $2), $3, $4 
                            FROM image_annotation_suggestion a 
							JOIN annotation_suggestion_data d ON d.image_annotation_suggestion_id = a.id 
							WHERE a.uuid = $5 AND d.uuid = $1
                          ON CONFLICT (annotation_suggestion_data_id, label_id)
                          DO UPDATE SET fingerprint_of_last_modification = $4, num_of_valid = image_annotation_suggestion_refinement.num_of_valid + 1
                          WHERE image_annotation_suggestion_refinement.annotation_suggestion_data_id = (SELECT d.id FROM annotation_suggestion_data d WHERE d.uuid = $1) 
                          AND image_annotation_suggestion_refinement.label_id = (SELECT l.id FROM label l WHERE l.uuid = $2)`, 
                               annotationDataId, item.LabelId, 1, clientFingerprint, annotationUuid)
		} else {
        	_, err = tx.Exec(`INSERT INTO image_annotation_refinement(annotation_data_id, label_id, num_of_valid, fingerprint_of_last_modification)
                            SELECT d.id, (SELECT l.id FROM label l WHERE l.uuid = $2), $3, $4 
                            FROM image_annotation a 
							JOIN annotation_data d ON d.image_annotation_id = a.id 
							WHERE a.uuid = $5 AND d.uuid = $1
                          ON CONFLICT (annotation_data_id, label_id)
                          DO UPDATE SET fingerprint_of_last_modification = $4, num_of_valid = image_annotation_refinement.num_of_valid + 1
                          WHERE image_annotation_refinement.annotation_data_id = (SELECT d.id FROM annotation_data d WHERE d.uuid = $1) 
                          AND image_annotation_refinement.label_id = (SELECT l.id FROM label l WHERE l.uuid = $2)`, 
                               annotationDataId, item.LabelId, 1, clientFingerprint, annotationUuid)
        }

        if err != nil {
            tx.Rollback()
            log.Error("[Add or Update annotation refinement] Couldn't update: ", err.Error())
            raven.CaptureError(err, nil)
            return err
        }
    }
    return nil
}

func (p *ImageMonkeyDatabase) AnnotationUuidIsASuggestion(annotationUuid string) (bool, error) {
	var isSuggestion bool = false
	err := p.db.QueryRow(`SELECT is_suggestion FROM
					      (
                           SELECT count(*) as count, false as is_suggestion
                           FROM image_annotation a 
                           WHERE a.uuid = $1::uuid

                           UNION ALL

                           SELECT count(*) as count, true as is_suggestion
                           FROM image_annotation_suggestion a
                           WHERE a.uuid = $1::uuid
                          ) q WHERE q.count > 0`, annotationUuid).Scan(&isSuggestion)
	if err != nil {
		return isSuggestion, err
	}

	return isSuggestion, nil
}

func (p *ImageMonkeyDatabase) AddOrUpdateRefinements(annotationUuid string, annotationDataId string, 
			annotationRefinementEntries []datastructures.AnnotationRefinementEntry, clientFingerprint string,
			isSuggestion bool) error {
    var err error

    tx, err := p.db.Begin()
    if err != nil {
        log.Error("[Add or Update annotation refinement] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }	

    err = addOrUpdateRefinementsInTransaction(tx, annotationUuid, annotationDataId, annotationRefinementEntries, clientFingerprint, isSuggestion)
    if err != nil { //transaction already rolled back, so we can return here
        return err
    }

    err = tx.Commit()
    if err != nil {
        log.Error("[Add or Update annotation refinement] Couldn't commit transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    return nil
}


func (p *ImageMonkeyDatabase) BatchAnnotationRefinement(annotationRefinementEntries []datastructures.BatchAnnotationRefinementEntry, 
		apiUser datastructures.APIUser) error {
    var err error

    tx, err := p.db.Begin()
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


func (p *ImageMonkeyDatabase) GetImagesForAutoAnnotation(labels []string) ([]datastructures.AutoAnnotationImage, error) {
    var autoAnnotationImages []datastructures.AutoAnnotationImage
    rows, err := p.db.Query(`SELECT i.key, i.width, i.height, json_agg(l.name)  FROM image i 
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

func (p *ImageMonkeyDatabase) GetBoundingBoxesForImageLabel(imageId string, label string) ([]image.Rectangle, error) {
    boundingBoxes := []image.Rectangle{}

    query := `WITH all_annotations AS (
                SELECT an.image_id as image_id, d.id as annotation_data_id, d.annotation as annotation, t.name as annotation_type
                FROM image_annotation an 
                JOIN annotation_data d ON d.image_annotation_id = an.id
                JOIN annotation_type t ON t.id = d.annotation_type_id
                JOIN image i ON i.id = an.image_id
                JOIN label_accessor acc ON acc.label_id = an.label_id
                WHERE i.key = $1 AND acc.accessor = $2
            ),
            ellipse_annotations AS (
                SELECT a.image_id, a.annotation_data_id as id, 
                ST_Envelope(Ellipse( (a.annotation->'left')::text::float, 
                         (a.annotation->'top')::text::float, 
                         2* (a.annotation->'rx')::text::float, 
                         2* (a.annotation->'ry')::text::float, 
                         CASE 
                            WHEN a.annotation->'angle' is null THEN 0 
                            ELSE (a.annotation->'angle')::text::float
                         END
                       )) as geom
                FROM all_annotations a
                WHERE annotation_type = 'ellipse'
            ),
            polygon_annotations AS (
              -- ST_MakePolygon might return a polygon with intersecting points. In order to fix that, one needs to call ST_MakeValid on the resulting polygon.
              --Unfortunately, this is _really_ slow (especially, if a lot of polygons are affected). In order to circumvent that, we create a ConvexHull around the
              --polygon. This works way faster and should also be precise enough for our purpose.
                SELECT q.image_id, q.annotation_data_id as id, ST_Envelope(ST_ConvexHull(ST_MakePolygon(ST_GeomFromText('LINESTRING(' || 
                                                                              string_agg((((q.annotation->'x')::text) || ' ' || ((q.annotation->'y')::text)), ',') 
                                                                              || ',' || (array_agg((q.annotation->'x')::text))[1] || ' ' || (array_agg((q.annotation->'y')::text))[1] 
                                                                              || ')')))) as geom
                FROM
                (
                    SELECT a.image_id, a.annotation_data_id, jsonb_array_elements(a.annotation->'points') as  annotation
                    FROM all_annotations a 
                    WHERE a.annotation_type = 'polygon' AND jsonb_array_length(a.annotation->'points') > 2
                ) q
                GROUP BY q.image_id, q.annotation_data_id
            ),
            rectangle_annotations AS (
                SELECT a.image_id, a.annotation_data_id as id, ST_Envelope(ST_MakePolygon(ST_MakeLine(
                   ARRAY[
                         ST_MakePoint((a.annotation->'left')::text::integer, (a.annotation->'top')::text::integer), 
                         ST_MakePoint((a.annotation->'left')::text::float + (a.annotation->'width')::text::float, (a.annotation->'top')::text::float),
                         ST_MakePoint((a.annotation->'left')::text::float + (a.annotation->'width')::text::float, 
                                                                (a.annotation->'top')::text::float + (a.annotation->'height')::text::float),
                         ST_MakePoint((a.annotation->'left')::text::float, (a.annotation->'top')::text::float + (a.annotation->'height')::text::float),
                         ST_MakePoint((a.annotation->'left')::text::float, (a.annotation->'top')::text::float)
                        ]))) as geom
                FROM all_annotations a 
                WHERE a.annotation_type = 'rect'
                --GROUP BY a.annotation_data_id, a.annotation
            ),
            all_annotation_areas AS (
                SELECT ST_AsGeoJSON(geom)::jsonb AS geom FROM polygon_annotations
                UNION 
                SELECT ST_AsGeoJSON(geom)::jsonb AS geom FROM rectangle_annotations
                UNION
                SELECT ST_AsGeoJSON(geom)::jsonb AS geom FROM ellipse_annotations
            )
            SELECT ((geom->'coordinates'->>0)::jsonb->>0)::jsonb->0 as x0,
                   ((geom->'coordinates'->>0)::jsonb->>0)::jsonb->1 as y0,
                   ((geom->'coordinates'->>0)::jsonb->>2)::jsonb->0 as x1,
                   ((geom->'coordinates'->>0)::jsonb->>2)::jsonb->1 as y1
            FROM all_annotation_areas`
    rows, err := p.db.Query(query, imageId, label)
    if err != nil {
        log.Error("[Get Bounding Boxes] Couldn't get bounding boxes for image label: ", err.Error())
        raven.CaptureError(err, nil)
        return boundingBoxes, err
    }

    defer rows.Close()

    for rows.Next() {
        var x0, y0, x1, y1 int
        err = rows.Scan(&x0, &y0, &x1, &y1)
        if err != nil {
            log.Error("[Get Bounding Boxes] Couldn't scan bounding boxes for image label: ", err.Error())
            raven.CaptureError(err, nil)
            return boundingBoxes, err
        }
        boundingBoxes = append(boundingBoxes, image.Rect(x0, y0, x1, y1))
    }

    return boundingBoxes, nil
}
