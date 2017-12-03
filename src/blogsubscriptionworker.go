package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"flag"
	"github.com/garyburd/redigo/redis"
	"time"
	"database/sql"
	_ "github.com/lib/pq"
	"encoding/json"
)

var db *sql.DB

func subscribe(email string) error {
	log.Debug("[Main] Got a new subscription: ", email)
	_,err := db.Exec(`INSERT INTO blog.subscription(email) VALUES ($1)
			 		  ON CONFLICT DO NOTHING`, email)
	if err != nil {
		log.Debug("[Main] Couldn't add subscription", err.Error())
		return err
	}

	return nil
}

func main(){
	fmt.Printf("Starting Blog Subscription Worker...\n")

	log.SetLevel(log.DebugLevel)

	redisAddress := flag.String("redis_address", ":6379", "Address to the Redis server")
	redisMaxConnections := flag.Int("redis_max_connections", 5, "Max connections to Redis")

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

	for {
		var data []byte
		retryImmediately := false

		redisConn := redisPool.Get()
		data, err := redis.Bytes(redisConn.Do("LPOP", "subscribe-to-blog"))
    	if err == nil {
    		retryImmediately = true

    		var blogSubscribeRequest BlogSubscribeRequest
	    	err = json.Unmarshal(data, &blogSubscribeRequest)
	    	if err == nil {
	    		subscribe(blogSubscribeRequest.Email)
	    	} else {
	    		log.Debug("[Main] Couldn't unmarshal request: ", err.Error())
	    	}
    	}

    	redisConn.Close()

		if !retryImmediately {
			time.Sleep((time.Second * 60)) //sleep for 60 seconds
		}
    }

}