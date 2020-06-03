package main

import (
	commons "github.com/bbernhard/imagemonkey-core/commons"
	"github.com/jackc/pgx/v4"
	log "github.com/sirupsen/logrus"
	"context"
	"flag"
	"os"
	"fmt"
	"github.com/jackc/pgtype"
	"golang.org/x/oauth2"
	"github.com/google/go-github/github"
)

var unrecoverableDeletedAnnotationData int = 0
var unrecoverableDeletedAnnotations int = 0
var githubIssueIds []int64
var numOfDeleteLabelSuggestions int = 0
var numOfDeletedImageLabelSuggestions int = 0
var numOfDeletedImageAnnotationSuggestionHistoryEntries int = 0
var typeOfScans []string = []string{"enable_indexscan", "enable_indexonlyscan"}

func removeElemFromSlice(s []int64, r int64) []int64 {
    for i, v := range s {
        if v == r {
            return append(s[:i], s[i+1:]...)
        }
    }
    return s
}

func removeImageAnnotationSuggestionHistoryEntry(tx pgx.Tx, imageAnnotationSuggestionId int64) error {
	var num int = 0
	err := tx.QueryRow(context.TODO(), `SELECT count(*) FROM image_annotation_suggestion_history WHERE id = $1`, imageAnnotationSuggestionId).Scan(&num)

	if num > 0 {
		_, err = tx.Exec(context.TODO(), `DELETE FROM image_annotation_suggestion_history WHERE id = $1`, imageAnnotationSuggestionId)
		numOfDeletedImageAnnotationSuggestionHistoryEntries += 1
	}
	return err
}

func disableTriggers(tx pgx.Tx) error {
	_, err := tx.Exec(context.TODO(), `ALTER TABLE image_annotation_suggestion DISABLE TRIGGER image_annotation_suggestion_versioning_trigger`)
	return err
}

func enableTriggers(tx pgx.Tx) error {
	_, err := tx.Exec(context.TODO(), `ALTER TABLE image_annotation_suggestion ENABLE TRIGGER image_annotation_suggestion_versioning_trigger`)
	return err
}

func persistDuplicateLabels(duplicateLabels []string) error {
	f, err := os.Create("/tmp/duplicate_labels.txt")
	if err != nil {
		return err
	}
	
	s := ""
	for _, duplicateLabel := range duplicateLabels {
		s += duplicateLabel + "\n"
	}

	_, err = f.WriteString(s)
	if err != nil {
		return err
	}

	return nil
}

func closeGithubIssue(githubIssueId int, repository string, githubProjectOwner string, githubApiToken string) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubApiToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	body := "Label is a duplicate."

	//create a new comment
	commentRequest := &github.IssueComment{
		Body:    github.String(body),
	}

	//we do not care whether we can successfully close the github issue..if it doesn't work, one can always close it
	//manually.
	_, _, err := client.Issues.CreateComment(ctx, githubProjectOwner, repository, githubIssueId, commentRequest)
	if err == nil { //if comment was successfully created, close issue
		issueRequest := &github.IssueRequest{
			State: github.String("closed"),
		}

		_, _, err = client.Issues.Edit(ctx, githubProjectOwner, repository, githubIssueId, issueRequest)
		return err
	} else {
		return err
	}

	return nil
}

//TODO: add unique constraint to label_suggestion table

//TODO: verify num of image labels + annotations before and after

func getNumOfImageAnnotationSuggestions(tx pgx.Tx) (int, error) {
	var num int
	err := tx.QueryRow(context.TODO(), `SELECT count(*) FROM image_annotation_suggestion`).Scan(&num)
	return num, err
}

func getNumOfImageAnnotationSuggestionHistoryEntries(tx pgx.Tx) (int, error) {
	var num int
	err := tx.QueryRow(context.TODO(), `SELECT count(*) FROM image_annotation_suggestion_history`).Scan(&num)
	return num, err
}

func getNumOfImages(tx pgx.Tx) (int, error) {
	var num int
	err := tx.QueryRow(context.TODO(), `SELECT count(*) FROM image`).Scan(&num)
	return num, err
}

