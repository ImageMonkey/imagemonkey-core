package imagemonkeydb

import (
    "../datastructures"
    "github.com/getsentry/raven-go"
    log "github.com/Sirupsen/logrus"
    "errors"
    "database/sql"
    "fmt"
    "encoding/json"
)


func (p *ImageMonkeyDatabase) GetNumOfDonatedImages() (int64, error) {
    var num int64
    err := p.db.QueryRow("SELECT count(*) FROM image").Scan(&num)
    if err != nil {
        log.Debug("[Fetch images] Couldn't get num of available images: ", err.Error())
        raven.CaptureError(err, nil)
        return 0, err
    }

    return num, nil
}

func (p *ImageMonkeyDatabase) GetImageDescriptionStatistics(period string) ([]datastructures.DataPoint, error) {
    imageDescriptionStatistics := []datastructures.DataPoint{}

    if period != "last-month" {
        return imageDescriptionStatistics, errors.New("Only last-month statistics are supported at the moment")
    }

    rows, err := p.db.Query(`WITH dates AS (
                            SELECT *
                            FROM generate_series((CURRENT_DATE - interval '1 month'), CURRENT_DATE, '1 day') date
                           ),
                           num_of_image_descriptions AS (
                            SELECT sys_period FROM image_description_history h
                            WHERE date(lower(h.sys_period)) IN (SELECT date FROM dates)
                            UNION ALL 
                            SELECT sys_period FROM image_description h1
                            WHERE date(lower(h1.sys_period)) IN (SELECT date FROM dates)
                           )
                          SELECT to_char(date(date), 'YYYY-MM-DD'),
                           ( SELECT count(*) FROM num_of_image_descriptions s
                             WHERE date(lower(s.sys_period)) = date(date) 
                           ) as num
                           FROM dates
                           GROUP BY date
                           ORDER BY date`)
    if err != nil {
        log.Debug("[Get Statistics] Couldn't get image description statistics: ", err.Error())
        raven.CaptureError(err, nil)
        return imageDescriptionStatistics, err
    }

    defer rows.Close()

    for rows.Next() {
        var datapoint datastructures.DataPoint
        err = rows.Scan(&datapoint.Date, &datapoint.Value)
        if err != nil {
            log.Debug("[Get Statistics] Couldn't scan image description row: ", err.Error())
            raven.CaptureError(err, nil)
            return imageDescriptionStatistics, err
        }

        imageDescriptionStatistics = append(imageDescriptionStatistics, datapoint)
    }

    return imageDescriptionStatistics, nil
}


