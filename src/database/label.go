package imagemonkeydb

import (
    "github.com/getsentry/raven-go"
    log "github.com/Sirupsen/logrus"
    "database/sql"
    "../datastructures"
    "encoding/json"
    "fmt"
)

func _getTotalLabelSuggestions(tx *sql.Tx) (int64, error) {
    var numOfTotalLabelSuggestions int64
    numOfTotalLabelSuggestions = 0

    rows, err := tx.Query(`SELECT count(*) FROM label_suggestion l`)
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

    rows, err := p.db.Query(`SELECT l.name FROM image_validation v 
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
     _, err := p.db.Exec(`INSERT INTO label_suggestion(name) VALUES($1)
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
    rows, err := p.db.Query(`SELECT pl.name 
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

func (p *ImageMonkeyDatabase) GetLabelSuggestions() ([]string, error) {
    var labelSuggestions []string

    rows, err := p.db.Query("SELECT name FROM label_suggestion")
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

func (p *ImageMonkeyDatabase) GetLabelAccessors() ([]string, error) {
    var labels []string
    rows, err := p.db.Query(`SELECT accessor FROM label_accessor`)
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

    rows, err := p.db.Query(query, queryValues...)
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