func getNumOfLabelSuggestions(tx pgx.Tx) (int, error) {
	var num int
	err := tx.QueryRow(context.TODO(), `SELECT count(*) from label_suggestion`).Scan(&num)
	return num, err
}

func getNumOfAnnotationSuggestionData(tx pgx.Tx) (int, error) {
	var num int
	err := tx.QueryRow(context.TODO(), `SELECT count(*) FROM annotation_suggestion_data`).Scan(&num)
	return num, err
}

func getNumOfImageLabelSuggestions(tx pgx.Tx) (int, error) {
	var num int
	err := tx.QueryRow(context.TODO(), `SELECT count(*) FROM image_label_suggestion`).Scan(&num)
	return num, err
}

func getNumOfProductiveLabelSuggestions(tx pgx.Tx) (int, error) {
	var num int
	err := tx.QueryRow(context.TODO(), `SELECT count(*) FROM trending_label_suggestion WHERE closed = true`).Scan(&num)
	return num, err
}

func unifyTrendingLabelSuggestions(tx pgx.Tx, sourceIds *pgtype.Int8Array) error {
	rows, err := tx.Query(context.TODO(), `SELECT id, github_issue_id 
											FROM trending_label_suggestion 
											WHERE label_suggestion_id = ANY($1)`, sourceIds)
	if err != nil {
		return err
	}

	defer rows.Close()

	trendingLabelSuggestionIds := []int64{}
	for rows.Next() {
		var trendingLabelSuggestionId int64
		var githubIssueId int64
		err := rows.Scan(&trendingLabelSuggestionId, &githubIssueId)
		if err != nil {
			return err
		}

		trendingLabelSuggestionIds = append(trendingLabelSuggestionIds, trendingLabelSuggestionId)
		githubIssueIds = append(githubIssueIds, githubIssueId)
	}

	for _, trendingLabelSuggestionId := range trendingLabelSuggestionIds {
		_, err = tx.Exec(context.TODO(), `DELETE FROM trending_label_suggestion 
											WHERE id = $1`, trendingLabelSuggestionId)
		if err != nil {
			return err
		}
	}
	
	return nil
}

func unifyImagesWithNoDuplicateLabels(tx pgx.Tx, imageIdsThatHaveNoDuplicateLabels []int64, sourceIds *pgtype.Int8Array, target int64) error {
	//unify images with no duplicate labels
	rows, err := tx.Query(context.TODO(), `SELECT id FROM image_label_suggestion i 
											WHERE label_suggestion_id = ANY($1) AND image_id = ANY($2)`, sourceIds, imageIdsThatHaveNoDuplicateLabels)
	if err != nil {
		return err
	}
	defer rows.Close()

	imageLabelSuggestionIds := []int64{}
	for rows.Next() {
		var imageLabelSuggestionId int64
		err := rows.Scan(&imageLabelSuggestionId)
		if err != nil {
			return err
		}
		imageLabelSuggestionIds = append(imageLabelSuggestionIds, imageLabelSuggestionId)
	}

	rows.Close()

	for _, imageLabelSuggestionId := range imageLabelSuggestionIds {
		_, err := tx.Exec(context.TODO(), `UPDATE image_label_suggestion SET label_suggestion_id = $1 WHERE id = $2`, target, imageLabelSuggestionId)
		if err != nil {
			return err
		}
	}

	return nil
}

