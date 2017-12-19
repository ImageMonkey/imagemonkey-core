package main

import(
	"fmt"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"flag"
)


func getLabelId(tx *sql.Tx, label string, sublabel string) int64 {
	var labelId int64
	labelId = -1
	if sublabel == "" {
		err := tx.QueryRow(`SELECT id FROM label WHERE name = $1 and parent_id is null`, label).Scan(&labelId)
		if err != nil {
			tx.Rollback()
			log.Fatal("hhhhhhh = ", err)
			panic(err)
		}
	} else {
		err := tx.QueryRow(`SELECT l.id FROM label l 
							JOIN label pl ON pl.id = l.parent_id
							WHERE l.name = $1 and pl.name = $2`, sublabel, label).Scan(&labelId)
		if err != nil {
			tx.Rollback()
			log.Fatal("kkkkkkkkkk = ", label, sublabel)
			panic(err)
		}
	}

	return labelId
}

func addAccessor(tx *sql.Tx, labelId int64, accessor string) error {
	var insertedId int64
	err := tx.QueryRow(`INSERT INTO label_accessor(label_id, accessor) VALUES($1, $2)
                       				ON CONFLICT (label_id, accessor) DO NOTHING RETURNING id`, labelId, accessor).Scan(&insertedId)

	if insertedId != -1 {
		fmt.Printf("Inserted label accessor %s for label with id %d\n", accessor, labelId)
	}

	fmt.Printf("ACCC = %s\n", accessor)

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
			log.Fatal("bbbbb = %s", err.Error())
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
						log.Fatal("aaa = %s", err.Error())
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
						log.Fatal("zzzz = %s", err.Error())
						panic(err)
					}

				}
			}
		}
	}
}

/*func populateAccessors(tx *sql.Tx, label string, sublabel string, labelMapEntry LabelMapEntry) error{
	fmt.Printf("Populating accessors...%d\n", len(labelMapEntry.Accessors))
	for _, accessor := range labelMapEntry.Accessors {
		//accessorStr := label
		//if accessor != "." {
		//	fmt.Printf("HERE\n")
		//	accessorStr = label + accessor
		//}
		var accessorStr string
		if sublabel == "" {
			accessorStr = label
			if accessor == "." {
				accessorStr = label
			}
		} else {
			accessorStr = label + sublabel


		labelId := getLabelId(tx, label, sublabel)
		if labelId == -1 {
			tx.Rollback()
			log.Fatal("Populating accessors...couldn't get label id")
		}

		var insertedId int64
		insertedId = -1
		err := tx.QueryRow(`INSERT INTO label_accessor(label_id, accessor) VALUES($1, $2)
                       				ON CONFLICT (label_id, accessor) DO NOTHING RETURNING id`, labelId, accessorStr).Scan(&insertedId)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
			panic(err)
		}

		if insertedId != -1 {
			fmt.Printf("Inserted label accessor %s for label with id %d\n", accessorStr, labelId)
		}
	}
	return nil
}*/

func main(){
	dryRun := flag.Bool("dry-run", true, "dry run")

	if *dryRun {
		fmt.Printf("Populating labels...\n")
	} else{
		fmt.Printf("Populating labels (dry run)...\n")
	}

	db, err := sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	tx, err := db.Begin()
    if err != nil {
    	fmt.Printf("Couldn't start transaction\n")
    	return
    }

    labelMap, _, err := getLabelMap("../wordlists/en/labels.json")
	if err != nil {
		fmt.Printf("Couldn't populate labels %s\n", err.Error())
		return
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
			fmt.Printf("Adding label %s\n", k)
			_,err := tx.Exec("INSERT INTO label(name) VALUES($1)", k)
			if err != nil {
				tx.Rollback()
				log.Fatal(err)
				panic(err)
			}
		} else {
			fmt.Printf("Skipping label %s, as it already exists\n", k)
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
					fmt.Printf("Adding label %s (parent: %s) \n", sublabel, k)
					_,err := tx.Exec(`INSERT INTO label(name, parent_id)
										SELECT $1, l.id FROM label l WHERE l.name = $2 AND l.parent_id is null`,
									sublabel, k)
					if err != nil {
						tx.Rollback()
						log.Fatal(err)
						panic(err)
					}
				} else {
					fmt.Printf("Skipping label %s (parent: %s), as it already exists\n", sublabel, k)
				}
			}

		}

		addAccessors(tx, k, val)
	} 

	//if ! *dryRun {
		err = tx.Commit()
		if err != nil {
			fmt.Printf("Couldn't commit changes\n")
			return
		}
	//}
}