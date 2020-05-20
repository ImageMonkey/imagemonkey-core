package main

import (
	commons "github.com/bbernhard/imagemonkey-core/commons"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"context"
	"flag"
	"github.com/jackc/pgtype"
	"errors"
)


func removeElemFromSlice(s []int64, r int64) []int64 {
    for i, v := range s {
        if v == r {
            return append(s[:i], s[i+1:]...)
        }
    }
    return s
}

//TODO: add unique constraint to label_suggestion table

func unifyDuplicateLabelSuggestions(tx pgx.Tx, source []int64, target int64) error {
	sourceIds := &pgtype.Int8Array{}
	sourceIds.Set(source)
	
	temp := []int64{}
	temp = append(temp, source...)
	temp = append(temp, target)
	allIds := &pgtype.Int8Array{}
	allIds.Set(temp)

	rows, err := tx.Query(context.TODO(), `SELECT count(*) FROM image_label_suggestion WHERE label_suggestion_id = ANY($1) GROUP BY image_id`, allIds)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var count int
		err := rows.Scan(&count)
		if err != nil {
			return err
		}

		if count > 1 {
			return errors.New("damn")
		}
	}

	rows.Close()
	
	rows, err = tx.Query(context.TODO(), `SELECT id FROM image_label_suggestion i WHERE label_suggestion_id = ANY($1)`, sourceIds)
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
	
	log.Info("here")
	log.Info(sourceIds)
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