func unifyImagesWithDuplicateLabels(tx pgx.Tx, imageIdsWithDuplicateLabels []int64, sourceIds *pgtype.Int8Array, allIds *pgtype.Int8Array, target int64) error {
	for _, imageIdWithDuplicateLabels := range imageIdsWithDuplicateLabels {
		rows, err := tx.Query(context.TODO(),
					`			SELECT id 
									FROM image_label_suggestion 
									WHERE image_id = $1 AND label_suggestion_id = ANY($2)`, imageIdWithDuplicateLabels, allIds)
		if err != nil {
			return err
		}

		imageLabelSuggestionIds := []int64{}
		for rows.Next() {
			var imageLabelSuggestionId int64
			err = rows.Scan(&imageLabelSuggestionId)
			if err != nil {
				return err
			}
			imageLabelSuggestionIds = append(imageLabelSuggestionIds, imageLabelSuggestionId)
		}
		rows.Close()

		//delate all occurences, except one
		for i := 0; i < len(imageLabelSuggestionIds)-1; i++ {
			numOfDeletedImageLabelSuggestions += 1
			_, err = tx.Exec(context.TODO(), `DELETE FROM image_label_suggestion WHERE id = $1`, imageLabelSuggestionIds[i])
			if err != nil {
				return err
			}
		}

		_, err = tx.Exec(context.TODO(), 
						`UPDATE image_label_suggestion 
							SET label_suggestion_id = $1 
							WHERE id = $2`, target, imageLabelSuggestionIds[len(imageLabelSuggestionIds)-1])
		if err != nil {
			return err
		}
	}

	return nil
}

func unifyImageLabelSuggestions(tx pgx.Tx, sourceIds *pgtype.Int8Array, allIds *pgtype.Int8Array, target int64) error {
	rows, err := tx.Query(context.TODO(), `SELECT a.image_id, count(*) FROM image_label_suggestion a
											WHERE label_suggestion_id = ANY($1) GROUP BY image_id`, allIds)
	if err != nil {
		return err
	}
	defer rows.Close()

	imageIdsThatHaveNoDuplicateLabels := []int64{}
	imageIdsWithDuplicateLabels := []int64{}
	for rows.Next() {
		var count int
		var imageId int64
		err := rows.Scan(&imageId, &count)
		if err != nil {
			return err
		}

		if count > 1 {
			imageIdsWithDuplicateLabels = append(imageIdsWithDuplicateLabels, imageId)
		} else {
			imageIdsThatHaveNoDuplicateLabels = append(imageIdsThatHaveNoDuplicateLabels, imageId)
		}
	}

	rows.Close()

	err = unifyImagesWithNoDuplicateLabels(tx, imageIdsThatHaveNoDuplicateLabels, sourceIds, target)
	if err != nil {
		return err
	}

	err = unifyImagesWithDuplicateLabels(tx, imageIdsWithDuplicateLabels, sourceIds, allIds, target)
	if err != nil {
		return err
	}

	return nil
}

func unifyImageAnnotationsWithNoDuplicateLabels(tx pgx.Tx, imageIdsThatHaveNoDuplicateLabels []int64, sourceIds *pgtype.Int8Array, target int64) error {
	rows, err := tx.Query(context.TODO(), `SELECT id 
											FROM image_annotation_suggestion a 
											WHERE label_suggestion_id = ANY($1)
											AND a.image_id = ANY($2)`, sourceIds, imageIdsThatHaveNoDuplicateLabels)
	if err != nil {
		return err
	}
	defer rows.Close()

	imageAnnotationLabelSuggestionIds := []int64{}
	for rows.Next() {
		var imageAnnotationLabelSuggestionId int64
		err := rows.Scan(&imageAnnotationLabelSuggestionId)
		if err != nil {
			return err
		}
		imageAnnotationLabelSuggestionIds = append(imageAnnotationLabelSuggestionIds, imageAnnotationLabelSuggestionId)
	}

	rows.Close()

	for _, imageAnnotationLabelSuggestionId := range imageAnnotationLabelSuggestionIds {
		_, err = tx.Exec(context.TODO(), `UPDATE image_annotation_suggestion
										SET label_suggestion_id = $1 WHERE id = $2`, target, imageAnnotationLabelSuggestionId)
		if err != nil {
			return err
		}
	}

	return nil
}