func (p *ImageMonkeyDatabase) GetAnnotationStatistics(period string) ([]datastructures.DataPoint, error) {
    annotationStatistics := []datastructures.DataPoint{}

    if period != "last-month" {
        return annotationStatistics, errors.New("Only last-month statistics are supported at the moment")
    }

    rows, err := p.db.Query(`WITH dates AS (
                            SELECT *
                            FROM generate_series((CURRENT_DATE - interval '1 month'), CURRENT_DATE, '1 day') date
                           ),
                           num_of_annotations AS (
                            SELECT sys_period FROM image_annotation_history h
                            WHERE date(lower(h.sys_period)) IN (SELECT date FROM dates)
                            UNION ALL 
                            SELECT sys_period FROM image_annotation h1
                            WHERE date(lower(h1.sys_period)) IN (SELECT date FROM dates)
                           )
                          SELECT to_char(date(date), 'YYYY-MM-DD'),
                           ( SELECT count(*) FROM num_of_annotations s
                             WHERE date(lower(s.sys_period)) = date(date) 
                           ) as num
                           FROM dates
                           GROUP BY date
                           ORDER BY date`)
    if err != nil {
        log.Debug("[Get Statistics] Couldn't get statistics: ", err.Error())
        raven.CaptureError(err, nil)
        return annotationStatistics, err
    }

    defer rows.Close()

    for rows.Next() {
        var datapoint datastructures.DataPoint
        err = rows.Scan(&datapoint.Date, &datapoint.Value)
        if err != nil {
            log.Debug("[Get Statistics] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return annotationStatistics, err
        }

        annotationStatistics = append(annotationStatistics, datapoint)
    }

    return annotationStatistics, nil
}

func (p *ImageMonkeyDatabase) GetValidationStatistics(period string) ([]datastructures.DataPoint, error) {
    validationStatistics := []datastructures.DataPoint{}

    if period != "last-month" {
        return validationStatistics, errors.New("Only last-month statistics are supported at the moment")
    }

    rows, err := p.db.Query(`WITH dates AS (
                            SELECT *
                            FROM generate_series((CURRENT_DATE - interval '1 month'), CURRENT_DATE, '1 day') date
                           ),
                           num_of_validations AS (
                            SELECT sys_period FROM image_validation_history h
                            WHERE date(lower(h.sys_period)) IN (SELECT date FROM dates)

                            UNION ALL

                            SELECT sys_period FROM image_validation v
                            WHERE date(lower(v.sys_period)) IN (SELECT date FROM dates) AND 
                            (v.num_of_valid > 0 OR v.num_of_invalid > 0)

                           )
                          SELECT to_char(date(date), 'YYYY-MM-DD'),
                           ( SELECT count(*) FROM num_of_validations s
                             WHERE date(lower(s.sys_period)) = date(date) 
                           ) as num
                           FROM dates
                           GROUP BY date
                           ORDER BY date`)
    if err != nil {
        log.Debug("[Get Statistics] Couldn't get statistics: ", err.Error())
        raven.CaptureError(err, nil)
        return validationStatistics, err
    }

    defer rows.Close()

    for rows.Next() {
        var datapoint datastructures.DataPoint
        err = rows.Scan(&datapoint.Date, &datapoint.Value)
        if err != nil {
            log.Debug("[Get Statistics] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return validationStatistics, err
        }

        validationStatistics = append(validationStatistics, datapoint)
    }

    return validationStatistics, nil
}

func (p *ImageMonkeyDatabase) GetLabeledObjectsStatistics(period string) ([]datastructures.DataPoint, error) {
    labeledObjectsStatistics := []datastructures.DataPoint{}

    if period != "last-month" {
        return labeledObjectsStatistics, errors.New("Only last-month statistics are supported at the moment")
    }

    rows, err := p.db.Query(`WITH dates AS (
                            SELECT *
                            FROM generate_series((CURRENT_DATE - interval '1 month'), CURRENT_DATE, '1 day') date
                           ),
                           num_of_validations AS (
                            SELECT sys_period FROM image_validation h
                            WHERE date(lower(h.sys_period)) IN (SELECT date FROM dates)
                            AND (num_of_valid = 0 AND num_of_invalid = 0)

                            UNION ALL
                            SELECT sys_period FROM image_validation_history h1
                            WHERE date(lower(h1.sys_period)) IN (SELECT date FROM dates)
                            AND (num_of_valid = 0 AND num_of_invalid = 0)

                            UNION ALL
                            SELECT sys_period FROM image_label_suggestion s
                            WHERE date(lower(s.sys_period)) IN (SELECT date FROM dates)
                           )
                          SELECT to_char(date(date), 'YYYY-MM-DD'),
                           ( SELECT count(*) FROM num_of_validations s
                             WHERE date(lower(s.sys_period)) = date(date) 
                           ) as num
                           FROM dates
                           GROUP BY date
                           ORDER BY date`)
    if err != nil {
        log.Error("[Get Label Statistics] Couldn't get statistics: ", err.Error())
        raven.CaptureError(err, nil)
        return labeledObjectsStatistics, err
    }

    defer rows.Close()

    for rows.Next() {
        var datapoint datastructures.DataPoint
        err = rows.Scan(&datapoint.Date, &datapoint.Value)
        if err != nil {
            log.Error("[Get Label Statistics] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return labeledObjectsStatistics, err
        }

        labeledObjectsStatistics = append(labeledObjectsStatistics, datapoint)
    }

    return labeledObjectsStatistics, nil
}


func (p *ImageMonkeyDatabase) GetAnnotationRefinementStatistics(period string) ([]datastructures.DataPoint, error) {
    var annotationRefinementStatistics []datastructures.DataPoint

    if period != "last-month" {
        return annotationRefinementStatistics, errors.New("Only last-month statistics are supported at the moment")
    }

    rows, err := p.db.Query(`WITH dates AS (
                            SELECT *
                            FROM generate_series((CURRENT_DATE - interval '1 month'), CURRENT_DATE, '1 day') date
                           ),
                           num_of_annotation_refinements AS (
                            SELECT sys_period FROM image_annotation_refinement_history h
                            WHERE date(lower(h.sys_period)) IN (SELECT date FROM dates)
                            UNION ALL 
                            SELECT sys_period FROM image_annotation_refinement h1
                            WHERE date(lower(h1.sys_period)) IN (SELECT date FROM dates)
                           )
                          SELECT to_char(date(date), 'YYYY-MM-DD'),
                           ( SELECT count(*) FROM num_of_annotation_refinements s
                             WHERE date(lower(s.sys_period)) = date(date) 
                           ) as num
                           FROM dates
                           GROUP BY date
                           ORDER BY date`)
    if err != nil {
        log.Debug("[Get Annotation Refinement Statistics] Couldn't get statistics: ", err.Error())
        raven.CaptureError(err, nil)
        return annotationRefinementStatistics, err
    }

    defer rows.Close()

    for rows.Next() {
        var datapoint datastructures.DataPoint
        err = rows.Scan(&datapoint.Date, &datapoint.Value)
        if err != nil {
            log.Debug("[Get Annotation Refinement Statistics] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return annotationRefinementStatistics, err
        }

        annotationRefinementStatistics = append(annotationRefinementStatistics, datapoint)
    }

    return annotationRefinementStatistics, nil
}


func (p *ImageMonkeyDatabase) GetUserStatistics(username string) (datastructures.UserStatistics, error) {
    var userStatistics datastructures.UserStatistics

    tx, err := p.db.Begin()
    if err != nil {
        log.Debug("[User Statistics] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return userStatistics, err
    }

    userStatistics.Total.Annotations = 0
    err = tx.QueryRow("SELECT count(*) FROM image_annotation").Scan(&userStatistics.Total.Annotations)
    if err != nil {
        tx.Rollback()
        log.Debug("[User Statistics] Couldn't get total annotations: ", err.Error())
        raven.CaptureError(err, nil)
        return userStatistics, err
    }


    userStatistics.User.Annotations = 0
    err = tx.QueryRow(`SELECT count(*) FROM user_image_annotation u
                       JOIN account a on u.account_id = a.id WHERE a.name = $1`, username).Scan(&userStatistics.User.Annotations)
    if err != nil {
        tx.Rollback()
        log.Debug("[User Statistics] Couldn't get user annotations: ", err.Error())
        raven.CaptureError(err, nil)
        return userStatistics, err
    }


    userStatistics.Total.Validations = 0
    err = tx.QueryRow("SELECT count(*) FROM image_validation").Scan(&userStatistics.Total.Validations)
    if err != nil {
        tx.Rollback()
        log.Debug("[User Statistics] Couldn't get total validations: ", err.Error())
        raven.CaptureError(err, nil)
        return userStatistics, err
    }

    userStatistics.User.Validations = 0
    err = tx.QueryRow(`SELECT count(*) FROM user_image_validation u
                       JOIN account a on u.account_id = a.id WHERE a.name = $1`, username).Scan(&userStatistics.User.Validations)
    if err != nil {
        tx.Rollback()
        log.Debug("[User Statistics] Couldn't get user validations: ", err.Error())
        raven.CaptureError(err, nil)
        return userStatistics, err
    }


    err = tx.Commit()
    if err != nil {
        log.Debug("[User Statistics] Couldn't commit transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return userStatistics, err
    }


    return userStatistics, nil
}

func (p *ImageMonkeyDatabase) Explore(words []string) (datastructures.Statistics, error) {
    statistics := datastructures.Statistics{}

    //use temporary map for faster lookup
    temp := make(map[string]datastructures.ValidationStat)

    tx, err := p.db.Begin()
    if err != nil {
        log.Debug("[Explore] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }
    
    rows, err := tx.Query(`SELECT CASE WHEN pl.name is null THEN l.name ELSE l.name || '/' || pl.name END, count(l.name), 
                           CASE 
                            WHEN SUM(v.num_of_valid + v.num_of_invalid) = 0 THEN 0 
                            ELSE (CAST (SUM(v.num_of_invalid) AS float)/(SUM(v.num_of_valid) + SUM(v.num_of_invalid))) 
                           END as error_rate, 
                           SUM(v.num_of_valid + v.num_of_invalid) as total_validations
                           FROM image_validation v 
                           JOIN label l ON v.label_id = l.id 
                           LEFT JOIN label pl on l.parent_id = pl.id
                           GROUP BY l.name, pl.name ORDER BY count(l.name) DESC`)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore data: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }
    defer rows.Close()

    for rows.Next() {
        var validationStat datastructures.ValidationStat
        err = rows.Scan(&validationStat.Label, &validationStat.Count, &validationStat.ErrorRate, &validationStat.TotalValidations)
        if err != nil {
            tx.Rollback()
            log.Debug("[Explore] Couldn't scan data row: ", err.Error())
            raven.CaptureError(err, nil)
            return statistics, err
        }

        temp[validationStat.Label] = validationStat
    }

    //add labels where we don't have a donation yet
    for _, value := range words {
        _, contains := temp[value]
        if !contains {
            var validationStat datastructures.ValidationStat
            validationStat.Label = value
            validationStat.Count = 0
            temp[value] = validationStat
        }
    }

    for _, value := range temp {
        statistics.Validations = append(statistics.Validations, value)
    }

    //get donations grouped by country
    donationsPerCountryRows, err := tx.Query(`SELECT country_code, count FROM donations_per_country ORDER BY count DESC`)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore data: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }
    defer donationsPerCountryRows.Close()

    for donationsPerCountryRows.Next() {
        var donationsPerCountryStat datastructures.DonationsPerCountryStat
        err = donationsPerCountryRows.Scan(&donationsPerCountryStat.CountryCode, &donationsPerCountryStat.Count)
        if err != nil {
            tx.Rollback()
            log.Debug("[Explore] Couldn't scan data row: ", err.Error())
            raven.CaptureError(err, nil)
            return statistics, err
        }

        statistics.DonationsPerCountry = append(statistics.DonationsPerCountry, donationsPerCountryStat)
    }


    //get validations grouped by country
    validationsPerCountryRows, err := tx.Query(`SELECT country_code, count FROM validations_per_country ORDER BY count DESC`)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore data: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }
    defer validationsPerCountryRows.Close()

    for validationsPerCountryRows.Next() {
        var validationsPerCountryStat datastructures.ValidationsPerCountryStat
        err = validationsPerCountryRows.Scan(&validationsPerCountryStat.CountryCode, &validationsPerCountryStat.Count)
        if err != nil {
            tx.Rollback()
            log.Debug("[Explore] Couldn't scan data row: ", err.Error())
            raven.CaptureError(err, nil)
            return statistics, err
        }

        statistics.ValidationsPerCountry = append(statistics.ValidationsPerCountry, validationsPerCountryStat)
    }

    //get annotations grouped by country
    annotationsPerCountryRows, err := tx.Query(`SELECT country_code, count FROM annotations_per_country ORDER BY count DESC`)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore data: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }
    defer annotationsPerCountryRows.Close()

    for annotationsPerCountryRows.Next() {
        var annotationsPerCountryStat datastructures.AnnotationsPerCountryStat
        err = annotationsPerCountryRows.Scan(&annotationsPerCountryStat.CountryCode, &annotationsPerCountryStat.Count)
        if err != nil {
            tx.Rollback()
            log.Debug("[Explore] Couldn't scan data row: ", err.Error())
            raven.CaptureError(err, nil)
            return statistics, err
        }

        statistics.AnnotationsPerCountry = append(statistics.AnnotationsPerCountry, annotationsPerCountryStat)
    }


    //get annotation refinements grouped by country
    annotationRefinementsPerCountryRows, err := tx.Query(`SELECT country_code, count FROM annotation_refinements_per_country ORDER BY count DESC`)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore data: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }
    defer annotationRefinementsPerCountryRows.Close()

    for annotationRefinementsPerCountryRows.Next() {
        var annotationRefinementsPerCountryStat datastructures.AnnotationRefinementsPerCountryStat
        err = annotationRefinementsPerCountryRows.Scan(&annotationRefinementsPerCountryStat.CountryCode, &annotationRefinementsPerCountryStat.Count)
        if err != nil {
            tx.Rollback()
            log.Debug("[Explore] Couldn't scan data row: ", err.Error())
            raven.CaptureError(err, nil)
            return statistics, err
        }

        statistics.AnnotationRefinementsPerCountry = append(statistics.AnnotationRefinementsPerCountry, annotationRefinementsPerCountryStat)
    }


    //get image descriptions grouped by country
    imageDescriptionsPerCountryRows, err := tx.Query(`SELECT country_code, count FROM image_descriptions_per_country ORDER BY count DESC`)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore data: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }
    defer imageDescriptionsPerCountryRows.Close()

    for imageDescriptionsPerCountryRows.Next() {
        var imageDescriptionsPerCountryStat datastructures.ImageDescriptionsPerCountryStat
        err = imageDescriptionsPerCountryRows.Scan(&imageDescriptionsPerCountryStat.CountryCode, &imageDescriptionsPerCountryStat.Count)
        if err != nil {
            tx.Rollback()
            log.Debug("[Explore] Couldn't scan data row: ", err.Error())
            raven.CaptureError(err, nil)
            return statistics, err
        }

        statistics.ImageDescriptionsPerCountry = append(statistics.ImageDescriptionsPerCountry, imageDescriptionsPerCountryStat)
    }

    //get all unlabeled donations
    err = tx.QueryRow(`SELECT count(i.id) from image i 
                        WHERE i.id NOT IN 
                        (
                            SELECT image_id FROM image_validation
                        ) AND i.id NOT IN (
                            SELECT image_id FROM image_label_suggestion
                        )`).Scan(&statistics.NumOfUnlabeledDonations)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't scan data row: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    statistics.AnnotationsPerApp, err = _exploreAnnotationsPerApp(tx)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore annotations per app: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    statistics.DonationsPerApp, err = _exploreDonationsPerApp(tx)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore donations per app: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    statistics.ValidationsPerApp, err = _exploreValidationsPerApp(tx)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't explore validations per app: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    statistics.NumOfDonations, err = _getTotalDonations(tx)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't get total donations: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    statistics.NumOfAnnotations, err = _getTotalAnnotations(tx, false)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't get total annotations: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    statistics.NumOfValidations, err = _getTotalValidations(tx)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't get total validations: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    statistics.NumOfAnnotationRefinements, err = _getTotalAnnotationRefinements(tx)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't get total annotation refinements: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    statistics.NumOfLabelSuggestions, err = _getTotalLabelSuggestions(tx)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't get total label suggestions: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    statistics.NumOfLabels, err = _getTotalLabels(tx)
    if err != nil {
        tx.Rollback()
        log.Debug("[Explore] Couldn't get total labels: ", err.Error())
        raven.CaptureError(err, nil)
        return statistics, err
    }

    return statistics, tx.Commit()
}

func _getTotalDonations(tx *sql.Tx) (int64, error) {
    var numOfTotalDonations int64
    numOfTotalDonations = 0

    rows, err := tx.Query(`SELECT count(*) FROM image i`)
    if err != nil {
        return numOfTotalDonations, nil
    }

    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&numOfTotalDonations)
        if err != nil {
            return numOfTotalDonations, err
        }
    }

    return numOfTotalDonations, nil
}

