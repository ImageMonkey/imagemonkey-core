package main

import (
	commons "github.com/bbernhard/imagemonkey-core/commons"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"context"
	"flag"
	"github.com/jackc/pgtype"
	//"errors"
)

var unrecoverableDeletedAnnotations int = 0

func removeElemFromSlice(s []int64, r int64) []int64 {
    for i, v := range s {
        if v == r {
            return append(s[:i], s[i+1:]...)
        }
    }
    return s
}

//TODO: add unique constraint to label_suggestion table

//TODO: verify num of image labels + annotations before and after

func unifyTrendingLabelSuggestions(tx pgx.Tx, sourceIds *pgtype.Int8Array) error {
	rows, err := tx.Query(context.TODO(), `SELECT id, github_issue_id 
											FROM trending_label_suggestion 
											WHERE label_suggestion_id = ANY($1)`, sourceIds)
	if err != nil {
		return err
	}

	defer rows.Close()

	trendingLabelSuggestionIds := []int64{}
	githubIssueIds := []int64{}
	for rows.Next() {
		var trendingLabelSuggestionId int64
		var githubIssueId int64
		err := rows.Scan(&trendingLabelSuggestionId, &githubIssueId)
		if err != nil {
			return err
		}

		trendingLabelSuggestionIds = append(trendingLabelSuggestionIds, trendingLabelSuggestionId)
		githubIssueIds = append(githubIssueIds, githubIssueId)
	}

	for _, trendingLabelSuggestionId := range trendingLabelSuggestionIds {
		_, err = tx.Exec(context.TODO(), `DELETE FROM trending_label_suggestion 
											WHERE id = $1`, trendingLabelSuggestionId)
		if err != nil {
			return err
		}
	}
	
	return nil
}

func unifyImagesWithNoDuplicateLabels(tx pgx.Tx, imageIdsThatHaveNoDuplicateLabels []int64, sourceIds *pgtype.Int8Array, target int64) error {
	//unify images with no duplicate labels
	rows, err := tx.Query(context.TODO(), `SELECT id FROM image_label_suggestion i 
											WHERE label_suggestion_id = ANY($1) AND image_id = ANY($2)`, sourceIds, imageIdsThatHaveNoDuplicateLabels)
	if err != nil {
		return err
	}
	defer rows.Close()

	imageLabelSuggestionIds := []int64{}
	for rows.Next() {
		var imageLabelSuggestionId int64
		err := rows.Scan(&imageLabelSuggestionId)
		if err != nil {
			return err
		}
		imageLabelSuggestionIds = append(imageLabelSuggestionIds, imageLabelSuggestionId)
	}

	rows.Close()

	for _, imageLabelSuggestionId := range imageLabelSuggestionIds {
		_, err := tx.Exec(context.TODO(), `UPDATE image_label_suggestion SET label_suggestion_id = $1 WHERE id = $2`, target, imageLabelSuggestionId)
		if err != nil {
			return err
		}
	}

	return nil
}

func unifyImagesWithDuplicateLabels(tx pgx.Tx, imageIdsWithDuplicateLabels []int64, sourceIds *pgtype.Int8Array, allIds *pgtype.Int8Array, target int64) error {
	for _, imageIdWithDuplicateLabels := range imageIdsWithDuplicateLabels {
		rows, err := tx.Query(context.TODO(),
					`			SELECT id 
									FROM image_label_suggestion 
									WHERE image_id = $1 AND label_suggestion_id = ANY($2)`, imageIdWithDuplicateLabels, allIds)
		if err != nil {
			return err
		}

		imageLabelSuggestionIds := []int64{}
		for rows.Next() {
			var imageLabelSuggestionId int64
			err = rows.Scan(&imageLabelSuggestionId)
			if err != nil {
				return err
			}
			imageLabelSuggestionIds = append(imageLabelSuggestionIds, imageLabelSuggestionId)
		}
		rows.Close()

		//delate all occurences, except one
		for i := 0; i < len(imageLabelSuggestionIds)-1; i++ {
			_, err = tx.Exec(context.TODO(), `DELETE FROM image_label_suggestion WHERE id = $1`, imageLabelSuggestionIds[i])
			if err != nil {
				return err
			}
		}

		_, err = tx.Exec(context.TODO(), 
						`UPDATE image_label_suggestion 
							SET label_suggestion_id = $1 
							WHERE id = $2`, target, imageLabelSuggestionIds[len(imageLabelSuggestionIds)-1])
		if err != nil {
			return err
		}
	}

	return nil
}

