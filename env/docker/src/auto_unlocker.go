package main

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/Sirupsen/logrus"
	"path/filepath"
	"net/http"
	"bytes"
	"flag"
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
    	log.Fatal("Couldn't auto-unlock image: ")
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
	            //log.Println("event:", event)
	            if event.Op == fsnotify.Create {
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