func _getTotalLabels(tx *sql.Tx) (int64, error) {
    var numOfTotalLabels int64
    numOfTotalLabels = 0

    rows, err := tx.Query(`SELECT count(*) FROM label l`)
    if err != nil {
        return numOfTotalLabels, nil
    }

    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&numOfTotalLabels)
        if err != nil {
            return numOfTotalLabels, err
        }
    }

    return numOfTotalLabels, nil
}

func _getTotalAnnotations(tx *sql.Tx, includeAutoGeneratedAnnotations bool) (int64, error) {
    var numOfAnnotations int64
    numOfAnnotations = 0

    q1 := ""
    if !includeAutoGeneratedAnnotations {
        q1 = "WHERE a.auto_generated = false"
    }

    q := fmt.Sprintf(`SELECT count(*) FROM image_annotation a %s`, q1)

    rows, err := tx.Query(q)
    if err != nil {
        return numOfAnnotations, nil
    }

    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&numOfAnnotations)
        if err != nil {
            return numOfAnnotations, err
        }
    }

    return numOfAnnotations, nil
}

func _getTotalValidations(tx *sql.Tx) (int64, error) {
    var numOfValidations int64
    numOfValidations = 0

    rows, err := tx.Query(`SELECT count(*) FROM image_validation v`)
    if err != nil {
        return numOfValidations, nil
    }

    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&numOfValidations)
        if err != nil {
            return numOfValidations, err
        }
    }

    return numOfValidations, nil
}