func unifyImageLabelSuggestions(tx pgx.Tx, sourceIds *pgtype.Int8Array, allIds *pgtype.Int8Array, target int64) error {
	rows, err := tx.Query(context.TODO(), `SELECT a.image_id, count(*) FROM image_label_suggestion a
											WHERE label_suggestion_id = ANY($1) GROUP BY image_id`, allIds)
	if err != nil {
		return err
	}
	defer rows.Close()

	imageIdsThatHaveNoDuplicateLabels := []int64{}
	imageIdsWithDuplicateLabels := []int64{}
	for rows.Next() {
		var count int
		var imageId int64
		err := rows.Scan(&imageId, &count)
		if err != nil {
			return err
		}

		if count > 1 {
			imageIdsWithDuplicateLabels = append(imageIdsWithDuplicateLabels, imageId)
		} else {
			imageIdsThatHaveNoDuplicateLabels = append(imageIdsThatHaveNoDuplicateLabels, imageId)
		}
	}

	rows.Close()

	err = unifyImagesWithNoDuplicateLabels(tx, imageIdsThatHaveNoDuplicateLabels, sourceIds, target)
	if err != nil {
		return err
	}

	err = unifyImagesWithDuplicateLabels(tx, imageIdsWithDuplicateLabels, sourceIds, allIds, target)
	if err != nil {
		return err
	}

	return nil
}

func unifyImageAnnotationsWithNoDuplicateLabels(tx pgx.Tx, imageIdsThatHaveNoDuplicateLabels []int64, sourceIds *pgtype.Int8Array, target int64) error {
	rows, err := tx.Query(context.TODO(), `SELECT id 
											FROM image_annotation_suggestion a 
											WHERE label_suggestion_id = ANY($1)
											AND a.image_id = ANY($2)`, sourceIds, imageIdsThatHaveNoDuplicateLabels)
	if err != nil {
		return err
	}
	defer rows.Close()

	imageAnnotationLabelSuggestionIds := []int64{}
	for rows.Next() {
		var imageAnnotationLabelSuggestionId int64
		err := rows.Scan(&imageAnnotationLabelSuggestionId)
		if err != nil {
			return err
		}
		imageAnnotationLabelSuggestionIds = append(imageAnnotationLabelSuggestionIds, imageAnnotationLabelSuggestionId)
	}

	rows.Close()

	for _, imageAnnotationLabelSuggestionId := range imageAnnotationLabelSuggestionIds {
		_, err = tx.Exec(context.TODO(), `UPDATE image_annotation_suggestion
										SET label_suggestion_id = $1 WHERE id = $2`, target, imageAnnotationLabelSuggestionId)
		if err != nil {
			return err
		}
	}

	return nil
}

func unifyImageAnnotationsWithDuplicateLabels(tx pgx.Tx, imageIdsWithDuplicateLabels []int64, allIds *pgtype.Int8Array, target int64) error {
	 for _, imageIdWithDuplicateLabels := range imageIdsWithDuplicateLabels {
		rows, err := tx.Query(context.TODO(),
					`			SELECT id 
									FROM image_annotation_suggestion 
									WHERE image_id = $1 AND label_suggestion_id = ANY($2)`, imageIdWithDuplicateLabels, allIds)
		if err != nil {
			return err
		}

		imageAnnotationSuggestionIds := []int64{}
		for rows.Next() {
			var imageAnnotationSuggestionId int64
			err = rows.Scan(&imageAnnotationSuggestionId)
			if err != nil {
				return err
			}
			imageAnnotationSuggestionIds = append(imageAnnotationSuggestionIds, imageAnnotationSuggestionId)
		}
		rows.Close()
		for _, imageAnnotationSuggestionId := range imageAnnotationSuggestionIds { 
			unrecoverableDeletedAnnotations += 1

			_, err = tx.Exec(context.TODO(), `DELETE FROM annotation_suggestion_data
												WHERE image_annotation_suggestion_id = $1`, imageAnnotationSuggestionId)
			if err != nil {
				return err
			}

			_, err = tx.Exec(context.TODO(), `DELETE FROM user_image_annotation_suggestion 
												WHERE image_annotation_suggestion_id = $1`, imageAnnotationSuggestionId)
			if err != nil {
				return err
			}

			_, err = tx.Exec(context.TODO(), `DELETE FROM image_annotation_suggestion WHERE id = $1`, imageAnnotationSuggestionId)
			if err != nil {
				return err
			}
		}
	 }

	return nil
}

