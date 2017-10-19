package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"strings"
	"fmt"
	"github.com/satori/go.uuid"
	"gopkg.in/h2non/filetype.v1"
	log "github.com/Sirupsen/logrus"
	"flag"
	"database/sql"
	"os"
	"strconv"
	"github.com/oschwald/geoip2-golang"
	"github.com/garyburd/redigo/redis"
	"encoding/json"
	"net"
)

var db *sql.DB
var geoipDb *geoip2.Reader

//Middleware to ensure that the correct X-Client-Id and X-Client-Secret are provided in the header
func ClientAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var clientId string
		var clientSecret string

		clientId = ""
		if values, _ := c.Request.Header["X-Client-Id"]; len(values) > 0 {
			clientId = values[0]
		}

		clientSecret = ""
		if values, _ := c.Request.Header["X-Client-Secret"]; len(values) > 0 {
			clientSecret = values[0]
		}

		if(!((clientSecret == X_CLIENT_SECRET) && (clientId == X_CLIENT_ID))) {
			c.String(401, "Please provide a valid client id and client secret")
			c.AbortWithStatus(401)
			return
		}

		c.Next()
	}
}

//CORS Middleware 
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	    c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-Id, Cache-Control, X-Requested-With")
	    c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH")

		if c.Request.Method == "OPTIONS" {
             c.AbortWithStatus(200)
         } else {
             c.Next()
         }
	}
}

func RequestId() gin.HandlerFunc {
	return func(c *gin.Context) {
		if values, _ := c.Request.Header["X-Request-Id"]; len(values) > 0 {
			if(values[0] != ""){
				c.Writer.Header().Set("X-Request-Id", values[0])
			}
		}

		c.Next()
	}
}


func pushCountryContributionToRedis(redisPool *redis.Pool, contributionsPerCountryRequest ContributionsPerCountryRequest) {
	serialized, err := json.Marshal(contributionsPerCountryRequest)
	if err != nil { 
		log.Debug("[Donate] Couldn't create contributions-per-country request: ", err.Error())
		return
	}

	redisConn := redisPool.Get()
	defer redisConn.Close()

	_, err = redisConn.Do("RPUSH", "contributions-per-country", serialized)
	if err != nil { //just log error, but not abort (it's just some statistical information)
		log.Debug("[Donate] Couldn't update contributions-per-country: ", err.Error())
		return
	}

}


