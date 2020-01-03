package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	commons "github.com/bbernhard/imagemonkey-core/commons"
	imagemonkeydb "github.com/bbernhard/imagemonkey-core/database"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	img "github.com/bbernhard/imagemonkey-core/image"
	parser "github.com/bbernhard/imagemonkey-core/parser/v2"
	"github.com/dgrijalva/jwt-go"
	"github.com/getsentry/raven-go"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/gomodule/redigo/redis"
	"github.com/h2non/filetype"
	"github.com/oschwald/geoip2-golang"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"html"
	"image"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var geoipDb *geoip2.Reader

func handleImageAnnotationsRequest(c *gin.Context, imageId string, imageMonkeyDatabase *imagemonkeydb.ImageMonkeyDatabase,
	authTokenHandler *AuthTokenHandler, apiBaseUrl string) {
	var apiUser datastructures.APIUser
	apiUser.ClientFingerprint = getBrowserFingerprint(c)
	apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

	_, err := uuid.FromString(imageId)
	if err != nil {
		c.JSON(422, gin.H{"error": "Couldn't process request - please provide a valid image id"})
		return
	}

	annotatedImages, err := imageMonkeyDatabase.GetAnnotations(apiUser, parser.ParseResult{}, imageId, apiBaseUrl)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "Couldn't process request - please try again later"})
		return
	}

	if len(annotatedImages) == 0 {
		imageExistsForUser, err := imageMonkeyDatabase.ImageExistsForUser(imageId, apiUser.Name)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "Couldn't process request - please try again later"})
			return
		}
		if !imageExistsForUser {
			c.JSON(422, gin.H{"error": "Couldn't process request - please provide a valid image id"})
			return
		}
	}

	c.JSON(200, annotatedImages)
}

func handleUnverifiedDonation(imageId string, action string,
	donationsDir string, unverifiedDonationsDir string, imageQuarantineDir string,
	imageMonkeyDatabase *imagemonkeydb.ImageMonkeyDatabase) (int, error) {
	//verify that uuid is a valid UUID (to prevent path injection)
	_, err := uuid.FromString(imageId)
	if err != nil {
		return 400, errors.New("Couldn't process request - not a valid image id")
	}

	if action == "good" {
		src := unverifiedDonationsDir + imageId
		dst := donationsDir + imageId
		err := os.Rename(src, dst)
		if err != nil {
			log.Debug("[Main] Couldn't move file ", src, " to ", dst)
			return 500, errors.New("Couldn't process request - please try again later")
		}

		err = imageMonkeyDatabase.UnlockImage(imageId)
		if err != nil {
			return 500, errors.New("Couldn't process request - please try again later")
		}
	} else if action == "bad" { //not handled at the moment, add later if needed

	} else if action == "delete" {
		err = imageMonkeyDatabase.DeleteImage(imageId)
		if err != nil {
			return 500, errors.New("Couldn't process request - please try again later")
		}

		dst := unverifiedDonationsDir + imageId
		err := os.Remove(dst)
		if err != nil {
			log.Debug("[Main] Couldn't remove file ", dst)
			return 500, errors.New("Couldn't process request - please try again later")
		}

	} else if action == "quarantine" {
		src := unverifiedDonationsDir + imageId
		dst := imageQuarantineDir + imageId
		err := os.Rename(src, dst)
		if err != nil {
			log.Debug("[Main] Couldn't move file ", src, " to ", dst)
			return 500, errors.New("Couldn't process request - please try again later")
		}

		err = imageMonkeyDatabase.PutImageInQuarantine(imageId)
		if err != nil {
			return 500, errors.New("Couldn't process request - please try again later")
		}
	} else {
		return 404, errors.New("Couldn't process request - invalid parameter")
	}

	return 201, nil
}

func IsImageCollectionNameValid(s string) bool {
	for _, c := range s {
		if !(c > 47 && c < 58) && // numeric (0-9)
			!(c > 64 && c < 91) && // upper alpha (A-Z)
			!(c > 96 && c < 123) && // lower alpha (a-z)
			!(c == 45) && //hyphen (-)
			!(c == 95) { //underline (_)
			return false
		}
	}
	return true
}

//Middleware to ensure that the correct X-Client-Id and X-Client-Secret are provided in the header
func ClientAuthMiddleware(xClientId string, xClientSecret string) gin.HandlerFunc {
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

		if !((clientSecret == xClientSecret) && (clientId == xClientId)) {
			c.String(401, "Please provide a valid client id and client secret")
			c.AbortWithStatus(401)
			return
		}

		c.Next()
	}
}

//CORS Middleware
func CorsMiddleware(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-Id, Cache-Control, X-Requested-With, X-Total-Count, X-Browser-Fingerprint, X-App-Identifier, Authorization, X-Api-Token, X-Moderation, X-Client-Id, X-Client-Secret")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, HEAD")

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

