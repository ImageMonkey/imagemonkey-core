package main

import (
	"time"
	log "github.com/sirupsen/logrus"
	"github.com/google/go-github/github"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/getsentry/raven-go"
	"flag"
	"context"
	"golang.org/x/oauth2"
	"fmt"
	"github.com/lib/pq"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	imagemonkeydb "github.com/bbernhard/imagemonkey-core/database"
	commons "github.com/bbernhard/imagemonkey-core/commons"
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

func handleRecurringLabelSuggestions() error {
	type ResultEntry struct {
		ImageId string
		Annotatable bool
		LabelMeEntry datastructures.LabelMeEntry
		ProductionLabelId int64
		LabelSuggestion string
	}

	tx, err := db.Begin()
    if err != nil {
    	log.Error("[Mark label suggestion as productive] Couldn't begin trensaction: ", err.Error())
        return err
    }

	rows, err := tx.Query(`SELECT s.id, i.key, ils.annotatable, l.name, COALESCE(pl.name, ''), t.productive_label_id, s.name
							FROM label_suggestion s
							JOIN trending_label_suggestion t ON t.label_suggestion_id = s.id
							JOIN image_label_suggestion ils ON ils.label_suggestion_id = s.id
							JOIN image i ON i.id = ils.image_id
							JOIN label l ON l.id = t.productive_label_id
							LEFT JOIN label pl ON l.parent_id = pl.id
							WHERE t.github_issue_id is not null AND t.productive_label_id is not null`)
	if err != nil {
		tx.Rollback()
		log.Error("[Mark label suggestions as productive] Couldn't get entries: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}
	defer rows.Close()

	labelSuggestionIds := []int64{}
	results := []ResultEntry{}
	for rows.Next() {
		var labelSuggestionId int64
		var label1 string
		var label2 string
		var resultEntry ResultEntry
		err = rows.Scan(&labelSuggestionId, &resultEntry.ImageId, &resultEntry.Annotatable, &label1, &label2, 
						&resultEntry.ProductionLabelId, &resultEntry.LabelSuggestion)
		if err != nil {
			tx.Rollback()
			log.Error("[Mark label suggestions as productive] Couldn't scan row: ", err.Error())
			raven.CaptureError(err, nil)
			return err
		}

		if label2 == "" {
            resultEntry.LabelMeEntry.Label = label1
        } else {
            resultEntry.LabelMeEntry.Label = label2
            resultEntry.LabelMeEntry.Sublabels = append(resultEntry.LabelMeEntry.Sublabels, 
            											datastructures.Sublabel{Name: label1})
        }
        results = append(results, resultEntry)
		labelSuggestionIds = append(labelSuggestionIds, labelSuggestionId)
	}
	rows.Close()

	if len(labelSuggestionIds) > 0 {
		for _, elem := range results {
			labels := []datastructures.LabelMeEntry{}
			labels = append(labels, elem.LabelMeEntry)
	    	
			numOfNotAnnotatable := 0
			if elem.Annotatable {
				numOfNotAnnotatable = 0
			} else {
				//if label is not annotatable, set num_of_not_annotatable to 10
				numOfNotAnnotatable = 10
			}
				
			_, err = imagemonkeydb.AddLabelsToImageInTransaction("", elem.ImageId, labels, 0, numOfNotAnnotatable, tx)  
			if err != nil {
				//transaction already rolled back in AddLabelsToImageInTransaction()
				log.Error("[Mark label suggestions as productive] Couldn't add labels: ", err.Error())
				raven.CaptureError(err, nil)
				return err
			}

			if len(elem.LabelMeEntry.Sublabels) == 0 {
				err = imagemonkeydb.MakeAnnotationsProductive(tx, elem.LabelSuggestion, elem.ProductionLabelId)
				if err != nil {
					tx.Rollback()
					log.Error("[Mark label suggestions as productive] Couldn't make annotations productive", err.Error())
					raven.CaptureError(err, nil)
					return err
				}
			}
		}


		//remove label suggestions
		_, err := tx.Exec(`DELETE FROM image_label_suggestion s
						   WHERE s.label_suggestion_id = ANY($1)`, pq.Array(labelSuggestionIds))
		if err != nil {
			tx.Rollback()
			log.Error("[Mark label suggestions as productive] Couldn't delete label suggestions: ", err.Error())
			raven.CaptureError(err, nil)
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Error("[Mark label suggestions as productive] Couldn't commit transaction: ", err.Error())
		return err
	}
	return nil
}

func getNewTrendingLabels(trendingLabelTreshold int) ([]TrendingLabel, error) {
	var trendingLabels []TrendingLabel

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

func createGithubTicket(trendingLabel TrendingLabel, repository string, githubProjectOwner string, githubApiToken string) (TrendingLabel, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubApiToken},
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

	issue, _, err := client.Issues.Create(ctx, githubProjectOwner, repository, issueRequest)

	if err == nil {
		trendingLabel.GithubIssue.Id = *issue.Number
		trendingLabel.GithubIssue.Exists = true
	}

	return trendingLabel, err
}

func addCommentToGithubTicket(trendingLabel TrendingLabel, repository string, 
		githubProjectOwner string, githubApiToken string) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubApiToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	body := fmt.Sprintf("New label count: %d" ,trendingLabel.Count)

	//create a new comment
	commentRequest := &github.IssueComment{
		Body:    github.String(body),
	}

	_, _, err := client.Issues.CreateComment(ctx, githubProjectOwner, repository, trendingLabel.GithubIssue.Id, commentRequest)

	return err
}

func updateSentTrendingLabelCount(trendingLabel TrendingLabel) error {
	_, err := db.Exec(`INSERT INTO trending_label_suggestion(label_suggestion_id, num_of_last_sent, github_issue_id, closed) VALUES($1, $2, $3, $4)
						ON CONFLICT(label_suggestion_id) DO UPDATE SET num_of_last_sent = $2, github_issue_id = $3`, 
						trendingLabel.Id, trendingLabel.Count, trendingLabel.GithubIssue.Id, false)
	return err
}

func main() {

	useSentry := flag.Bool("use_sentry", false, "Use Sentry for error logging")
	singleshot := flag.Bool("singleshot", false, "Terminate after work is done")
	repository := flag.String("repository", "imagemonkey-trending-labels-test", "Github repository")
	trendingLabelsTreshold := flag.Int("treshold", 20, "Trending labels treshold")
	useGithub := flag.Bool("use_github", true, "Create Issue in Issues tracker")
	flag.Parse()


	githubProjectOwner := ""
	githubApiToken := ""
	
	if *useGithub {
		githubProjectOwner = commons.MustGetEnv("GITHUB_PROJECT_OWNER")
		githubApiToken = commons.MustGetEnv("GITHUB_API_TOKEN")
	} else {
		githubProjectOwner = commons.GetEnv("GITHUB_PROJECT_OWNER")
		githubApiToken = commons.GetEnv("GITHUB_API_TOKEN")
	}

	if *useSentry {
		log.Info("Setting Sentry DSN")
		sentryDsn := commons.MustGetEnv("SENTRY_DSN")
		raven.SetDSN(sentryDsn)
		raven.SetEnvironment("trending-labels")
		raven.CaptureMessage("Starting up trending-labels worker", nil)
	}
	log.Info("[Main] Starting up Trending Labels Worker...")

	imageMonkeyDbConnectionString := commons.MustGetEnv("IMAGEMONKEY_DB_CONNECTION_STRING")	

	var err error
	db, err = sql.Open("postgres", imageMonkeyDbConnectionString)
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

	for {
		trendingLabels, err := getNewTrendingLabels(*trendingLabelsTreshold)
		if err != nil {
			log.Error("[Main] Couldn't get trending labels: ", err.Error())
			raven.CaptureError(err, nil)
		} else {
			for _, trendingLabel := range trendingLabels {
				log.Info("[Main] Detected a new trending label: ", trendingLabel.Name)
				var githubErr error
				if !trendingLabel.GithubIssue.Exists {
					githubErr = nil
					var t TrendingLabel
					if *useGithub {
						//there is a new trending label...create a github ticket for that
						log.Info("[Main] Creating Github ticket for trending label: ", trendingLabel.Name)
						t, githubErr = createGithubTicket(trendingLabel, *repository, githubProjectOwner, githubApiToken)
						if githubErr != nil {
							log.Error("[Main] Couldn't create github issue for trending label: ", err.Error())
							raven.CaptureError(err, nil)
						}
					} else {
						t = trendingLabel
					}

					if githubErr == nil {
						err := updateSentTrendingLabelCount(t)
						if err != nil {
							log.Error("[Main] Couldn't mark trending label as sent: ", err.Error())
							raven.CaptureError(err, nil)
						}
					}
				} else { //ticket exists, just add a comment
					githubErr = nil
					if *useGithub {
						githubErr = addCommentToGithubTicket(trendingLabel, *repository, githubProjectOwner, githubApiToken)
						if githubErr != nil {
							log.Error("[Main] Couldn't update trending label count for trending label: ", err.Error())
							raven.CaptureError(err, nil)
						} 
					}

					if githubErr == nil {
						err := updateSentTrendingLabelCount(trendingLabel)
						if err != nil {
							log.Error("[Main] Couldn't mark trending label as sent: ", err.Error())
							raven.CaptureError(err, nil)
						}
					}
				}
			}
		}

		//in case someone adds a trending label that was already made productive, we can 
		//transition the label suggestion automatically to productive. 
		err = handleRecurringLabelSuggestions()
		if err != nil {
			log.Error("[Main] Couldn't mark trending labels as productive: ", err.Error())
		}

		if *singleshot {
			return
		}

		time.Sleep((time.Second * 120)) //sleep for 120 seconds
    }
}
