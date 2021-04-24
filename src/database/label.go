package imagemonkeydb

import (
	"context"
	"encoding/json"
	"fmt"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	"github.com/getsentry/raven-go"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
)

func _getTotalLabelSuggestions(tx pgx.Tx) (int64, error) {
	var numOfTotalLabelSuggestions int64
	numOfTotalLabelSuggestions = 0

	rows, err := tx.Query(context.TODO(), `SELECT count(*) FROM label_suggestion l`)
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

func (p *ImageMonkeyDatabase) GetMostPopularLabels(limit int32) ([]string, error) {
	var labels []string

	rows, err := p.db.Query(context.TODO(),
		`SELECT l.name FROM image_validation v 
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

func (p *ImageMonkeyDatabase) AddLabelSuggestion(suggestedLabel string) error {
	_, err := p.db.Exec(context.TODO(),
		`INSERT INTO label_suggestion(name, uuid)
	 						SELECT $1, uuid_generate_v4()
                       ON CONFLICT (name) DO NOTHING`, suggestedLabel)
	if err != nil {
		log.Debug("[Add label suggestion] Couldn't insert: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	return nil
}

func (p *ImageMonkeyDatabase) GetLabelCategories() ([]string, error) {
	var labels []string
	rows, err := p.db.Query(context.TODO(),
		`SELECT pl.name 
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

func (p *ImageMonkeyDatabase) GetLabelSuggestions(includeUnlocked bool) ([]string, error) {
	labelSuggestions := []string{}

	q1 := ""
	if !includeUnlocked {
		q1 = `LEFT JOIN trending_label_suggestion t ON t.label_suggestion_id = l.id
			  WHERE t.closed = false OR t.closed is null`
	}

	q := fmt.Sprintf(`SELECT l.name FROM label_suggestion l %s`, q1)

	rows, err := p.db.Query(context.TODO(), q)
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

func (p *ImageMonkeyDatabase) GetLabelSuggestionsUsage() ([]datastructures.ImageLabelSuggestion, error) {
	labelSuggestions := []datastructures.ImageLabelSuggestion{}

	q := fmt.Sprintf(`WITH num_labels AS (
						SELECT l.id AS label_suggestion_id, l.name AS label_name, count(*) AS num_labels
						FROM image_label_suggestion i
						JOIN label_suggestion l ON l.id = i.label_suggestion_id
						GROUP BY l.id
					), num_annotations AS (
						SELECT l.id AS label_suggestion_id, count(*) AS num_annotations
						FROM image_annotation_suggestion i
						JOIN label_suggestion l ON l.id = i.label_suggestion_id
						GROUP BY l.id
					)
					SELECT l.name, COALESCE(n1.num_labels, 0), COALESCE(n2.num_annotations, 0)
					FROM label_suggestion l
					LEFT JOIN num_labels n1 ON n1.label_suggestion_id = l.id
					LEFT JOIN num_annotations n2 ON n2.label_suggestion_id = l.id`)

	rows, err := p.db.Query(context.TODO(), q)
	if err != nil {
		log.Error("[Get Label Suggestions Usage] Couldn't get label suggestions: ", err.Error())
		raven.CaptureError(err, nil)
		return labelSuggestions, err
	}

	defer rows.Close()

	for rows.Next() {
		var labelSuggestion datastructures.ImageLabelSuggestion
		err := rows.Scan(&labelSuggestion.Name, &labelSuggestion.NumOfLabels, &labelSuggestion.NumOfAnnotations)
		if err != nil {
			log.Error("[Get Label Suggestions Usage] Couldn't scan label suggestions: ", err.Error())
			raven.CaptureError(err, nil)
			return labelSuggestions, err
		}

		labelSuggestions = append(labelSuggestions, labelSuggestion)
	}

	return labelSuggestions, nil
}

func (p *ImageMonkeyDatabase) GetLabelAccessors() ([]string, error) {
	var labels []string
	rows, err := p.db.Query(context.TODO(), `SELECT accessor FROM label_accessor`)
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

func (p *ImageMonkeyDatabase) GetLabelAccessorDetails(labelType string) ([]datastructures.LabelAccessorDetail, error) {
	var labelAccessorDetails []datastructures.LabelAccessorDetail
	var queryValues []interface{}

	q1 := ""
	if labelType != "" {
		q1 = "WHERE l.label_type = $1"
		queryValues = append(queryValues, labelType)
	}

	query := fmt.Sprintf(`SELECT COALESCE(json_agg(json_build_object('accessor', acc.accessor, 'parent_accessor', pacc.accessor)) 
                                            FILTER (WHERE l.id is not null), '[]'::json)
                             FROM label_accessor acc 
                             JOIN label l ON l.id = acc.label_id
                             LEFT JOIN label pl ON pl.id = l.parent_id
                             LEFT JOIN label_accessor pacc ON pl.id = pacc.label_id
                             %s`, q1)

	rows, err := p.db.Query(context.TODO(), query, queryValues...)
	if err != nil {
		log.Error("[Get detailed label accessors] Couldn't get accessors: ", err.Error())
		raven.CaptureError(err, nil)
		return labelAccessorDetails, err
	}
	defer rows.Close()

	if rows.Next() {
		var bytes []byte
		err = rows.Scan(&bytes)
		if err != nil {
			log.Error("[Get detailed label accessors] Couldn't scan row: ", err.Error())
			raven.CaptureError(err, nil)
			return labelAccessorDetails, err
		}

		err = json.Unmarshal(bytes, &labelAccessorDetails)
		if err != nil {
			log.Error("[Get detailed label accessors] Couldn't unmarshal result: ", err.Error())
			raven.CaptureError(err, nil)
			return labelAccessorDetails, err
		}
	}

	return labelAccessorDetails, nil
}

func (p *ImageMonkeyDatabase) GetLabelAccessorsMapping() (json.RawMessage, error) {
	var res json.RawMessage

	rows, err := p.db.Query(context.TODO(),
		`SELECT json_object_agg(a.accessor, CASE 
                                    WHEN pl.name is not null THEN 
                                        (json_build_object('label', pl.name, 'sublabel', l.name)) 
                                    ELSE 
                                        (json_build_object('label', l.name, 'sublabel', pl.name)) 
                                        END)
                             FROM label_accessor a 
                             JOIN label l ON l.id = a.label_id
                             LEFT JOIN label pl ON pl.id = l.parent_id
                             WHERE l.label_type = 'normal' OR l.label_type = 'meta'`)
	if err != nil {
		log.Error("[Get Label Accessors Mapping] Couldn't get label accessors mapping: ", err.Error())
		raven.CaptureError(err, nil)
		return res, err
	}

	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&res)
		if err != nil {
			log.Error("[Get Label Accessors Mapping] Couldn't scan row: ", err.Error())
			raven.CaptureError(err, nil)
			return res, err
		}
	}
	return res, nil
}
