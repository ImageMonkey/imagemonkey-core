package main

import (
	log "github.com/sirupsen/logrus"
	commons "github.com/bbernhard/imagemonkey-core/commons"
	_ "github.com/lib/pq"
	"gopkg.in/src-d/go-git.v4"
	"time"
	"database/sql"
	"flag"
	"os"
	"os/exec"
	"strconv"
)

var db *sql.DB

type LabelsDownloader struct {
	repositoryUrl string
	downloadLocation string
}

func NewLabelsDownloader(repositoryUrl string, downloadLocation string) *LabelsDownloader {
	return &LabelsDownloader{
		repositoryUrl: repositoryUrl,
		downloadLocation: downloadLocation,
	}
}

func (p *LabelsDownloader) Download() error {
	os.RemoveAll(p.downloadLocation)
	_, err := git.PlainClone(p.downloadLocation, false, &git.CloneOptions{
		URL:      p.repositoryUrl,
		Progress: os.Stdout,
	})
	return err
}

type CommandsRunner struct {
	directory string
}

func NewCommandsRunner(directory string) *CommandsRunner {
	return &CommandsRunner{
		directory: directory,
	}
}


func (p *CommandsRunner) MakeTrendingLabelsProductive(name string, renameTo string, dryRun bool, autoCloseGithubIssue bool) error {
	cmd := exec.Command("./make_trending_labels_productive", "-trendinglabel", name, "-renameto", renameTo, 
							"-dryrun=" +strconv.FormatBool(dryRun), "autoclose="+strconv.FormatBool(autoCloseGithubIssue))
	cmd.Dir = p.directory
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}

	// Wait for the process to finish
	done := make(chan error, 1)
	go func() {
	    done <- cmd.Wait()
	}()
	select {
	case err := <-done:
	    return err
	}
}

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
						   WHERE t.closed = false AND 
						   (b.state = 'accepted' OR b.state='pending' OR b.state='building' 
						   	OR b.state='merged')`)
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
	labelsRepositoryUrl := flag.String("labels_repository_url", "https://github.com/bbernhard/imagemonkey-trending-labels-test", "Labels Repository URL")
	labelsLocation := flag.String("labels_dir", "/tmp/labels", "Labels Location")
	backupLocation := flag.String("backup_dir", "/tmp/labels.bak", "Backup Location")
	binariesLocation := flag.String("binaries_dir", "./", "Binaries Location")
	autoCloseGithubIssue := flag.Bool("autoclose_github_issue" , false, "automatically close trending label github issue")
	singleshot := flag.Bool("singleshot", true, "singleshot")

	flag.Parse()
	
	var err error
	//open database and make sure that we can ping it
	imageMonkeyDbConnectionString := commons.MustGetEnv("IMAGEMONKEY_DB_CONNECTION_STRING")
	db, err = sql.Open("postgres", imageMonkeyDbConnectionString)
	if err != nil {
		log.Fatal("Couldn't open database: ", err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Couldn't ping database: ", err.Error())
	}
	defer db.Close()
	
	labelsDownloader := NewLabelsDownloader(*labelsRepositoryUrl, *labelsLocation)
	commandsRunner := NewCommandsRunner(*binariesLocation)
	
	labelsPath := *labelsLocation + "/en/labels.jsonnet"
	labelRefinementsPath := *labelsLocation + "/en/label-refinements.json"
	metalabelsPath := *labelsLocation + "/en/metalabels.jsonnet" 
	labelsPopulator := commons.NewLabelsPopulator(imageMonkeyDbConnectionString, labelsPath, labelRefinementsPath, metalabelsPath) 
	err = labelsPopulator.Load()
	if err != nil {
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
			return
		}

		if len(trendingLabelsForDeployment) > 0 {
			err := createBackup(*labelsLocation, *backupLocation)
			if err != nil {
				log.Error("Couldn't create backup: ", err.Error())
				return
			}

			log.Info("Downloading")
			err = labelsDownloader.Download()
			if err != nil {
				log.Error("Couldn't download labels: ", err.Error())
				log.Info("Restoring backup again..")
				os.RemoveAll(*labelsLocation)
				err = restoreBackup(*backupLocation, *labelsLocation)
				if err != nil {
					log.Error("Couldn't restore backup: ", err.Error())	
				}
				return
			}

			log.Info("Populating labels...")
			err = labelsPopulator.Populate(false)
			if err != nil {
				log.Error("Couldn't populate labels: ", err.Error())
				log.Info("Restoring backup again..")
				os.RemoveAll(*labelsLocation)
				err = restoreBackup(*backupLocation, *labelsLocation)
				if err != nil {
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
					err = commandsRunner.MakeTrendingLabelsProductive(trendingLabel.Name, trendingLabel.RenameTo, dryRun, *autoCloseGithubIssue)
					if err != nil {
						log.Error("Couldn't make trending label productive: ", err.Error()) 
						return
					}
				}
			}

			log.Info("All done...removing backup")
			err = removeBackup(*backupLocation)
			if err != nil {
				log.Error("Couldn't remove backup: ", err.Error())
				return
			}
		}
	}
}