func unifyImageAnnotationSuggestions(tx pgx.Tx, sourceIds *pgtype.Int8Array, allIds *pgtype.Int8Array, target int64) error {
	rows, err := tx.Query(context.TODO(), `SELECT a.image_id, count(*) FROM image_annotation_suggestion a
											WHERE label_suggestion_id = ANY($1) GROUP BY image_id`, allIds)
	if err != nil {
		return err
	}
	defer rows.Close()

	imageIdsThatHaveNoDuplicateLabels := []int64{}
	imageIdsWithDuplicateLabels := []int64{}
	for rows.Next() {
		var count int
		var imageId int64
		err := rows.Scan(&imageId, &count)
		if err != nil {
			return err
		}

		if count > 1 {
			imageIdsWithDuplicateLabels = append(imageIdsWithDuplicateLabels, imageId)
		} else {
			imageIdsThatHaveNoDuplicateLabels = append(imageIdsThatHaveNoDuplicateLabels, imageId)
		}
	}

	rows.Close()


	err = unifyImageAnnotationsWithNoDuplicateLabels(tx, imageIdsThatHaveNoDuplicateLabels, sourceIds, target)
	if err != nil {
		return err
	}

	err = unifyImageAnnotationsWithDuplicateLabels(tx, imageIdsWithDuplicateLabels, allIds, target)
	if err != nil {
		return err
	}
	
	return nil
}

func unifyDuplicateLabelSuggestions(tx pgx.Tx, source []int64, target int64) error {
	sourceIds := &pgtype.Int8Array{}
	sourceIds.Set(source)

	temp := []int64{}
	temp = append(temp, source...)
	temp = append(temp, target)
	allIds := &pgtype.Int8Array{}
	allIds.Set(temp)

	err := unifyImageLabelSuggestions(tx, sourceIds, allIds, target)
	if err != nil {
		return err
	}

	err = unifyImageAnnotationSuggestions(tx, sourceIds, allIds, target)
	if err != nil {
		return err
	}

	err = unifyTrendingLabelSuggestions(tx, sourceIds)
	if err != nil {
		return err
	}
	
	_, err = tx.Exec(context.TODO(), `DELETE FROM label_suggestion WHERE id = ANY($1)`, sourceIds)
	if err != nil {
		return err
	}

	return nil
}

func getDuplicateLabelSuggestions(tx pgx.Tx) ([]string, error) {
	duplicateLabelSuggestions := []string{}
	rows, err := tx.Query(context.TODO(), `SELECT name FROM label_suggestion GROUP BY name HAVING COUNT(name) > 1`)
	if err != nil {
		return duplicateLabelSuggestions, err
	}
	defer rows.Close()

	for rows.Next() {
		var duplicateLabelSuggestion string
		err := rows.Scan(&duplicateLabelSuggestion)
		if err != nil {
			return duplicateLabelSuggestions, err
		}

		duplicateLabelSuggestions = append(duplicateLabelSuggestions, duplicateLabelSuggestion)
	}

	return duplicateLabelSuggestions, nil
}

func getLabelSuggestionIdsForLabelSuggestion(tx pgx.Tx, labelSuggestion string) ([]int64, error) {
	duplicateLabelSuggestionIds := []int64{}
	rows, err := tx.Query(context.TODO(), `SELECT id FROM label_suggestion WHERE name = $1`, labelSuggestion)
	if err != nil {
		return duplicateLabelSuggestionIds, err
	}
	defer rows.Close()

	for rows.Next() {
		var duplicateLabelSuggestionId int64
		err := rows.Scan(&duplicateLabelSuggestionId)
		if err != nil {
			return duplicateLabelSuggestionIds, err
		}

		duplicateLabelSuggestionIds = append(duplicateLabelSuggestionIds, duplicateLabelSuggestionId)
	}

	return duplicateLabelSuggestionIds, nil
}

