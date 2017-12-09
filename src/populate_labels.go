package main

import(
	"fmt"
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

func main(){
	fmt.Printf("Populating labels...\n")

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
	if(err != nil){
		fmt.Printf("Couldn't populate labels\n")
		return
	}

	for k := range labelMap {
		val := labelMap[k]

		rows, err := tx.Query("SELECT COUNT(id) FROM label WHERE name = $1", k)
		if(err != nil){
			tx.Rollback()
			log.Fatal(err)
			panic(err)
		}
		if(!rows.Next()){
			tx.Rollback()
			log.Fatal(err)
			panic(err)
		}

		numOfLabels := 0
		err = rows.Scan(&numOfLabels)
		if(err != nil){
			tx.Rollback()
			log.Fatal(err)
			panic(err)
		}

		rows.Close()

		if(numOfLabels == 0){
			fmt.Printf("Adding label %s\n", k)
			_,err := tx.Exec("INSERT INTO label(name) VALUES($1)", k)
			if(err != nil){
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
				if(err != nil){
					tx.Rollback()
					log.Fatal(err)
					panic(err)
				}
				if(!rows.Next()){
					tx.Rollback()
					log.Fatal(err)
					panic(err)
				}

				numOfLabels := 0
				err = rows.Scan(&numOfLabels)
				if(err != nil){
					tx.Rollback()
					log.Fatal(err)
					panic(err)
				}

				rows.Close()

				if(numOfLabels == 0){
					fmt.Printf("Adding label %s (parent: %s) \n", sublabel, k)
					_,err := tx.Exec(`INSERT INTO label(name, parent_id)
										SELECT $1, l.id FROM label l WHERE l.name = $2 AND l.parent_id is null`,
									sublabel, k)
					if(err != nil){
						tx.Rollback()
						log.Fatal(err)
						panic(err)
					}
				} else {
					fmt.Printf("Skipping label %s (parent: %s), as it already exists\n", sublabel, k)
				}

			}

		}
	} 

	err = tx.Commit()
	if(err != nil){
		fmt.Printf("Couldn't commit changes\n")
		return
	}
}