func _getTotalAnnotationRefinements(tx *sql.Tx) (int64, error) {
    var numOfAnnotationRefinements int64
    numOfAnnotationRefinements = 0

    rows, err := tx.Query(`SELECT count(*) FROM image_annotation_refinement r`)
    if err != nil {
        return numOfAnnotationRefinements, nil
    }

    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&numOfAnnotationRefinements)
        if err != nil {
            return numOfAnnotationRefinements, err
        }
    }

    return numOfAnnotationRefinements, nil
}

func _exploreAnnotationsPerApp(tx *sql.Tx) ([]datastructures.AnnotationsPerAppStat, error) {
    var annotationsPerApp []datastructures.AnnotationsPerAppStat

    //get annotations grouped by app
    annotationsPerAppRows, err := tx.Query(`SELECT app_identifier, count FROM annotations_per_app ORDER BY count DESC`)
    if err != nil {
        return annotationsPerApp, err
    }
    defer annotationsPerAppRows.Close()

    for annotationsPerAppRows.Next() {
        var annotationsPerAppStat datastructures.AnnotationsPerAppStat
        err = annotationsPerAppRows.Scan(&annotationsPerAppStat.AppIdentifier, &annotationsPerAppStat.Count)
        if err != nil {
            return annotationsPerApp, err
        }

        annotationsPerApp = append(annotationsPerApp, annotationsPerAppStat)
    }

    return annotationsPerApp, nil
}

