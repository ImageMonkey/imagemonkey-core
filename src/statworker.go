package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"flag"
	"github.com/gomodule/redigo/redis"
	"time"
	"encoding/json"
	"github.com/getsentry/raven-go"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	imagemonkeydb "github.com/bbernhard/imagemonkey-core/database"
	commons "github.com/bbernhard/imagemonkey-core/commons"
)

func main(){
	fmt.Printf("Starting Statistics Worker...\n")

	log.SetLevel(log.DebugLevel)

	redisAddress := flag.String("redis_address", ":6379", "Address to the Redis server")
	redisMaxConnections := flag.Int("redis_max_connections", 10, "Max connections to Redis")
	singleshot := flag.Bool("singleshot", false, "Terminate after work is done")
	useSentry := flag.Bool("use_sentry", false, "Use Sentry for error logging")
	maxNumOfDatabaseConnections := flag.Int("db_max_connections", 5, "Max. number of database connections")

	flag.Parse()

	var err error
	
	imageDbConnectionString := commons.MustGetEnv("IMAGEMONKEY_DB_CONNECTION_STRING")
	imageMonkeyDatabase := imagemonkeydb.NewImageMonkeyDatabase()
	err = imageMonkeyDatabase.Open(imageDbConnectionString, int32(*maxNumOfDatabaseConnections))
	if err != nil {
		log.Fatal("[Main] Couldn't ping ImageMonkey database: ", err.Error())
	}
	defer imageMonkeyDatabase.Close()

	if *useSentry {
		sentryDsn := commons.MustGetEnv("SENTRY_DSN")
		log.Debug("Setting Sentry DSN")
		raven.SetDSN(sentryDsn)
		raven.SetEnvironment("statworker")
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
	    	var contributionsPerCountryRequest datastructures.ContributionsPerCountryRequest
	    	err = json.Unmarshal(data, &contributionsPerCountryRequest)
	    	if err != nil{
	    		retryImmediately = false
	    		log.Debug("[Main] Couldn't unmarshal contributions_per_country request: ", err.Error())
	    		
	    	} else {
		    	err = imageMonkeyDatabase.UpdateContributionsPerCountry(contributionsPerCountryRequest.Type, 
		    									 						contributionsPerCountryRequest.CountryCode)
		    	if err != nil {
		    		retryImmediately = false
		    	}
			}
		}

		data, err = redis.Bytes(redisConn.Do("LPOP", "contributions-per-app"))
		if err == nil {
			retryImmediately = true
	    	var contributionsPerAppRequest datastructures.ContributionsPerAppRequest
	    	err = json.Unmarshal(data, &contributionsPerAppRequest)
	    	if err != nil{
	    		retryImmediately = false
	    		log.Debug("[Main] Couldn't unmarshal contributions_per_app request: ", err.Error())
	    		
	    	} else {
		    	err = imageMonkeyDatabase.UpdateContributionsPerApp(contributionsPerAppRequest.Type, 
		    									 					contributionsPerAppRequest.AppIdentifier)
		    	if err != nil {
		    		retryImmediately = false
		    	}
			}
		}

		redisConn.Close()

		if !retryImmediately {
			if *singleshot {
				return
			}
			time.Sleep((time.Second * 2)) //sleep for two seconds
		}

	}

}
