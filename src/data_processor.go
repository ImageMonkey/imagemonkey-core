package main

import (
	"time"
	log "github.com/sirupsen/logrus"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/getsentry/raven-go"
	"flag"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"os"
	"encoding/json"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	commons "github.com/bbernhard/imagemonkey-core/commons"
)

var db *sql.DB

func removeOldImageAnnotationCoverage(updateAnnotationCoverageRequest datastructures.UpdateAnnotationCoverageRequest, tx *sql.Tx) error {
	var queryValues []interface{}
	q1 := ""
	if updateAnnotationCoverageRequest.Uuid != "" {
		if updateAnnotationCoverageRequest.Type == "image" {
			q1 = "USING image i WHERE i.id = c.image_id AND i.key = $1"
		} else if updateAnnotationCoverageRequest.Type == "annotation" {
			q1 = "USING image_annotation a WHERE c.image_id = a.image_id AND a.uuid = $1"
		}
		queryValues = append(queryValues, updateAnnotationCoverageRequest.Uuid)
	}

	q := fmt.Sprintf(`DELETE FROM image_annotation_coverage c %s`, q1)
    _, err := tx.Exec(q, queryValues...)
    if err != nil {
    	tx.Rollback()
    	log.Error("[Removing old Image Annotation coverage] Couldn't remove old image annotation coverage: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    return nil
}

func calculateImageAnnotationCoverage(updateAnnotationCoverageRequest datastructures.UpdateAnnotationCoverageRequest) error {
	var queryValues []interface{}
	q1 := ""
	if updateAnnotationCoverageRequest.Uuid != "" {
		if updateAnnotationCoverageRequest.Type == "image" {
			q1 = "$1"
		} else if updateAnnotationCoverageRequest.Type == "annotation" {
			q1 = `(SELECT i.key 
				  FROM image i
				  JOIN image_annotation a ON a.image_id = i.id
				  WHERE a.uuid = $1)`
		}
		queryValues = append(queryValues, updateAnnotationCoverageRequest.Uuid)
	}

	q := fmt.Sprintf(`INSERT INTO image_annotation_coverage(image_id, area, annotated_percentage)
						SELECT image_id, area, annotated_percentage 
			   				FROM sp_get_image_annotation_coverage(%s)`, q1)

	tx, err := db.Begin()
    if err != nil {
    	log.Error("[Calculating Image Annotation coverage] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    err = removeOldImageAnnotationCoverage(updateAnnotationCoverageRequest, tx)
	if err != nil { //transaction already rolled back, so nothing to do here
		return err
	}

	
	_, err = tx.Exec(q, queryValues...)
	if err != nil {
		tx.Rollback()
		log.Error("[Calculating Image Annotation coverage] Couldn't calculate image annotation coverage", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	err = tx.Commit()
    if err != nil {
    	log.Error("[Calculating Image Annotation coverage] Couldn't commit transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    return nil
}

func main() {
	maintenanceModeFile := flag.String("maintenance_mode_file", "../maintenance.tmp", "maintenance mode file")
	singleshot := flag.Bool("singleshot", false, "Terminate after work is done")
	useSentry := flag.Bool("use_sentry", false, "Use Sentry for error logging")
	redisAddress := flag.String("redis_address", ":6379", "Address to the Redis server")
	redisMaxConnections := flag.Int("redis_max_connections", 5, "Max connections to Redis")

	flag.Parse()

	if *useSentry {
		sentryDsn := commons.MustGetEnv("SENTRY_DSN")
		raven.SetDSN(sentryDsn)
		raven.SetEnvironment("data-processor")
	}

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

	//create redis pool
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", *redisAddress)

		if err != nil {
			log.Fatal("[Main] Couldn't dial redis: ", err.Error())
		}

		return c, err
	}, *redisMaxConnections)
	defer redisPool.Close()

	//if file exists, start in maintenance mode
	maintenanceMode := false
	if _, err := os.Stat(*maintenanceModeFile); err == nil {
		maintenanceMode = true
		log.Info("[Main] Starting data processor (maintenance mode)...")
	} else {
		log.Info("[Main] Starting data processor...")
	}

	if !maintenanceMode {
		//on startup, do a full re-calculate
		for {
			err := calculateImageAnnotationCoverage(datastructures.UpdateAnnotationCoverageRequest{})
			if err == nil { 
				log.Info("[Main] Completely re-calculated image annotation coverage")
				break
			}
		}

		if *singleshot {
			return
		}

		//then do only a incremental re-calculate
		for {
			retryImmediately := false

			redisConn := redisPool.Get()
			data, err := redis.Bytes(redisConn.Do("LPOP", commons.UPDATE_IMAGE_ANNOTATION_COVERAGE_TOPIC))
	    	if err == nil { //data available
	    		retryImmediately = true //in case there is data available, try it again immediatelly to get more data

	    		var updateAnnotationCoverageRequest datastructures.UpdateAnnotationCoverageRequest
		    	err = json.Unmarshal(data, &updateAnnotationCoverageRequest)
		    	if err != nil{
		    		retryImmediately = false //in case of an error, wait a bit (maybe it recovers in the meanwhile)
		    		log.Error("[Main] Couldn't unmarshal request: ", err.Error())
		    		raven.CaptureError(err, nil)
		    	} else {
		    		err := calculateImageAnnotationCoverage(updateAnnotationCoverageRequest)
		    		if err == nil {
		    			retryImmediately = false //in case of an error, wait a bit (maybe it recovers in the meanwhile)
		    			log.Info("[Main] Re-calculated image annotation coverage for ", 
		    						updateAnnotationCoverageRequest.Type, " with id: ", updateAnnotationCoverageRequest.Uuid)
		    		}
		    	}
	    	}

	    	redisConn.Close()

			if !retryImmediately {
				time.Sleep((time.Second * 2)) //sleep for two seconds
			}


		}
		
	} else {
		select{} //sleep forever, without eating CPU
	}

}
