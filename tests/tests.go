package tests

import (
	log "github.com/sirupsen/logrus"
	_"github.com/lib/pq"
	"flag"
	"strings"
	commons "github.com/bbernhard/imagemonkey-core/commons"
)

var db *ImageMonkeyDatabase

const BASE_URL string = "http://127.0.0.1:8081/"
const API_VERSION string = "v1"
var UNVERIFIED_DONATIONS_DIR string = "../unverified_donations/"
var DONATIONS_DIR string = "../donations/"
var X_CLIENT_ID string
var X_CLIENT_SECRET string
var DB_PORT string = "5432"

func init() {
	unverifiedDonationsDir := flag.String("unverified_donations_dir", "../unverified_donations/", "Path to unverified donations directory")
	donationsDir := flag.String("donations_dir", "../donations/", "Path to donations directory")
	flag.Parse()

	UNVERIFIED_DONATIONS_DIR = *unverifiedDonationsDir
	DONATIONS_DIR = *donationsDir

	X_CLIENT_ID = commons.MustGetEnv("X_CLIENT_ID")
	X_CLIENT_SECRET = commons.MustGetEnv("X_CLIENT_SECRET")

	tokens := strings.Split(commons.MustGetEnv("IMAGEMONKEY_DB_CONNECTION_STRING"), " ")
	for _,token := range tokens {
		if(strings.HasPrefix(token, "port=")) {
			DB_PORT = strings.Replace(token, "port=", "", 1)
		}
	}

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