func isLabelSuggestionIdProductive(tx pgx.Tx, labelSuggestionId int64) (bool, error) {	
	rows, err := tx.Query(context.TODO(),
						`SELECT closed FROM trending_label_suggestion WHERE label_suggestion_id = $1`, labelSuggestionId)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		var productive bool
		err := rows.Scan(&productive)
		return productive, err
	}
	return false, nil
}

func main() {
	dryRun := flag.Bool("dryrun", true, "Do a dry run")
	
	flag.Parse()
	
	imageMonkeyDbConnectionString := commons.MustGetEnv("IMAGEMONKEY_DB_CONNECTION_STRING")
	db, err := pgx.Connect(context.TODO(), imageMonkeyDbConnectionString)
	if err != nil {
		log.Fatal("Couldn't begin transaction: ", err.Error())
	}

	tx, err := db.Begin(context.TODO())
	if err != nil {
		log.Fatal("Couldn't begin transaction: ", err.Error())
	}

	duplicateLabelSuggestions, err := getDuplicateLabelSuggestions(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't get duplicate label suggestions: ", err.Error())
	}

	for _, duplicateLabelSuggestion := range duplicateLabelSuggestions {
		labelSuggestionIds, err := getLabelSuggestionIdsForLabelSuggestion(tx, duplicateLabelSuggestion)
		if err != nil {
			tx.Rollback(context.TODO())
			log.Fatal("Couldn't get label suggestion ids: ", err.Error())
		}

		
		var productiveLabelSuggestionId int64 = -1
		for _, labelSuggestionId := range labelSuggestionIds {
			isProductive, err := isLabelSuggestionIdProductive(tx, labelSuggestionId)
			if err != nil {
				tx.Rollback(context.TODO())
				log.Fatal("Couldn't get label suggestion id info: ", err.Error())
			}

			if isProductive {
				if productiveLabelSuggestionId != -1 {
					tx.Rollback(context.TODO())
					log.Fatal("more than one productive label!")
				}
				productiveLabelSuggestionId = labelSuggestionId;
			}
		}

		if productiveLabelSuggestionId != -1 {
			err = unifyDuplicateLabelSuggestions(tx, removeElemFromSlice(labelSuggestionIds, productiveLabelSuggestionId), productiveLabelSuggestionId)
			if err != nil {
				tx.Rollback(context.TODO())
				log.Fatal("Couldn't unify: ", err.Error())
			}
		} else {
			err = unifyDuplicateLabelSuggestions(tx, labelSuggestionIds[1:], labelSuggestionIds[0])
			if err != nil {
				tx.Rollback(context.TODO())
				log.Fatal("Couldn't unify: ", err.Error())
			}
		}
	}

	//just a sanity check if everything is clean now
	duplicateLabelSuggestionsAfterUnification, err := getDuplicateLabelSuggestions(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("[Verify] Couldn't get duplicate label suggestions: ", err.Error())
	}
	if len(duplicateLabelSuggestionsAfterUnification) > 0 {
		tx.Rollback(context.TODO())
		log.Fatal("Verification failed. There are still duplicates!")
	}
	log.Info("Verification successful")
	log.Info("")
	log.Info("Statistics:")
	log.Info("Unrecoverable deleted annotations: ", unrecoverableDeletedAnnotations)
	log.Info("-----------------------------------")
	log.Info("")

	if *dryRun {
		log.Info("Just a dry run..rolling back transaction")
		err := tx.Rollback(context.TODO())
		if err != nil {
			log.Fatal("Couldn't rollback transaction: ", err.Error())
		}
	} else {
		err := tx.Commit(context.TODO())
		if err != nil {
			log.Fatal("Couldn't commit transaction: ", err.Error())
		}
		log.Info("done")
	}

}
