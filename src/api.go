package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
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
	"errors"
	"html"
	"net/url"
	"image"
	"github.com/nfnt/resize"
	"bytes"
	"image/gif"
    "image/jpeg"
    "image/png"
    "strings"
    "encoding/base64"
    "github.com/dgrijalva/jwt-go"
    "golang.org/x/crypto/bcrypt"
    "time"
    "github.com/getsentry/raven-go"
    "./datastructures"
	//"gopkg.in/h2non/bimg.v1"
)

var db *sql.DB
var geoipDb *geoip2.Reader

func ResizeImage(path string, width uint, height uint) ([]byte, string, error){
	buf := new(bytes.Buffer) 
	imgFormat := ""

	file, err := os.Open(path)
	if err != nil {
		log.Debug("[Resize Image Handler] Couldn't open image: ", err.Error())
		return buf.Bytes(), imgFormat, err
	}

	// decode jpeg into image.Image
	img, format, err := image.Decode(file)
	if err != nil {
		log.Debug("[Resize Image Handler] Couldn't decode image: ", err.Error())
		return buf.Bytes(), imgFormat, err
	}
	file.Close()

	resizedImg := resize.Resize(width, height, img, resize.NearestNeighbor)


	if format == "png" {
		err = png.Encode(buf, resizedImg)
		if err != nil {
			log.Debug("[Resize Image Handler] Couldn't encode image: ", err.Error())
	    	return buf.Bytes(), imgFormat, err
		}
	} else if format == "gif" {
		err = gif.Encode(buf, resizedImg, nil)
		if err != nil {
			log.Debug("[Resize Image Handler] Couldn't encode image: ", err.Error())
	    	return buf.Bytes(), imgFormat, err
		}
	} else {
		err = jpeg.Encode(buf, resizedImg, nil)
		if err != nil {
			log.Debug("[Resize Image Handler] Couldn't encode image: ", err.Error())
	    	return buf.Bytes(), imgFormat, err
		}
	}
	imgFormat = format

	return buf.Bytes(), imgFormat, nil
}

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
	    c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-Id, Cache-Control, X-Requested-With, X-Browser-Fingerprint, X-App-Identifier, Authorization, X-Api-Token, X-Moderation")
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
			if values[0] != "" {
				c.Writer.Header().Set("X-Request-Id", values[0])
			}
		}

		c.Next()
	}
}

func isModerationRequest(c *gin.Context) bool {
	if values, _ := c.Request.Header["X-Moderation"]; len(values) > 0 {
		if values[0] == "true" {
			return true
		}
	}

	return false
}


func pushCountryContributionToRedis(redisPool *redis.Pool, contributionsPerCountryRequest datastructures.ContributionsPerCountryRequest) {
	serialized, err := json.Marshal(contributionsPerCountryRequest)
	if err != nil { 
		log.Debug("[Push Contributions per Country to Redis] Couldn't create contributions-per-country request: ", err.Error())
		return
	}

	redisConn := redisPool.Get()
	defer redisConn.Close()

	_, err = redisConn.Do("RPUSH", "contributions-per-country", serialized)
	if err != nil { //just log error, but not abort (it's just some statistical information)
		log.Debug("[Push Contributions per Country to Redis] Couldn't update contributions-per-country: ", err.Error())
		return
	}
}

/*func annotationsValid(annotations []json.RawMessage) error{
	for _, r := range annotations {
		var obj map[string]interface{}
        err := json.Unmarshal(r, &obj)
        if err != nil {
            return err
        }

        shapeType := obj["type"]
        if shapeType == "rect" {
        	var rectangleAnnotation RectangleAnnotation 
			err = json.Unmarshal(r, &rectangleAnnotation)
			if err != nil {
				return err
			}
        } else if shapeType == "ellipse" {
        	var ellipsisAnnotation EllipsisAnnotation 
			err = json.Unmarshal(r, &ellipsisAnnotation)
			if err != nil {
				return err
			}
        } else if shapeType == "polygon" {
        	var polygonAnnotation PolygonAnnotation 
			err = json.Unmarshal(r, &polygonAnnotation)
			if err != nil {
				return err
			}
        } else {
        	return errors.New("Invalid shape type")
        }
	}

	if len(annotations) == 0 {
		return errors.New("annotations missing")
	}

	return nil
}*/

func annotationsValid(annotations []json.RawMessage) error{
	for _, r := range annotations {
		var obj map[string]interface{}
        err := json.Unmarshal(r, &obj)
        if err != nil {
            return err
        }

        shapeType := obj["type"]
        if shapeType == "rect" {
        	var rectangleAnnotation datastructures.RectangleAnnotation 
        	decoder := json.NewDecoder(bytes.NewReader([]byte(r)))
        	decoder.DisallowUnknownFields() //throw an error in case of an unknown field 
        	err = decoder.Decode(&rectangleAnnotation)
        	if err != nil {
        		raven.CaptureError(err, nil)
        		return err
        	}
        } else if shapeType == "ellipse" {
        	var ellipsisAnnotation datastructures.EllipsisAnnotation 
			decoder := json.NewDecoder(bytes.NewReader([]byte(r)))
        	decoder.DisallowUnknownFields() //throw an error in case of an unknown field 
        	err = decoder.Decode(&ellipsisAnnotation)
        	if err != nil {
        		raven.CaptureError(err, nil)
        		return err
        	}
        } else if shapeType == "polygon" {
        	var polygonAnnotation datastructures.PolygonAnnotation 
			decoder := json.NewDecoder(bytes.NewReader([]byte(r)))
        	decoder.DisallowUnknownFields() //throw an error in case of an unknown field 
        	err = decoder.Decode(&polygonAnnotation)
        	if err != nil {
        		raven.CaptureError(err, nil)
        		return err
        	}
        } else {
        	return errors.New("Invalid shape type")
        }
	}

	if len(annotations) == 0 {
		return errors.New("annotations missing")
	}

	return nil
}


func getBrowserFingerprint(c *gin.Context) string {
	browserFingerprint := ""
	if values, _ := c.Request.Header["X-Browser-Fingerprint"]; len(values) > 0 {
		browserFingerprint = values[0]
	}

	return browserFingerprint
}

func getAppIdentifier(c *gin.Context) string {
	appIdentifier := ""
	if values, _ := c.Request.Header["X-App-Identifier"]; len(values) > 0 {
		appIdentifier = values[0]
	}

	return appIdentifier
}

func getUsernameFromContext(c *gin.Context, authTokenHandler *AuthTokenHandler) (string, error) {
	username := ""
	accessTokenInfo := authTokenHandler.GetAccessTokenInfo(c)

	if accessTokenInfo.Username == "" { //in case access token is missing, try if there is an api token
		if !accessTokenInfo.Empty {
			return "", errors.New("The provided Access Token is either invalid or was revoked")
		}

		apiTokenInfo := authTokenHandler.GetAPITokenInfo(c)
		username = apiTokenInfo.Username
		if username == "" && !apiTokenInfo.Empty {
			return "", errors.New("The provided API Token is either invalid or was revoked")
		}
	} else {
		username = accessTokenInfo.Username
	}

	return username, nil
}

