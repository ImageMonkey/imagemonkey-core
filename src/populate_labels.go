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

	words, err := getWordLists("../wordlists/en/misc.txt")
	if(err != nil){
		fmt.Printf("Couldn't populate labels\n")
		return
	}

	for _, word := range words {
		rows, err := db.Query("SELECT COUNT(id) FROM label WHERE name = $1", word)
		if(err != nil){
			log.Fatal(err)
			panic(err)
		}
		if(!rows.Next()){
			log.Fatal(err)
			panic(err)
		}

		numOfLabels := 0
		err = rows.Scan(&numOfLabels)
		if(err != nil){
			log.Fatal(err)
			panic(err)
		}

		if(numOfLabels == 0){
			fmt.Printf("Adding label %s\n", word)
			_,err := db.Exec("INSERT INTO label(name) VALUES($1)", word)
			if(err != nil){
				log.Fatal(err)
				panic(err)
			}
		}

	}
}