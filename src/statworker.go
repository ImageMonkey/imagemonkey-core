package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"flag"
	"database/sql"
	"github.com/garyburd/redigo/redis"
	"time"
	"encoding/json"
)

var db *sql.DB

func main(){
	fmt.Printf("Starting Statistics Worker...\n")

	log.SetLevel(log.DebugLevel)

	redisAddress := flag.String("redis_address", ":6379", "Address to the Redis server")
	redisMaxConnections := flag.Int("redis_max_connections", 10, "Max connections to Redis")

	flag.Parse()

	var err error

	//open database and make sure that we can ping it
	db, err = sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
	if err != nil {
		log.Fatal("[Main] Couldn't open database: ", err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("[Main] Couldn't ping database: ", err.Error())
	}

	//create redis pool
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", *redisAddress)

		if err != nil {
			log.Fatal("[Main] Couldn't dial redis: ", err.Error())
		}

		return c, err
	}, *redisMaxConnections)
	defer redisPool.Close()

	for {
		var data []byte
		retryImmediately := false

		redisConn := redisPool.Get()
		data, err := redis.Bytes(redisConn.Do("LPOP", "contributions-per-country"))
    	if err == nil {
    		retryImmediately = true
	    	var contributionsPerCountryRequest ContributionsPerCountryRequest
	    	err = json.Unmarshal(data, &contributionsPerCountryRequest)
	    	if err != nil{
	    		retryImmediately = false
	    		log.Debug("[Main] Couldn't unmarshal contributions_per_country request: ", err.Error())
	    		
	    	} else {
		    	err = updateContributionsPerCountry(contributionsPerCountryRequest.Type, 
		    									 	contributionsPerCountryRequest.CountryCode)
		    	if err != nil {
		    		retryImmediately = false
		    	}
			}
		}

		data, err = redis.Bytes(redisConn.Do("LPOP", "contributions-per-app"))
		if err == nil {
			retryImmediately = true
	    	var contributionsPerAppRequest ContributionsPerAppRequest
	    	err = json.Unmarshal(data, &contributionsPerAppRequest)
	    	if err != nil{
	    		retryImmediately = false
	    		log.Debug("[Main] Couldn't unmarshal contributions_per_app request: ", err.Error())
	    		
	    	} else {
		    	err = updateContributionsPerApp(contributionsPerAppRequest.Type, 
		    									 	contributionsPerAppRequest.AppIdentifier)
		    	if err != nil {
		    		retryImmediately = false
		    	}
			}
		}

		redisConn.Close()

		if !retryImmediately {
			time.Sleep((time.Second * 2)) //sleep for two seconds
		}

	}

}