func donate(c *gin.Context, username string, imageSource datastructures.ImageSource, labelMap map[string]datastructures.LabelMapEntry, dir string, 
				redisPool *redis.Pool, statisticsPusher *StatisticsPusher, geodb *geoip2.Reader, autoUnlock bool) {
	label := c.PostForm("label")

	file, header, err := c.Request.FormFile("image")
	if err != nil {
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

    if !filetype.IsImage(fileHeader) {
        c.JSON(422, gin.H{"error": "Unsopported MIME type detected"})
        return
    }

    //check if image already exists by using an image hash
    imageInfo, err := getImageInfo(file)
    if err != nil {
        c.JSON(500, gin.H{"error": "Couldn't add photo - please try again later"})
        return
    }
    exists, err := imageExists(imageInfo.Hash)
    if err != nil {
        c.JSON(500, gin.H{"error": "Couldn't add photo - please try again later"})
        return 
    }
    if exists {
        c.JSON(409, gin.H{"error": "Couldn't add photo - image already exists"})
        return
    }

    if label != "" { //allow unlabeled donation. If label is provided it needs to be valid!
        if !isLabelValid(labelMap, label, []datastructures.Sublabel{}) {
        	c.JSON(409, gin.H{"error": "Couldn't add photo - invalid label"})
        	return
        }
    }

    addSublabels := false
    temp := c.PostForm("add_sublabels")
    if temp == "true" {
        addSublabels = true
    }


	labelMapEntry := labelMap[label]
	if !addSublabels {
		labelMapEntry.LabelMapEntries = nil
	}
	var labelMeEntry datastructures.LabelMeEntry
	var labelMeEntries []datastructures.LabelMeEntry
	labelMeEntry.Label = label
	labelMeEntry.Annotatable = true //assume that the label that was directly provided together with the donation is annotatable 
	for key, _ := range labelMapEntry.LabelMapEntries {
		labelMeEntry.Sublabels = append(labelMeEntry.Sublabels, datastructures.Sublabel{Name: key})
	}
	labelMeEntries = append(labelMeEntries, labelMeEntry)

    //image doesn't already exist, so save it and add it to the database
	u, err := uuid.NewV4()
	if err != nil {
		c.JSON(500, gin.H{"error": "Couldn't set add photo - please try again later"})	
		return
	}
	uuid := u.String()
	err = c.SaveUploadedFile(header, (dir + uuid))
	if err != nil {
		c.JSON(500, gin.H{"error": "Couldn't set add photo - please try again later"})	
		return
	}

	browserFingerprint := getBrowserFingerprint(c)

	imageInfo.Source = imageSource
	imageInfo.Name = uuid

	err = addDonatedPhoto(username, imageInfo, autoUnlock, browserFingerprint, labelMeEntries)
	if err != nil {
		c.JSON(500, gin.H{"error": "Couldn't add photo - please try again later"})	
		return
	}

	//get client IP address and try to determine country
	var contributionsPerCountryRequest datastructures.ContributionsPerCountryRequest
	contributionsPerCountryRequest.Type = "donation"
	contributionsPerCountryRequest.CountryCode = "--"
	ip := net.ParseIP(getIPAddress(c.Request))
	if ip != nil {
		record, err := geodb.Country(ip)
		if err != nil { //just log, but don't abort...it's just for statistics
			log.Debug("[Donation] Couldn't determine geolocation from ", err.Error())
				
		} else {
			contributionsPerCountryRequest.CountryCode = record.Country.IsoCode
		}
	}
	pushCountryContributionToRedis(redisPool, contributionsPerCountryRequest)
	statisticsPusher.PushAppAction(getAppIdentifier(c), "donation")

	c.JSON(http.StatusOK, gin.H{"uuid": uuid})
}

func IsFilenameValid(filename string) bool {
	_, err := uuid.FromString(filename)

	if err != nil {
		return false
	}
	return true
	/*for _, ch := range filename {
		if ((ch >= 'A') && (ch <= 'Z')) || ((ch >= 'a') && (ch <= 'z')) || ((ch >= '0') && (ch <= '9')) || (ch == '-') {
			continue
		}
		return false
	}
	return true*/
}


func main(){
	fmt.Printf("Starting API Service...\n")

	log.SetLevel(log.DebugLevel)

	releaseMode := flag.Bool("release", false, "Run in release mode")
	wordlistPath := flag.String("wordlist", "../wordlists/en/labels.json", "Path to label map")
	donationsDir := flag.String("donations_dir", "../donations/", "Location of the uploaded donations")
	unverifiedDonationsDir := flag.String("unverified_donations_dir", "../unverified_donations/", "Location of the uploaded but unverified donations")
	imageQuarantineDir := flag.String("image_quarantine_dir", "../quarantine/", "Location of the images that are put in quarantine")
	redisAddress := flag.String("redis_address", ":6379", "Address to the Redis server")
	redisMaxConnections := flag.Int("redis_max_connections", 500, "Max connections to Redis")
	geoIpDbPath := flag.String("geoip_db", "../geoip_database/GeoLite2-Country.mmdb", "Path to the GeoIP database")
	labelExamplesDir := flag.String("examples_dir", "../label_examples/", "Location of the label examples")
	userProfilePicturesDir := flag.String("avatars_dir", "../avatars/", "Avatars directory")
	useSentry := flag.Bool("use_sentry", false, "Use Sentry for error logging")
	maintenanceModeFile := flag.String("maintenance_mode_file", "../maintenance.tmp", "maintenance mode file")
	//labelGraphDefinitionPath := flag.String("label_graph_def", "../wordlists/en/graph.dot", "Path to the label graph definition")
	labelGraphDefinitionsPath := flag.String("label_graph_definitions", "../wordlists/en/graphdefinitions", "Path to the label graph definitions")
	listenPort := flag.Int("listen_port", 8081, "Specify the listen port")
	apiBaseUrl := flag.String("api_base_url", "http://127.0.0.1:8081/", "API Base URL")

	flag.Parse()
	if *releaseMode {
		fmt.Printf("[Main] Starting gin in release mode!\n")
		gin.SetMode(gin.ReleaseMode)
	}

	if *useSentry {
		fmt.Printf("Setting Sentry DSN\n")
		raven.SetDSN(SENTRY_DSN)
		raven.SetEnvironment("api")

		raven.CaptureMessage("Starting up api worker", nil)
	}

	log.Debug("[Main] Reading Label Map")
	labelMap, words, err := getLabelMap(*wordlistPath)
	if err != nil {
		fmt.Printf("[Main] Couldn't read label map...terminating!")
		log.Fatal(err)
	}

	//if the mostPopularLabels gets extended with sublabels, 
	//adapt getRandomGroupedImages() and validateImages() also!
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

	labelGraphRepository := NewLabelGraphRepository(*labelGraphDefinitionsPath)
	err = labelGraphRepository.Load()
	if err != nil {
		log.Fatal("[Main] Couldn't load label graph repository: ", err.Error())
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

	statisticsPusher := NewStatisticsPusher(redisPool)
	err = statisticsPusher.Load()
	if err != nil {
		log.Fatal("[Main] Couldn't load statistics pusher: ", err.Error())
	}

	sampleExportQueries := GetSampleExportQueries()
	authTokenHandler := NewAuthTokenHandler()

	//if file exists, start in maintenance mode
	maintenanceMode := false
	if _, err := os.Stat(*maintenanceModeFile); err == nil {
		maintenanceMode = true
		log.Info("[Main] Starting in maintenance mode")
	}



	router := gin.Default()
	if maintenanceMode {
		router.NoRoute(func(c *gin.Context) {
    		c.JSON(302, gin.H{"error": "Sorry for the inconvenience but we're performing some maintenance at the moment. We'll be back shortly."})
		})	
	} else {
		router.Use(CorsMiddleware())
		router.Use(RequestId())

		//serve images in "donations" directory with the possibility to scale images
		//before serving them
		router.GET("/v1/donation/:imageid", func(c *gin.Context) {
			params := c.Request.URL.Query()
			imageId := c.Param("imageid")

			var width uint
			width = 0
			if temp, ok := params["width"]; ok {
				n, err := strconv.ParseUint(temp[0], 10, 32)
			    if err == nil {
			        width = uint(n)
			    }
			}

			var height uint
			height = 0
			if temp, ok := params["height"]; ok {
				n, err := strconv.ParseUint(temp[0], 10, 32)
			    if err == nil {
	            	height = uint(n)
			    }
			}

			if !IsFilenameValid(imageId) {
				c.String(404, "Invalid filename")
				return
			} 

			unlocked, err := isImageUnlocked(imageId) 
			if err != nil {
				c.String(500, "Couldn't process request, please try again later")
				return
			}
			if !unlocked {
				c.String(404, "Couldn't access image, as image is still in locked mode")
				return
			}

			imgBytes, format, err := ResizeImage((*donationsDir + imageId), width, height)
			if err != nil {
				log.Debug("[Serving Donation] Couldn't serve donation: ", err.Error())
				c.String(500, "Couldn't process request, please try again later")
				return

			}

			c.Writer.Header().Set("Content-Type", ("image/" + format))
	        c.Writer.Header().Set("Content-Length", strconv.Itoa(len(imgBytes)))
	        _, err = c.Writer.Write(imgBytes) 
	        if err != nil {
	            log.Debug("[Serving Donation] Couldn't serve donation: ", err.Error())
	            c.String(500, "Couldn't process request, please try again later")
	            return
	        }
		})


		//serve images in "donations" directory with the possibility to scale images
		//before serving them
		router.GET("/v1/unverified-donation/:imageid", func(c *gin.Context) {
			params := c.Request.URL.Query()
			imageId := c.Param("imageid")

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfoFromUrl(c).Username

			var width uint
			width = 0
			if temp, ok := params["width"]; ok {
				n, err := strconv.ParseUint(temp[0], 10, 32)
			    if err == nil {
			        width = uint(n)
			    }
			}

			var height uint
			height = 0
			if temp, ok := params["height"]; ok {
				n, err := strconv.ParseUint(temp[0], 10, 32)
			    if err == nil {
	            	height = uint(n)
			    }
			}

			if !IsFilenameValid(imageId) {
				c.String(404, "Invalid filename")
				return
			} 

			unlocked, err := isOwnDonation(imageId, apiUser.Name) 
			if err != nil {
				c.String(500, "Couldn't process request, please try again later")
				return
			}
			if !unlocked {
				c.String(403, "You do not have the appropriate permissions to access the image")
				return
			}

			imgBytes, format, err := ResizeImage((*unverifiedDonationsDir + imageId), width, height)
			if err != nil {
				log.Debug("[Serving unverified Donation] Couldn't serve donation: ", err.Error())
				c.String(500, "Couldn't process request, please try again later")
				return

			}

			c.Writer.Header().Set("Content-Type", ("image/" + format))
	        c.Writer.Header().Set("Content-Length", strconv.Itoa(len(imgBytes)))
	        _, err = c.Writer.Write(imgBytes) 
	        if err != nil {
	            log.Debug("[Serving unverified Donation] Couldn't serve donation: ", err.Error())
	            c.String(500, "Couldn't process request, please try again later")
	            return
	        }
		})



		//the following endpoints are secured with a client id + client secret. 
		//that's mostly because currently each donation needs to be unlocked manually. 
		//(as we want to make sure that we don't accidentally host inappropriate content, like nudity)
		clientAuth := router.Group("/")
		clientAuth.Use(RequestId())
		clientAuth.Use(ClientAuthMiddleware())
		{
			clientAuth.Static("./v1/unverified/donation", *unverifiedDonationsDir)
			clientAuth.GET("/v1/internal/unverified-donations", func(c *gin.Context) {

				imageProvider := getParamFromUrlParams(c, "image_provider", "")
				shuffle := getParamFromUrlParams(c, "shuffle", "")
				orderRandomly := false
				if shuffle == "true" {
					orderRandomly = true
				}

				images, err := getAllUnverifiedImages(imageProvider, orderRandomly)
				use(images)
				if err != nil {
					c.JSON(500, gin.H{"error" : "Couldn't process request - please try again later"}) 
				} else {
					c.JSON(http.StatusOK, images)
				}
			})

			clientAuth.POST("/v1/internal/labelme/donate",  func(c *gin.Context) {
				imageSourceUrl := c.PostForm("image_source_url")

				var imageSource datastructures.ImageSource
				imageSource.Provider = "labelme"
				imageSource.Url = imageSourceUrl
				imageSource.Trusted = true


				var dir string
				var autoUnlock bool

				autoUnlockStr := c.PostForm("auto_unlock")
				if autoUnlockStr == "yes" {
					autoUnlock = true
					dir = *donationsDir
				} else {
					autoUnlock = false
					dir = *unverifiedDonationsDir
				}
				

				//we trust images from the labelme database, so we automatically save them
				//into the donations folder and unlock them per default.
				donate(c, "", imageSource, labelMap, dir, redisPool, statisticsPusher, geoipDb, autoUnlock)
			})

			clientAuth.POST("/v1/internal/auto-annotate/:imageid",  func(c *gin.Context) {
				imageId := c.Param("imageid")
				if imageId == "" {
					c.JSON(422, gin.H{"error": "invalid request - image id missing"})
					return
				}

				var annotations datastructures.Annotations
				err := c.BindJSON(&annotations)
				if err != nil {
					c.JSON(422, gin.H{"error": "invalid request - annotations missing"})
					return
				}

				err = annotationsValid(annotations.Annotations)
				if err != nil {
					c.JSON(422, gin.H{"error": "invalid request - annotations invalid"})
					return
				}

				var apiUser datastructures.APIUser
				apiUser.ClientFingerprint = ""
				apiUser.Name = ""

				_, err = addAnnotations(apiUser, imageId, annotations, true)
				if(err != nil){
					c.JSON(500, gin.H{"error": "Couldn't add annotations - please try again later"})
					return
				}
				c.JSON(201, nil)
			})

			clientAuth.GET("/v1/internal/auto-annotation",  func(c *gin.Context) {
				params := c.Request.URL.Query()

				labelsStr := ""
				if temp, ok := params["label"]; ok {
					labelsStr = temp[0]
				}

				if labelsStr == "" {
					c.JSON(422, gin.H{"error": "Please provide at least one label"})
					return
				}

				labels := strings.Split(labelsStr, ",")

				images, err := getImagesForAutoAnnotation(labels)
				if err != nil {
					c.JSON(500, gin.H{"error": "Couldn't get images - please try again later"})
					return
				}

				c.JSON(http.StatusOK, images)
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

				if param == "good" {
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
				} else if param == "bad" { //not handled at the moment, add later if needed

				} else if param == "delete" {
					err = deleteImage(imageId)
					if err != nil {
						c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
						return
					}

					dst := *unverifiedDonationsDir + imageId
					err := os.Remove(dst)
					if err != nil {
						log.Debug("[Main] Couldn't remove file ", dst)
						c.JSON(500, gin.H{"error" : "Couldn't process request - please try again later"})
						return
					}

				} else if param == "quarantine" {
					src := *unverifiedDonationsDir + imageId
					dst := *imageQuarantineDir + imageId
					err := os.Rename(src, dst)
					if err != nil {
						log.Debug("[Main] Couldn't move file ", src, " to ", dst)
						c.JSON(500, gin.H{"error" : "Couldn't process request - please try again later"})
						return
					}

					err = putImageInQuarantine(imageId)
					if err != nil {
						c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
						return
					}
				} else{
					c.JSON(404, gin.H{"error": "Couldn't process request - invalid parameter"})
					return
				}

				c.JSON(http.StatusOK, nil)
			})
		}

		router.GET("/v1/validation", func(c *gin.Context) {
			params := c.Request.URL.Query()

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

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

			imageId := ""
			if temp, ok := params["image_id"]; ok {
				imageId = temp[0]
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
				labelId, err := getLabelIdFromUrlParams(params) 
				if err != nil {
					c.JSON(422, gin.H{"error": "label id needs to be an integer"})
					return
				}

				image, err := getImageToValidate(imageId, labelId, apiUser.Name)
				if err != nil {
					c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
					return
				}

				if image.Id == "" {
					c.JSON(422, gin.H{"error": "Couldn't process request - empty result set"})
					return
				}
				
				c.JSON(http.StatusOK, gin.H{"image" : gin.H{ "uuid": image.Id, 
															 "provider": image.Provider, 
															 "unlocked": image.Unlocked, 
															 "url": getImageUrlFromImageId(*apiBaseUrl, image.Id, image.Unlocked),
														   }, 
											"label": image.Label, "sublabel": image.Sublabel, "num_yes": image.Validation.NumOfValid, 
											"num_no": image.Validation.NumOfInvalid, "uuid": image.Validation.Id })
			}
		})

		router.POST("/v1/donation/:imageid/labelme", func(c *gin.Context) {
			imageId := c.Param("imageid")

			var labels []datastructures.LabelMeEntry
			if c.BindJSON(&labels) != nil {
				c.JSON(400, gin.H{"error": "Couldn't process request - labels missing"})
				return
			}

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)

			apiUser.Name, err = getUsernameFromContext(c, authTokenHandler)
			if err != nil {
				c.JSON(401, gin.H{"error": err.Error()})
				return
			} 
			
			err := addLabelsToImage(apiUser, labelMap, imageId, labels)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, nil)
		})

		router.GET("/v1/donation/:imageid/labels", func(c *gin.Context) {
			imageId := c.Param("imageid")

			img, err := getImageToLabel(imageId, "")
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			c.JSON(http.StatusOK, img.AllLabels)
		})

		router.GET("/v1/donation/:imageid/validations/unannotated", func(c *gin.Context) {
			imageId := c.Param("imageid")

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			ids, err := getUnannotatedValidations(apiUser, imageId)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(http.StatusOK, ids)
		})

		router.GET("/v1/donation/:imageid/annotations", func(c *gin.Context) {
			imageId := c.Param("imageid")

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			annotatedImages, err := getAnnotations(apiUser, ParseResult{}, imageId, *apiBaseUrl)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if len(annotatedImages) == 0 {
				c.JSON(422, gin.H{"error": "Couldn't process request - missing result set"})
				return
			}

			c.JSON(200, annotatedImages)
		})

		router.POST("/v1/validation/:validationid/validate/:param", func(c *gin.Context) {
			validationId := c.Param("validationid")
			param := c.Param("param")

			if param != "yes" && param != "no" {
				c.JSON(404, nil)
				return
			}

			var imageValidationBatch datastructures.ImageValidationBatch 
			imageValidationBatch.Validations = append(imageValidationBatch.Validations, datastructures.ImageValidation {Uuid: validationId, Valid: param})


			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			moderatorAction := false
			if isModerationRequest(c) {
				if apiUser.Name != "" {
					userInfo, err := getUserInfo(apiUser.Name)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't process request - please try again later"})
						return
					}

					if userInfo.Permissions != nil && userInfo.Permissions.CanRemoveLabel {
						moderatorAction = true
					}
				}
			}

			err := validateImages(apiUser, imageValidationBatch, moderatorAction)
			if(err != nil){
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't process request - please try again later"})
				return
			} 

			//get client IP address and try to determine country
			var contributionsPerCountryRequest datastructures.ContributionsPerCountryRequest
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
			statisticsPusher.PushAppAction(getAppIdentifier(c), "validation")

			c.JSON(http.StatusOK, nil)
		})

		router.GET("/v1/labelme", func(c *gin.Context) {
			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			imageId := getParamFromUrlParams(c, "image_id", "")

			image, err := getImageToLabel(imageId, apiUser.Name)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if image.Id == "" {
				c.JSON(422, gin.H{"error": "Couldn't process request - empty result set"})
			} else {
				imageUrl := getImageUrlFromImageId(*apiBaseUrl, image.Id, image.Unlocked)

				c.JSON(http.StatusOK, gin.H{"image": gin.H{"uuid": image.Id, "provider": image.Provider, 
															"url": imageUrl, "unlocked": image.Unlocked,
															"width": image.Width, "height": image.Height}, 
											"all_labels": image.AllLabels})
			}
		})

		router.PATCH("/v1/refine", func(c *gin.Context) {
			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			var annotationRefinements []datastructures.BatchAnnotationRefinementEntry
			err := c.BindJSON(&annotationRefinements)
			if err != nil {
				c.JSON(422, gin.H{"error": "Couldn't process request - invalid request"})
				return
			}

			err = batchAnnotationRefinement(annotationRefinements, apiUser)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			c.JSON(204, nil)
		});

		router.GET("/v1/donations/labels", func(c *gin.Context) {
			query := getParamFromUrlParams(c, "query", "")

			orderRandomly := false
			shuffle := getParamFromUrlParams(c, "shuffle", "")
		    if shuffle == "true" {
		    	orderRandomly = true
		    }

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			if query == "" {
	        	c.JSON(422, gin.H{"error": "Couldn't process request - query missing"})
				return
	        }

			query, err = url.QueryUnescape(query)
		    if err != nil {
		        c.JSON(422, gin.H{"error": "Couldn't process request - please provide a valid query"})
				return
		    }

			queryParser := NewQueryParserV2(query)
	        parseResult, err := queryParser.Parse(1)
	        if err != nil {
	            c.JSON(422, gin.H{"error": err.Error()})
	            return
	        }

			imageInfos, err := getImagesLabels(apiUser, parseResult, *apiBaseUrl, orderRandomly)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			c.JSON(http.StatusOK, imageInfos)

			
		})

		router.PATCH("/v1/validation/validate", func(c *gin.Context) {
			var imageValidationBatch datastructures.ImageValidationBatch
			
			if c.BindJSON(&imageValidationBatch) != nil {
				c.JSON(400, gin.H{"error": "Couldn't process request - invalid patch"})
				return
			}

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			err := validateImages(apiUser, imageValidationBatch, false)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			//get client IP address and try to determine country
			var contributionsPerCountryRequest datastructures.ContributionsPerCountryRequest
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
			statisticsPusher.PushAppAction(getAppIdentifier(c), "validation")

			c.JSON(200, nil)
		})

		router.GET("/v1/export", func(c *gin.Context) {
			query, annotationsOnly, err := getExploreUrlParams(c)
			if err != nil {
				c.JSON(422, gin.H{"error": err.Error()})
				return
			}

			queryParser := NewQueryParser(query)
	        parseResult, err := queryParser.Parse(1)
	        if err != nil {
	            c.JSON(422, gin.H{"error": err.Error()})
	            return
	        }

	        jsonData, err := export(parseResult, annotationsOnly)
	        if err == nil {
	            c.JSON(http.StatusOK, jsonData)
	            return
	        }
	        
	        c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't export data, please try again later."})
	        return
		})


		router.GET("/v1/export/sample", func(c *gin.Context) {
			q := sampleExportQueries[random(0, len(sampleExportQueries) - 1)]
			c.JSON(http.StatusOK, q)
		})

		router.GET("/v1/label", func(c *gin.Context) {
			params := c.Request.URL.Query()
			if temp, ok := params["detailed"]; ok {
				if temp[0] == "true" {
					c.JSON(http.StatusOK, labelMap)
					return
				}
			}
			c.JSON(http.StatusOK, words)
		})

		/*router.GET("/v1/label/search", func(c *gin.Context) {
			params := c.Request.URL.Query()
			if temp, ok := params["q"]; ok {
				res, err := autocompleteLabel(temp[0])
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"Error": "Couldn't process request, please try again later."})
					return
				}
				c.JSON(http.StatusOK, res)
				return
			}

			c.JSON(422, gin.H{"error": "please provide a search query"})
		})*/


		router.GET("/v1/label/random", func(c *gin.Context) {
			label := words[random(0, len(words) - 1)]
			c.JSON(http.StatusOK, gin.H{"label": label})
		})

		router.Static("/v1/label/example", *labelExamplesDir)		

		router.POST("/v1/label/suggest", func(c *gin.Context) {
			type SuggestedLabel struct {
			    Name string `json:"label"`
			}
			var suggestedLabel SuggestedLabel

			if c.BindJSON(&suggestedLabel) != nil {
				c.JSON(422, gin.H{"error": "Couldn't process request - label missing"})
				return
			}

			escapedLabel := html.EscapeString(suggestedLabel.Name)
			err = addLabelSuggestion(escapedLabel)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			c.JSON(200, nil)
		})

		/*router.GET("/v1/label/suggest", func(c *gin.Context) {
			tags := ""
			params := c.Request.URL.Query()
			temp, ok := params["tags"] 
			if !ok {
				c.JSON(422, gin.H{"error": "Couldn't process request - no tags specified"})
				return
			}
			tags = temp[0]
			suggestedTags, err := getLabelSuggestions(tags)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(http.StatusOK, suggestedTags)
		})*/

		router.GET("/v1/label/popular", func(c *gin.Context) {
			popularLabels, err := getMostPopularLabels(10) //limit to 10
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(http.StatusOK, popularLabels)
		})

		router.GET("/v1/label/graph/:labelgraphname", func(c *gin.Context) {
			labelGraphName := c.Param("labelgraphname")

			labelGraph, err := labelGraphRepository.Get(labelGraphName) 
			if err != nil {
				c.JSON(422, gin.H{"error": "Couldn't process request - invalid label graph name"})
				return
			}

			labelGraphJson, err := labelGraph.GetJson()
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"graph" : labelGraphJson, "metadata": labelGraph.GetMetadata()})
		})

		router.POST("/v1/label/graph-editor/evaluate", func(c *gin.Context) {
			type LabelGraphInput struct {
			    Data string `json:"data"`
			}

			var labelGraphInput LabelGraphInput
			if c.BindJSON(&labelGraphInput) != nil {
				c.JSON(422, gin.H{"error": "Couldn't process request - 'data' missing"})
				return
			}

			var data []byte
			data, err = base64.StdEncoding.DecodeString(labelGraphInput.Data)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			labelGraph := NewLabelGraph("", LabelGraphMappingEntry{})
			err := labelGraph.LoadFromString(string(data)) 
			if err != nil {
				c.JSON(422, gin.H{"error": "Couldn't process request"})
				return
			}

			labelGraphJson, err := labelGraph.GetJson()
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(http.StatusOK, labelGraphJson)
		})

		router.GET("/v1/label/graph/:labelgraphname/query-builder", func(c *gin.Context) {
			params := c.Request.URL.Query()

			labelGraphName := c.Param("labelgraphname")

			identifier := ""
			if temp, ok := params["identifier"]; ok {
				identifier, err = url.QueryUnescape(temp[0])
				if err != nil {
					c.JSON(422, gin.H{"error": "Couldn't process request - please provide a valid identifier"})
					return
				}
			}

			if identifier == "" {
				c.JSON(422, gin.H{"error": "Couldn't process request - please provide a valid identifier"})
				return
			}

			labelGraph, err := labelGraphRepository.Get(labelGraphName) 
			if err != nil {
				c.JSON(422, gin.H{"error": "Couldn't process request - invalid label graph name"})
				return
			}

			q := ""
			children := labelGraph.GetChildren(identifier)
			for _, child := range children {
				if child == nil { 
					continue
				}
				uuid := child.Attrs["id"]
				if uuid == "" {
					continue
				}

				uuid, err = strconv.Unquote(uuid)
				if err != nil {
					c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
					return
				}

				if q == "" {
					q += uuid
				} else {
					q += " | " + uuid
				}
			}

			c.JSON(http.StatusOK, gin.H{"query": q})
		})

		router.POST("/v1/donate", func(c *gin.Context) {
			var imageSource datastructures.ImageSource
			imageSource.Provider = "donation"
			imageSource.Trusted = false

			username, err := getUsernameFromContext(c, authTokenHandler)
			if err != nil {
				c.JSON(401, gin.H{"error": err.Error()})
				return
			}


			donate(c, username, imageSource, labelMap, *unverifiedDonationsDir, redisPool, statisticsPusher, geoipDb, false)
		})

		router.POST("/v1/report/:imageid", func(c *gin.Context) {
			imageId := c.Param("imageid")

			var report datastructures.Report
			if(c.BindJSON(&report) != nil){
				c.JSON(422, gin.H{"error": "reason missing - please provide a valid 'reason'"})
				return
			}

			s := "Someone reported a violation (uuid: " + imageId + ", reason: " + report.Reason + ")" 
			raven.CaptureMessage(s, nil)

			err := reportImage(imageId, report.Reason)
			if(err != nil){
				c.JSON(500, gin.H{"error": "Couldn't report image - please try again later"})
				return
			}
			c.JSON(http.StatusOK, nil)
		})

		router.GET("/v1/validations/unannotated", func(c *gin.Context) {
			query := getParamFromUrlParams(c, "query", "")
			if query != "" {
				query, err = url.QueryUnescape(query)
		        if err != nil {
		            c.JSON(422, gin.H{"error": "please provide a valid query"})
					return
		        }

		        queryParser := NewQueryParserV2(query)
		        parseResult, err := queryParser.Parse(1)
		        if err != nil {
		            c.JSON(422, gin.H{"error": err.Error()})
		            return
		        }

		        orderRandomly := false
		        shuffle := getParamFromUrlParams(c, "shuffle", "")
		        if shuffle == "true" {
		        	orderRandomly = true
		        }

		        var apiUser datastructures.APIUser
				apiUser.ClientFingerprint = getBrowserFingerprint(c)
				apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

				if len(parseResult.queryValues) == 0 {
					c.JSON(422, gin.H{"error": "Couldn't process request - please provide a valid query!"})
					return	
				}

		        annotationTasks, err := getAvailableAnnotationTasks(apiUser, parseResult, orderRandomly, *apiBaseUrl)
		        if err != nil {
		        	c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
					return
		        }

				c.JSON(http.StatusOK, annotationTasks)	
				return	        
		    } 

		    c.JSON(422, gin.H{"error": "please provide a valid query"})
		})

		router.POST("/v1/validation/:validationid/blacklist-annotation", func(c *gin.Context) {
			validationId := c.Param("validationid")

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			if apiUser.Name == "" {
				c.JSON(401, gin.H{"error": "Authentication required"})
				return
			}

			err := blacklistForAnnotation(validationId, apiUser)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't blacklist annotation - please try again later"})
				return
			}

			c.JSON(http.StatusOK, nil)
		})

		router.POST("/v1/validation/:validationid/not-annotatable", func(c *gin.Context) {
			validationId := c.Param("validationid")

			err := markValidationAsNotAnnotatable(validationId)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't mark validation as not-annotatable - please try again later"})
				return
			}

			c.JSON(http.StatusOK, nil)
		})

		router.PUT("/v1/annotation/:annotationid", func(c *gin.Context) {
			annotationId := c.Param("annotationid")

			var annotations datastructures.Annotations
			err := c.BindJSON(&annotations)
			if err != nil {
				c.JSON(422, gin.H{"error": "invalid request - annotations missing"})
				return
			}

			err = annotationsValid(annotations.Annotations)
			if err != nil {
				c.JSON(422, gin.H{"error": "invalid request - annotations invalid"})
				return
			}

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			err = updateAnnotation(apiUser, annotationId, annotations)
			if(err != nil){
				c.JSON(500, gin.H{"error": "Couldn't update annotation - please try again later"})
				return
			}


			//get client IP address and try to determine country
			var contributionsPerCountryRequest datastructures.ContributionsPerCountryRequest
			contributionsPerCountryRequest.Type = "annotation"
			contributionsPerCountryRequest.CountryCode = "--"
			ip := net.ParseIP(getIPAddress(c.Request))
			if ip != nil {
				record, err := geoipDb.Country(ip)
				if err != nil { //just log, but don't abort...it's just for statistics
					log.Debug("[Annotate] Couldn't determine geolocation from ", err.Error())
					
				} else {
					contributionsPerCountryRequest.CountryCode = record.Country.IsoCode
				}
			}
			pushCountryContributionToRedis(redisPool, contributionsPerCountryRequest)
			statisticsPusher.PushAppAction(getAppIdentifier(c), "annotation")

			c.JSON(201, nil)
		})

		router.POST("/v1/annotation/:annotationid/validate/:param", func(c *gin.Context) {
			annotationId := c.Param("annotationid")
			param := c.Param("param")

			parameter := false
			if param == "yes" {
				parameter = true
			} else if param == "no" {
				parameter = false
			} else{
				c.JSON(404, nil)
				return
			}

			var labelValidationEntry datastructures.LabelValidationEntry
			if c.BindJSON(&labelValidationEntry) != nil {
				c.JSON(400, gin.H{"error": "Couldn't process request - please provide valid label(s)"})
				return
			}

			if ((labelValidationEntry.Label == "") && (labelValidationEntry.Sublabel == "")) {
				c.JSON(400, gin.H{"error": "Please provide a valid label"})
				return
			}

			browserFingerprint := getBrowserFingerprint(c)

			err := validateAnnotatedImage(browserFingerprint, annotationId, labelValidationEntry, parameter)
			if(err != nil){
				c.JSON(http.StatusInternalServerError, gin.H{"Error": "Database Error: Couldn't update data"})
				return
			} 

			//get client IP address and try to determine country
			var contributionsPerCountryRequest datastructures.ContributionsPerCountryRequest
			contributionsPerCountryRequest.Type = "validation"
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
			statisticsPusher.PushAppAction(getAppIdentifier(c), "validation")

			c.JSON(http.StatusOK, nil)
		})

		router.POST("/v1/annotate/:imageid", func(c *gin.Context) {
			imageId := c.Param("imageid")
			if imageId == "" {
				c.JSON(422, gin.H{"error": "invalid request - image id missing"})
				return
			}

			var annotations datastructures.Annotations
			err := c.BindJSON(&annotations)
			if err != nil {
				c.JSON(422, gin.H{"error": "invalid request - annotations missing"})
				return
			}

			err = annotationsValid(annotations.Annotations)
			if err != nil {
				c.JSON(422, gin.H{"error": "invalid request - annotations invalid"})
				return
			}

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username


			annotationId, err := addAnnotations(apiUser, imageId, annotations, false)
			if(err != nil){
				c.JSON(500, gin.H{"error": "Couldn't add annotations - please try again later"})
				return
			}


			//get client IP address and try to determine country
			var contributionsPerCountryRequest datastructures.ContributionsPerCountryRequest
			contributionsPerCountryRequest.Type = "annotation"
			contributionsPerCountryRequest.CountryCode = "--"
			ip := net.ParseIP(getIPAddress(c.Request))
			if ip != nil {
				record, err := geoipDb.Country(ip)
				if err != nil { //just log, but don't abort...it's just for statistics
					log.Debug("[Annotate] Couldn't determine geolocation from ", err.Error())
					
				} else {
					contributionsPerCountryRequest.CountryCode = record.Country.IsoCode
				}
			}
			pushCountryContributionToRedis(redisPool, contributionsPerCountryRequest)
			statisticsPusher.PushAppAction(getAppIdentifier(c), "annotation")

			c.Writer.Header().Set("Location", (*apiBaseUrl + "/v1/annotation?annotation_id=" + annotationId))
			c.JSON(201, nil)
		})

		router.GET("/v1/annotate", func(c *gin.Context) {
			params := c.Request.URL.Query()

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username
			

			addAutoAnnotations := false
			if temp, ok := params["add_auto_annotations"]; ok {
				if temp[0] == "true" {
					addAutoAnnotations = true
				}
			}

			labelId, err := getLabelIdFromUrlParams(params) 
			if err != nil {
				c.JSON(422, gin.H{"error": "label id needs to be an integer"})
				return
			}

			validationId := getValidationIdFromUrlParams(params)

			img, err := getImageForAnnotation(apiUser.Name, addAutoAnnotations, validationId, labelId)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if img.Id == "" {
				c.JSON(422, gin.H{"error": "Couldn't process request - missing result set"})
				return
			}

			img.Url = getImageUrlFromImageId(*apiBaseUrl, img.Id, img.Unlocked)
			
			c.JSON(http.StatusOK, img)
		})

		router.GET("/v1/annotations", func(c *gin.Context) {
			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			query := getParamFromUrlParams(c, "query", "")
			query, err = url.QueryUnescape(query)
	        if err != nil {
	            c.JSON(422, gin.H{"error": "invalid query"})
	            return
	        }

			queryParser := NewQueryParser(query)
	        parseResult, err := queryParser.Parse(1)
	        if err != nil {
	            c.JSON(422, gin.H{"error": err.Error()})
	            return
	        }

			annotatedImages, err := getAnnotations(apiUser, parseResult, "", *apiBaseUrl)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			c.JSON(200, annotatedImages)
		})

		router.GET("/v1/annotation", func(c *gin.Context) {
			params := c.Request.URL.Query()

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			log.Debug(apiUser.Name)
			

			autoGenerated := false
			if temp, ok := params["auto_generated"]; ok {
				if temp[0] == "true" {
					autoGenerated = true
				}
			}

			annotationId := getParamFromUrlParams(c, "annotation_id", "")

			rev := getParamFromUrlParams(c, "rev", "-1")
			revision, err := strconv.ParseInt(rev, 10, 32)
			if err != nil {
				c.JSON(422, gin.H{"error": "Invalid request - please provide a valid revision"})
				return
			}

			annotatedImage, err := getAnnotatedImage(apiUser, annotationId, autoGenerated, int32(revision))
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if annotatedImage.Id == "" {
				c.JSON(422, gin.H{"error": "Couldn't process request - missing result set"})
				return
			}

			annotatedImage.Image.Url = getImageUrlFromImageId(*apiBaseUrl, annotatedImage.Image.Id, annotatedImage.Image.Unlocked)


			c.JSON(http.StatusOK, annotatedImage)
		})

		router.GET("/v1/quiz-refine", func(c *gin.Context) {
			randomImage, err := getRandomAnnotationForQuizRefinement()
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if randomImage.Image.Uuid == "" {
				c.JSON(422, gin.H{"error": "Couldn't process request - missing result set"})
				return
			}

			c.JSON(200, randomImage)
		})

		router.GET("/v1/refine", func(c *gin.Context) {
			annotationDataId := getParamFromUrlParams(c, "annotation_data_id", "")

			var parseResult ParseResult
			query := getParamFromUrlParams(c, "query", "")
			if query != "" {
				query, err = url.QueryUnescape(query)
		        if err != nil {
		            c.JSON(422, gin.H{"error": "invalid query"})
		            return
		        }

				queryParser := NewQueryParserV2(query)
		        parseResult, err = queryParser.Parse(1)
		        if err != nil {
		            c.JSON(422, gin.H{"error": err.Error()})
		            return
		        }
		    }

		    if query == "" && annotationDataId == "" {
		    	c.JSON(422, gin.H{"error": "Couldn't process request - invalid request"})
		    	return
		    }
		    
			annotations, err := getAnnotationsForRefinement(parseResult, *apiBaseUrl, annotationDataId)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if len(annotations) == 0 {
				if annotationDataId != "" {
					c.JSON(422, gin.H{"error": "Couldn't process request - missing result set"})
					return
				}
			}

			if annotationDataId != "" {
				c.JSON(200, annotations[0]) 
				return
			}

			c.JSON(200, annotations) 
		})

		router.POST("/v1/annotation/:annotationid/refine/:annotationdataid", func(c *gin.Context) {
			annotationId := c.Param("annotationid")
			if annotationId == "" {
				c.JSON(422, gin.H{"error": "Invalid request - please provide a valid annotation id"})
				return
			}

			annotationDataId := c.Param("annotationdataid")
			if annotationDataId == "" {
				c.JSON(422, gin.H{"error": "Invalid request - please provide a valid annotation data id"})
				return
			}

			var annotationRefinementEntries []datastructures.AnnotationRefinementEntry
			if c.BindJSON(&annotationRefinementEntries) != nil {
				c.JSON(400, gin.H{"error": "Couldn't process request - please provide a valid label id"})
				return
			}

			browserFingerprint := getBrowserFingerprint(c)

			err := addOrUpdateRefinements(annotationId, annotationDataId, annotationRefinementEntries, browserFingerprint)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't add annotation refinement - please try again later"})
				return
			}

			c.JSON(201, nil)
		})

		router.POST("/v1/blog/subscribe", func(c *gin.Context) {
			var blogSubscribeRequest datastructures.BlogSubscribeRequest
			if c.BindJSON(&blogSubscribeRequest) != nil {
				c.JSON(400, gin.H{"error": "Couldn't process request - please provide a valid email address"})
				return
			}


			redisConn := redisPool.Get()
			defer redisConn.Close()

			serialized, err := json.Marshal(blogSubscribeRequest)
			if err != nil { 
				log.Debug("[Subscribe to blog] Couldn't create subscribe-to-blog request: ", err.Error())
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}


			_, err = redisConn.Do("RPUSH", "subscribe-to-blog", serialized)
			if err != nil {
				log.Debug("[Subscribe to blog] Couldn't subscribe to blog: ", err.Error())
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			c.JSON(http.StatusOK, nil)
		})

		router.POST("/v1/login", func(c *gin.Context) {
			type MyCustomClaims struct {
				Username string `json:"username"`
				Created int64 `json:"created"`
				jwt.StandardClaims
			}


			auth := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)

	        if len(auth) != 2 || auth[0] != "Basic" {
	            c.JSON(422, gin.H{"error": "Authorization failed"})
	            return
	        }

	        payload, err := base64.StdEncoding.DecodeString(auth[1])
	        if err != nil {
	        	c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
	            return
	        }
			pair := strings.SplitN(string(payload), ":", 2)

			userExists, err := userExists(pair[0])
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
	            return
			}

			if !userExists {
				c.JSON(401, gin.H{"error": "Invalid username or password"})
	            return
			}


			hashedPassword, err := getHashedPasswordForUser(pair[0])
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
	            return
			}

			err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(pair[1]))
			if err == nil { //nil means password match
				now := time.Now()
				expirationTime := now.Add(time.Hour * 24 * 7)

				claims := MyCustomClaims{
					pair[0],
					now.Unix(),
					jwt.StandardClaims{
						ExpiresAt: expirationTime.Unix(),
						Issuer: "imagemonkey-api",
					},
				}

				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

				tokenString, err := token.SignedString([]byte(JWT_SECRET))
				if err != nil {
					c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
	            	return
				}

				err = addAccessToken(pair[0], tokenString, expirationTime.Unix())
				if err != nil {
					c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
	            	return
				}


				c.JSON(http.StatusOK, gin.H{"token": tokenString})
				return
			}

			c.JSON(401, nil)
		})

		router.POST("/v1/logout", func(c *gin.Context) {
			accessTokenInfo := authTokenHandler.GetAccessTokenInfo(c)
			err := removeAccessToken(accessTokenInfo.Token)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
	            return
			}

			c.JSON(200, nil)
		})

		router.POST("/v1/signup", func(c *gin.Context) {
			var userSignupRequest datastructures.UserSignupRequest
			
			if c.BindJSON(&userSignupRequest) != nil {
				c.JSON(400, gin.H{"error": "Couldn't process request - invalid data"})
				return
			}

			if ((userSignupRequest.Username == "") || (userSignupRequest.Password == "") || (userSignupRequest.Email == "")) {
				c.JSON(422, gin.H{"error": "Invalid data"})
	            return
			}

			if(!isAlphaNumeric(userSignupRequest.Username)){
				c.JSON(422, gin.H{"error": "Username contains invalid characters"})
	            return
			}

			userExists, err := userExists(userSignupRequest.Username)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
	            return
			}

			if userExists {
				c.JSON(409, gin.H{"error": "Username already taken"})
	            return
			}

			emailExists, err := emailExists(userSignupRequest.Email)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
	            return
			}

			if emailExists {
				c.JSON(409, gin.H{"error": "There already exists a username with this email address"})
	            return
			}

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userSignupRequest.Password), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
	            return
			}

			err = createUser(userSignupRequest.Username, hashedPassword, userSignupRequest.Email)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
	            return
			}

			c.JSON(201, nil)
		})

		router.GET("/v1/user/:username/profile", func(c *gin.Context) {
			username := c.Param("username")
			if username == "" {
				c.JSON(422, gin.H{"error": "Invalid request - username missing"})
				return
			}

			userExists, err := userExists(username) 
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
	            return
			}

			if !userExists {
				c.JSON(422, gin.H{"error": "Invalid request - username doesn't exist"})
				return
			}

			userStatistics, err := getUserStatistics(username)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
	            return
			}

			c.JSON(200, gin.H{"statistics": userStatistics})
		})

		router.POST("/v1/user/:username/password_reset", func(c *gin.Context) {
			username := c.Param("username")
			if username == "" {
				c.JSON(422, gin.H{"error": "Invalid request - username missing"})
				return
			}

			c.JSON(201, nil)
		})

		router.POST("/v1/user/:username/profile/change_picture", func(c *gin.Context) {
			username := c.Param("username")
			if username == "" {
				c.JSON(422, gin.H{"error": "Invalid request - username missing"})
				return
			}

			accessTokenInfo := authTokenHandler.GetAccessTokenInfo(c)
			if !accessTokenInfo.Valid {
				c.JSON(403, gin.H{"error": "Please provide a valid access token"})
				return
			}

			if accessTokenInfo.Username != username {
				c.JSON(403, gin.H{"error": "Permission denied"})
				return
			}

			file, header, err := c.Request.FormFile("image")
			if(err != nil){
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

	        u, err := uuid.NewV4()
	        if err != nil {
	        	c.JSON(500, gin.H{"error": "Couldn't set profile picture - please try again later"})	
				return
	        }
	        uuid := u.String()
			err = c.SaveUploadedFile(header, (*userProfilePicturesDir + uuid))
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't set profile picture - please try again later"})	
				return
			}

			oldProfilePicture, err := changeProfilePicture(username, uuid)
			if(err != nil){
				c.JSON(500, gin.H{"error": "Couldn't set profile picture - please try again later"})	
				return
			}

			//if there exists an old profile picture, remove it.
			//we don't care if we can't remove it for some reason (it's not worth to fail the request just because of that) 
			if oldProfilePicture != "" {
				dst := *userProfilePicturesDir + oldProfilePicture
				os.Remove(dst)
			}

			c.JSON(201, nil)

		})

		router.POST("/v1/user/:username/api-token", func(c *gin.Context) {
			type ApiTokenRequest struct {
				Description string `json:"description"`
			}
			username := c.Param("username")

			if username == "" {
				c.JSON(422, gin.H{"error": "Invalid request - username missing"})
				return
			}

			var apiTokenRequest ApiTokenRequest
			if c.BindJSON(&apiTokenRequest) != nil {
				c.JSON(422, gin.H{"error": "Invalid request - description missing"})
				return
			}

			accessTokenInfo := authTokenHandler.GetAccessTokenInfo(c)

			if accessTokenInfo.Username != username { 
				c.JSON(422, gin.H{"error": "Invalid access token"})
				return
			}

			apiToken, err := generateApiToken(username, apiTokenRequest.Description)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't generate API token - please try again later"})	
				return
			}

			c.JSON(201, apiToken)
		})

		router.POST("/v1/user/:username/api-token/:token/revoke", func(c *gin.Context) {
			username := c.Param("username")
			token := c.Param("token")

			if username == "" {
				c.JSON(422, gin.H{"error": "Invalid request - username missing"})
				return
			}

			accessTokenInfo := authTokenHandler.GetAccessTokenInfo(c)

			if accessTokenInfo.Username != username { 
				c.JSON(422, gin.H{"error": "Invalid access token"})
				return
			}

			revoked, err := revokeApiToken(username, token)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't revoke API token - please try again later"})	
				return
			}

			if !revoked {
				c.JSON(400, gin.H{"error": "Couldn't revoke API token - invalid token"})	
				return
			}

			c.JSON(200, nil)
		})

		router.GET("/v1/user/:username/profile/avatar", func(c *gin.Context) {
			params := c.Request.URL.Query()
			username := c.Param("username")
			if username == "" {
				c.JSON(422, gin.H{"error": "Invalid request - username missing"})
				return
			}

			userInfo, err := getUserInfo(username)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if userInfo.Name == "" {
				c.JSON(422, gin.H{"error": "Incorrect username"})
				return
			}


			var width uint
			width = 0
			if temp, ok := params["width"]; ok {
				n, err := strconv.ParseUint(temp[0], 10, 32)
			    if err == nil {
			        width = uint(n)
			    }
			}

			var height uint
			height = 0
			if temp, ok := params["height"]; ok {
				n, err := strconv.ParseUint(temp[0], 10, 32)
			    if err == nil {
	            	height = uint(n)
			    }
			}

			fname := userInfo.ProfilePicture
			if userInfo.ProfilePicture == "" {
				fname = "default.png"
			}

			imgBytes, format, err := ResizeImage((*userProfilePicturesDir + fname), width, height)
			if err != nil {
				log.Debug("[Serving Avatar] Couldn't serve avatar: ", err.Error())
				c.String(500, "Couldn't process request - please try again later")
				return

			}

			c.Writer.Header().Set("Content-Type", ("image/" + format))
	        c.Writer.Header().Set("Content-Length", strconv.Itoa(len(imgBytes)))
	        _, err = c.Writer.Write(imgBytes) 
	        if err != nil {
	            log.Debug("[Serving Avatar] Couldn't serve avatar: ", err.Error())
	            c.String(500, "Couldn't process request - please try again later")
	            return
	        }
		})


		router.GET("/v1/statistics/annotation", func(c *gin.Context) {
			//currently only last-month is allowed as period
			statistics, err := getAnnotationStatistics("last-month")
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, gin.H{"statistics": statistics, "period": "last-month"})
		})

		router.GET("/v1/statistics/validation", func(c *gin.Context) {
			//currently only last-month is allowed as period
			statistics, err := getValidationStatistics("last-month")
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, gin.H{"statistics": statistics, "period": "last-month"})
		})

		/*router.GET("/v1/activity/validation", func(c *gin.Context) {
			//currently only last-month is allowed as period
			activity, err := getValidationActivity("last-month")
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, gin.H{"activity": activity, "period": "last-month"})
		})

		router.GET("/v1/activity/annotation", func(c *gin.Context) {
			//currently only last-month is allowed as period
			activity, err := getAnnotationActivity("last-month")
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, gin.H{"activity": activity, "period": "last-month"})
		})*/

		router.GET("/v1/activity", func(c *gin.Context) {
			//currently only last-month is allowed as period
			activity, err := getActivity("last-month")
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, gin.H{"activity": activity, "period": "last-month"})
		})
	}

	router.Run(":" + strconv.FormatInt(int64(*listenPort), 10))
}