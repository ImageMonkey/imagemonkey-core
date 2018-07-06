package tests

import (
	log "github.com/Sirupsen/logrus"
	_"github.com/lib/pq"
)

var db *ImageMonkeyDatabase

const BASE_URL string = "http://127.0.0.1:8081/"
const API_VERSION string = "v1"

func init() {
	db = NewImageMonkeyDatabase()
	err := db.Initialize()
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	err = db.Open()
	if err != nil {
		log.Fatal("[Main] Couldn't open database: ", err.Error())
	}
}