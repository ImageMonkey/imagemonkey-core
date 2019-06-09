package tests

import (
	log "github.com/sirupsen/logrus"
	_"github.com/lib/pq"
	"flag"
)

var db *ImageMonkeyDatabase

const BASE_URL string = "http://127.0.0.1:8081/"
const API_VERSION string = "v1"
var UNVERIFIED_DONATIONS_DIR string = "../unverified_donations/"
var DONATIONS_DIR string = "../donations/"

func init() {
	unverifiedDonationsDir := flag.String("unverified_donations_dir", "../unverified_donations/", "Path to unverified donations directory")
	donationsDir := flag.String("donations_dir", "../donations/", "Path to donations directory")
	flag.Parse()

	UNVERIFIED_DONATIONS_DIR = *unverifiedDonationsDir
	DONATIONS_DIR = *donationsDir

	db = NewImageMonkeyDatabase()
	err := db.Initialize()
	if err != nil {
		log.Fatal("[Main] Couldn't initialize database: ", err.Error())
		panic(err)
	}

	err = db.Open()
	if err != nil {
		log.Fatal("[Main] Couldn't open database: ", err.Error())
	}
}
