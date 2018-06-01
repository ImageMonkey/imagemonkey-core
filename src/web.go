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
	"github.com/getsentry/raven-go"
	"html"
	"time"
	"strings"
	"path/filepath"
	"strconv"
)

var db *sql.DB


func GetTemplates(path string, funcMap template.FuncMap)  (*template.Template, error) {
    templ := template.New("main").Funcs(funcMap)
    err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
        if strings.Contains(path, ".html") {
            _, err = templ.ParseFiles(path)
            if err != nil {
                return err
            }
        }

        return err
    })

    return templ, err
}

func main() {
	fmt.Printf("Starting Web Service...\n")

	log.SetLevel(log.DebugLevel)

	releaseMode := flag.Bool("release", false, "Run in release mode")
	wordlistPath := flag.String("wordlist", "../wordlists/en/labels.json", "Path to labels map")
	donationsDir := flag.String("donations_dir", "../donations/", "Location of the uploaded and verified donations")
	apiBaseUrl := flag.String("api_base_url", "http://127.0.0.1:8081", "API Base URL")
	playgroundBaseUrl := flag.String("playground_base_url", "http://127.0.0.1:8082", "Playground Base URL")
	htmlDir := flag.String("html_dir", "../html/templates/", "Location of the html directory")
	maintenanceModeFile := flag.String("maintenance_mode_file", "../maintenance.tmp", "maintenance mode file")
	useSentry := flag.Bool("use_sentry", false, "Use Sentry for error logging")
	listenPort := flag.Int("listen_port", 8080, "Specify the listen port")

	webAppIdentifier := "edd77e5fb6fc0775a00d2499b59b75d"
	browserExtensionAppIdentifier := "adf78e53bd6fc0875a00d2499c59b75"

	sessionCookieHandler := NewSessionCookieHandler()

	flag.Parse()
	if *releaseMode {
		fmt.Printf("Starting gin in release mode!\n")
		gin.SetMode(gin.ReleaseMode)
	}

	if *useSentry {
		fmt.Printf("Setting Sentry DSN\n")
		raven.SetDSN(SENTRY_DSN)
		raven.SetEnvironment("web")

		raven.CaptureMessage("Starting up web worker", nil)
	}

	var tmpl *template.Template

	funcMap := template.FuncMap{
	    //simple round function
	    //be careful: only works for POSITIVE float values
	    "round" : func(f float32, places int) (float64) {
		    shift := math.Pow(10, float64(places))
		    return math.Floor((float64(f) * shift) + .5) / shift;    
		},
		"htmlEscape" : func(s string) string {
			return html.EscapeString(s)
		},
		"elideRight" : func(s string) string {
			if len(s) > 15 {
				return s[:14] + "..."
			}

			return s
		},
		"unixTimestampToDateStr" : func(t int64) string {
			d := time.Unix(t, 0)
			return fmt.Sprintf("%d-%02d-%02d", d.Year(), d.Month(), d.Day())
		},
		/*"executeTemplate": func(name string) string {
    		buf := &bytes.Buffer{}
    		_ = tmpl.ExecuteTemplate(buf, name, nil)
    		return buf.String()
		},*/
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


	//if file exists, start in maintenance mode
	maintenanceMode := false
	if _, err := os.Stat(*maintenanceModeFile); err == nil {
		maintenanceMode = true
		log.Info("[Main] Starting in maintenance mode")
	}

	tmpl, err = GetTemplates(*htmlDir, funcMap)
	if err != nil {
		log.Fatal("[Main] Couldn't parse templates", err.Error())
	}

	router := gin.Default()
	router.SetHTMLTemplate(tmpl)
	router.Static("./js", "../js") //serve javascript files
	router.Static("./css", "../css") //serve css files

	if maintenanceMode {
		router.NoRoute(func(c *gin.Context) {
    		c.HTML(http.StatusOK, "maintenance.html", gin.H{
    			"title": "ImageMonkey Maintenance",
    		})
		})	
	} else {
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
				"annotationStatistics": pick(getAnnotationStatistics("last-month"))[0],
				"validationStatistics": pick(getValidationStatistics("last-month"))[0],
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
				"labelSuggestions": pick(getLabelSuggestions())[0],
				"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
			})
		})


		router.GET("/annotate", func(c *gin.Context) {
			params := c.Request.URL.Query()

			sessionInformation := sessionCookieHandler.GetSessionInformation(c)
			

			labelId, err := getLabelIdFromUrlParams(params)
			if err != nil {
				c.JSON(422, gin.H{"error": "label id needs to be an integer"})
				return
			}


			c.HTML(http.StatusOK, "annotate.html", gin.H{
				"title": "Annotate",
				"randomImage": pick(getRandomUnannotatedImage(sessionInformation.Username, true, labelId))[0],
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

			labelId, err := getLabelIdFromUrlParams(params)
			if err != nil {
				c.JSON(422, gin.H{"error": "label id needs to be an integer"})
				return
			}


			c.HTML(http.StatusOK, "validate.html", gin.H{
				"title": "Validate Label",
				"randomImage": getRandomImage(labelId),
				"activeMenuNr": 5,
				"showHeader": showHeader,
				"showFooter": showFooter,
				"onlyOnce": onlyOnce,
				"apiBaseUrl": apiBaseUrl,
				"appIdentifier": appIdentifier,
				"callback": callback,
				"labelId": labelId,
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
			type QueryInfo struct {
				Query string
				AnnotationsOnly bool
			}

			var queryInfo QueryInfo

			queryInfo.Query, queryInfo.AnnotationsOnly, _ = getExploreUrlParams(c)

			c.HTML(http.StatusOK, "explore.html", gin.H{
				"title": "Explore Dataset",
				"activeMenuNr": 9,
				"apiBaseUrl": apiBaseUrl,
				"labelAccessors": pick(getLabelAccessors())[0],
				"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
				"queryInfo": queryInfo,
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
			sessionInformation := sessionCookieHandler.GetSessionInformation(c)

			//when logged in, redirect to profile page
			if(sessionInformation.LoggedIn){
				redirectUrl := "/profile/" + sessionInformation.Username
				c.Redirect(302, redirectUrl)
			} else {
				c.HTML(http.StatusOK, "login.html", gin.H{
					"title": "Login",
					"apiBaseUrl": apiBaseUrl,
					"activeMenuNr": 12,
					"sessionInformation": sessionInformation,
				})
			}
		})

		router.GET("/signup", func(c *gin.Context) {
			sessionInformation := sessionCookieHandler.GetSessionInformation(c)
			//when logged in, redirect to profile page
			if(sessionInformation.LoggedIn){
				redirectUrl := "/profile/" + sessionInformation.Username
				c.Redirect(302, redirectUrl)
			} else {
				c.HTML(http.StatusOK, "signup.html", gin.H{
					"title": "Sign Up",
					"apiBaseUrl": apiBaseUrl,
					"activeMenuNr": -1,
					"sessionInformation": sessionInformation,
				})
			}
		})

		router.GET("/profile/:username", func(c *gin.Context) {
			username := c.Param("username")

			userInfo, _ := getUserInfo(username)
			if userInfo.Name == "" {
				c.String(404, "404 page not found")
				return
			}

			sessionInformation := sessionCookieHandler.GetSessionInformation(c)

			var apiTokens []APIToken
			if sessionInformation.Username == userInfo.Name { //only fetch API tokens in case it's our own profile
				apiTokens, err = getApiTokens(username)
				if err != nil {
					c.String(500, "Internal server error - please try again later")
					return
				}
			}

			c.HTML(http.StatusOK, "profile.html", gin.H{
				"title": "Profile",
				"apiBaseUrl": apiBaseUrl,
				"activeMenuNr": -1,
				"statistics": pick(getUserStatistics(username))[0],
				"userInfo": userInfo,
				"sessionInformation": sessionInformation,
				"apiTokens": apiTokens,
			})
		})

		router.GET("/libraries", func(c *gin.Context) {
			c.HTML(http.StatusOK, "libraries.html", gin.H{
				"title": "Libraries",
				"apiBaseUrl": apiBaseUrl,
				"activeMenuNr": 13,
				"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
			})
		})

		router.GET("/graph", func(c *gin.Context) {
			params := c.Request.URL.Query()

			labelGraphName := "main"
			if temp, ok := params["name"]; ok {
				labelGraphName = temp[0]
			}

			title := "Label Graph"
			editorMode := false
			if temp, ok := params["editor"]; ok {
				if temp[0] == "true" {
					editorMode = true
					title = "Label Graph Editor"
				}
			}


			c.HTML(http.StatusOK, "graph.html", gin.H{
				"title": title,
				"apiBaseUrl": apiBaseUrl,
				"activeMenuNr": 14,
				"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
				"defaultLabelGraphName": labelGraphName,
				"editorMode" : editorMode,
			})
		})

		/*router.GET("/reset_password", func(c *gin.Context) {
			c.HTML(http.StatusOK, "reset_password.html", gin.H{
				"title": "Profile",
				"apiBaseUrl": apiBaseUrl,
				"activeMenuNr": -1,
				"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
			})
		})*/
	}

	router.Run(":" + strconv.FormatInt(int64(*listenPort), 10))
}
