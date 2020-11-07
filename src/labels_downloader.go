package main

import (
	log "github.com/sirupsen/logrus"
	commons "github.com/bbernhard/imagemonkey-core/commons"
	clients "github.com/bbernhard/imagemonkey-core/clients"
	"github.com/getsentry/raven-go"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4"
	"time"
	"flag"
	"os"
	"strconv"
	"context"
)

var db *pgx.Conn

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
	Name                string `json:"name"`
	RenameTo            string `json:"rename_to"`
	BotTaskId           int64  `json:"bot_task_id"`
	ParentLabelId       int64  `json:"parent_label_id"`
}

func getLabelUuid(name string, parentLabelId int64) (string, error) {
	var uuid string
	err := db.QueryRow(context.TODO(),
				`SELECT l.uuid
					FROM label l
					JOIN label pl ON l.parent_id = pl.id
					WHERE l.name = $1 AND pl.id = $2`, name, parentLabelId).Scan(&uuid)
	if err != nil {
		return "", err
	}

	return uuid, nil
}

func getTrendingLabelsForDeployment() ([]TrendingLabel, error) {
	trendingLabels := []TrendingLabel{}

	rows, err := db.Query(context.TODO(),
						  `SELECT b.id, s.name, b.rename_to, COALESCE(pl.id, -1) 
					  	   FROM trending_label_suggestion t
					  	   JOIN trending_label_bot_task b ON b.trending_label_suggestion_id = t.id 
					  	   JOIN label_suggestion s ON s.id = t.label_suggestion_id
						   LEFT JOIN label pl ON pl.id = b.parent_label_id
						   WHERE t.closed = false AND b.state='merged'`)
	if err != nil {
		return trendingLabels, err
	}

	defer rows.Close()

	for rows.Next() {
		var trendingLabel TrendingLabel
		err = rows.Scan(&trendingLabel.BotTaskId, &trendingLabel.Name, &trendingLabel.RenameTo, &trendingLabel.ParentLabelId)
		if err != nil {
			return trendingLabels, err
		}

		trendingLabels = append(trendingLabels, trendingLabel)
	}

	return trendingLabels, nil
}

func restoreBackupDueToError(labelsDir string, backupPath string) {
	log.Info("Restoring backup again..")
	os.RemoveAll(labelsDir)
	err := restoreBackup(backupPath, labelsDir)
	if err != nil {
		raven.CaptureError(err, nil)
		log.Error("Couldn't restore backup: ", err.Error())	
	}
}

func setTrendingLabelBotTaskStateProductive(id int64) error {
	_, err := db.Exec(context.TODO(),
					   `UPDATE trending_label_bot_task 
					  	SET state = 'productive'
						WHERE id = $1`, id)
	return err
}

