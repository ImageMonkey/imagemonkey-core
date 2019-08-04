package main

import (
	log "github.com/sirupsen/logrus"
	commons "github.com/bbernhard/imagemonkey-core/commons"
	clients "github.com/bbernhard/imagemonkey-core/clients"
	"github.com/getsentry/raven-go"
	_ "github.com/lib/pq"	
	"time"
	"database/sql"
	"flag"
	"os"
	"strconv"
)

var db *sql.DB

func createBackup(from string, to string) error {
	return os.Rename(from, to)
}

func restoreBackup(from string, to string) error {
	return os.Rename(from, to)
}

func removeBackup(backup string) error {
	return os.RemoveAll(backup)
}

type TrendingLabel struct {
	Name       string `json:"name"`
	RenameTo   string `json:"rename_to"`
	BotTaskId  int64  `json:"bot_task_id"`
}

func getTrendingLabelsForDeployment() ([]TrendingLabel, error) {
	trendingLabels := []TrendingLabel{}

	rows, err := db.Query(`SELECT b.id, s.name, b.rename_to
					  	   FROM trending_label_suggestion t
					  	   JOIN trending_label_bot_task b ON b.trending_label_suggestion_id = t.id 
					  	   JOIN label_suggestion s ON s.id = t.label_suggestion_id
						   WHERE t.closed = false AND b.state='merged'`)
	if err != nil {
		return trendingLabels, err
	}

	defer rows.Close()

	for rows.Next() {
		var trendingLabel TrendingLabel
		err = rows.Scan(&trendingLabel.BotTaskId, &trendingLabel.Name, &trendingLabel.RenameTo)
		if err != nil {
			return trendingLabels, err
		}

		trendingLabels = append(trendingLabels, trendingLabel)
	}

	return trendingLabels, nil
}

func main() {
	labelsRepositoryUrl := flag.String("labels_repository_url", "https://github.com/bbernhard/imagemonkey-labels-test", "Labels Repository URL")
	labelsDir := flag.String("labels_dir", "/tmp/labels", "Labels Location")
	backupDir := flag.String("backup_dir", "/tmp/labels-backup", "Backup Location")
	autoCloseGithubIssue := flag.Bool("autoclose_github_issue" , false, "automatically close trending label github issue")
	singleshot := flag.Bool("singleshot", true, "singleshot")
	useSentry := flag.Bool("use_sentry", false, "Use Sentry")

	flag.Parse()

	if *useSentry {
		log.Info("Setting Sentry DSN")
		raven.SetDSN(commons.MustGetEnv("SENTRY_DSN"))
		raven.SetEnvironment("labels-downloader")

		raven.CaptureMessage("Starting up labels downloader", nil)
	}
	
	var err error
	//open database and make sure that we can ping it
	imageMonkeyDbConnectionString := commons.MustGetEnv("IMAGEMONKEY_DB_CONNECTION_STRING")
	db, err = sql.Open("postgres", imageMonkeyDbConnectionString)
	if err != nil {
		raven.CaptureError(err, nil)
		log.Fatal("Couldn't open database: ", err.Error())
	}

	err = db.Ping()
	if err != nil {
		raven.CaptureError(err, nil)
		log.Fatal("Couldn't ping database: ", err.Error())
	}
	defer db.Close()
	
	labelsDownloader := clients.NewLabelsDownloader(*labelsRepositoryUrl, *labelsDir)
	
	labelsPath := *labelsDir + "/en/labels.jsonnet"
	labelRefinementsPath := *labelsDir + "/en/label-refinements.json"
	metalabelsPath := *labelsDir + "/en/metalabels.jsonnet" 
	labelsPopulator := clients.NewLabelsPopulatorClient(imageMonkeyDbConnectionString, labelsPath, labelRefinementsPath, metalabelsPath) 
	err = labelsPopulator.Load()
	if err != nil {
		raven.CaptureError(err, nil)
		log.Fatal(err.Error())
	}

	makeLabelsProductive := clients.NewMakeLabelsProductiveClient(imageMonkeyDbConnectionString, labelsPath, metalabelsPath, false, *autoCloseGithubIssue)
	err = makeLabelsProductive.Load()
	if err != nil {
		raven.CaptureError(err, nil)
		log.Fatal(err.Error())
	}

	firstIteration := true
	for {
		if !firstIteration {
			if *singleshot {
				return
			}
			time.Sleep(1 * time.Second)
		} else {	
			firstIteration = false
		}
		
		trendingLabelsForDeployment, err := getTrendingLabelsForDeployment()
		if err != nil {
			log.Error("Couldn't get trending labels for deployment: ", err.Error())
			raven.CaptureError(err, nil)
			return
		}

		if len(trendingLabelsForDeployment) > 0 {
			backupPath := *backupDir + "/" + strconv.FormatInt(time.Now().Unix(), 10)
			
			err := createBackup(*labelsDir, backupPath)
			if err != nil {
				log.Error("Couldn't create backup: ", err.Error())
				raven.CaptureError(err, nil)
				return
			}

			log.Info("Downloading")
			err = labelsDownloader.Download()
			if err != nil {
				raven.CaptureError(err, nil)
				log.Error("Couldn't download labels: ", err.Error())
				log.Info("Restoring backup again..")
				os.RemoveAll(*labelsDir)
				err = restoreBackup(backupPath, *labelsDir)
				if err != nil {
					raven.CaptureError(err, nil)
					log.Error("Couldn't restore backup: ", err.Error())	
				}
				return
			}

			log.Info("Populating labels...")
			err = labelsPopulator.Populate(false)
			if err != nil {
				raven.CaptureError(err, nil)
				log.Error("Couldn't populate labels: ", err.Error())
				log.Info("Restoring backup again..")
				os.RemoveAll(*labelsDir)
				err = restoreBackup(backupPath, *labelsDir)
				if err != nil {
					raven.CaptureError(err, nil)
					log.Error("Couldn't restore backup: ", err.Error())	
				}
				return
			}
			
			log.Info("Making trending labels productive...dryrun")
			dryRun := true
			for i := 0; i < 2; i++ {
				if i == 1 {
					log.Info("Making trending labels productive")
					dryRun = false
				}
				for _, trendingLabel := range trendingLabelsForDeployment {
					err = makeLabelsProductive.DoIt(trendingLabel.Name, trendingLabel.RenameTo, dryRun)
					if err != nil {
						raven.CaptureError(err, nil)
						log.Error("Couldn't make trending label productive: ", err.Error()) 
						return
					}
				}
			}

			log.Info("All done...removing backup")
			err = removeBackup(backupPath)
			if err != nil {
				raven.CaptureError(err, nil)
				log.Error("Couldn't remove backup: ", err.Error())
				return
			}
		}
	}
}