func pushAnnotationCoverageUpdateRequestToRedis(redisPool *redis.Pool, uuid string, t string) {
	updateAnnotationCoverageRequest := datastructures.UpdateAnnotationCoverageRequest{Uuid: uuid, Type: t}
	serialized, err := json.Marshal(updateAnnotationCoverageRequest)
	if err != nil {
		log.Debug("[Notify data processor] Couldn't create update annotation coverage request: ", err.Error())
		raven.CaptureError(err, nil)
		return
	}

	redisConn := redisPool.Get()
	defer redisConn.Close()

	_, err = redisConn.Do("RPUSH", commons.UPDATE_IMAGE_ANNOTATION_COVERAGE_TOPIC, serialized)
	if err != nil {
		log.Debug("[Notify data processor] Couldn't update annotation coverage: ", err.Error())
		raven.CaptureError(err, nil)
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
}*/

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

func donate(c *gin.Context, db *imagemonkeydb.ImageMonkeyDatabase, username string, imageSource datastructures.ImageSource,
	labelMap map[string]datastructures.LabelMapEntry, metalabels *commons.MetaLabels, dir string,
	redisPool *redis.Pool, statisticsPusher *commons.StatisticsPusher, geodb *geoip2.Reader, autoUnlock bool) {
	label := c.PostForm("label")

	if imageSource.Provider == "imagehunt" {
		if label == "" {
			c.JSON(400, gin.H{"error": "Please provide a label"})
			return
		} else {
			if !commons.IsLabelValid(labelMap, metalabels, label, []datastructures.Sublabel{}) {
				c.JSON(400, gin.H{"error": "Please provide a valid label"})
				return
			}
		}
	}

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
	imageInfo, err := commons.GetImageInfo(file)
	if err != nil {
		c.JSON(500, gin.H{"error": "Couldn't add photo - please try again later"})
		return
	}
	exists, err := db.ImageExists(imageInfo.Hash)
	if err != nil {
		c.JSON(500, gin.H{"error": "Couldn't add photo - please try again later"})
		return
	}
	if exists {
		c.JSON(409, gin.H{"error": "Couldn't add photo - image already exists"})
		return
	}

	addSublabels := false
	temp := c.PostForm("add_sublabels")
	if temp == "true" {
		addSublabels = true
	}

	imageCollectionName := c.PostForm("image_collection")
	if !IsImageCollectionNameValid(imageCollectionName) {
		c.JSON(400, gin.H{"error": "Couldn't process request - image collection name contains unsopported characters"})
		return
	}

	var labelMeEntry datastructures.LabelMeEntry
	labelMeEntries := []datastructures.LabelMeEntry{}
	labelMeEntry.Label = label

	if label != "" {
		if commons.IsLabelValid(labelMap, metalabels, label, []datastructures.Sublabel{}) {
			labelMapEntry := labelMap[label]
			if !addSublabels {
				labelMapEntry.LabelMapEntries = nil
			}
			labelMeEntry.Annotatable = true //assume that the label that was directly provided together with the donation is annotatable
			for key, _ := range labelMapEntry.LabelMapEntries {
				labelMeEntry.Sublabels = append(labelMeEntry.Sublabels, datastructures.Sublabel{Name: key})
			}
		} else { //labels that are not already productive, can only be added if authenticated
			if username == "" {
				c.JSON(401, gin.H{"error": "Authentication required"})
				return
			}
		}
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

	var apiUser datastructures.APIUser
	apiUser.ClientFingerprint = browserFingerprint
	apiUser.Name = username

	e := db.AddDonatedPhoto(apiUser, imageInfo, autoUnlock, labelMeEntries,
		imageCollectionName, labelMap, metalabels)
	if e != nil {
		switch e.(type) {
		case *imagemonkeydb.InvalidImageCollectionInputError:
			c.JSON(404, gin.H{"error": "Couldn't add photo - image collection doesn't exist"})
			return
		default:
			c.JSON(500, gin.H{"error": "Couldn't add photo - please try again later"})
			return
		}
	}

	//get client IP address and try to determine country
	var contributionsPerCountryRequest datastructures.ContributionsPerCountryRequest
	contributionsPerCountryRequest.Type = "donation"
	contributionsPerCountryRequest.CountryCode = "--"
	ip := net.ParseIP(commons.GetIPAddress(c.Request))
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

func main() {
	fmt.Printf("Starting API Service...\n")

	log.SetLevel(log.DebugLevel)

	releaseMode := flag.Bool("release", false, "Run in release mode")
	wordlistPath := flag.String("wordlist", "../wordlists/en/labels.jsonnet", "Path to label map")
	labelRefinementsPath := flag.String("label_refinements", "../wordlists/en/label-refinements.json", "Path to label refinements")
	metalabelsPath := flag.String("metalabels", "../wordlists/en/metalabels.jsonnet", "Path to metalabels")
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
	corsAllowOrigin := flag.String("cors_allow_origin", "*", "CORS Access-Control-Allow-Origin")
	imageHuntAssetsDir := flag.String("imagehunt_assets_dir", "../img/game-assets/", "ImageHunt Game Assets Directory")

	sentryEnvironment := "api"

	flag.Parse()
	if *releaseMode {
		fmt.Printf("[Main] Starting gin in release mode!\n")
		gin.SetMode(gin.ReleaseMode)
	}

	sentryDsn := ""
	if *useSentry {
		fmt.Printf("Setting Sentry DSN\n")
		sentryDsn = commons.MustGetEnv("SENTRY_DSN")
		raven.SetEnvironment(sentryEnvironment)
		raven.SetDSN(sentryDsn)

		raven.CaptureMessage("Starting up api worker", nil)
	}

	log.Debug("[Main] Reading Label Map")
	labelRepository := commons.NewLabelRepository(*wordlistPath)
	err := labelRepository.Load()
	if err != nil {
		fmt.Printf("[Main] Couldn't read label map...terminating!")
		log.Fatal(err)
	}
	labelMap := labelRepository.GetMapping()
	words := labelRepository.GetWords()

	log.Debug("[Main] Reading Metalabel Map")
	metaLabels := commons.NewMetaLabels(*metalabelsPath)
	err = metaLabels.Load()
	if err != nil {
		fmt.Printf("[Main] Couldn't read metalabel map...terminating!")
		log.Fatal(err)
	}

	log.Debug("[Main] Reading label refinements")
	labelRefinementsMap, err := commons.GetLabelRefinementsMap(*labelRefinementsPath)
	if err != nil {
		fmt.Printf("[Main] Couldn't read label refinements: %s...terminating!", *labelRefinementsPath)
		log.Fatal(err)
	}

	imageMonkeyDbConnectionString := commons.MustGetEnv("IMAGEMONKEY_DB_CONNECTION_STRING")
	imageMonkeyDatabase := imagemonkeydb.NewImageMonkeyDatabase()
	err = imageMonkeyDatabase.Open(imageMonkeyDbConnectionString)
	if err != nil {
		log.Fatal("[Main] Couldn't ping ImageMonkey database: ", err.Error())
	}

	if *useSentry {
		imageMonkeyDatabase.InitializeSentry(sentryDsn, sentryEnvironment)
	}
	defer imageMonkeyDatabase.Close()

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

	redisConn := redisPool.Get()
	defer redisConn.Close()

	psc := redis.PubSubConn{Conn: redisConn}
	defer psc.Close()

	if err := psc.Subscribe(redis.Args{}.AddFlat([]string{"tasks"})...); err != nil {
		log.Fatal("Couldn't subscribe to topic 'tasks': ", err.Error())
	}

	done := make(chan error, 1)

	go func() {
		for {
			switch n := psc.Receive().(type) {
			case error:
				done <- n
				return
			case redis.Message:
				if n.Channel == "tasks" {
					if string(n.Data) == "reloadlabels" {
						log.Info("[Main] Reloading labels")
						err := labelRepository.Load()
						if err != nil {
							log.Error("Couldn't read label map: ", err.Error())
							raven.CaptureError(err, nil)
						}
						labelMap = labelRepository.GetMapping()
						words = labelRepository.GetWords()

						err = metaLabels.Load()
						if err != nil {
							log.Error("Couldn't read metalabels map: ", err.Error())
							raven.CaptureError(err, nil)
						}

						labelRefinementsMap, err = commons.GetLabelRefinementsMap(*labelRefinementsPath)
						if err != nil {
							log.Error("Couldn't read label refinements: ", err.Error())
							raven.CaptureError(err, nil)
						}
					} else if string(n.Data) == "reconnectdb" {
						log.Info("Reconnecting to Database")
						log.Info(string(n.Data))
						imageMonkeyDatabase.Close()
						err = imageMonkeyDatabase.Open(imageMonkeyDbConnectionString)
						if err != nil {
							raven.CaptureError(err, nil)
							log.Fatal("[Main] Couldn't ping ImageMonkey database: ", err.Error())
						}
					}
				}
			case redis.Subscription:
				switch n.Count {
				case 0:
					// Return from the goroutine when all channels are unsubscribed.
					done <- nil
					return
				}
			}
		}
	}()

	statisticsPusher := commons.NewStatisticsPusher(redisPool)
	err = statisticsPusher.Load()
	if err != nil {
		log.Fatal("[Main] Couldn't load statistics pusher: ", err.Error())
	}

	sampleExportQueries := commons.GetSampleExportQueries()

	jwtSecret := commons.MustGetEnv("JWT_SECRET")
	authTokenHandler := NewAuthTokenHandler(imageMonkeyDatabase, jwtSecret)

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
		router.Use(CorsMiddleware(*corsAllowOrigin))
		router.Use(RequestId())

		//serve images in "donations" directory with the possibility to scale images
		//before serving them
		router.GET("/v1/donation/:imageid", func(c *gin.Context) {
			params := c.Request.URL.Query()
			imageId := c.Param("imageid")

			var width int
			width = 0
			if temp, ok := params["width"]; ok {
				n, err := strconv.ParseUint(temp[0], 10, 32)
				if err == nil {
					width = int(n)
				}
			}

			var height int
			height = 0
			if temp, ok := params["height"]; ok {
				n, err := strconv.ParseUint(temp[0], 10, 32)
				if err == nil {
					height = int(n)
				}
			}

			if !IsFilenameValid(imageId) {
				c.String(404, "Invalid filename")
				return
			}

			unlocked, err := imageMonkeyDatabase.IsImageUnlocked(imageId)
			if err != nil {
				c.String(500, "Couldn't process request, please try again later")
				return
			}
			if !unlocked {
				c.String(404, "Couldn't access image, as image is still in locked mode")
				return
			}

			var imgBytes []byte
			var format string
			imageRegions := []image.Rectangle{}

			highlightAnnotations := commons.GetParamFromUrlParams(c, "highlight", "")
			highlightAnnotations, err = url.QueryUnescape(highlightAnnotations)
			if err != nil {
				c.String(422, "Couldn't process request, please provide a valid 'highlight' parameter")
				return
			}

			if highlightAnnotations != "" {
				imageRegions, err = imageMonkeyDatabase.GetBoundingBoxesForImageLabel(imageId, highlightAnnotations)
				if err != nil {
					c.String(500, "Couldn't process request, please try again later")
					return
				}
			} /* else {
				imageRegions, err = commons.GetImageRegionsFromUrlParams(c)
				if len(imageRegions) == 0 {
					imgBytes, format, err = img.ResizeImage((*donationsDir + imageId), width, height)
					if err != nil {
						log.Error("[Serving Resized Donation] Couldn't serve donation: ", err.Error())
						c.String(500, "Couldn't process request, please try again later")
						return
					}
				} else {


					var errorType commons.ExtractRoIFromImageErrorType
					imgBytes, format, errorType, err = commons.ExtractRoIFromImage((*donationsDir + imageId), imageRegion)
					if errorType == commons.ExtractRoIFromImageInternalError {
						log.Error("[Serving RoI of Donation] Couldn't serve donation: ", err.Error())
						c.String(500, "Couldn't process request, please try again later")
						return
					} else if errorType == commons.ExtractRoIFromImageInvalidRegionError {
						c.String(422, "Couldn't process request - invalid region")
					}
				}
			}*/

			if len(imageRegions) > 0 {
				imgBytes, err = img.HighlightAnnotationsInImage((*donationsDir + imageId), imageRegions, int(width), int(height))
				format = "jpg"
				if err != nil {
					log.Error("[Serving RoI of Donation] Couldn't serve donation: ", err.Error())
					c.String(500, "Couldn't process request, please try again later")
					return
				}
			} else {
				//imageRegions, err = commons.GetImageRegionsFromUrlParams(c)
				//if len(imageRegions) == 0 {
				imgBytes, format, err = img.ResizeImage((*donationsDir + imageId), width, height)
				if err != nil {
					log.Error("[Serving Resized Donation] Couldn't serve donation: ", err.Error())
					c.String(500, "Couldn't process request, please try again later")
					return
				}
				//}
			}

			if len(imageRegions) == 0 {
				//tell the CDN to cache images for 2 months and the browser to cache it for 1 week.
				//we do that only for images that are unmodified.
				c.Writer.Header().Set("Cache-Control", "public,s-maxage=5260000,max-age=604800")
			} else {
				//do not cache images that we have modified
				c.Writer.Header().Set("Cache-Control", "no-cache")
			}

			c.Writer.Header().Set("Content-Type", ("image/" + format))
			c.Writer.Header().Set("Content-Length", strconv.Itoa(len(imgBytes)))
			_, err = c.Writer.Write(imgBytes)
			if err != nil {
				log.Error("[Serving Donation] Couldn't serve donation: ", err.Error())
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

			var width int
			width = 0
			if temp, ok := params["width"]; ok {
				n, err := strconv.ParseUint(temp[0], 10, 32)
				if err == nil {
					width = int(n)
				}
			}

			var height int
			height = 0
			if temp, ok := params["height"]; ok {
				n, err := strconv.ParseUint(temp[0], 10, 32)
				if err == nil {
					height = int(n)
				}
			}

			if !IsFilenameValid(imageId) {
				c.String(404, "Invalid filename")
				return
			}

			isOwnDonation, err := imageMonkeyDatabase.IsOwnDonation(imageId, apiUser.Name)
			if err != nil {
				c.String(500, "Couldn't process request, please try again later")
				return
			}
			if !isOwnDonation {
				//check if user has unlock permission
				userInfo, err := imageMonkeyDatabase.GetUserInfo(apiUser.Name)
				if err != nil {
					c.String(500, "Couldn't process request, please try again later")
					return
				}

				if userInfo.Permissions == nil || !userInfo.Permissions.CanUnlockImage {
					c.String(403, "You do not have the appropriate permissions to access the image")
					return
				}
			}

			imgBytes, format, err := img.ResizeImage((*unverifiedDonationsDir + imageId), width, height)
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

		router.GET("/v1/unverified-donation/:imageid/annotations", func(c *gin.Context) {
			imageId := c.Param("imageid")

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfoFromUrl(c).Username

			isOwnDonation, err := imageMonkeyDatabase.IsOwnDonation(imageId, apiUser.Name)
			if err != nil {
				c.String(500, "Couldn't process request, please try again later")
				return
			}

			if !isOwnDonation {
				c.String(403, "You do not have the appropriate permissions to access the image")
				return
			}

			handleImageAnnotationsRequest(c, imageId, imageMonkeyDatabase, authTokenHandler, *apiBaseUrl)
		})

		clientId := commons.MustGetEnv("X_CLIENT_ID")
		clientSecret := commons.MustGetEnv("X_CLIENT_SECRET")

		//the following endpoints are secured with a client id + client secret.
		//that's mostly because currently each donation needs to be unlocked manually.
		//(as we want to make sure that we don't accidentally host inappropriate content, like nudity)
		clientAuth := router.Group("/")
		clientAuth.Use(RequestId())
		clientAuth.Use(ClientAuthMiddleware(clientId, clientSecret))
		{
			clientAuth.Static("./v1/unverified/donation", *unverifiedDonationsDir)
			clientAuth.GET("/v1/internal/unverified-donations", func(c *gin.Context) {

				imageProvider := commons.GetParamFromUrlParams(c, "image_provider", "")
				shuffle := commons.GetParamFromUrlParams(c, "shuffle", "")
				orderRandomly := false
				if shuffle == "true" {
					orderRandomly = true
				}

				limitBy := -1
				limit := commons.GetParamFromUrlParams(c, "limit", "")
				if limit != "" {
					limitBy, err = strconv.Atoi(limit)
					if err != nil {
						c.JSON(422, gin.H{"error": "Invalid request - please provide a valid limit"})
						return
					}
				}

				images, err := imageMonkeyDatabase.GetAllUnverifiedImages(imageProvider, orderRandomly, limitBy)
				if err != nil {
					c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				} else {
					c.JSON(http.StatusOK, images)
				}
			})

			clientAuth.GET("/v1/internal/statistics/pg", func(c *gin.Context) {
				var apiUser datastructures.APIUser
				apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

				hasPermissions := false
				if apiUser.Name != "" {
					userInfo, err := imageMonkeyDatabase.GetUserInfo(apiUser.Name)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't process request - please try again later"})
						return
					}

					if userInfo.Permissions != nil && userInfo.Permissions.CanAccessPgStat {
						hasPermissions = true
					}
				}


				if !hasPermissions {
					c.JSON(403, gin.H{"error": "You do not have the appropriate permissions to access this information"})
					return
				}

				res, err := imageMonkeyDatabase.GetPgStatStatements()
				if err != nil {
					c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
					return
				}

				c.JSON(200, res)
			})

			clientAuth.POST("/v1/internal/labelme/donate", func(c *gin.Context) {
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
				donate(c, imageMonkeyDatabase, "", imageSource, labelMap, metaLabels, dir, redisPool,
					statisticsPusher, geoipDb, autoUnlock)
			})

			clientAuth.POST("/v1/internal/auto-annotate/:imageid", func(c *gin.Context) {
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

				annotationsValidator := commons.NewAnnotationsValidator(annotations.Annotations)
				err = annotationsValidator.Parse()
				if err != nil {
					c.JSON(422, gin.H{"error": "invalid request - annotations invalid"})
					return
				}

				isSuggestion := false
				if !labelRepository.Contains(annotations.Label, annotations.Sublabel) {
					isSuggestion = true
				}

				var apiUser datastructures.APIUser
				apiUser.ClientFingerprint = ""
				apiUser.Name = ""

				var annos []datastructures.AnnotationsContainer
				annos = append(annos, datastructures.AnnotationsContainer{
					Annotations:        annotations,
					AllowedRefinements: annotationsValidator.GetRefinements(),
					AutoGenerated:      true,
					IsSuggestion:       isSuggestion,
				})

				_, err = imageMonkeyDatabase.AddAnnotations(apiUser, imageId, annos)
				if err != nil {
					switch err.(type) {
					case *imagemonkeydb.AuthenticationRequiredError:
						c.JSON(401, gin.H{"error": "Couldn't add annotations - you need to be authenticated to perform this action"})
						return
					default:
						c.JSON(500, gin.H{"error": "Couldn't add annotations - please try again later"})
						return
					}
				}
				c.JSON(201, nil)
			})

			clientAuth.GET("/v1/internal/auto-annotation", func(c *gin.Context) {
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

				images, err := imageMonkeyDatabase.GetImagesForAutoAnnotation(labels)
				if err != nil {
					c.JSON(500, gin.H{"error": "Couldn't get images - please try again later"})
					return
				}

				c.JSON(http.StatusOK, images)
			})

			clientAuth.POST("/v1/unverified/donation/:imageid/:param", func(c *gin.Context) {
				imageId := c.Param("imageid")
				action := c.Param("param")

				retCode, err := handleUnverifiedDonation(imageId, action, *donationsDir,
					*unverifiedDonationsDir, *imageQuarantineDir, imageMonkeyDatabase)
				if err != nil {
					c.JSON(retCode, gin.H{"error": err})
					return
				}
				c.JSON(retCode, nil)
			})

			clientAuth.PATCH("/v1/unverified/donation", func(c *gin.Context) {
				var lockedImages []datastructures.LockedImageAction

				if c.BindJSON(&lockedImages) != nil {
					c.JSON(422, gin.H{"error": "Couldn't process request - invalid request"})
					return
				}

				for _, val := range lockedImages {
					retCode, err := handleUnverifiedDonation(val.ImageId, val.Action, *donationsDir,
						*unverifiedDonationsDir, *imageQuarantineDir, imageMonkeyDatabase)
					if err != nil {
						c.JSON(retCode, gin.H{"error": err})
						return
					}
				}
				c.JSON(204, nil)
			})
		}

		router.GET("/v1/validation", func(c *gin.Context) {
			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			validationId := commons.GetParamFromUrlParams(c, "validation_id", "")

			image, err := imageMonkeyDatabase.GetImageToValidate(validationId, apiUser.Name)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if image.Id == "" {
				c.JSON(422, gin.H{"error": "Couldn't process request - empty result set"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"image": gin.H{"uuid": image.Id,
				"provider": image.Provider,
				"unlocked": image.Unlocked,
				"url":      commons.GetImageUrlFromImageId(*apiBaseUrl, image.Id, image.Unlocked),
			},
				"label": image.Label, "sublabel": image.Sublabel, "num_yes": image.Validation.NumOfValid,
				"num_no": image.Validation.NumOfInvalid, "uuid": image.Validation.Id})
		})

		router.GET("/v1/validation/:validationid", func(c *gin.Context) {
			validationId := c.Param("validationid")

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			image, err := imageMonkeyDatabase.GetImageToValidate(validationId, apiUser.Name)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if image.Id == "" {
				c.JSON(422, gin.H{"error": "Couldn't process request - empty result set"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"image": gin.H{"uuid": image.Id,
				"provider": image.Provider,
				"unlocked": image.Unlocked,
				"url":      commons.GetImageUrlFromImageId(*apiBaseUrl, image.Id, image.Unlocked),
			},
				"label": image.Label, "sublabel": image.Sublabel, "num_yes": image.Validation.NumOfValid,
				"num_no": image.Validation.NumOfInvalid, "uuid": image.Validation.Id})
		})

		router.POST("/v1/donation/:imageid/description", func(c *gin.Context) {
			imageId := c.Param("imageid")

			var descriptions []datastructures.ImageDescription
			if c.BindJSON(&descriptions) != nil {
				c.JSON(400, gin.H{"error": "Couldn't process request - description missing"})
				return
			}

			err := imageMonkeyDatabase.AddImageDescriptions(imageId, descriptions)
			if err == imagemonkeydb.AddImageDescriptionInternalError {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			} else if err == imagemonkeydb.AddImageDescriptionInvalidLanguage {
				c.JSON(400, gin.H{"error": "Couldn't process request - invalid language"})
				return
			} else if err == imagemonkeydb.AddImageDescriptionInvalidImageDescription {
				c.JSON(400, gin.H{"error": "Couldn't process request - invalid image description"})
				return
			}

			//get client IP address and try to determine country
			contributionsPerCountryRequest := datastructures.ContributionsPerCountryRequest{Type: "image-description",
				CountryCode: "--"}
			ip := net.ParseIP(commons.GetIPAddress(c.Request))
			if ip != nil {
				record, err := geoipDb.Country(ip)
				if err != nil { //just log, but don't abort...it's just for statistics
					log.Debug("[Donation] Couldn't determine geolocation from ", err.Error())

				} else {
					contributionsPerCountryRequest.CountryCode = record.Country.IsoCode
				}
			}

			for _ = range descriptions {
				pushCountryContributionToRedis(redisPool, contributionsPerCountryRequest)
			}

			c.JSON(201, nil)
		})

		router.POST("/v1/donation/:imageid/description/:descriptionid/unlock", func(c *gin.Context) {
			imageId := c.Param("imageid")
			descriptionId := c.Param("descriptionid")

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			hasModeratorPermissions := false
			if isModerationRequest(c) {
				if apiUser.Name != "" {
					userInfo, err := imageMonkeyDatabase.GetUserInfo(apiUser.Name)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't process request - please try again later"})
						return
					}

					if userInfo.Permissions != nil && userInfo.Permissions.CanUnlockImageDescription {
						hasModeratorPermissions = true
					}
				}
			}

			if !hasModeratorPermissions {
				c.JSON(401, gin.H{"error": "Authentication required"})
				return
			}

			errCode := imageMonkeyDatabase.UnlockImageDescription(apiUser, imageId, descriptionId)
			if errCode == imagemonkeydb.UnlockImageDescriptionInternalError {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			} else if errCode == imagemonkeydb.UnlockImageDescriptionInvalidId {
				c.JSON(404, gin.H{"error": "Couldn't process request - resource doesn't exist"})
				return
			}
			c.JSON(201, nil)
		})

		router.POST("/v1/donation/:imageid/description/:descriptionid/lock", func(c *gin.Context) {
			imageId := c.Param("imageid")
			descriptionId := c.Param("descriptionid")

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			hasModeratorPermissions := false
			if isModerationRequest(c) {
				if apiUser.Name != "" {
					userInfo, err := imageMonkeyDatabase.GetUserInfo(apiUser.Name)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't process request - please try again later"})
						return
					}

					if userInfo.Permissions != nil && userInfo.Permissions.CanUnlockImageDescription {
						hasModeratorPermissions = true
					}
				}
			}

			if !hasModeratorPermissions {
				c.JSON(401, gin.H{"error": "Authentication required"})
				return
			}

			errCode := imageMonkeyDatabase.LockImageDescription(apiUser, imageId, descriptionId)
			if errCode == imagemonkeydb.LockImageDescriptionInternalError {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			} else if errCode == imagemonkeydb.LockImageDescriptionInvalidId {
				c.JSON(404, gin.H{"error": "Couldn't process request - resource doesn't exist"})
				return
			}
			c.JSON(201, nil)
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

			err := imageMonkeyDatabase.AddLabelsToImage(apiUser, labelMap, metaLabels, imageId, labels)
			if err != nil {
				switch err.(type) {
				case *imagemonkeydb.AuthenticationRequiredError:
					c.JSON(401, gin.H{"error": "Couldn't process request - you need to be authenticated to perform this action"})
					return
				default:
					c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
					return
				}
			}
			c.JSON(200, nil)
		})

		router.GET("/v1/donation/:imageid/labels", func(c *gin.Context) {
			imageId := c.Param("imageid")

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			includeOnlyUnlockedLabels := false
			temp := commons.GetParamFromUrlParams(c, "only_unlocked_labels", "")
			if temp == "true" {
				includeOnlyUnlockedLabels = true
			}

			img, err := imageMonkeyDatabase.GetImageToLabel(imageId, apiUser.Name, includeOnlyUnlockedLabels)
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

			ids, err := imageMonkeyDatabase.GetUnannotatedValidations(apiUser, imageId)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(http.StatusOK, ids)
		})

		router.GET("/v1/donation/:imageid/annotations", func(c *gin.Context) {
			imageId := c.Param("imageid")
			handleImageAnnotationsRequest(c, imageId, imageMonkeyDatabase, authTokenHandler, *apiBaseUrl)
		})

		router.GET("/v1/donation/:imageid/annotations/coverage", func(c *gin.Context) {
			imageId := c.Param("imageid")

			imageCoverages, err := imageMonkeyDatabase.GetAnnotationCoverage(imageId)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if len(imageCoverages) == 0 {
				c.JSON(422, gin.H{"error": "Couldn't process request - missing result set"})
				return
			}

			c.JSON(200, imageCoverages[0])
		})

		router.POST("/v1/validation/:validationid/validate/:param", func(c *gin.Context) {
			validationId := c.Param("validationid")
			param := c.Param("param")

			if param != "yes" && param != "no" {
				c.JSON(404, nil)
				return
			}

			var imageValidationBatch datastructures.ImageValidationBatch
			imageValidationBatch.Validations = append(imageValidationBatch.Validations, datastructures.ImageValidation{Uuid: validationId, Valid: param})

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			moderatorAction := false
			if isModerationRequest(c) {
				if apiUser.Name != "" {
					userInfo, err := imageMonkeyDatabase.GetUserInfo(apiUser.Name)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't process request - please try again later"})
						return
					}

					if userInfo.Permissions != nil && userInfo.Permissions.CanRemoveLabel {
						moderatorAction = true
					}
				}
			}

			err := imageMonkeyDatabase.ValidateImages(apiUser, imageValidationBatch, moderatorAction)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			//get client IP address and try to determine country
			var contributionsPerCountryRequest datastructures.ContributionsPerCountryRequest
			contributionsPerCountryRequest.Type = "validation"
			contributionsPerCountryRequest.CountryCode = "--"
			ip := net.ParseIP(commons.GetIPAddress(c.Request))
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

			imageId := commons.GetParamFromUrlParams(c, "image_id", "")

			image, err := imageMonkeyDatabase.GetImageToLabel(imageId, apiUser.Name, false)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if image.Id == "" {
				c.JSON(422, gin.H{"error": "Couldn't process request - empty result set"})
			} else {
				imageUrl := commons.GetImageUrlFromImageId(*apiBaseUrl, image.Id, image.Unlocked)

				c.JSON(http.StatusOK, gin.H{"image": gin.H{"uuid": image.Id, "provider": image.Provider,
					"url": imageUrl, "unlocked": image.Unlocked,
					"width": image.Width, "height": image.Height,
					"descriptions": image.ImageDescriptions},
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

			err = imageMonkeyDatabase.BatchAnnotationRefinement(annotationRefinements, apiUser)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			//get client IP address and try to determine country
			contributionsPerCountryRequest := datastructures.ContributionsPerCountryRequest{Type: "annotation-refinement",
				CountryCode: "--"}
			ip := net.ParseIP(commons.GetIPAddress(c.Request))
			if ip != nil {
				record, err := geoipDb.Country(ip)
				if err != nil { //just log, but don't abort...it's just for statistics
					log.Debug("[Donation] Couldn't determine geolocation from ", err.Error())

				} else {
					contributionsPerCountryRequest.CountryCode = record.Country.IsoCode
				}
			}
			pushCountryContributionToRedis(redisPool, contributionsPerCountryRequest)

			c.JSON(204, nil)
		})

		router.HEAD("/v1/donations/unprocessed-descriptions", func(c *gin.Context) {
			if values, _ := c.Request.Header["X-Total-Count"]; len(values) > 0 {
				num, err := imageMonkeyDatabase.GetNumOfUnprocessedDescriptions()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't process request - please try again later"})
					return
				}
				c.Writer.Header().Set("X-Total-Count", strconv.Itoa(num))
				c.Writer.Header().Set("Access-Control-Expose-Headers", "X-Total-Count")
				return
			}
		})

		router.GET("/v1/donations/unprocessed-descriptions", func(c *gin.Context) {
			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			hasModeratorPermissions := false
			if isModerationRequest(c) {
				if apiUser.Name != "" {
					userInfo, err := imageMonkeyDatabase.GetUserInfo(apiUser.Name)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't process request - please try again later"})
						return
					}

					if userInfo.Permissions != nil && userInfo.Permissions.CanUnlockImageDescription {
						hasModeratorPermissions = true
					}
				}
			}

			if !hasModeratorPermissions {
				c.JSON(401, gin.H{"error": "Authentication required"})
				return
			}

			descriptionsPerImage, err := imageMonkeyDatabase.GetUnprocessedImageDescriptions()
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if len(descriptionsPerImage) == 0 {
				c.JSON(200, make([]string, 0))
				return
			}

			c.JSON(200, descriptionsPerImage)
		})

		router.GET("/v1/donations/labels", func(c *gin.Context) {
			query := commons.GetParamFromUrlParams(c, "query", "")

			orderRandomly := false
			shuffle := commons.GetParamFromUrlParams(c, "shuffle", "")
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

			queryParser := parser.NewQueryParser(query)
			queryParser.AllowImageWidth(true)
			queryParser.AllowImageHeight(true)
			queryParser.AllowAnnotationCoverage(true)
			queryParser.AllowImageCollection(true)
			queryParser.AllowImageHasLabels(true)
			parseResult, err := queryParser.Parse()
			if err != nil {
				c.JSON(422, gin.H{"error": err.Error()})
				return
			}

			imageInfos, err := imageMonkeyDatabase.GetImagesLabels(apiUser, parseResult, *apiBaseUrl, orderRandomly)
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

			err := imageMonkeyDatabase.ValidateImages(apiUser, imageValidationBatch, false)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			//get client IP address and try to determine country
			var contributionsPerCountryRequest datastructures.ContributionsPerCountryRequest
			contributionsPerCountryRequest.Type = "validation"
			contributionsPerCountryRequest.CountryCode = "--"
			ip := net.ParseIP(commons.GetIPAddress(c.Request))
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

			c.JSON(204, nil)
		})

		router.GET("/v1/export", func(c *gin.Context) {
			query, annotationsOnly, err := commons.GetExploreUrlParams(c)
			if err != nil {
				c.JSON(422, gin.H{"error": err.Error()})
				return
			}

			queryParser := parser.NewQueryParser(query)
			queryParser.SetVersion(1)
			parseResult, err := queryParser.Parse()
			if err != nil {
				c.JSON(422, gin.H{"error": err.Error()})
				return
			}

			jsonData, err := imageMonkeyDatabase.Export(parseResult, annotationsOnly)
			if err == nil {
				c.JSON(http.StatusOK, jsonData)
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't export data, please try again later."})
			return
		})

		router.GET("/v1/export/sample", func(c *gin.Context) {
			q := sampleExportQueries[commons.Random(0, len(sampleExportQueries)-1)]
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
			label := words[commons.Random(0, len(words)-1)]
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
			err = imageMonkeyDatabase.AddLabelSuggestion(escapedLabel)
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
			popularLabels, err := imageMonkeyDatabase.GetMostPopularLabels(10) //limit to 10
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(http.StatusOK, popularLabels)
		})

		router.GET("/v1/label/accessors", func(c *gin.Context) {
			detailed := commons.GetParamFromUrlParams(c, "detailed", "")

			if detailed == "true" {
				labelType := commons.GetParamFromUrlParams(c, "label_type", "")
				if labelType != "" && labelType != "normal" && labelType != "refinement" && labelType != "refinement_category" {
					c.JSON(422, gin.H{"error": "Couldn't process request - invalid 'label_type'"})
					return
				}

				labelAccessors, err := imageMonkeyDatabase.GetLabelAccessorDetails(labelType)
				if err != nil {
					c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
					return
				}
				c.JSON(http.StatusOK, labelAccessors)
				return
			}

			labelAccessors, err := imageMonkeyDatabase.GetLabelAccessors()
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(http.StatusOK, labelAccessors)
		})

		router.GET("/v1/label/plurals", func(c *gin.Context) {
			pluralsMapping := labelRepository.GetPluralsMapping()
			c.JSON(200, pluralsMapping)
		})

		router.GET("/v1/label/refinements", func(c *gin.Context) {
			c.JSON(http.StatusOK, labelRefinementsMap)
		})

		router.GET("/v1/label/suggestions", func(c *gin.Context) {
			includeUnlockedStr := commons.GetParamFromUrlParams(c, "include_unlocked", "true")
			includeUnlocked := true
			if includeUnlockedStr == "false" {
				includeUnlocked = false
			}
			
			labelSuggestions, err := imageMonkeyDatabase.GetLabelSuggestions(includeUnlocked)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(http.StatusOK, labelSuggestions)
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

			c.JSON(http.StatusOK, gin.H{"graph": labelGraphJson, "metadata": labelGraph.GetMetadata()})
		})

		router.GET("/v1/label/graph/:labelgraphname/definition", func(c *gin.Context) {
			labelGraphName := c.Param("labelgraphname")

			labelGraph, err := labelGraphRepository.Get(labelGraphName)
			if err != nil {
				c.JSON(422, gin.H{"error": "Couldn't process request - invalid label graph name"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"definition": labelGraph.GetDefinition()})
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

			donate(c, imageMonkeyDatabase, username, imageSource, labelMap, metaLabels,
				*unverifiedDonationsDir, redisPool, statisticsPusher, geoipDb, false)
		})

		router.POST("/v1/report/:imageid", func(c *gin.Context) {
			imageId := c.Param("imageid")

			var report datastructures.Report
			if c.BindJSON(&report) != nil {
				c.JSON(422, gin.H{"error": "reason missing - please provide a valid 'reason'"})
				return
			}

			s := "Someone reported a violation (uuid: " + imageId + ", reason: " + report.Reason + ")"
			raven.CaptureMessage(s, nil)

			err := imageMonkeyDatabase.ReportImage(imageId, report.Reason)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't report image - please try again later"})
				return
			}
			c.JSON(http.StatusOK, nil)
		})

		router.GET("/v1/validations", func(c *gin.Context) {
			query := commons.GetParamFromUrlParams(c, "query", "")
			if query != "" {
				query, err = url.QueryUnescape(query)
				if err != nil {
					c.JSON(422, gin.H{"error": "please provide a valid query"})
					return
				}

				queryParser := parser.NewQueryParser(query)
				queryParser.AllowImageHeight(true)
				queryParser.AllowImageWidth(true)
				queryParser.AllowAnnotationCoverage(true)
				parseResult, err := queryParser.Parse()
				if err != nil {
					c.JSON(422, gin.H{"error": err.Error()})
					return
				}

				orderRandomly := false
				shuffle := commons.GetParamFromUrlParams(c, "shuffle", "")
				if shuffle == "true" {
					orderRandomly = true
				}

				var apiUser datastructures.APIUser
				apiUser.ClientFingerprint = getBrowserFingerprint(c)
				apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

				if len(parseResult.QueryValues) == 0 {
					c.JSON(422, gin.H{"error": "Couldn't process request - please provide a valid query!"})
					return
				}

				validations, err := imageMonkeyDatabase.GetImagesForValidation(apiUser, parseResult, orderRandomly, *apiBaseUrl)
				if err != nil {
					c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
					return
				}

				c.JSON(http.StatusOK, validations)
				return
			}

			c.JSON(422, gin.H{"error": "please provide a valid query"})
		})

		router.GET("/v1/validations/unannotated", func(c *gin.Context) {
			query := commons.GetParamFromUrlParams(c, "query", "")
			if query != "" {
				query, err = url.QueryUnescape(query)
				if err != nil {
					c.JSON(422, gin.H{"error": "please provide a valid query"})
					return
				}

				queryParser := parser.NewQueryParser(query)
				queryParser.AllowImageWidth(true)
				queryParser.AllowImageHeight(true)
				queryParser.AllowAnnotationCoverage(true)
				queryParser.AllowImageCollection(true)
				parseResult, err := queryParser.Parse()
				if err != nil {
					c.JSON(422, gin.H{"error": err.Error()})
					return
				}

				orderRandomly := false
				shuffle := commons.GetParamFromUrlParams(c, "shuffle", "")
				if shuffle == "true" {
					orderRandomly = true
				}

				var apiUser datastructures.APIUser
				apiUser.ClientFingerprint = getBrowserFingerprint(c)
				apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

				if len(parseResult.QueryValues) == 0 {
					c.JSON(422, gin.H{"error": "Couldn't process request - please provide a valid query!"})
					return
				}
				
				includeLabelSuggestions := false
				if apiUser.Name != "" {
					includeLabelSuggestions = true
				}

				annotationTasks, err := imageMonkeyDatabase.GetAvailableAnnotationTasks(apiUser, parseResult, orderRandomly, *apiBaseUrl, includeLabelSuggestions)
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

			err := imageMonkeyDatabase.BlacklistForAnnotation(validationId, apiUser)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't blacklist annotation - please try again later"})
				return
			}

			c.JSON(http.StatusOK, nil)
		})

		router.POST("/v1/validation/:validationid/not-annotatable", func(c *gin.Context) {
			validationId := c.Param("validationid")

			err := imageMonkeyDatabase.MarkValidationAsNotAnnotatable(validationId)
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

			annotationsValidator := commons.NewAnnotationsValidator(annotations.Annotations)
			err = annotationsValidator.Parse()
			if err != nil {
				c.JSON(422, gin.H{"error": "invalid request - annotations invalid"})
				return
			}

			isSuggestion, err := imageMonkeyDatabase.AnnotationUuidIsASuggestion(annotationId)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't add annotation refinement - please try again later"})
				return
			}

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			annotationsContainer := datastructures.AnnotationsContainer{
				Annotations:        annotations,
				AllowedRefinements: annotationsValidator.GetRefinements(),
				AutoGenerated:      false,
				IsSuggestion:       isSuggestion,
			}
			err = imageMonkeyDatabase.UpdateAnnotation(apiUser, annotationId, annotationsContainer)
			if err != nil {
				switch err.(type) {
				case *imagemonkeydb.AuthenticationRequiredError:
					c.JSON(401, gin.H{"error": "Couldn't update annotation - you need to be authenticated to perform this action"})
					return
				default:
					c.JSON(500, gin.H{"error": "Couldn't update annotation - please try again later"})
					return
				}
			}

			//get client IP address and try to determine country
			var contributionsPerCountryRequest datastructures.ContributionsPerCountryRequest
			contributionsPerCountryRequest.Type = "annotation"
			contributionsPerCountryRequest.CountryCode = "--"
			ip := net.ParseIP(commons.GetIPAddress(c.Request))
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

			pushAnnotationCoverageUpdateRequestToRedis(redisPool, annotationId, "annotation")

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
			} else {
				c.JSON(404, nil)
				return
			}

			var labelValidationEntry datastructures.LabelValidationEntry
			if c.BindJSON(&labelValidationEntry) != nil {
				c.JSON(400, gin.H{"error": "Couldn't process request - please provide valid label(s)"})
				return
			}

			if (labelValidationEntry.Label == "") && (labelValidationEntry.Sublabel == "") {
				c.JSON(400, gin.H{"error": "Please provide a valid label"})
				return
			}

			browserFingerprint := getBrowserFingerprint(c)

			err := imageMonkeyDatabase.ValidateAnnotatedImage(browserFingerprint, annotationId, labelValidationEntry, parameter)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"Error": "Database Error: Couldn't update data"})
				return
			}

			//get client IP address and try to determine country
			var contributionsPerCountryRequest datastructures.ContributionsPerCountryRequest
			contributionsPerCountryRequest.Type = "validation"
			contributionsPerCountryRequest.CountryCode = "--"
			ip := net.ParseIP(commons.GetIPAddress(c.Request))
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

		router.POST("/v1/donation/:imageid/annotate", func(c *gin.Context) {
			imageId := c.Param("imageid")
			if imageId == "" {
				c.JSON(422, gin.H{"error": "invalid request - image id missing"})
				return
			}

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			var annotations []datastructures.Annotations
			err := c.BindJSON(&annotations)
			if err != nil {
				c.JSON(422, gin.H{"error": "invalid request - annotations missing"})
				return
			}

			var annos []datastructures.AnnotationsContainer
			for _, annotation := range annotations {

				annotationsValidator := commons.NewAnnotationsValidator(annotation.Annotations)
				err = annotationsValidator.Parse()
				if err != nil {
					c.JSON(422, gin.H{"error": "invalid request - annotations invalid"})
					return
				}

				if annotation.Sublabel == "" && metaLabels.Contains(annotation.Label) {
					c.JSON(400, gin.H{"error": "Couldn't add annotations - it's not possible to annotate metalabels"})
					return
				}

				isSuggestion := false
				if !labelRepository.Contains(annotation.Label, annotation.Sublabel) {
					isSuggestion = true
				}

				annos = append(annos, datastructures.AnnotationsContainer{
					Annotations:        annotation,
					AllowedRefinements: annotationsValidator.GetRefinements(),
					AutoGenerated:      false,
					IsSuggestion:		isSuggestion,
				})
			}
			annotationIds, err := imageMonkeyDatabase.AddAnnotations(apiUser, imageId, annos)
			if err != nil {
				switch err.(type) {
				case *imagemonkeydb.AuthenticationRequiredError:
					c.JSON(401, gin.H{"error": "Couldn't add annotations - you need to be authenticated to perform this action"})
					return
				default:
					c.JSON(500, gin.H{"error": "Couldn't add annotations - please try again later"})
					return
				}
			}

			//get client IP address and try to determine country
			var contributionsPerCountryRequest datastructures.ContributionsPerCountryRequest
			contributionsPerCountryRequest.Type = "annotation"
			contributionsPerCountryRequest.CountryCode = "--"
			ip := net.ParseIP(commons.GetIPAddress(c.Request))
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

			pushAnnotationCoverageUpdateRequestToRedis(redisPool, imageId, "image")

			if len(annotationIds) == 1 {
				c.Writer.Header().Set("Location", (*apiBaseUrl + "v1/annotation?annotation_id=" + annotationIds[0]))
				c.JSON(201, nil)
			}
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

			labelId, err := commons.GetLabelIdFromUrlParams(params)
			if err != nil {
				c.JSON(422, gin.H{"error": "label id needs to be an integer"})
				return
			}

			validationId := commons.GetValidationIdFromUrlParams(params)

			img, err := imageMonkeyDatabase.GetImageForAnnotation(apiUser.Name, addAutoAnnotations, validationId, labelId)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if img.Id == "" {
				c.JSON(422, gin.H{"error": "Couldn't process request - missing result set"})
				return
			}

			img.Url = commons.GetImageUrlFromImageId(*apiBaseUrl, img.Id, img.Unlocked)

			c.JSON(http.StatusOK, img)
		})

		router.GET("/v1/annotations", func(c *gin.Context) {
			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			query := commons.GetParamFromUrlParams(c, "query", "")
			query, err = url.QueryUnescape(query)
			if err != nil {
				c.JSON(422, gin.H{"error": "Please provide a valid query"})
				return
			}

			if query == "" {
				c.JSON(422, gin.H{"error": "Please provide a valid query"})
				return
			}

			queryParser := parser.NewQueryParser(query)
			queryParser.SetVersion(1)
			queryParser.AllowImageHeight(true)
			queryParser.AllowImageWidth(true)
			queryParser.AllowAnnotationCoverage(true)
			queryParser.AllowImageCollection(true)
			parseResult, err := queryParser.Parse()
			if err != nil {
				c.JSON(422, gin.H{"error": err.Error()})
				return
			}

			annotatedImages, err := imageMonkeyDatabase.GetAnnotations(apiUser, parseResult, "", *apiBaseUrl)
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

			autoGenerated := false
			if temp, ok := params["auto_generated"]; ok {
				if temp[0] == "true" {
					autoGenerated = true
				}
			}

			annotationId := commons.GetParamFromUrlParams(c, "annotation_id", "")

			rev := commons.GetParamFromUrlParams(c, "rev", "-1")
			revision, err := strconv.ParseInt(rev, 10, 32)
			if err != nil {
				c.JSON(422, gin.H{"error": "Invalid request - please provide a valid revision"})
				return
			}

			annotatedImage, err := imageMonkeyDatabase.GetAnnotatedImage(apiUser, annotationId, autoGenerated, int32(revision))
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if annotatedImage.Id == "" {
				c.JSON(422, gin.H{"error": "Couldn't process request - missing result set"})
				return
			}

			annotatedImage.Image.Url = commons.GetImageUrlFromImageId(*apiBaseUrl, annotatedImage.Image.Id, annotatedImage.Image.Unlocked)

			c.JSON(http.StatusOK, annotatedImage)
		})

		router.GET("/v1/quiz-refine", func(c *gin.Context) {
			randomImage, err := imageMonkeyDatabase.GetRandomAnnotationForQuizRefinement()
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
			annotationDataId := commons.GetParamFromUrlParams(c, "annotation_data_id", "")

			var parseResult parser.ParseResult
			query := commons.GetParamFromUrlParams(c, "query", "")
			if query != "" {
				query, err = url.QueryUnescape(query)
				if err != nil {
					c.JSON(422, gin.H{"error": "invalid query"})
					return
				}

				queryParser := parser.NewQueryParser(query)
				parseResult, err = queryParser.Parse()
				if err != nil {
					c.JSON(422, gin.H{"error": err.Error()})
					return
				}
			}

			if query == "" && annotationDataId == "" {
				c.JSON(422, gin.H{"error": "Couldn't process request - invalid request"})
				return
			}

			annotations, err := imageMonkeyDatabase.GetAnnotationsForRefinement(parseResult, *apiBaseUrl, annotationDataId)
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
				c.JSON(422, gin.H{"error": "Couldn't process request - please provide a valid annotation id"})
				return
			}

			_, err := uuid.FromString(annotationId)
			if err != nil {
				c.JSON(422, gin.H{"error": "Couldn't process request - please provide a valid annotation id"})
				return
			}

			isSuggestion, err := imageMonkeyDatabase.AnnotationUuidIsASuggestion(annotationId)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't add annotation refinement - please try again later"})
				return
			}

			annotationDataId := c.Param("annotationdataid")
			if annotationDataId == "" {
				c.JSON(422, gin.H{"error": "Couldn't process request - please provide a valid annotation data id"})
				return
			}

			var annotationRefinementEntries []datastructures.AnnotationRefinementEntry
			if c.BindJSON(&annotationRefinementEntries) != nil {
				c.JSON(400, gin.H{"error": "Couldn't process request - please provide a valid label id"})
				return
			}

			browserFingerprint := getBrowserFingerprint(c)
			err = imageMonkeyDatabase.AddOrUpdateRefinements(annotationId, annotationDataId, annotationRefinementEntries, browserFingerprint, isSuggestion)
			if err != nil {
				switch err.(type) {
				case *imagemonkeydb.InvalidLabelIdError:
					c.JSON(400, gin.H{"error": "Couldn't add annotation refinement - please provide a valid label id"})
					return
				default:
					c.JSON(500, gin.H{"error": "Couldn't add annotation refinement - please try again later"})
					return
				}
			}

			//get client IP address and try to determine country
			contributionsPerCountryRequest := datastructures.ContributionsPerCountryRequest{Type: "annotation-refinement",
				CountryCode: "--"}
			ip := net.ParseIP(commons.GetIPAddress(c.Request))
			if ip != nil {
				record, err := geoipDb.Country(ip)
				if err != nil { //just log, but don't abort...it's just for statistics
					log.Debug("[Donation] Couldn't determine geolocation from ", err.Error())

				} else {
					contributionsPerCountryRequest.CountryCode = record.Country.IsoCode
				}
			}
			pushCountryContributionToRedis(redisPool, contributionsPerCountryRequest)

			c.JSON(201, nil)
		})

		router.GET("/v1/trendinglabels", func(c *gin.Context) {
			trendingLabels, err := imageMonkeyDatabase.GetTrendingLabels()
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, trendingLabels)
		})

		router.POST("/v1/trendinglabels/:trendinglabel/accept", func(c *gin.Context) {
			trendingLabel := c.Param("trendinglabel")
			if trendingLabel == "" {
				c.JSON(400, gin.H{"error": "Couldn't process request - please provide a valid trending label"})
				return
			}

			var labelDetails datastructures.AcceptTrendingLabel
			err := c.BindJSON(&labelDetails)
			if err != nil {
				c.JSON(400, gin.H{"error": "Couldn't process request - please provide a label type"})
				return
			}
			if (labelDetails.Label.Type != "normal") && (labelDetails.Label.Type != "meta") {
				c.JSON(400, gin.H{"error": "Couldn't process request - please provide a label type"})
				return
			}

			if labelDetails.Label.RenameTo == "" {
				c.JSON(400, gin.H{"error": "Couldn't process request - rename_to cannot be empty!"})
				return
			}

			if labelDetails.Label.Type == "normal" && labelDetails.Label.Plural == "" {
				c.JSON(400, gin.H{"error": "Couldn't process request - plural cannot be empty!"})
				return
			}

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			if apiUser.Name == "" {
				c.JSON(401, gin.H{"error": "Please authenticate first"})
				return
			}

			userInfo, err := imageMonkeyDatabase.GetUserInfo(apiUser.Name)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			err = imageMonkeyDatabase.AcceptTrendingLabel(trendingLabel, labelDetails.Label.Type,
				labelDetails.Label.Description, labelDetails.Label.Plural,
				labelDetails.Label.RenameTo, userInfo)
			if err != nil {
				switch err.(type) {
				case *imagemonkeydb.InvalidTrendingLabelError:
					c.JSON(404, gin.H{"error": "Couldn't accept trending label - please provide a valid trending label"})
					return
				default:
					c.JSON(500, gin.H{"error": "Couldn't accept trending label - please try again later"})
					return
				}

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
				Created  int64  `json:"created"`
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

			userExists, err := imageMonkeyDatabase.UserExists(pair[0])
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if !userExists {
				c.JSON(401, gin.H{"error": "Invalid username or password"})
				return
			}

			hashedPassword, err := imageMonkeyDatabase.GetHashedPasswordForUser(pair[0])
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
						Issuer:    "imagemonkey-api",
					},
				}

				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

				tokenString, err := token.SignedString([]byte(jwtSecret))
				if err != nil {
					c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
					return
				}

				err = imageMonkeyDatabase.AddAccessToken(pair[0], tokenString, expirationTime.Unix())
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
			err := imageMonkeyDatabase.RemoveAccessToken(accessTokenInfo.Token)
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

			if (userSignupRequest.Username == "") || (userSignupRequest.Password == "") || (userSignupRequest.Email == "") {
				c.JSON(422, gin.H{"error": "Invalid data"})
				return
			}

			if !commons.IsAlphaNumeric(userSignupRequest.Username) {
				c.JSON(422, gin.H{"error": "Username contains invalid characters"})
				return
			}

			userExists, err := imageMonkeyDatabase.UserExists(userSignupRequest.Username)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if userExists {
				c.JSON(409, gin.H{"error": "Username already taken"})
				return
			}

			emailExists, err := imageMonkeyDatabase.EmailExists(userSignupRequest.Email)
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

			err = imageMonkeyDatabase.CreateUser(userSignupRequest.Username, hashedPassword, userSignupRequest.Email)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			c.JSON(201, nil)
		})

		router.GET("/v1/user/:username/imagecollections", func(c *gin.Context) {
			username := c.Param("username")

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			if apiUser.Name == "" {
				c.JSON(401, gin.H{"error": "Please authenticate first"})
				return
			}

			if username != apiUser.Name {
				c.JSON(403, gin.H{"error": "You are not allowed to perform this action"})
				return
			}

			imageCollections, err := imageMonkeyDatabase.GetImageCollections(apiUser, *apiBaseUrl)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			c.JSON(200, imageCollections)
		})

		router.POST("/v1/user/:username/imagecollection/:name/image/:imageid", func(c *gin.Context) {
			imageCollectionName := c.Param("name")
			imageId := c.Param("imageid")
			username := c.Param("username")

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			if apiUser.Name == "" {
				c.JSON(401, gin.H{"error": "Please authenticate first"})
				return
			}

			if username != apiUser.Name {
				c.JSON(403, gin.H{"error": "You are not allowed to perform this action"})
				return
			}

			if imageCollectionName == "" {
				c.JSON(400, gin.H{"error": "Couldn't process request - image collection name is empty"})
				return
			}

			if imageId == "" {
				c.JSON(400, gin.H{"error": "Couldn't process request - image id is empty"})
				return
			}

			err := imageMonkeyDatabase.AddImageToImageCollection(apiUser, imageCollectionName, imageId, true)
			if err != nil {
				switch err.(type) {
				case *imagemonkeydb.InvalidImageCollectionInputError:
					c.JSON(400, gin.H{"error": "Couldn't process request - invalid input"})
					return
				case *imagemonkeydb.DuplicateImageCollectionError:
					c.JSON(409, gin.H{"error": "Couldn't add image to collection - resource already exists"})
					return
				default:
					c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
					return
				}
			}

			c.JSON(201, nil)
		})

		router.POST("/v1/user/:username/imagecollection", func(c *gin.Context) {
			var imageCollection datastructures.ImageCollection

			username := c.Param("username")

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			if apiUser.Name == "" {
				c.JSON(401, gin.H{"error": "Please authenticate first"})
				return
			}

			if username != apiUser.Name {
				c.JSON(403, gin.H{"error": "You are not allowed to perform this action"})
				return
			}

			if c.BindJSON(&imageCollection) != nil {
				c.JSON(400, gin.H{"error": "Couldn't process request - invalid data"})
				return
			}

			if imageCollection.Name == "" {
				c.JSON(400, gin.H{"error": "Couldn't process request - image collection name is empty"})
				return
			}

			if !IsImageCollectionNameValid(imageCollection.Name) {
				c.JSON(400, gin.H{"error": "Couldn't process request - image collection name contains unsopported characters"})
				return
			}

			err := imageMonkeyDatabase.AddImageCollection(apiUser, imageCollection.Name, imageCollection.Description)
			if err != nil {
				switch err.(type) {
				case *imagemonkeydb.DuplicateImageCollectionError:
					c.JSON(409, gin.H{"error": "Couldn't process request - image collection already exists"})
					return
				default:
					c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
					return
				}
			}

			c.JSON(201, nil)
		})

		router.GET("/v1/user/:username/profile", func(c *gin.Context) {
			username := c.Param("username")
			if username == "" {
				c.JSON(422, gin.H{"error": "Invalid request - username missing"})
				return
			}

			userExists, err := imageMonkeyDatabase.UserExists(username)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if !userExists {
				c.JSON(422, gin.H{"error": "Invalid request - username doesn't exist"})
				return
			}

			userStatistics, err := imageMonkeyDatabase.GetUserStatistics(username)
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

			oldProfilePicture, err := imageMonkeyDatabase.ChangeProfilePicture(username, uuid)
			if err != nil {
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
			username := c.Param("username")

			if username == "" {
				c.JSON(422, gin.H{"error": "Invalid request - username missing"})
				return
			}

			var apiTokenRequest datastructures.ApiTokenRequest
			if c.BindJSON(&apiTokenRequest) != nil {
				c.JSON(422, gin.H{"error": "Invalid request - description missing"})
				return
			}

			accessTokenInfo := authTokenHandler.GetAccessTokenInfo(c)

			if accessTokenInfo.Username != username {
				c.JSON(422, gin.H{"error": "Invalid access token"})
				return
			}

			apiToken, err := imageMonkeyDatabase.GenerateApiToken(jwtSecret, username, apiTokenRequest.Description)
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

			revoked, err := imageMonkeyDatabase.RevokeApiToken(username, token)
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

			userInfo, err := imageMonkeyDatabase.GetUserInfo(username)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			if userInfo.Name == "" {
				c.JSON(422, gin.H{"error": "Incorrect username"})
				return
			}

			var width int
			width = 0
			if temp, ok := params["width"]; ok {
				n, err := strconv.ParseUint(temp[0], 10, 32)
				if err == nil {
					width = int(n)
				}
			}

			var height int
			height = 0
			if temp, ok := params["height"]; ok {
				n, err := strconv.ParseUint(temp[0], 10, 32)
				if err == nil {
					height = int(n)
				}
			}

			fname := userInfo.ProfilePicture
			if userInfo.ProfilePicture == "" {
				fname = "default.png"
			}

			imgBytes, format, err := img.ResizeImage((*userProfilePicturesDir + fname), width, height)
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

		router.GET("/v1/statistics", func(c *gin.Context) {
			statistics, err := imageMonkeyDatabase.Explore(words)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, statistics)
		})

		router.GET("/v1/statistics/donations", func(c *gin.Context) {
			//currently only last-month is allowed as period
			statistics, err := imageMonkeyDatabase.GetDonationsStatistics("last-month")
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, gin.H{"statistics": statistics, "period": "last-month"})
		})

		router.GET("/v1/statistics/annotations", func(c *gin.Context) {
			//currently only last-month is allowed as period
			statistics, err := imageMonkeyDatabase.GetAnnotationStatistics("last-month")
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, gin.H{"statistics": statistics, "period": "last-month"})
		})

		router.GET("/v1/statistics/annotations/count", func(c *gin.Context) {
			minProbabilityStr := commons.GetParamFromUrlParams(c, "min_probability", "0")
			minCountStr := commons.GetParamFromUrlParams(c, "min_count", "0")

			minCount, err := strconv.Atoi(minCountStr)
			if err != nil {
				c.JSON(422, gin.H{"error": "Couldn't process request - invalid 'min_count' parameter"})
				return
			}

			minProbability, err := strconv.ParseFloat(minProbabilityStr, 64)
			if err != nil {
				c.JSON(422, gin.H{"error": "Couldn't process request - invalid 'min_probability' parameter"})
				return
			}

			annotationsCount, err := imageMonkeyDatabase.GetAnnotationsCount(minProbability, minCount)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, annotationsCount)
		})

		router.GET("/v1/statistics/annotations/refinements", func(c *gin.Context) {
			//currently only last-month is allowed as period
			statistics, err := imageMonkeyDatabase.GetAnnotationRefinementStatistics("last-month")
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, gin.H{"statistics": statistics, "period": "last-month"})
		})

		router.GET("/v1/statistics/validations", func(c *gin.Context) {
			//currently only last-month is allowed as period
			statistics, err := imageMonkeyDatabase.GetValidationStatistics("last-month")
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, gin.H{"statistics": statistics, "period": "last-month"})
		})

		router.GET("/v1/statistics/validations/count", func(c *gin.Context) {
			minProbabilityStr := commons.GetParamFromUrlParams(c, "min_probability", "0")
			minCountStr := commons.GetParamFromUrlParams(c, "min_count", "0")

			minCount, err := strconv.Atoi(minCountStr)
			if err != nil {
				c.JSON(422, gin.H{"error": "Couldn't process request - invalid 'min_count' parameter"})
				return
			}

			minProbability, err := strconv.ParseFloat(minProbabilityStr, 64)
			if err != nil {
				c.JSON(422, gin.H{"error": "Couldn't process request - invalid 'min_probability' parameter"})
				return
			}

			validationsCount, err := imageMonkeyDatabase.GetValidationsCount(minProbability, minCount)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, validationsCount)
		})

		router.GET("/v1/statistics/annotated", func(c *gin.Context) {
			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			annotatedStatistics, err := imageMonkeyDatabase.GetAnnotatedStatistics(apiUser, true)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			c.JSON(200, annotatedStatistics)
		})

		router.GET("/v1/statistics/validated", func(c *gin.Context) {
			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			validatedStatistics, err := imageMonkeyDatabase.GetValidatedStatistics(apiUser)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			c.JSON(200, validatedStatistics)
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
			activity, err := imageMonkeyDatabase.GetActivity("last-month")
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, gin.H{"activity": activity, "period": "last-month"})
		})

		router.GET("/v1/user/:username/games/imagehunt/tasks", func(c *gin.Context) {
			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			username := c.Param("username")

			if (apiUser.Name == "") || (apiUser.Name != username) {
				c.JSON(403, gin.H{"error": "You do not have the appropriate permissions to access this information"})
				return
			}

			imageHuntTasks, err := imageMonkeyDatabase.GetImageHuntTasks(apiUser, *apiBaseUrl)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, imageHuntTasks)
		})

		router.GET("/v1/user/:username/games/imagehunt/stats", func(c *gin.Context) {
			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			username := c.Param("username")
			utcOffset, err := commons.GetIntParamFromUrlParams(c, "utc_offset", 0)
			if err != nil {
				c.JSON(400, gin.H{"error": "Please provide a valid utc offset"})
				return
			}

			if (apiUser.Name == "") || (apiUser.Name != username) {
				c.JSON(403, gin.H{"error": "You do not have the appropriate permissions to access this information"})
				return
			}

			imageHuntStats, err := imageMonkeyDatabase.GetImageHuntStats(apiUser, *apiBaseUrl+"v1/games/imagehunt/assets/", len(labelMap), utcOffset)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			c.JSON(200, imageHuntStats)
		})

		router.POST("/v1/games/imagehunt/donate", func(c *gin.Context) {
			var imageSource datastructures.ImageSource
			imageSource.Provider = "imagehunt"
			imageSource.Trusted = false

			var apiUser datastructures.APIUser
			apiUser.ClientFingerprint = getBrowserFingerprint(c)
			apiUser.Name = authTokenHandler.GetAccessTokenInfo(c).Username

			if apiUser.Name == "" {
				c.JSON(403, gin.H{"error": "You do not have the appropriate permissions to access this information"})
				return
			}

			donate(c, imageMonkeyDatabase, apiUser.Name, imageSource, labelMap, metaLabels,
				*unverifiedDonationsDir, redisPool, statisticsPusher, geoipDb, false)
		})

		router.Static("/v1/games/imagehunt/assets", *imageHuntAssetsDir)
	}

	if *corsAllowOrigin == "*" {
		corsWarning := "CORS Access-Control-Allow-Origin is set to '*' - which is a potential security risk."
		corsWarning += "DO NOT RUN THE SERVICE IN PRODUCTION WITH THIS CONFIGURATION!"
		log.Info(corsWarning)
	}

	router.Run(":" + strconv.FormatInt(int64(*listenPort), 10))
}
