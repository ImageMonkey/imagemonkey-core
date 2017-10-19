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

		redisConn := redisPool.Get()

		data, err := redis.Bytes(redisConn.Do("LPOP", "contributions-per-country"))
    	if err != nil {
    		redisConn.Close()
    		time.Sleep((time.Second * 2)) //nothing in queue, sleep for two seconds
    		continue
    	}

    	var contributionsPerCountryRequest ContributionsPerCountryRequest
    	err = json.Unmarshal(data, &contributionsPerCountryRequest)
    	if err != nil{
    		log.Debug("[Main] Couldn't unmarshal: ", err.Error())
    		redisConn.Close()
			continue
    	}

    	err = updateContributionsPerCountry(contributionsPerCountryRequest.Type, 
    									 	contributionsPerCountryRequest.CountryCode)
    	if err != nil {
    		redisConn.Close()
    		continue
    	}

		redisConn.Close()
	}

}