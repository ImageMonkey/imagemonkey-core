package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
	"context"
	"os/exec"
	"errors"
	"time"
	"github.com/jackc/pgx/v4"
)


func runSnapshotScript(snapshotScript string, annotationId string, outputPath string) error {
	// Start a process
	cmd := exec.Command("node", snapshotScript, annotationId, outputPath)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return err
	}

	// Wait for the process to finish or kill it after a timeout:
	done := make(chan error, 1)
	go func() {
	    done <- cmd.Wait()
	}()
	select {
	case <-time.After(60 * time.Second):
	    err := cmd.Process.Kill()
	    return errors.New("process killed as timeout reached: " + err.Error())
	case err := <-done:
	    return err
	}

	return nil
}


func getAnnotationIds(conn *pgx.Conn) ([]string, error) {
	
	annotationIds := []string{}
	
	rows, err := conn.Query(context.TODO(), 
								`SELECT a.uuid 
								 FROM image_annotation a
								 JOIN image i ON i.id = a.image_id
								 WHERE i.unlocked = true`)
	if err != nil {
		return annotationIds, err
	}

	for rows.Next() {
		var annotationId string
		err := rows.Scan(&annotationId)
		if err != nil {
			return annotationIds, err
		}

		annotationIds = append(annotationIds, annotationId)
	}

	return annotationIds, nil
}

func main() {
	outputFolder := flag.String("output-folder", "", "Output Folder")
	snapshotScriptPath := flag.String("snapshot-script", "", "Path to snapshot script")

	flag.Parse()

	if *outputFolder == "" {
		log.Fatal("Please provide a output folder with --output-folder!")
	}

	if _, err := os.Stat(*outputFolder); os.IsNotExist(err) {
		log.Fatal("Folder ", *outputFolder, " does not exist!")
	}

	if *snapshotScriptPath == "" {
		log.Fatal("Please provide a snapshot script with --snapshot-script!")
	}

	if _, err := os.Stat(*snapshotScriptPath); os.IsNotExist(err) {
		log.Fatal("Snapshot script ", *snapshotScriptPath, " doesn't exist!")
	}

	conn, err := pgx.Connect(context.TODO(), os.Getenv("IMAGEMONKEY_DB_CONNECTION_STRING"))
	if err != nil {
		log.Fatal("Unable to connect to database: ", err.Error())
	}
	defer conn.Close(context.TODO())


	annotationIds, err := getAnnotationIds(conn)
	if err != nil {
		log.Fatal("Couldn't get annotation ids: ", err.Error())
	}

	for _, annotationId := range annotationIds {
		path := *outputFolder + "/" + annotationId + ".png"
		err := runSnapshotScript(*snapshotScriptPath, annotationId, path)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}