func main() {
	labelsRepositoryUrl := flag.String("labels_repository_url", "https://github.com/bbernhard/imagemonkey-labels-test", "Labels Repository URL")
	trendingLabelsRepositoryName := flag.String("trending_labels_repository_name" , "", "Trending Labels Repository Name")
	trendingLabelsRepositoryOwner := flag.String("trending_labels_repository_owner", "", "Trending Labels Repository Owner")
	labelsDir := flag.String("labels_dir", "/tmp/labels", "Labels Location")
	downloadDir := flag.String("download_dir", "/tmp/labels", "Download Location")
	backupDir := flag.String("backup_dir", "/tmp/labels-backup", "Backup Location")
	autoCloseGithubIssue := flag.Bool("autoclose_github_issue" , false, "automatically close trending label github issue")
	singleshot := flag.Bool("singleshot", true, "singleshot")
	useSentry := flag.Bool("use_sentry", false, "Use Sentry")
	useBackupTimestamp := flag.Bool("use_backup_timestamp", true, "Create backups with unix timestamps")
	pollingInterval := flag.Int("polling_interval", 1, "Polling Interval")
	redisAddress := flag.String("redis_address", ":6379", "Address to the Redis server")
	redisMaxConnections := flag.Int("redis_max_connections", 5, "Max connections to Redis")

	flag.Parse()

	log.SetLevel(log.DebugLevel)
	log.Info("Starting ImageMonkey labels downloader")

	githubApiToken := ""
	if *trendingLabelsRepositoryName != "" {
		githubApiToken = commons.MustGetEnv("IMAGEMONKEY_BOT_GITHUB_API_TOKEN")
	}

	if *useSentry {
		log.Info("Setting Sentry DSN")
		raven.SetDSN(commons.MustGetEnv("SENTRY_DSN"))
		raven.SetEnvironment("labels-downloader")

		raven.CaptureMessage("Starting up labels downloader", nil)
	}

	//create redis pool
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", *redisAddress)

		if err != nil {
			log.Fatal("[Main] Couldn't dial redis: ", err.Error())
		}

		return c, err
	}, *redisMaxConnections)
	defer redisPool.Close()

	
	var err error
	//open database and make sure that we can ping it
	imageMonkeyDbConnectionString := commons.MustGetEnv("IMAGEMONKEY_DB_CONNECTION_STRING")
	db, err = pgx.Connect(context.Background(), imageMonkeyDbConnectionString)
	if err != nil {
		raven.CaptureError(err, nil)
		log.Fatal("Couldn't open database: ", err.Error())
	}

	err = db.Ping(context.Background())
	if err != nil {
		raven.CaptureError(err, nil)
		log.Fatal("Couldn't ping database: ", err.Error())
	}
	defer db.Close(context.Background())
	
	labelsDownloader := clients.NewLabelsDownloader(*labelsRepositoryUrl, *downloadDir)
	
	labelsPath := *labelsDir + "/en/labels.jsonnet"
	autogeneratedBaseLabelsPath := *labelsDir + "/en/includes/labels/autogenerated"
	autogeneratedLabelsPath := autogeneratedBaseLabelsPath + ".libsonnet"
	autogeneratedBaseMetaLabelsPath := *labelsDir + "/en/includes/metalabels/autogenerated"
	autogeneratedMetaLabelsPath := autogeneratedBaseMetaLabelsPath + ".libsonnet"
	labelRefinementsPath := *labelsDir + "/en/label-refinements.json"
	metalabelsPath := *labelsDir + "/en/metalabels.jsonnet" 

	firstIteration := true
	for {
		if !firstIteration {
			if *singleshot {
				return
			}
			time.Sleep(time.Duration(*pollingInterval) * time.Second)
		} else {	
			firstIteration = false
		}
		
		log.Debug("Checking for trending labels in merged state")
		trendingLabelsForDeployment, err := getTrendingLabelsForDeployment()
		if err != nil {
			log.Error("Couldn't get trending labels for deployment: ", err.Error())
			raven.CaptureError(err, nil)
		}

		if err == nil && len(trendingLabelsForDeployment) > 0 {
			backupPath := *backupDir
			if *useBackupTimestamp {
				backupPath = backupPath + "/" + strconv.FormatInt(time.Now().Unix(), 10)

			}

			log.Info("Downloading")
			err = labelsDownloader.Download()
			if err != nil {
				raven.CaptureError(err, nil)
				log.Error("Couldn't download labels: ", err.Error())
				restoreBackupDueToError(*labelsDir, backupPath)	
				continue
			}

			err := createBackup(*labelsDir, backupPath)
			if err != nil {
				log.Error("Couldn't create backup: ", err.Error())
				raven.CaptureError(err, nil)
				continue
			}

			//move downloaded labels to labels folder
			err = os.Rename(*downloadDir, *labelsDir)
			if err != nil {
				log.Error("Couldn't move labels folder to final destination: ", err.Error())
				raven.CaptureError(err, nil)
				restoreBackupDueToError(*labelsDir, backupPath)
				continue
			}

			log.Info("Merging labels into one file...")
			labelsDirectoryMerger := commons.NewLabelsDirectoryMerger(autogeneratedBaseLabelsPath, autogeneratedLabelsPath)
			err = labelsDirectoryMerger.Merge()
			if err != nil {
				raven.CaptureError(err, nil)
				log.Error("Couldn't merge labels into one file: ", err.Error())
				restoreBackupDueToError(*labelsDir, backupPath)
				return
			}

			log.Info("Merging metalabels into one file...")
			metaLabelsDirectoryMerger := commons.NewMetaLabelsDirectoryMerger(autogeneratedBaseMetaLabelsPath, autogeneratedMetaLabelsPath)
			err = metaLabelsDirectoryMerger.Merge()
			if err != nil {
				raven.CaptureError(err, nil)
				log.Error("Couldn't merge metalabels into one file: ", err.Error())
				restoreBackupDueToError(*labelsDir, backupPath)
				continue
			}

			log.Info("Populating labels...")
			labelsPopulator := clients.NewLabelsPopulatorClient(imageMonkeyDbConnectionString, labelsPath, 
									labelRefinementsPath, metalabelsPath) 
			err = labelsPopulator.Load()
			if err != nil {
				raven.CaptureError(err, nil)
				log.Error("Couldn't load labels populator: ", err.Error())
				restoreBackupDueToError(*labelsDir, backupPath)
				continue
			}

			err = labelsPopulator.Populate(false)
			if err != nil {
				raven.CaptureError(err, nil)
				log.Error("Couldn't populate labels: ", err.Error())
				restoreBackupDueToError(*labelsDir, backupPath)
				continue
			}
			
			log.Info("Making trending labels productive...dryrun")
			makeLabelsProductive := clients.NewMakeLabelsProductiveClient(imageMonkeyDbConnectionString, labelsPath, 
											metalabelsPath, false, *autoCloseGithubIssue)

			makeLabelsProductive.SetGithubRepository(*trendingLabelsRepositoryName)
			makeLabelsProductive.SetGithubRepositoryOwner(*trendingLabelsRepositoryOwner)
			makeLabelsProductive.SetGithubApiToken(githubApiToken)

			err = makeLabelsProductive.Load()
			if err != nil {
				raven.CaptureError(err, nil)
				log.Error("Couldn't load make labels productive client: ", err.Error())
				restoreBackupDueToError(*labelsDir, backupPath)
				continue
			}
			
			dryRun := true
			for i := 0; i < 2; i++ {
				if i == 1 {
					log.Info("Making trending labels productive")
					dryRun = false
				}
				for _, trendingLabel := range trendingLabelsForDeployment {
					renameTo := trendingLabel.RenameTo
					if trendingLabel.ParentLabelId != -1 {
						renameTo, err = getLabelUuid(trendingLabel.RenameTo, trendingLabel.ParentLabelId)
						if err != nil {
							raven.CaptureError(err, nil)
							log.Error("Couldn't get label uuid: ", err.Error()) 
							continue
						}
					}

					err = makeLabelsProductive.DoIt(trendingLabel.Name, renameTo, dryRun)
					if err != nil {
						raven.CaptureError(err, nil)
						log.Error("Couldn't make trending label productive: ", err.Error()) 
						continue
					}

					err = setTrendingLabelBotTaskStateProductive(trendingLabel.BotTaskId)
					if err != nil {
						raven.CaptureError(err, nil)
						log.Error("Couldn't set trending label bot task state to 'productive' for id ", trendingLabel.BotTaskId, ": ", err.Error())
						continue
					}
				}
			}

			log.Info("All done...removing backup")
			err = removeBackup(backupPath)
			if err != nil {
				raven.CaptureError(err, nil)
				log.Error("Couldn't remove backup: ", err.Error())
			}
			
			redisConn := redisPool.Get()
			_, err = redisConn.Do("PUBLISH", "tasks", "reloadlabels")
			if err != nil {
				raven.CaptureError(err, nil)
				log.Error("Couldn't publish message: ", err.Error())
			}
			redisConn.Close()
		}
	}
}
