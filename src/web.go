package main

	
import (
	"net/http"
	"html/template"
	"github.com/gin-gonic/gin"
	"fmt"
	"os"
	log "github.com/Sirupsen/logrus"
	"flag"
	"database/sql"
	"math"
	"strconv"
)

var db *sql.DB

func main() {
	fmt.Printf("Starting Web Service...\n")

	log.SetLevel(log.DebugLevel)

	fmt.Printf("Setting environment variable for sentry\n")
	os.Setenv("SENTRY_DSN", WEB_SENTRY_DSN)

	releaseMode := flag.Bool("release", false, "Run in release mode")
	wordlistPath := flag.String("wordlist", "../wordlists/en/labels.json", "Path to labels map")
	donationsDir := flag.String("donations_dir", "../donations/", "Location of the uploaded and verified donations")
	apiBaseUrl := flag.String("api_base_url", "http://127.0.0.1:8081", "API Base URL")
	playgroundBaseUrl := flag.String("playground_base_url", "http://127.0.0.1:8082", "Playground Base URL")
	htmlDir := flag.String("html_dir", "../html/templates/", "Location of the html directory")

	webAppIdentifier := "edd77e5fb6fc0775a00d2499b59b75d"
	browserExtensionAppIdentifier := "adf78e53bd6fc0875a00d2499c59b75"


	sessionCookieHandler := NewSessionCookieHandler()

	flag.Parse()
	if(*releaseMode){
		fmt.Printf("Starting gin in release mode!\n")
		gin.SetMode(gin.ReleaseMode)
	}

	funcMap := template.FuncMap{
	    //simple round function
	    //be careful: only works for POSITIVE float values
	    "round" : func(f float32, places int) (float64) {
		    shift := math.Pow(10, float64(places))
		    return math.Floor((float64(f) * shift) + .5) / shift;    
		},
	}

	log.Debug("[Main] Reading Label Map")
	labelMap, words, err := getLabelMap(*wordlistPath)
	if err != nil {
		fmt.Printf("[Main] Couldn't read label map...terminating!")
		log.Fatal(err)
	}

	//open database and make sure that we can ping it
	db, err = sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
	if err != nil {
		log.Fatal("[Main] Couldn't open database: ", err.Error())
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("[Main] Couldn't ping database: ", err.Error())
	}

	tmpl := template.Must(template.New("main").Funcs(funcMap).ParseGlob(*htmlDir + "*"))


	router := gin.Default()
	router.SetHTMLTemplate(tmpl)
	router.Static("./js", "../js") //serve javascript files
	router.Static("./css", "../css") //serve css files
	router.Static("./img", "../img") //serve images
	router.Static("./api", "../html/static/api")
	router.Static("./donations", *donationsDir) //serve doncations
	router.Static("./blog", "../html/static/blog")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "ImageMonkey",
			"activeMenuNr": 1,
			"numOfDonations": pick(getNumOfDonatedImages())[0],
			"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
			"apiBaseUrl": apiBaseUrl,
		})
	})
	router.GET("/donate", func(c *gin.Context) {
		c.HTML(http.StatusOK, "donate.html", gin.H{
			"title": "Donate Image",
			"randomWord": words[random(0, len(words) - 1)],
			"activeMenuNr": 2,
			"apiBaseUrl": apiBaseUrl,
			"words": words,
			"appIdentifier": webAppIdentifier,
			"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
		})
	})

	router.GET("/label", func(c *gin.Context) {
		c.HTML(http.StatusOK, "label.html", gin.H{
			"title": "Add Labels",
			"image": pick(getImageToLabel())[0],
			"activeMenuNr": 3,
			"apiBaseUrl": apiBaseUrl,
			"labels": labelMap,
			"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
		})
	})


	router.GET("/annotate", func(c *gin.Context) {
		params := c.Request.URL.Query()
		

		var labelId int64
		labelId = -1
		if temp, ok := params["label_id"]; ok {
			labelId, err = strconv.ParseInt(temp[0], 10, 64)
			if err != nil {
				c.JSON(422, gin.H{"error": "label id needs to be an integer"})
				return
			}
		}


		c.HTML(http.StatusOK, "annotate.html", gin.H{
			"title": "Annotate",
			"randomImage": pick(getRandomUnannotatedImage(true, labelId))[0],
			"activeMenuNr": 4,
			"apiBaseUrl": apiBaseUrl,
			"appIdentifier": webAppIdentifier,
			"playgroundBaseUrl": playgroundBaseUrl,
			"labelId": labelId,
			"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
		})
	})

	router.GET("/verify", func(c *gin.Context) {
		params := c.Request.URL.Query()

		appIdentifier := webAppIdentifier
		
		showHeader := true
		if temp, ok := params["show_header"]; ok {
			if temp[0] == "false" {
				showHeader = false
			}
		}

		showFooter := true
		if temp, ok := params["show_footer"]; ok {
			if temp[0] == "false" {
				showFooter = false
			}
		}

		onlyOnce := false
		if temp, ok := params["only_once"]; ok {
			if temp[0] == "true" {
				onlyOnce = true
			}
		}

		callback := false
		if temp, ok := params["callback"]; ok {
			if temp[0] == "true" {
				callback = true
			}
		}

		if temp, ok := params["browser_extension"]; ok {
			if temp[0] == "true" {
				appIdentifier = browserExtensionAppIdentifier
			}
		}


		c.HTML(http.StatusOK, "validate.html", gin.H{
			"title": "Validate Label",
			"randomImage": getRandomImage(),
			"activeMenuNr": 5,
			"showHeader": showHeader,
			"showFooter": showFooter,
			"onlyOnce": onlyOnce,
			"apiBaseUrl": apiBaseUrl,
			"appIdentifier": appIdentifier,
			"callback": callback,
			"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
		})
	})
	router.GET("/verify_annotation", func(c *gin.Context) {
		c.HTML(http.StatusOK, "validate_annotations.html", gin.H{
			"title": "Validate Annotations",
			"randomImage": pick(getRandomAnnotatedImage(false))[0],
			"activeMenuNr": 6,
			"apiBaseUrl": apiBaseUrl,
			"appIdentifier": webAppIdentifier,
			"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
		})
	})
	router.GET("/quiz", func(c *gin.Context) {
		c.HTML(http.StatusOK, "quiz.html", gin.H{
			"title": "Quiz",
			"randomQuiz": "",
			"randomAnnotatedImage": pick(getRandomAnnotationForRefinement())[0],
			"activeMenuNr": 7,
			"apiBaseUrl": apiBaseUrl,
			"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
		})
	})	
	router.GET("/statistics", func(c *gin.Context) {
		c.HTML(http.StatusOK, "statistics.html", gin.H{
			"title": "Statistics",
			"words": words,
			"activeMenuNr": 8,
			"statistics": pick(explore(words))[0],
			"apiBaseUrl": apiBaseUrl,
			"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
		})
	})
	router.GET("/explore", func(c *gin.Context) {
		c.HTML(http.StatusOK, "explore.html", gin.H{
			"title": "Explore Dataset",
			"activeMenuNr": 9,
			"apiBaseUrl": apiBaseUrl,
			"labelAccessors": pick(getLabelAccessors())[0],
			"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
		})
	})
	router.GET("/apps", func(c *gin.Context) {
		c.HTML(http.StatusOK, "mobile.html", gin.H{
			"title": "Mobile Apps & Extensions",
			"activeMenuNr": 10,
			"apiBaseUrl": apiBaseUrl,
			"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
		})
	})
	router.GET("/playground", func(c *gin.Context) {
		c.HTML(http.StatusOK, "playground.html", gin.H{
			"title": "Playground",
			"activeMenuNr": 11,
			"apiBaseUrl": apiBaseUrl,
			"playgroundPredictBaseUrl": playgroundBaseUrl,
			"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
		})
	})
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login",
			"apiBaseUrl": apiBaseUrl,
			"activeMenuNr": 12,
			"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
		})
	})

	router.GET("/signup", func(c *gin.Context) {
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"title": "Sign Up",
			"apiBaseUrl": apiBaseUrl,
			"activeMenuNr": -1,
			"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
		})
	})

	router.GET("/profile/:username", func(c *gin.Context) {
		username := c.Param("username")

		exists, _ := userExists(username)
		if !exists {
			c.String(404, "404 page not found")
			return
		}

		c.HTML(http.StatusOK, "profile.html", gin.H{
			"title": "Profile",
			"apiBaseUrl": apiBaseUrl,
			"activeMenuNr": -1,
			"statistics": pick(getUserStatistics(username))[0],
			"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
		})
	})

	router.Run(":8080")
}