func main(){
	fmt.Printf("Starting API Service...\n")

	log.SetLevel(log.DebugLevel)

	releaseMode := flag.Bool("release", false, "Run in release mode")
	wordlistDir := flag.String("wordlist", "../wordlists/en/misc.txt", "Path to wordlist")
	donationsDir := flag.String("donations_dir", "../donations/", "Location of the uploaded donations")
	unverifiedDonationsDir := flag.String("unverified_donations_dir", "../unverified_donations/", "Location of the uploaded but unverified donations")
	redisAddress := flag.String("redis_address", ":6379", "Address to the Redis server")
	redisMaxConnections := flag.Int("redis_max_connections", 500, "Max connections to Redis")
	geoIpDbPath := flag.String("geoip_db", "../geoip_database/GeoLite2-Country.mmdb", "Path to the GeoIP database")

	flag.Parse()
	if(*releaseMode){
		fmt.Printf("[Main] Starting gin in release mode!\n")
		gin.SetMode(gin.ReleaseMode)
	}

	fmt.Printf("Reading wordlists...\n")
	words, err := getWordLists(*wordlistDir)
	if(err != nil){
		fmt.Printf("[Main] Couldn't read wordlists...terminating!")
		log.Fatal(err)
	}

	mostPopularLabels := []string{"cat", "dog"}

	//open database and make sure that we can ping it
	db, err = sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
	if err != nil {
		log.Fatal("[Main] Couldn't open database: ", err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("[Main] Couldn't ping database: ", err.Error())
	}


	//open geoip database
	geoipDb, err := geoip2.Open(*geoIpDbPath)
	if err != nil {
		log.Fatal("[Main] Couldn't read geoip database: ", err.Error())
	}
	defer geoipDb.Close()

	//create redis pool
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", *redisAddress)

		if err != nil {
			log.Fatal("[Main] Couldn't dial redis: ", err.Error())
		}

		return c, err
	}, *redisMaxConnections)
	defer redisPool.Close()


	router := gin.Default()
	router.Use(CorsMiddleware())
	router.Use(RequestId())
	router.Static("./v1/donation", *donationsDir) //serve static images

	//the following endpoints are secured with a client id + client secret. 
	//that's mostly because currently each donation needs to be unlocked manually. 
	//(as we want to make sure that we don't accidentally host inappropriate content, like nudity)
	clientAuth := router.Group("/")
	clientAuth.Use(RequestId())
	clientAuth.Use(ClientAuthMiddleware())
	{
		clientAuth.Static("./v1/unverified/donation", *unverifiedDonationsDir)
		clientAuth.GET("/v1/unverified/donation", func(c *gin.Context) {
			images, err := getAllUnverifiedImages()
			use(images)
			if err != nil {
				c.JSON(500, gin.H{"error" : "Couldn't process request - please try again later"}) 
			} else {
				c.JSON(http.StatusOK, images)
			}
		})

		clientAuth.POST("/v1/unverified/donation/:imageid/:param",  func(c *gin.Context) {
			imageId := c.Param("imageid")

			//verify that uuid is a valid UUID (to prevent path injection)
			_, err := uuid.FromString(imageId)
			if err != nil {
				c.JSON(400, gin.H{"error" : "Couldn't process request - not a valid image id"})
				return
			}

			param := c.Param("param")
			isBad := true
			if param == "bad" {
				isBad = true
			} else if param == "good" {
				isBad = false
			} else{
				c.JSON(404, gin.H{"error": "Couldn't process request - invalid parameter"})
				return
			}

			if !isBad {
				src := *unverifiedDonationsDir + imageId
				dst := *donationsDir + imageId
				err := os.Rename(src, dst)
				if err != nil {
					log.Debug("[Main] Couldn't move file ", src, " to ", dst)
					c.JSON(500, gin.H{"error" : "Couldn't process request - please try again later"})
					return
				}

				err = unlockImage(imageId)
				if err != nil {
					c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
					return
				}
			}

			c.JSON(http.StatusOK, nil)
		})
	}

	router.GET("/v1/validate", func(c *gin.Context) {
		params := c.Request.URL.Query()

		grouped := false
		if temp, ok := params["grouped"]; ok {
			if temp[0] == "true" {
				grouped = true
			}
		}

		limit := 20
		if temp, ok := params["limit"]; ok {
			limit, err = strconv.Atoi(temp[0])
			if err != nil {
				c.JSON(400, gin.H{"error": "Couldn't process request - invalid limit parameter"})
				return
			}
		}

		if grouped {
			pos := 0
			if len(mostPopularLabels) > 0 {
				pos = random(0, (len(mostPopularLabels) - 1))
			}
				

			randomGroupedImages, err := getRandomGroupedImages(mostPopularLabels[pos], limit)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"label": mostPopularLabels[pos], "donations": randomGroupedImages})

		} else {
			randomImage := getRandomImage()
			c.JSON(http.StatusOK, gin.H{"uuid": randomImage.Id, "label": randomImage.Label, "provider": randomImage.Provider})
		}
	})

	router.POST("/v1/donation/:imageid/validate/:param", func(c *gin.Context) {
		imageId := c.Param("imageid")
		param := c.Param("param")

		parameter := false
		if(param == "yes"){
			parameter = true
		} else if(param == "no"){
			parameter = false
		} else{
			c.JSON(404, nil)
			return
		}

		err := validateDonatedPhoto(imageId, parameter)
		if(err != nil){
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Database Error: Couldn't update data"})
			return
		} 

		//get client IP address and try to determine country
		var contributionsPerCountryRequest ContributionsPerCountryRequest
		contributionsPerCountryRequest.Type = "validation"
		contributionsPerCountryRequest.CountryCode = "--"
		ip := net.ParseIP(getIPAddress(c.Request))
		if ip != nil {
			record, err := geoipDb.Country(ip)
			if err != nil { //just log, but don't abort...it's just for statistics
				log.Debug("[Validation] Couldn't determine geolocation from ", err.Error())
				
			} else {
				contributionsPerCountryRequest.CountryCode = record.Country.IsoCode
			}
		}
		pushCountryContributionToRedis(redisPool, contributionsPerCountryRequest)

		c.JSON(http.StatusOK, nil)
	})

	router.PATCH("/v1/donation/validate", func(c *gin.Context) {
		var imageValidationBatch []ImageValidation
		
		if c.BindJSON(&imageValidationBatch) != nil {
			c.JSON(400, gin.H{"error": "Couldn't process request - invalid patch"})
			return
		}

		err := validateImages(imageValidationBatch)
		if err != nil {
			c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
			return
		}

		//get client IP address and try to determine country
		var contributionsPerCountryRequest ContributionsPerCountryRequest
		contributionsPerCountryRequest.Type = "validation"
		contributionsPerCountryRequest.CountryCode = "--"
		ip := net.ParseIP(getIPAddress(c.Request))
		if ip != nil {
			record, err := geoipDb.Country(ip)
			if err != nil { //just log, but don't abort...it's just for statistics
				log.Debug("[Validation] Couldn't determine geolocation from ", err.Error())
				
			} else {
				contributionsPerCountryRequest.CountryCode = record.Country.IsoCode
			}
		}
		pushCountryContributionToRedis(redisPool, contributionsPerCountryRequest)

		c.JSON(200, nil)
	})

	router.GET("/v1/export", func(c *gin.Context) {
		tags := ""
		params := c.Request.URL.Query()
		if temp, ok := params["tags"]; ok {
			tags = temp[0]
			jsonData, err := export(strings.Split(tags, ","))
			if(err == nil){
				c.JSON(http.StatusOK, jsonData)
				return
			} else{
				c.JSON(http.StatusInternalServerError, gin.H{"Error": "Couldn't export data, please try again later."})
				return
			}
		} else {
			c.JSON(422, gin.H{"error": "no tags specified"})
			return
		}
	})

	router.GET("/v1/label", func(c *gin.Context) {
		c.JSON(http.StatusOK, words)
	})


	router.GET("/v1/label/random", func(c *gin.Context) {
		label := words[random(0, len(words) - 1)].Name
		c.JSON(http.StatusOK, gin.H{"label": label})
	})

	router.POST("/v1/donate", func(c *gin.Context) {
		label := c.PostForm("label")

		file, header, err := c.Request.FormFile("image")
		if(err != nil){
			fmt.Printf("err = %s", err.Error())
			c.JSON(400, gin.H{"error": "Picture is missing"})
			return
		}

		// Create a buffer to store the header of the file
        fileHeader := make([]byte, 512)

        // Copy the file header into the buffer
        if _, err := file.Read(fileHeader); err != nil {
        	c.JSON(422, gin.H{"error": "Unable to detect MIME type"})
        	return
        }

        // set position back to start.
        if _, err := file.Seek(0, 0); err != nil {
        	c.JSON(422, gin.H{"error": "Unable to detect MIME type"})
        	return
        }

        if(!filetype.IsImage(fileHeader)){
        	c.JSON(422, gin.H{"error": "Unsopported MIME type detected"})
        	return
        }

        //check if image already exists by using an image hash
        hash, err := hashImage(file)
        if(err != nil){
        	c.JSON(500, gin.H{"error": "Couldn't add photo - please try again later"})
        	return 
        }
        exists, err := imageExists(hash)
        if(err != nil){
        	c.JSON(500, gin.H{"error": "Couldn't add photo - please try again later"})
        	return
        }
        if(exists){
        	c.JSON(409, gin.H{"error": "Couldn't add photo - image already exists"})
        	return
        }

        //image doesn't already exist, so save it and add it to the database
		uuid := uuid.NewV4().String()
		c.SaveUploadedFile(header, (*unverifiedDonationsDir + uuid))

		err = addDonatedPhoto(uuid, hash, label)
		if(err != nil){
			c.JSON(500, gin.H{"error": "Couldn't add photo - please try again later"})	
			return
		}

		//get client IP address and try to determine country
		var contributionsPerCountryRequest ContributionsPerCountryRequest
		contributionsPerCountryRequest.Type = "donation"
		contributionsPerCountryRequest.CountryCode = "--"
		ip := net.ParseIP(getIPAddress(c.Request))
		if ip != nil {
			record, err := geoipDb.Country(ip)
			if err != nil { //just log, but don't abort...it's just for statistics
				log.Debug("[Donation] Couldn't determine geolocation from ", err.Error())
				
			} else {
				contributionsPerCountryRequest.CountryCode = record.Country.IsoCode
			}
		}
		pushCountryContributionToRedis(redisPool, contributionsPerCountryRequest)

		c.JSON(http.StatusOK, nil)
	})

	router.POST("/v1/report/:imageid", func(c *gin.Context) {
		imageId := c.Param("imageid")

		var report Report
		if(c.BindJSON(&report) != nil){
			c.JSON(422, gin.H{"error": "reason missing - please provide a valid 'reason'"})
			return
		}
		err := reportImage(imageId, report.Reason)
		if(err != nil){
			c.JSON(500, gin.H{"error": "Couldn't report image - please try again later"})
			return
		}
		c.JSON(http.StatusOK, nil)
	})

	router.POST("/v1/annotation/:imageid/validate/:param", func(c *gin.Context) {
		imageId := c.Param("imageid")
		param := c.Param("param")

		parameter := false
		if(param == "yes"){
			parameter = true
		} else if(param == "no"){
			parameter = false
		} else{
			c.JSON(404, nil)
			return
		}

		err := validateAnnotatedImage(imageId, parameter)
		if(err != nil){
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Database Error: Couldn't update data"})
			return
		} 

		//get client IP address and try to determine country
		var contributionsPerCountryRequest ContributionsPerCountryRequest
		contributionsPerCountryRequest.Type = "annotation"
		contributionsPerCountryRequest.CountryCode = "--"
		ip := net.ParseIP(getIPAddress(c.Request))
		if ip != nil {
			record, err := geoipDb.Country(ip)
			if err != nil { //just log, but don't abort...it's just for statistics
				log.Debug("[Annotation] Couldn't determine geolocation from ", err.Error())
				
			} else {
				contributionsPerCountryRequest.CountryCode = record.Country.IsoCode
			}
		}
		pushCountryContributionToRedis(redisPool, contributionsPerCountryRequest)

		c.JSON(http.StatusOK, nil)
	})

	router.POST("/v1/annotate/:imageid", func(c *gin.Context) {
		imageId := c.Param("imageid")
		if(imageId == ""){
			c.JSON(422, gin.H{"error": "invalid request - image id missing"})
			return
		}

		var annotations []Annotation
		err := c.BindJSON(&annotations)
		if(err != nil){
			c.JSON(422, gin.H{"error": "invalid request - annotations missing"})
			return
		}

		err = addAnnotations(imageId, annotations)
		if(err != nil){
			c.JSON(500, gin.H{"error": "Couldn't add annotations - please try again later"})
			return
		}
	})

	router.GET("/v1/annotate", func(c *gin.Context) {
		randomImage := getRandomUnannotatedImage()
		c.JSON(http.StatusOK, gin.H{"uuid": randomImage.Id, "label": randomImage.Label, "provider": randomImage.Provider})
	})

	router.GET("/v1/annotation", func(c *gin.Context) {
		randomImage := getRandomAnnotatedImage()
		c.JSON(http.StatusOK, gin.H{"uuid": randomImage.Id, "label": randomImage.Label, "provider": randomImage.Provider, "annotations": randomImage.Annotations})
	})

	router.Run(":8081")
}