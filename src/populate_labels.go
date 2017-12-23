package main

import(
	"database/sql"
	_ "github.com/lib/pq"
	log "github.com/Sirupsen/logrus"
	"flag"
)


func getLabelId(tx *sql.Tx, label string, sublabel string) int64 {
	var labelId int64
	labelId = -1
	if sublabel == "" {
		err := tx.QueryRow(`SELECT id FROM label WHERE name = $1 and parent_id is null`, label).Scan(&labelId)
		if err != nil {
			tx.Rollback()
			log.Fatal("Couldn't get label id: ", err)
			panic(err)
		}
	} else {
		err := tx.QueryRow(`SELECT l.id FROM label l 
							JOIN label pl ON pl.id = l.parent_id
							WHERE l.name = $1 and pl.name = $2`, sublabel, label).Scan(&labelId)
		if err != nil {
			tx.Rollback()
			log.Fatal("Couldn't get label id: ", label, sublabel)
			panic(err)
		}
	}

	return labelId
}

func addAccessor(tx *sql.Tx, labelId int64, accessor string) error {
	var insertedId int64
	insertedId = -1
	err := tx.QueryRow(`INSERT INTO label_accessor(label_id, accessor) VALUES($1, $2)
                       				ON CONFLICT (label_id, accessor) DO NOTHING RETURNING id`, labelId, accessor).Scan(&insertedId)

	if insertedId != -1 {
		log.Info("Inserted label accessor ", accessor, " for label with id ", labelId)
	} else {
		err = nil //a negative id means, that the label already exists
	}

	return err
}

func addAccessors(tx *sql.Tx, label string, val LabelMapEntry) {
	for _, accessor := range val.Accessors {

		labelId := getLabelId(tx, label, "")
		if labelId == -1 {
			tx.Rollback()
			log.Fatal("Populating accessors...couldn't get label id")
		}

		accessorStr := accessor
		if accessor == "." {
			accessorStr = label
		}

		err := addAccessor(tx, labelId, accessorStr)
		if err != nil {
			tx.Rollback()
			log.Fatal("Couldn't add accessor: ", err.Error())
			panic(err)
		}


		if len(val.LabelMapEntries) != 0 {
			for sublabel, subval  := range val.LabelMapEntries {
				for _, subaccessor := range subval.Accessors {
					labelId := getLabelId(tx, label, sublabel)
					if labelId == -1 {
						tx.Rollback()
						log.Fatal("Populating accessors...couldn't get label id")
					}

					subaccessorStr := accessorStr + subaccessor + "='" + sublabel + "'" 
					if subaccessor == "." {
						subaccessorStr = accessorStr + sublabel
					}

					err := addAccessor(tx, labelId, subaccessorStr)
					if err != nil {
						tx.Rollback()
						log.Fatal("Couldn't add accessor for sublabel: ", err.Error())
						panic(err)
					}
				}
			}
		}

		//quiz
		for _, quizEntry := range val.Quiz {
			for _, quizEntryAccessor := range quizEntry.Accessors {
				for _, answer := range quizEntry.Answers {

					sublabel := answer.Name
					if label == answer.Name {
						sublabel = ""
					}

					quizEntryAccessorStr := accessorStr + quizEntryAccessor + "='" + answer.Name + "'"
					
					labelId := getLabelId(tx, label, sublabel)
					if labelId == -1 {
						tx.Rollback()
						log.Fatal("Populating accessors...couldn't get label id")
					}

					err := addAccessor(tx, labelId, quizEntryAccessorStr)
					if err != nil {
						tx.Rollback()
						log.Fatal("Couldn't add accessor for quiz entry: ", err.Error())
						panic(err)
					}

				}
			}
		}
	}
}

func main(){
	dryRun := flag.Bool("dryrun", true, "dry run")
	debug := flag.Bool("debug", false, "debug")

	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	if *dryRun {
		log.Info("Populating labels (dry run)...")
	} else{
		log.Info("Populating labels...")
	}

	db, err := sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	tx, err := db.Begin()
    if err != nil {
    	log.Fatal("Couldn't start transaction: ", err.Error())
    }

    labelMap, _, err := getLabelMap("../wordlists/en/labels.json")
	if err != nil {
		log.Fatal("Couldn't get label map: ", err.Error())
	}

	for k := range labelMap {
		val := labelMap[k]

		rows, err := tx.Query("SELECT COUNT(id) FROM label WHERE name = $1", k)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
			panic(err)
		}
		if !rows.Next() {
			tx.Rollback()
			log.Fatal(err)
			panic(err)
		}

		numOfLabels := 0
		err = rows.Scan(&numOfLabels)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
			panic(err)
		}

		rows.Close()

		if numOfLabels == 0 {
			log.Info("Adding label ", k)
			_,err := tx.Exec("INSERT INTO label(name) VALUES($1)", k)
			if err != nil {
				tx.Rollback()
				log.Fatal(err)
				panic(err)
			}
		} else {
			log.Debug("Skipping label ", k, " as it already exists")
		}


		if len(val.LabelMapEntries) != 0 {
			for sublabel := range val.LabelMapEntries {
				rows, err := tx.Query("select count(l.id) from label l join label pl on l.parent_id = pl.id WHERE l.name = $1 AND pl.name = $2", sublabel, k)
				if err != nil {
					tx.Rollback()
					log.Fatal(err)
					panic(err)
				}
				if !rows.Next() {
					tx.Rollback()
					log.Fatal(err)
					panic(err)
				}

				numOfLabels := 0
				err = rows.Scan(&numOfLabels)
				if err != nil {
					tx.Rollback()
					log.Fatal(err)
					panic(err)
				}

				rows.Close()

				if numOfLabels == 0 {
					log.Info("Adding label ", sublabel, " (parent: ", k, " )")
					_,err := tx.Exec(`INSERT INTO label(name, parent_id)
										SELECT $1, l.id FROM label l WHERE l.name = $2 AND l.parent_id is null`,
									sublabel, k)
					if err != nil {
						tx.Rollback()
						log.Fatal(err)
						panic(err)
					}
				} else {
					log.Debug("Skipping label ", sublabel, " (parent: ", k, " ), as it already exists")
				}
			}

		}

		addAccessors(tx, k, val)
	} 

	if ! *dryRun {
		err = tx.Commit()
		if err != nil {
			log.Fatal("Couldn't commit changes: ", err.Error())
		}
	} else {
		tx.Rollback()
		log.Info("Rolling back transaction...it was only a dry run.")
	}
}