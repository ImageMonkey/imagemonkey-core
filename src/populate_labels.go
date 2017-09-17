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

	words, err := getWordLists("../wordlists/en/misc.txt")
	if(err != nil){
		fmt.Printf("Couldn't populate labels\n")
		return
	}

	for _, word := range words {
		fmt.Printf("word = %s\n", word.Name)
		rows, err := tx.Query("SELECT COUNT(id) FROM label WHERE name = $1", word.Name)
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
			fmt.Printf("Adding label %s\n", word.Name)
			_,err := tx.Exec("INSERT INTO label(name) VALUES($1)", word.Name)
			if(err != nil){
				tx.Rollback()
				log.Fatal(err)
				panic(err)
			}
		}
	}

	err = tx.Commit()
	if(err != nil){
		fmt.Printf("Couldn't commit changes\n")
		return
	}
}