package main

import (
	commons "github.com/bbernhard/imagemonkey-core/commons"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"context"
	"flag"
	"html"
)

type LabelSuggestionEntry struct {
	Id int64
	Name string
}

func fixHtmlEncoding(tx pgx.Tx) error {
	rows, err := tx.Query(context.TODO(), `SELECT id, name::text
											FROM label_suggestion 
											WHERE (name LIKE '%&gt;%') OR (name LIKE '%&lt;%')
											OR (name LIKE '%&amp;%') OR (name LIKE '%&quot;%')
											OR (name LIKE '%&#39;%')`)
	if err != nil {
		return err
	}
	defer rows.Close()

	labelSuggestionEntries := []LabelSuggestionEntry{}
	for rows.Next() {
		var labelSuggestionEntry LabelSuggestionEntry
		err = rows.Scan(&labelSuggestionEntry.Id, &labelSuggestionEntry.Name)
		if err != nil {
			return err
		}
		labelSuggestionEntries = append(labelSuggestionEntries, labelSuggestionEntry)
	}

	rows.Close()

	for _, labelSuggestionEntry := range labelSuggestionEntries {
		fixedName := html.UnescapeString(labelSuggestionEntry.Name)

		log.Info("Renaming ", labelSuggestionEntry.Name, " to ", fixedName)
		_, err = tx.Exec(context.TODO(), `UPDATE label_suggestion SET name = $1 WHERE id = $2`, fixedName, labelSuggestionEntry.Id)
		if err != nil {
			return err
		}
	}
	
	return nil
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

	err = fixHtmlEncoding(tx)
	if err != nil {
		log.Error("Couldn't fix html encoded string: ", err.Error())
		err := tx.Rollback(context.TODO())
		if err != nil {
			log.Fatal("Couldn't rollback transaction: ", err.Error())
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