func _exploreDonationsPerApp(tx *sql.Tx) ([]datastructures.DonationsPerAppStat, error) {
    var donationsPerApp []datastructures.DonationsPerAppStat

    //get donations grouped by app
    donationsPerAppRows, err := tx.Query(`SELECT app_identifier, count FROM donations_per_app ORDER BY count DESC`)
    if err != nil {
        return donationsPerApp, err
    }
    defer donationsPerAppRows.Close()

    for donationsPerAppRows.Next() {
        var donationsPerAppStat datastructures.DonationsPerAppStat
        err = donationsPerAppRows.Scan(&donationsPerAppStat.AppIdentifier, &donationsPerAppStat.Count)
        if err != nil {
            return donationsPerApp, err
        }

        donationsPerApp = append(donationsPerApp, donationsPerAppStat)
    }

    return donationsPerApp, nil
}

func _exploreValidationsPerApp(tx *sql.Tx) ([]datastructures.ValidationsPerAppStat, error) {
    var validationsPerApp []datastructures.ValidationsPerAppStat

    //get validations grouped by app
    validationsPerAppRows, err := tx.Query(`SELECT app_identifier, count FROM validations_per_app ORDER BY count DESC`)
    if err != nil {
        return validationsPerApp, err
    }
    defer validationsPerAppRows.Close()

    for validationsPerAppRows.Next() {
        var validationsPerAppStat datastructures.ValidationsPerAppStat
        err = validationsPerAppRows.Scan(&validationsPerAppStat.AppIdentifier, &validationsPerAppStat.Count)
        if err != nil {
            return validationsPerApp, err
        }

        validationsPerApp = append(validationsPerApp, validationsPerAppStat)
    }

    return validationsPerApp, nil
}

func (p *ImageMonkeyDatabase) UpdateContributionsPerCountry(contributionType string, countryCode string) error {
    if contributionType == "donation" {
        _, err := p.db.Exec(`INSERT INTO donations_per_country (country_code, count)
                            VALUES ($1, $2) ON CONFLICT (country_code)
                            DO UPDATE SET count = donations_per_country.count + 1`, countryCode, 1)
        if err != nil {
            log.Debug("[Update Contributions per Country] Couldn't insert into/update donations_per_country: ", err.Error())
            return err
        }
    } else if contributionType == "validation" {
        _, err := p.db.Exec(`INSERT INTO validations_per_country (country_code, count)
                            VALUES ($1, $2) ON CONFLICT (country_code)
                            DO UPDATE SET count = validations_per_country.count + 1`, countryCode, 1)
        if err != nil {
            log.Debug("[Update Contributions per Country] Couldn't insert into/update validations_per_country: ", err.Error())
            return err
        }
    } else if contributionType == "annotation" {
        _, err := p.db.Exec(`INSERT INTO annotations_per_country (country_code, count)
                            VALUES ($1, $2) ON CONFLICT (country_code)
                            DO UPDATE SET count = annotations_per_country.count + 1`, countryCode, 1)
        if err != nil {
            log.Debug("[Update Contributions per Country] Couldn't insert into/update annotations_per_country: ", err.Error())
            return err
        }
    } else if contributionType == "annotation-refinement" {
        _, err := p.db.Exec(`INSERT INTO annotation_refinements_per_country (country_code, count)
                            VALUES ($1, $2) ON CONFLICT (country_code)
                            DO UPDATE SET count = annotation_refinements_per_country.count + 1`, countryCode, 1)
        if err != nil {
            log.Debug("[Update Contributions per Country] Couldn't insert into/update annotation_refinements_per_country: ", err.Error())
            return err
        }
    } else if contributionType == "image-description" {
        _, err := p.db.Exec(`INSERT INTO image_descriptions_per_country (country_code, count)
                            VALUES ($1, $2) ON CONFLICT (country_code)
                            DO UPDATE SET count = image_descriptions_per_country.count + 1`, countryCode, 1)
        if err != nil {
            log.Debug("[Update Contributions per Country] Couldn't insert into/update image_descriptions_per_country: ", err.Error())
            return err
        }
    }

    return nil
}

