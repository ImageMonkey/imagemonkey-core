package main

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"net/http"
	"bytes"
	"flag"
	"time"
)


func unlockImage(imageId string) {
	url := "http://127.0.0.1:8081/v1/unverified/donation/" + imageId + "/good"

    req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(`{}`)))
    req.Header.Set("X-Client-Secret", X_CLIENT_SECRET)
    req.Header.Set("X-Client-Id", X_CLIENT_ID)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Fatal("Couldn't auto-unlock image: ", err.Error())
    }
    defer resp.Body.Close()

    if resp.StatusCode != 201 {
    	log.Fatal("Couldn't auto-unlock image. Status code: ", resp.StatusCode)
    }
}

func main() {

	path := flag.String("path", "../unverified_donations/", "path to monitor")

	flag.Parse()

	log.Info("Starting auto unlocker..")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
	    log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
	    for {
	        select {
	        case event := <-watcher.Events:
	            if event.Op == fsnotify.Create {
	            	//this is a pretty ugly fix for a race condition. 
	            	//when a new image gets uploaded, the file is first written
	            	//to the filesystem, before the image entry gets added to the database. 
	            	//so we sleep a bit here in order to make sure that the image entry
	            	//is for sure in the database. this is pretty ugly and error prone, but
	            	//for now it works
	            	time.Sleep(500 * time.Millisecond)

	            	fname := filepath.Base(event.Name)
	            	log.Info("detected new file: ", fname)
	            	unlockImage(fname)
	            }
	        case err := <-watcher.Errors:
	            log.Fatal(err)
	        }
	    }
	}()

	err = watcher.Add(*path)
	if err != nil {
	    log.Fatal(err)
	}
	<-done
}
