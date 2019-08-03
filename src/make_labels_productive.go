package main

import (
	log "github.com/sirupsen/logrus"
	"flag"
	"database/sql"
	commons "github.com/bbernhard/imagemonkey-core/commons"
	clients "github.com/bbernhard/imagemonkey-core/clients"
)

var db *sql.DB


func main() {
	log.SetLevel(log.DebugLevel)

	trendingLabel := flag.String("trendinglabel", "", "The name of the trending label that should be made productive")
	renameTo := flag.String("renameto", "", "Rename the label")
	wordlistPath := flag.String("wordlist", "../wordlists/en/labels.jsonnet", "Path to label map")
	metalabelsPath := flag.String("metalabels", "../wordlists/en/metalabels.jsonnet", "Path to metalabels map")
	dryRun := flag.Bool("dryrun", true, "Specifies whether this is a dryrun or not")
	autoCloseIssue := flag.Bool("autoclose", true, "Automatically close issue")
	githubRepository := flag.String("repository", "", "Github repository")
	strict := flag.Bool("strict", false, "strict label matching")

	flag.Parse()

	githubProjectOwner := ""
	githubApiToken := ""
	if *autoCloseIssue {
		githubProjectOwner = commons.MustGetEnv("GITHUB_PROJECT_OWNER")
		githubApiToken = commons.MustGetEnv("GITHUB_API_TOKEN")
	}

	imageMonkeyDbConnectionString := commons.MustGetEnv("IMAGEMONKEY_DB_CONNECTION_STRING")

	makeLabelsProductiveClient := clients.NewMakeLabelsProductiveClient(imageMonkeyDbConnectionString, *wordlistPath, *metalabelsPath, *strict, *autoCloseIssue)
	makeLabelsProductiveClient.SetGithubRepository(*githubRepository)
	makeLabelsProductiveClient.SetGithubRepositoryOwner(githubProjectOwner)
	makeLabelsProductiveClient.SetGithubApiToken(githubApiToken)

	err := makeLabelsProductiveClient.Load()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = makeLabelsProductiveClient.DoIt(*trendingLabel, *renameTo, *dryRun)
	if err != nil {
		log.Fatal(err.Error())
	}
}