func (p *ImageMonkeyDatabase) UpdateContributionsPerApp(contributionType string, appIdentifier string) error {
    if contributionType == "donation" {
        _, err := p.db.Exec(`INSERT INTO donations_per_app (app_identifier, count)
                            VALUES ($1, $2) ON CONFLICT (app_identifier)
                            DO UPDATE SET count = donations_per_app.count + 1`, appIdentifier, 1)
        if err != nil {
            log.Debug("[Update Contributions per App] Couldn't insert into/update donations_per_app: ", err.Error())
            return err
        }
    } else if contributionType == "validation" {
        _, err := p.db.Exec(`INSERT INTO validations_per_app (app_identifier, count)
                            VALUES ($1, $2) ON CONFLICT (app_identifier)
                            DO UPDATE SET count = validations_per_app.count + 1`, appIdentifier, 1)
        if err != nil {
            log.Debug("[Update Contributions per App] Couldn't insert into/update validations_per_app: ", err.Error())
            return err
        }
    } else if contributionType == "annotation" {
        _, err := p.db.Exec(`INSERT INTO annotations_per_app (app_identifier, count)
                            VALUES ($1, $2) ON CONFLICT (app_identifier)
                            DO UPDATE SET count = annotations_per_app.count + 1`, appIdentifier, 1)
        if err != nil {
            log.Debug("[Update Contributions per App] Couldn't insert into/update annotations_per_app: ", err.Error())
            return err
        }
    }

    return nil
}