func unifyImageAnnotationsWithDuplicateLabels(tx pgx.Tx, imageIdsWithDuplicateLabels []int64, allIds *pgtype.Int8Array, target int64) error {
	 for _, imageIdWithDuplicateLabels := range imageIdsWithDuplicateLabels {
		rows, err := tx.Query(context.TODO(),
					`			SELECT id 
									FROM image_annotation_suggestion 
									WHERE image_id = $1 AND label_suggestion_id = ANY($2)`, imageIdWithDuplicateLabels, allIds)
		if err != nil {
			return err
		}

		imageAnnotationSuggestionIds := []int64{}
		for rows.Next() {
			var imageAnnotationSuggestionId int64
			err = rows.Scan(&imageAnnotationSuggestionId)
			if err != nil {
				return err
			}
			imageAnnotationSuggestionIds = append(imageAnnotationSuggestionIds, imageAnnotationSuggestionId)
		}
		rows.Close()
		for _, imageAnnotationSuggestionId := range imageAnnotationSuggestionIds { 
			unrecoverableDeletedAnnotations += 1
			var unrecoverableDeletedAnnotationDataTemp int

			err = tx.QueryRow(context.TODO(), 
						`SELECT count(*) FROM annotation_suggestion_data WHERE image_annotation_suggestion_id = $1`, 
							imageAnnotationSuggestionId).Scan(&unrecoverableDeletedAnnotationDataTemp)
			if err != nil {
				return err
			}

			unrecoverableDeletedAnnotationData += unrecoverableDeletedAnnotationDataTemp
			
			_, err = tx.Exec(context.TODO(), `DELETE FROM annotation_suggestion_data
												WHERE image_annotation_suggestion_id = $1`, imageAnnotationSuggestionId)
			if err != nil {
				return err
			}

			_, err = tx.Exec(context.TODO(), `DELETE FROM user_image_annotation_suggestion 
												WHERE image_annotation_suggestion_id = $1`, imageAnnotationSuggestionId)
			if err != nil {
				return err
			}

			_, err = tx.Exec(context.TODO(), `DELETE FROM image_annotation_suggestion WHERE id = $1`, imageAnnotationSuggestionId)
			if err != nil {
				return err
			}

			err = removeImageAnnotationSuggestionHistoryEntry(tx, imageAnnotationSuggestionId)
			if err != nil {
				return err
			}
		}
	 }

	return nil
}

func unifyImageAnnotationSuggestions(tx pgx.Tx, sourceIds *pgtype.Int8Array, allIds *pgtype.Int8Array, target int64) error {
	rows, err := tx.Query(context.TODO(), `SELECT a.image_id, count(*) FROM image_annotation_suggestion a
											WHERE label_suggestion_id = ANY($1) GROUP BY image_id`, allIds)
	if err != nil {
		return err
	}
	defer rows.Close()

	imageIdsThatHaveNoDuplicateLabels := []int64{}
	imageIdsWithDuplicateLabels := []int64{}
	for rows.Next() {
		var count int
		var imageId int64
		err := rows.Scan(&imageId, &count)
		if err != nil {
			return err
		}

		if count > 1 {
			imageIdsWithDuplicateLabels = append(imageIdsWithDuplicateLabels, imageId)
		} else {
			imageIdsThatHaveNoDuplicateLabels = append(imageIdsThatHaveNoDuplicateLabels, imageId)
		}
	}

	rows.Close()


	err = unifyImageAnnotationsWithNoDuplicateLabels(tx, imageIdsThatHaveNoDuplicateLabels, sourceIds, target)
	if err != nil {
		return err
	}

	err = unifyImageAnnotationsWithDuplicateLabels(tx, imageIdsWithDuplicateLabels, allIds, target)
	if err != nil {
		return err
	}
	
	return nil
}

func unifyDuplicateLabelSuggestions(tx pgx.Tx, source []int64, target int64) error {
	sourceIds := &pgtype.Int8Array{}
	sourceIds.Set(source)

	temp := []int64{}
	temp = append(temp, source...)
	temp = append(temp, target)
	allIds := &pgtype.Int8Array{}
	allIds.Set(temp)

	err := unifyImageLabelSuggestions(tx, sourceIds, allIds, target)
	if err != nil {
		return err
	}

	err = unifyImageAnnotationSuggestions(tx, sourceIds, allIds, target)
	if err != nil {
		return err
	}

	err = unifyTrendingLabelSuggestions(tx, sourceIds)
	if err != nil {
		return err
	}
	
	numOfDeleteLabelSuggestions += len(source)
	_, err = tx.Exec(context.TODO(), `DELETE FROM label_suggestion WHERE id = ANY($1)`, sourceIds)
	if err != nil {
		return err
	}

	return nil
}

