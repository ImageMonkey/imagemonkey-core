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
	"fmt"
)

var db *sql.DB

type TrendingLabel struct {
	Name string `json:"name"`
	Id string `json:"id"`
	Count int32 `json:"count"`

	GithubIssue struct {
        Id int `json:"id"`
        Exists bool `json:"exists"`
    } `json:"github_issue"`
}

func getNewTrendingLabels() ([]TrendingLabel, error) {
	var trendingLabels []TrendingLabel

	trendingLabelTreshold := 20
	rows, err := db.Query(`SELECT s.id, s.name, COUNT(t.id), COALESCE(github_issue_id, -1) FROM label_suggestion s 
							JOIN image_label_suggestion i ON i.label_suggestion_id = s.id
							LEFT JOIN trending_label_suggestion t ON t.label_suggestion_id = s.id
							GROUP BY s.name, s.id, num_of_last_sent, github_issue_id
							HAVING COUNT(*) > (COALESCE(num_of_last_sent, 0) + $1)`, trendingLabelTreshold) 
	if err != nil {
		return trendingLabels, err
	}
	defer rows.Close()

	for rows.Next() {
		var trendingLabel TrendingLabel 
		err := rows.Scan(&trendingLabel.Id, &trendingLabel.Name, &trendingLabel.Count, &trendingLabel.GithubIssue.Id)
		if err != nil {
			return trendingLabels, err
		}

		if trendingLabel.GithubIssue.Id == -1 {
			trendingLabel.GithubIssue.Exists = false
		} else {
			trendingLabel.GithubIssue.Exists = true
		}

		trendingLabels = append(trendingLabels, trendingLabel)
	}

	return trendingLabels, nil
}

func createGithubTicket(trendingLabel TrendingLabel, repository string) (TrendingLabel, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GITHUB_API_TOKEN},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	title := trendingLabel.Name + " is now trending"
	body := ""

	//create a new Issue
	issueRequest := &github.IssueRequest{
		Title:    github.String(title),
		Body:    github.String(body),
	}

	issue, _, err := client.Issues.Create(ctx, GITHUB_PROJECT_OWNER, repository, issueRequest)

	if err == nil {
		trendingLabel.GithubIssue.Id = *issue.Number
		trendingLabel.GithubIssue.Exists = true
	}

	return trendingLabel, err
}

func addCommentToGithubTicket(trendingLabel TrendingLabel, repository string) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: GITHUB_API_TOKEN},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	body := fmt.Sprintf("New label count: %d" ,trendingLabel.Count)

	//create a new comment
	commentRequest := &github.IssueComment{
		Body:    github.String(body),
	}

	_, _, err := client.Issues.CreateComment(ctx, GITHUB_PROJECT_OWNER, repository, trendingLabel.GithubIssue.Id, commentRequest)

	return err
}

func updateSentTrendingLabelCount(trendingLabel TrendingLabel) error {
	_, err := db.Exec(`INSERT INTO trending_label_suggestion(label_suggestion_id, num_of_last_sent, github_issue_id) VALUES($1, $2, $3)
						ON CONFLICT(label_suggestion_id) DO UPDATE SET num_of_last_sent = $2, github_issue_id = $3`, 
						trendingLabel.Id, trendingLabel.Count, trendingLabel.GithubIssue.Id)
	return err
}

func main() {

	useSentry := flag.Bool("use_sentry", false, "Use Sentry for error logging")
	repository := flag.String("repository", "imagemonkey-trending-labels-test", "Github repository")
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

	/*var trendingLabel TrendingLabel
	trendingLabel.Name = "bla"
	err = createGithubTicket(trendingLabel, repository)
	if err != nil {
		log.Info(err.Error())
	}*/

	for {
		trendingLabels, err := getNewTrendingLabels()
		if err != nil {
			log.Error("[Main] Couldn't get trending labels: ", err.Error())
			raven.CaptureError(err, nil)
		} else {
			for _, trendingLabel := range trendingLabels {
				log.Info("[Main] Detected a new trending label: ", trendingLabel.Name)

				if !trendingLabel.GithubIssue.Exists {
					//there is a new trending label...create a github ticket for that
					log.Info("[Main] Creating Github ticket for trending label: ", trendingLabel.Name)
					tl, err := createGithubTicket(trendingLabel, *repository)
					if err != nil {
						log.Error("[Main] Couldn't create github issue for trending label: ", err.Error())
						raven.CaptureError(err, nil)
					} else {
						err := updateSentTrendingLabelCount(tl)
						if err != nil {
							log.Error("[Main] Couldn't mark trending label as sent: ", err.Error())
							raven.CaptureError(err, nil)
						}
					}
				} else { //ticket exists, just add a comment
					err = addCommentToGithubTicket(trendingLabel, *repository)
					if err != nil {
						log.Error("[Main] Couldn't update trending label count for trending label: ", err.Error())
						raven.CaptureError(err, nil)
					} else {
						err := updateSentTrendingLabelCount(trendingLabel)
						if err != nil {
							log.Error("[Main] Couldn't mark trending label as sent: ", err.Error())
							raven.CaptureError(err, nil)
						}
					}
				}
			}
		}
		time.Sleep((time.Second * 60)) //sleep for 60 seconds
    }
}