func (p *ImageMonkeyDatabase) GetAnnotatedStatistics(apiUser datastructures.APIUser, excludeMetalabels bool) ([]datastructures.AnnotatedStat, error) {
    annotatedStats := []datastructures.AnnotatedStat{}
    var queryValues []interface{}

    includeOwnImageDonations := ""
    if apiUser.Name != "" {
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
        queryValues = append(queryValues, apiUser.Name)
    }

    q1 := ""
    if excludeMetalabels {
        q1 = "WHERE l.label_type != 'meta'"
    }



    q := fmt.Sprintf(`WITH num_validations AS (
                        SELECT v.label_id, COUNT(v.label_id) as num
                        FROM image_validation v
                        JOIN image i ON v.image_id = i.id
                        WHERE (i.unlocked = true %s)
                        GROUP BY v.label_id
                     ),
                     num_annotations AS (
                        SELECT a.label_id, COUNT(a.label_id) as num
                        FROM image_annotation a
                        JOIN image i ON a.image_id = i.id
                        WHERE (i.unlocked = true %s) AND a.auto_generated = false
                        GROUP BY a.label_id
                     )
                     SELECT l.uuid, acc.accessor, COALESCE(v.num, 0) as num_total, COALESCE(a.num, 0) as num_completed
                     FROM num_validations v
                     JOIN label_accessor acc ON acc.label_id = v.label_id
                     JOIN label l ON acc.label_id = l.id
                     LEFT JOIN num_annotations a ON a.label_id = acc.label_id
                     %s
                     ORDER BY 
                        CASE 
                            WHEN v.num = 0 THEN 0
                            ELSE a.num/v.num
                        END DESC`, 
                     includeOwnImageDonations, includeOwnImageDonations, q1)

    rows, err := p.db.Query(q, queryValues...)
    if err != nil {
        log.Debug("[Get Annotated Statistics] Couldn't get annotated statistics: ", err.Error())
        raven.CaptureError(err, nil)
        return annotatedStats, err
    }

    defer rows.Close()

    for rows.Next() {
        var annotatedStat datastructures.AnnotatedStat
        err = rows.Scan(&annotatedStat.Label.Id, &annotatedStat.Label.Name, &annotatedStat.Num.Total, &annotatedStat.Num.Completed)
        if err != nil {
            log.Debug("[Get Annotated Statistics] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return annotatedStats, err
        }

        annotatedStats = append(annotatedStats, annotatedStat)
    }

    return annotatedStats, nil
}




func (p *ImageMonkeyDatabase) GetValidatedStatistics(apiUser datastructures.APIUser) ([]datastructures.ValidatedStat, error) {
    var validatedStats []datastructures.ValidatedStat
    var queryValues []interface{}

    includeOwnImageDonations := ""
    if apiUser.Name != "" {
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
        queryValues = append(queryValues, apiUser.Name)
    }



    q := fmt.Sprintf(`WITH validated_images AS (
                        SELECT v.label_id as label_id, count(*) as num
                        FROM image_validation v 
                        JOIN image i ON v.image_id = i.id
                        WHERE v.num_of_valid <> 0 OR v.num_of_invalid <> 0 AND (i.unlocked = true %s)
                        GROUP BY v.label_id
                     ),
                     total_image_validations AS (
                        SELECT v.label_id as label_id, count(*) as num
                        FROM image_validation v 
                        JOIN image i ON v.image_id = i.id
                        WHERE (i.unlocked = true %s)
                        GROUP BY v.label_id
                     )
                     SELECT l.uuid, acc.accessor, COALESCE(t.num, 0) as num_total, COALESCE(v.num, 0) as num_completed
                     FROM total_image_validations t
                     JOIN label_accessor acc ON acc.label_id = t.label_id
                     JOIN label l ON acc.label_id = l.id
                     LEFT JOIN validated_images v ON v.label_id = t.label_id
                     ORDER BY 
                        CASE 
                            WHEN v.num = 0 THEN 0
                            ELSE v.num/t.num
                        END DESC`, 
                     includeOwnImageDonations, includeOwnImageDonations)

    rows, err := p.db.Query(q, queryValues...)
    if err != nil {
        log.Debug("[Get Validated Statistics] Couldn't get validated statistics: ", err.Error())
        raven.CaptureError(err, nil)
        return validatedStats, err
    }

    defer rows.Close()

    for rows.Next() {
        var validatedStat datastructures.ValidatedStat
        err = rows.Scan(&validatedStat.Label.Id, &validatedStat.Label.Name, &validatedStat.Num.Total, &validatedStat.Num.Completed)
        if err != nil {
            log.Debug("[Get Validated Statistics] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return validatedStats, err
        }

        validatedStats = append(validatedStats, validatedStat)
    }

    return validatedStats, nil
}


func (p *ImageMonkeyDatabase) GetActivity(period string) ([]datastructures.Activity, error) {
    var activity []datastructures.Activity

    if period != "last-month" {
        return activity, errors.New("Only last-month statistics are supported at the moment")
    }

    rows, err := p.db.Query(`SELECT l.name, i.key, q.type, date(q.dt), i.width, i.height, 
                           (d.annotation || ('{"type":"' || t.name || '"}')::jsonb)::jsonb as annotation, q.activity_name 
                           FROM
                            (
                                (
                                    (
                                        SELECT label_id, image_id, 'created' as type, lower(a.sys_period) as dt, 
                                        a.id as annotation_id, 'annotation' as activity_name
                                        FROM image_annotation a 
                                        WHERE id NOT IN ( SELECT id FROM image_annotation_history h
                                                          WHERE h.label_id = a.label_id and a.image_id = h.image_id
                                                        )
                                        AND 
                                        (
                                                date(lower(a.sys_period)) <= CURRENT_DATE 
                                                AND 
                                                date(lower(a.sys_period)) >= (CURRENT_DATE - interval '1 month')
                                        )
                                    )

                                    UNION

                                    (
                                        SELECT label_id, image_id, 'created' as type, lower(h.sys_period) as dt, 
                                        h.id as annotation_id, 'annotation' as activity_name
                                        FROM image_annotation_history h
                                        WHERE h.num_of_valid = 0 AND h.num_of_invalid = 0
                                        AND 
                                        (
                                            date(upper(h.sys_period)) <= CURRENT_DATE
                                            AND 
                                            date(upper(h.sys_period)) >= (CURRENT_DATE - interval '1 month')
                                        )
                                    )

                                    UNION ALL

                                    (
                                        SELECT a.label_id, a.image_id, 'verified' as type, upper(h.sys_period) as dt, 
                                        h.id as annotation_id, 'annotation' as activity_name
                                        FROM image_annotation_history h
                                        JOIN image_annotation a 
                                        ON a.id = h.id
                                        AND 
                                        (
                                            date(upper(h.sys_period)) <= CURRENT_DATE 
                                            AND 
                                            date(upper(h.sys_period)) >= (CURRENT_DATE - interval '1 month')
                                        )
                                    )
                                )


                                UNION ALL
                                (
                                    (
                                        SELECT label_id, image_id, 'created' as type, lower(v.sys_period) as dt, 
                                        null::bigint as annotation_id, 'validation' as activity_name
                                        FROM image_validation v 
                                        WHERE id NOT IN ( SELECT id FROM image_validation_history h
                                                          WHERE h.label_id = v.label_id and v.image_id = h.image_id
                                                        )
                                        AND 
                                        (
                                                date(lower(v.sys_period)) <= CURRENT_DATE 
                                                AND 
                                                date(lower(v.sys_period)) >= (CURRENT_DATE - interval '1 month')
                                        )
                                    )

                                    UNION

                                    (
                                        SELECT label_id, image_id, 'created' as type, lower(h.sys_period) as dt, 
                                        null::bigint as annotation_id, 'validation' as activity_name
                                        FROM image_validation_history h
                                        WHERE h.num_of_valid = 0 AND h.num_of_invalid = 0
                                        AND 
                                        (
                                            date(upper(h.sys_period)) <= CURRENT_DATE
                                            AND 
                                            date(upper(h.sys_period)) >= (CURRENT_DATE - interval '1 month')
                                        )
                                    )

                                    UNION ALL

                                    (
                                        SELECT v.label_id, v.image_id, 'verified' as type, upper(h.sys_period) as dt, 
                                        null::bigint as annotation_id, 'validation' as activity_name
                                        FROM image_validation_history h
                                        JOIN image_validation v 
                                        ON v.id = h.id
                                        AND 
                                        (
                                            date(upper(h.sys_period)) <= CURRENT_DATE 
                                            AND 
                                            date(upper(h.sys_period)) >= (CURRENT_DATE - interval '1 month')
                                        )
                                    )
                                )
                            ) q
                            JOIN label l ON q.label_id = l.id
                            JOIN image i ON q.image_id = i.id
                            LEFT JOIN annotation_data d ON q.annotation_id = d.image_annotation_id
                            LEFT JOIN annotation_type t ON d.annotation_type_id = t.id
                            WHERE i.unlocked = true
                            order by dt desc`)
    if err != nil {
        log.Debug("[Get Activity] Couldn't get activity: ", err.Error())
        raven.CaptureError(err, nil)
        return activity, err
    }

    defer rows.Close()

    for rows.Next() {
        var a datastructures.Activity
        var annotation []byte
        err = rows.Scan(&a.Image.Label, &a.Image.Id, &a.Type, &a.Date, &a.Image.Width, &a.Image.Height, &annotation, &a.Name)
        if err != nil {
            log.Debug("[Get Activity] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return activity, err
        }

        if len(annotation) > 0 {
            err := json.Unmarshal(annotation, &a.Image.Annotation)
            if err != nil {
                log.Debug("[Get Activity] Couldn't unmarshal annotations: ", err.Error())
                raven.CaptureError(err, nil)
                return activity, err
            }
        }

        activity = append(activity, a)
    }

    return activity, nil
}


func (p *ImageMonkeyDatabase) GetNumOfAnnotatedImages() (int64, error) {
    var num int64
    err := p.db.QueryRow("SELECT count(*) FROM image_annotation").Scan(&num)
    if err != nil {
        log.Debug("[Fetch images] Couldn't get num of annotated images: ", err.Error())
        raven.CaptureError(err, nil)
        return 0, err
    }

    return num, nil
}


func (p *ImageMonkeyDatabase) GetNumOfValidatedImages() (int64, error) {
    var num int64
    err := p.db.QueryRow("SELECT count(*) FROM image_validation").Scan(&num)
    if err != nil {
        log.Debug("[Fetch images] Couldn't get num of validated images: ", err.Error())
        raven.CaptureError(err, nil)
        return 0, err
    }

    return num, nil
}

func (p *ImageMonkeyDatabase) GetValidationsCount(minProbability float64, minCount int) ([]datastructures.ValidationCount, error) {
    validationCounts := []datastructures.ValidationCount{}

    rows, err := p.db.Query(`SELECT a.accessor, COUNT(*) 
                             FROM
                             (
                                SELECT v.id as validation_id, v.label_id as label_id 
                                FROM image_validation v
                                JOIN image i on i.id = v.image_id 
                                WHERE i.unlocked = true
                                GROUP BY v.id, v.label_id 
                                HAVING (SUM(v.num_of_valid)/NULLIF((SUM(v.num_of_valid) + SUM(v.num_of_invalid)), 0)) >= $1::float
                             ) q
                             JOIN label_accessor a ON a.label_id = q.label_id
                             GROUP BY a.accessor
                             HAVING COUNT(*) >= $2`, minProbability, minCount)

    if err != nil {
        log.Error("[Num of validations] Couldn't get num of validations: ", err.Error())
        raven.CaptureError(err, nil)
        return validationCounts, err
    }

    defer rows.Close()

    for rows.Next() {
        var validationCount datastructures.ValidationCount

        err = rows.Scan(&validationCount.Label, &validationCount.Count)
        if err != nil {
            log.Error("[Num of validations] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return validationCounts, err
        }

        validationCounts = append(validationCounts, validationCount)
    }

    return validationCounts, nil
}


func (p *ImageMonkeyDatabase) GetAnnotationsCount(minProbability float64, minCount int) ([]datastructures.AnnotationCount, error) {
    annotationCounts := []datastructures.AnnotationCount{}

    rows, err := p.db.Query(`SELECT a.accessor, COUNT(*) 
                             FROM
                             (
                                SELECT a.id as annotation_id, a.label_id as label_id 
                                FROM image_annotation a
                                JOIN image i on i.id = a.image_id
                                WHERE i.unlocked = true and a.auto_generated = false
                                GROUP BY a.id, a.label_id 
                                HAVING COALESCE((SUM(a.num_of_valid)/NULLIF((SUM(a.num_of_valid) + SUM(a.num_of_invalid)), 0)), 0) >= $1::float
                             ) q
                             JOIN label_accessor a ON a.label_id = q.label_id
                             GROUP BY a.accessor
                             HAVING COUNT(*) >= $2`, minProbability, minCount)

    if err != nil {
        log.Error("[Num of annotations] Couldn't get num of annotations: ", err.Error())
        raven.CaptureError(err, nil)
        return annotationCounts, err
    }

    defer rows.Close()

    for rows.Next() {
        var annotationCount datastructures.AnnotationCount

        err = rows.Scan(&annotationCount.Label, &annotationCount.Count)
        if err != nil {
            log.Error("[Num of annotations] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return annotationCounts, err
        }

        annotationCounts = append(annotationCounts, annotationCount)
    }

    return annotationCounts, nil
}