func getDuplicateLabelSuggestions(tx pgx.Tx) ([]string, error) {
	duplicateLabelSuggestions := []string{}
	rows, err := tx.Query(context.TODO(), `SELECT name FROM label_suggestion GROUP BY name HAVING COUNT(name) > 1`)
	if err != nil {
		return duplicateLabelSuggestions, err
	}
	defer rows.Close()

	for rows.Next() {
		var duplicateLabelSuggestion string
		err := rows.Scan(&duplicateLabelSuggestion)
		if err != nil {
			return duplicateLabelSuggestions, err
		}

		duplicateLabelSuggestions = append(duplicateLabelSuggestions, duplicateLabelSuggestion)
	}

	return duplicateLabelSuggestions, nil
}

func getLabelSuggestionIdsForLabelSuggestion(tx pgx.Tx, labelSuggestion string) ([]int64, error) {
	duplicateLabelSuggestionIds := []int64{}
	rows, err := tx.Query(context.TODO(), `SELECT id FROM label_suggestion WHERE name = $1`, labelSuggestion)
	if err != nil {
		return duplicateLabelSuggestionIds, err
	}
	defer rows.Close()

	for rows.Next() {
		var duplicateLabelSuggestionId int64
		err := rows.Scan(&duplicateLabelSuggestionId)
		if err != nil {
			return duplicateLabelSuggestionIds, err
		}

		duplicateLabelSuggestionIds = append(duplicateLabelSuggestionIds, duplicateLabelSuggestionId)
	}

	return duplicateLabelSuggestionIds, nil
}

func isLabelSuggestionIdProductive(tx pgx.Tx, labelSuggestionId int64) (bool, error) {	
	rows, err := tx.Query(context.TODO(),
						`SELECT closed FROM trending_label_suggestion WHERE label_suggestion_id = $1`, labelSuggestionId)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		var productive bool
		err := rows.Scan(&productive)
		return productive, err
	}
	return false, nil
}

func disableIndexScans(tx pgx.Tx) error {
	for _, typeOfScan := range typeOfScans {
		q := fmt.Sprintf(`SET %s = OFF`, typeOfScan)
		_, err := tx.Exec(context.TODO(), q)
		if err != nil {
			return err
		}
	}
	return nil
}

