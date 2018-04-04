package main

import (
	"time"
	log "github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/getsentry/raven-go"
	"flag"
	"context"
	"golang.org/x/oauth2"
)

var db *sql.DB

type TrendingLabel struct {
	Name string `json:"name"`
	Id string `json:"id"`
}

func getNewTrendingLabels() ([]TrendingLabel, error) {
	var trendingLabels []TrendingLabel

	trendingLabelTreshold := 20
	rows, err := db.Query(`SELECT s.id, s.name FROM label_suggestion s 
						   JOIN image_label_suggestion i ON i.label_suggestion_id = s.id
						   WHERE s.id NOT IN (
						   	 SELECT label_suggestion_id FROM trending_label_suggestion
						   ) 
						   GROUP BY s.name, s.id
						   HAVING COUNT(*) > $1`, trendingLabelTreshold) 
	if err != nil {
		return trendingLabels, err
	}
	defer rows.Close()

	for rows.Next() {
		var trendingLabel TrendingLabel 
		err := rows.Scan(&trendingLabel.Id, &trendingLabel.Name)
		if err != nil {
			return trendingLabels, err
		}

		trendingLabels = append(trendingLabels, trendingLabel)
	}

	return trendingLabels, nil
}

func createGithubTicket(trendingLabel TrendingLabel, repository string) error {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GITHUB_API_TOKEN},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	//create a new Issue
	issueRequest := &github.IssueRequest{
		Title:    github.String("test"),
		Body:    github.String("test"),
	}

	_, _, err := client.Issues.Create(ctx, "bbernhard", repository, issueRequest)

	return err
}

func markTrendingLabelAsSent(trendingLabel TrendingLabel) error {
	_, err := db.Exec("INSERT INTO trending_label_suggestion(label_suggestion_id, sent) VALUES($1, $2)", 
						trendingLabel.Id, true)
	return err
}

func main() {

	useSentry := flag.Bool("use_sentry", false, "Use Sentry for error logging")
	flag.Parse()

	if *useSentry {
		log.Info("Setting Sentry DSN")
		raven.SetDSN(SENTRY_DSN)
		raven.SetEnvironment("trending-labels")
		raven.CaptureMessage("Starting up trending-labels worker", nil)
	}
	log.Info("[Main] Starting up Trending Labels Worker...")

	var err error
	db, err = sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
	if err != nil {
		raven.CaptureError(err, nil)
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		raven.CaptureError(err, nil)
		log.Fatal("[Main] Couldn't ping database: ", err.Error())
	}
	defer db.Close()

	var trendingLabel TrendingLabel
	trendingLabel.Name = "bla"
	err = createGithubTicket(trendingLabel, "imagemonkey-trending-labels-test")
	if err != nil {
		log.Info(err.Error())
	}

	for {
		trendingLabels, err := getNewTrendingLabels()
		if err != nil {
			log.Error("[Main] Couldn't get trending labels: ", err.Error())
			raven.CaptureError(err, nil)
		} else {
			for _, trendingLabel := range trendingLabels {
				log.Info("[Main] Detected a new trending label: ", trendingLabel.Name)
				log.Info("[Main] Creating Github ticket for trending label: ", trendingLabel.Name)
				//there is a new trending label...create a github ticket for that
				err := createGithubTicket(trendingLabel, "imagemonkey-trending-labels-test")
				if err != nil {
					log.Error("[Main] Couldn't create github issue for trending label: ", err.Error())
					raven.CaptureError(err, nil)
				} else {
					err := markTrendingLabelAsSent(trendingLabel)
					if err != nil {
						log.Error("[Main] Couldn't mark trending label as sent: ", err.Error())
						raven.CaptureError(err, nil)
					}
				}
			}
		}
		time.Sleep((time.Second * 60)) //sleep for 60 seconds
    }
}