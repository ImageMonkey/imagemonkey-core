package main

import (
	commons "github.com/bbernhard/imagemonkey-core/commons"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"context"
	"flag"
)

func unifyDuplicateLabelSuggestions(souce []int64, target int64) {
	
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
		} else {
		}
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