func enableIndexScans(tx pgx.Tx) error {
	for _, typeOfScan := range typeOfScans {
		q := fmt.Sprintf(`SET %s = ON`, typeOfScan)
		_, err := tx.Exec(context.TODO(), q)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	dryRun := flag.Bool("dryrun", true, "Do a dry run")
	closeIssue := flag.Bool("close-github-issue", false, "Close github issue")
	githubRepository := flag.String("repository", "", "Github repository")
	
	flag.Parse()

	githubProjectOwner := ""
	githubApiToken := ""
	if *closeIssue {
		githubApiToken = commons.MustGetEnv("GITHUB_API_TOKEN")
		githubProjectOwner = commons.MustGetEnv("GITHUB_PROJECT_OWNER")

		if *githubRepository == "" {
			log.Fatal("Please provide a github repository!")
		}
	}

	imageMonkeyDbConnectionString := commons.MustGetEnv("IMAGEMONKEY_DB_CONNECTION_STRING")
	db, err := pgx.Connect(context.TODO(), imageMonkeyDbConnectionString)
	if err != nil {
		log.Fatal("Couldn't begin transaction: ", err.Error())
	}

	tx, err := db.Begin(context.TODO())
	if err != nil {
		log.Fatal("Couldn't begin transaction: ", err.Error())
	}

	err = disableIndexScans(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't disable index scans: ", err.Error())
	}

	err = disableTriggers(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't disable triggers: ", err.Error())
	}

	numOfImagesBefore, err := getNumOfImages(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't get num of images: ", err.Error())
	}

	numOfImageAnnotationSuggestionsBefore, err := getNumOfImageAnnotationSuggestions(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't get num of image annotation suggestions: ", err.Error())
	}

	numOfLabelSuggestionsBefore, err := getNumOfLabelSuggestions(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't get num of label suggestions: ", err.Error())
	}

	numOfAnnotationSuggestionDataBefore, err := getNumOfAnnotationSuggestionData(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't get num of annotation suggestion data: ", err.Error())
	}

	numOfProductiveLabelSuggestionsBefore, err := getNumOfProductiveLabelSuggestions(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't get num of productive label suggestions: ", err.Error())
	}

	numOfImageLabelSuggestionsBefore, err := getNumOfImageLabelSuggestions(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't get num of image label suggestions: ", err.Error())
	}

	numOfImageAnnotationSuggestionHistoryEntriesBefore, err := getNumOfImageAnnotationSuggestionHistoryEntries(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't get num of image annotation suggestion history entries: ", err.Error())
	}

	duplicateLabelSuggestions, err := getDuplicateLabelSuggestions(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't get duplicate label suggestions: ", err.Error())
	}

	for _, duplicateLabelSuggestion := range duplicateLabelSuggestions {
		labelSuggestionIds, err := getLabelSuggestionIdsForLabelSuggestion(tx, duplicateLabelSuggestion)
		if err != nil {
			tx.Rollback(context.TODO())
			log.Fatal("Couldn't get label suggestion ids: ", err.Error())
		}

		
		var productiveLabelSuggestionId int64 = -1
		for _, labelSuggestionId := range labelSuggestionIds {
			isProductive, err := isLabelSuggestionIdProductive(tx, labelSuggestionId)
			if err != nil {
				tx.Rollback(context.TODO())
				log.Fatal("Couldn't get label suggestion id info: ", err.Error())
			}

			if isProductive {
				if productiveLabelSuggestionId != -1 {
					tx.Rollback(context.TODO())
					log.Fatal("more than one productive label!")
				}
				productiveLabelSuggestionId = labelSuggestionId;
			}
		}

		if productiveLabelSuggestionId != -1 {
			err = unifyDuplicateLabelSuggestions(tx, removeElemFromSlice(labelSuggestionIds, productiveLabelSuggestionId), productiveLabelSuggestionId)
			if err != nil {
				tx.Rollback(context.TODO())
				log.Fatal("Couldn't unify: ", err.Error())
			}
		} else {
			err = unifyDuplicateLabelSuggestions(tx, labelSuggestionIds[1:], labelSuggestionIds[0])
			if err != nil {
				tx.Rollback(context.TODO())
				log.Fatal("Couldn't unify: ", err.Error())
			}
		}
	}

	//just a sanity check if everything is clean now
	duplicateLabelSuggestionsAfterUnification, err := getDuplicateLabelSuggestions(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("[Verify] Couldn't get duplicate label suggestions: ", err.Error())
	}
	if len(duplicateLabelSuggestionsAfterUnification) > 0 {
		tx.Rollback(context.TODO())
		log.Fatal("Verification failed. There are still duplicates!")
	}

	numOfImagesAfter, err := getNumOfImages(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't get num of images: ", err.Error())
	}

	if numOfImagesBefore != numOfImagesAfter {
		tx.Rollback(context.TODO())
		log.Fatal("Num of images doesn't match!")
	}

	numOfImageAnnotationSuggestionsAfter, err := getNumOfImageAnnotationSuggestions(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't get num of image annotation suggestions: ", err.Error())
	}

	if numOfImageAnnotationSuggestionsBefore != (numOfImageAnnotationSuggestionsAfter + unrecoverableDeletedAnnotations) {
		tx.Rollback(context.TODO())
		log.Fatal("Num of image annotation suggestions do not match!")
	}

	numOfAnnotationSuggestionDataAfter, err := getNumOfAnnotationSuggestionData(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't get num of annotation suggestion data: ", err.Error())
	}

	if numOfAnnotationSuggestionDataBefore != (numOfAnnotationSuggestionDataAfter +  unrecoverableDeletedAnnotationData) {
		tx.Rollback(context.TODO())
		log.Fatal("fail: ")
	}

	numOfLabelSuggestionsAfter, err := getNumOfLabelSuggestions(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't get num of label suggestions: ", err.Error())
	}

	if numOfLabelSuggestionsBefore != (numOfLabelSuggestionsAfter + numOfDeleteLabelSuggestions)  {
		tx.Rollback(context.TODO())
		log.Fatal("Num of label suggestions do not match!")
	}

	numOfProductiveLabelSuggestionsAfter, err := getNumOfProductiveLabelSuggestions(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't get num of productive label suggestions: ", err.Error())
	}

	if numOfProductiveLabelSuggestionsBefore != numOfProductiveLabelSuggestionsAfter {
		tx.Rollback(context.TODO())
		log.Fatal("Num of production label suggestions do not match!")
	}

	numOfImageLabelSuggestionsAfter, err := getNumOfImageLabelSuggestions(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't get num of image label suggestions")
	}

	if numOfImageLabelSuggestionsBefore != (numOfImageLabelSuggestionsAfter + numOfDeletedImageLabelSuggestions) {
		log.Fatal("Num of image label suggestions do not match!")
	}

	numOfImageAnnotationSuggestionHistoryEntriesAfter, err := getNumOfImageAnnotationSuggestionHistoryEntries(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't get num of image annotation suggestion history entries: ", err.Error())
	}

	if numOfImageAnnotationSuggestionHistoryEntriesBefore != 
		(numOfImageAnnotationSuggestionHistoryEntriesAfter + numOfDeletedImageAnnotationSuggestionHistoryEntries) {
		tx.Rollback(context.TODO())
		log.Fatal("Num of image annotation suggestion history entries do not match!")
	}

	log.Info("Verification successful")
	log.Info("")
	log.Info("Statistics:")
	log.Info("Unrecoverable deleted annotations: ", unrecoverableDeletedAnnotations)
	log.Info("Unrecoverable deleted annotation data: ", unrecoverableDeletedAnnotationData)
	log.Info("Num of deleted label suggestions: ", numOfDeleteLabelSuggestions)
	log.Info("Num of deleted image label suggestions: ", numOfDeletedImageLabelSuggestions)
	log.Info("Num of deleted image annotation suggestion history entries: ", numOfDeletedImageAnnotationSuggestionHistoryEntries)
	log.Info("-----------------------------------")
	log.Info("")

	log.Info("The following ", len(githubIssueIds) , " github issues can be closed:")
	for _, githubIssueId := range githubIssueIds {
		log.Info(githubIssueId)
	}

	err = enableTriggers(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't disable triggers: ", err.Error())
	}

	err = enableIndexScans(tx)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't enable index scans: ", err.Error())
	}

	err = persistDuplicateLabels(duplicateLabelSuggestions)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Fatal("Couldn't persist duplicate labels: ", err.Error())
	}

	if *dryRun {
		log.Info("Just a dry run..rolling back transaction")
		err := tx.Rollback(context.TODO())
		if err != nil {
			log.Fatal("Couldn't rollback transaction: ", err.Error())
		}
	} else {
		err := tx.Commit(context.TODO())
		if err != nil {
			log.Fatal("Couldn't commit transaction: ", err.Error())
		}


		unclosedGithubIssues := []int64{}
		if *closeIssue {
			for _, githubIssueId := range githubIssueIds {
				err := closeGithubIssue(int(githubIssueId), *githubRepository, githubProjectOwner, githubApiToken)
				if err != nil {
					unclosedGithubIssues = append(unclosedGithubIssues, githubIssueId)
				}
			}
		}

		log.Error("The following github issues couldn't be closed: ")
		for _, unclosedGithubIssue := range unclosedGithubIssues {
			log.Error(unclosedGithubIssue)
		}

		log.Info("done")
	}

}
