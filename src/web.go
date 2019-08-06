package main

	
import (
	"net/http"
	"html/template"
	"github.com/gin-gonic/gin"
	"fmt"
	"os"
	log "github.com/sirupsen/logrus"
	"flag"
	//"database/sql"
	"math"
	"github.com/getsentry/raven-go"
	"html"
	"time"
	"strings"
	"path/filepath"
	"strconv"
	"errors"
	"net/http/httputil"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	commons "github.com/bbernhard/imagemonkey-core/commons"
	//"net/url"
	imagemonkeydb "github.com/bbernhard/imagemonkey-core/database"
	languages "github.com/bbernhard/imagemonkey-core/languages"
	img "github.com/bbernhard/imagemonkey-core/image" 
	"github.com/NYTimes/gziphandler"
	"github.com/bbernhard/gin-wraphh"
	"compress/gzip"
	"net/url"
	"io/ioutil"
)

func ShowErrorPage(c *gin.Context) {
	c.HTML(404, "404.html", gin.H{
		"title": "Page not found",
	})
}


func GetTemplates(path string, funcMap template.FuncMap)  (*template.Template, error) {
    templ := template.New("main").Funcs(funcMap)
    err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
        if strings.Contains(path, ".html") || strings.Contains(path, ".js") {
            _, err = templ.ParseFiles(path)
            if err != nil {
                return err
            }
        }

        return err
    })

    return templ, err
}

func GetImages(p string) (map[string]string, error) {
	files := make(map[string]string, 0)
	err := filepath.Walk(p, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() || (f.IsDir() && (path != p)) {
			rel, err := filepath.Rel(p, path)
			if err != nil {
				return err
			}
			files[filepath.ToSlash(rel)] = path
		}
		return err
	})

	if err != nil {
		return files, err
	}

	return files, nil
}


func ReverseProxy(target string, sessionCookieHandler *SessionCookieHandler, 
					imageMonkeyDatabase *imagemonkeydb.ImageMonkeyDatabase) gin.HandlerFunc {
    return func(c *gin.Context) {
    	sessionInformation := sessionCookieHandler.GetSessionInformation(c)

    	hasPermission := false
		if sessionInformation.LoggedIn {
			userInfo, _ := imageMonkeyDatabase.GetUserInfo(sessionInformation.Username)
			if userInfo.IsModerator && userInfo.Permissions != nil && userInfo.Permissions.CanMonitorSystem {
				hasPermission = true
			}
		}

		if hasPermission {
	        director := func(req *http.Request) {
	            req.URL.Scheme = "http"
	            req.URL.Host = target
	            req.Host = ""
	            req.URL.Path = ""
	        }
	        proxy := &httputil.ReverseProxy{Director: director}
	        proxy.ServeHTTP(c.Writer, c.Request)
	    } else {
	    	ShowErrorPage(c)
	    }
    }
}

func main() {
	fmt.Printf("Starting Web Service...\n")

	log.SetLevel(log.DebugLevel)

	releaseMode := flag.Bool("release", false, "Run in release mode")
	wordlistPath := flag.String("wordlist", "../wordlists/en/labels.jsonnet", "Path to labels map")
	metalabelsPath := flag.String("metalabels", "../wordlists/en/metalabels.jsonnet", "Path to metalabels")
	labelRefinementsPath := flag.String("label_refinements", "../wordlists/en/label-refinements.json", "Path to label refinements")
	donationsDir := flag.String("donations_dir", "../donations/", "Location of the uploaded and verified donations")
	apiBaseUrl := flag.String("api_base_url", "http://127.0.0.1:8081", "API Base URL")
	playgroundBaseUrl := flag.String("playground_base_url", "http://127.0.0.1:8082", "Playground Base URL")
	htmlDir := flag.String("html_dir", "../html/templates/", "Location of the html directory")
	maintenanceModeFile := flag.String("maintenance_mode_file", "../maintenance.tmp", "maintenance mode file")
	useSentry := flag.Bool("use_sentry", false, "Use Sentry for error logging")
	listenPort := flag.Int("listen_port", 8080, "Specify the listen port")
	publicBackupsPath := flag.String("public_backups_path", "../public_backups/public_backups.json", "Path to public backups")
	netdataUrl := flag.String("netdata_url", "127.0.0.1:19999", "Netdata Monitoring URL")
	modelsPath :=   flag.String("models_path", "https://raw.githubusercontent.com/bbernhard/imagemonkey-models/master/models.json", 
								"Path to the pre-trained models")
	gzipCompress := flag.Bool("gzip_compress", true, "Use Gzip Compression")
	localSentryDsn := flag.String("local_sentry_dsn", "http://sentry:sentry@127.0.0.1:8080/sentry", "local Sentry DSN")
	labelsRepositoryUrl := flag.String("labels_repository_url", "https://github.com/bbernhard/imagemonkey-labels-test", "Labels Repository")

	webAppIdentifier := "edd77e5fb6fc0775a00d2499b59b75d"
	browserExtensionAppIdentifier := "adf78e53bd6fc0875a00d2499c59b75"
	sentryEnvironment := "web"

	flag.Parse()
	if *releaseMode {
		fmt.Printf("Starting gin in release mode!\n")
		gin.SetMode(gin.ReleaseMode)
	}

	sentryDsn := ""
	if *useSentry {
		fmt.Printf("Setting Sentry DSN\n")
		sentryDsn = commons.MustGetEnv("SENTRY_DSN")
		raven.SetDSN(sentryDsn)
		raven.SetEnvironment(sentryEnvironment)

		raven.CaptureMessage("Starting up web worker", nil)
	} else { //also disable local sentry reporting
		*localSentryDsn = ""
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
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
	        if len(values)%2 != 0 {
	            return nil, errors.New("invalid dict call")
	        }
	        dict := make(map[string]interface{}, len(values)/2)
	        for i := 0; i < len(values); i+=2 {
	            key, ok := values[i].(string)
	            if !ok {
	                return nil, errors.New("dict keys must be strings")
	            }
	            dict[key] = values[i+1]
	        }
	        return dict, nil
	    },
	    "loop": func(min int32, n int32) []int32 {
            arr := make([]int32, n-min+1)
		    for i := range arr {
		        arr[i] = int32(min) + int32(i)
		    }
		    return arr
        },
		/*"executeTemplate": func(name string) string {
    		buf := &bytes.Buffer{}
    		_ = tmpl.ExecuteTemplate(buf, name, nil)
    		return buf.String()
		},*/
	}

	clientSecret := commons.MustGetEnv("X_CLIENT_SECRET")
	clientId := commons.MustGetEnv("X_CLIENT_ID")

	log.Debug("[Main] Reading labels")
	labelRepository := commons.NewLabelRepository(*wordlistPath)
	err := labelRepository.Load()
	if err != nil {
		fmt.Printf("[Main] Couldn't read labels: %s...terminating!",*wordlistPath)
		log.Fatal(err)
	}

	words := labelRepository.GetWords()

	log.Debug("[Main] Reading label refinements")
	labelRefinementsMap, err := commons.GetLabelRefinementsMap(*labelRefinementsPath)
	if err != nil {
		fmt.Printf("[Main] Couldn't read label refinements: %s...terminating!", *labelRefinementsPath)
		log.Fatal(err)
	}

	log.Debug("[Main] Reading Metalabel Map")
	metaLabels := commons.NewMetaLabels(*metalabelsPath)
	err = metaLabels.Load()
	if err != nil {
		fmt.Printf("[Main] Couldn't read metalabel map...terminating!")
		log.Fatal(err)
	}

	log.Debug("[Main] Reading public backups")
	publicBackups, err := commons.GetPublicBackups(*publicBackupsPath)
	if err != nil {
		fmt.Printf("[Main] Couldn't read public backups: %s...terminating!", *publicBackupsPath)
		log.Fatal(err)
	}

	imageDbConnectionString := commons.MustGetEnv("IMAGEMONKEY_DB_CONNECTION_STRING")

	imageMonkeyDatabase := imagemonkeydb.NewImageMonkeyDatabase()
	err = imageMonkeyDatabase.Open(imageDbConnectionString)
	if err != nil {
		log.Fatal("[Main] Couldn't ping ImageMonkey database: ", err.Error())
	}
	defer imageMonkeyDatabase.Close()

	if *useSentry {
		imageMonkeyDatabase.InitializeSentry(sentryDsn, sentryEnvironment)
	}
	
	jwtSecret := commons.MustGetEnv("JWT_SECRET")
	sessionCookieHandler := NewSessionCookieHandler(imageMonkeyDatabase, jwtSecret)

	availableModels, err := commons.GetAvailableModels(*modelsPath)
	if err != nil {
		log.Fatal("[Main] Couldn't get available models: ", err.Error())
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

	imgs, err := GetImages("../img")
	if err != nil {
		log.Fatal("[Main] Couldn't read images", err.Error())
	}

	router := gin.Default()
	router.SetHTMLTemplate(tmpl)

	if *gzipCompress {
		gzipHandler, err := gziphandler.GzipHandlerWithOpts(gziphandler.CompressionLevel(gzip.DefaultCompression), 
															gziphandler.ContentTypes([]string{"text/html", "text/css", "application/javascript", "text/javascript"}))
		if err != nil {
			log.Fatal("[Main] Couldn't initialize Gzip Handler", err.Error())
		}
		router.Use(wraphh.WrapHH(gzipHandler))
	}

	
	router.Static("./js", "../js") //serve javascript files
	router.Static("./css", "../css") //serve css files

	if maintenanceMode {
		router.NoRoute(func(c *gin.Context) {
    		c.HTML(http.StatusOK, "maintenance.html", gin.H{
    			"title": "ImageMonkey Maintenance",
    		})
		})	
	} else {
		//router.Static("./img", "../img") //serve images
		router.GET("/img/*name", func(c *gin.Context) { //serve images
			imageName := strings.Trim(c.Param("name"), "/")
			params := c.Request.URL.Query()

			if _, ok := imgs[imageName]; !ok {
    			ShowErrorPage(c)
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

			imgBytes, format, err := img.ResizeImage(("../img/" + imageName), width, height)
			if err != nil {
				log.Error("[Serving Image] Couldn't serve img: ", err.Error())
				c.String(500, "Couldn't process request - please try again later")
				return

			}

			if format == "svg" {
				format = "svg+xml"
			}

			c.Writer.Header().Set("Content-Type", ("image/" + format))
	        c.Writer.Header().Set("Content-Length", strconv.Itoa(len(imgBytes)))
	        _, err = c.Writer.Write(imgBytes) 
	        if err != nil {
	            log.Error("[Serving Image] Couldn't serve img: ", err.Error())
	            c.String(500, "Couldn't process request - please try again later")
	            return
	        }
		})


		//router.Static("./api", "../html/static/api")
		router.Static("./donations", *donationsDir) //DEPRECTATED; USE /donation API endpoint 
		router.Static("./blog", "../html/static/blog")
		router.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"title": "ImageMonkey",
				"activeMenuNr": 1,
				"numOfDonations": commons.Pick(imageMonkeyDatabase.GetNumOfDonatedImages())[0],
				"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
				"apiBaseUrl": apiBaseUrl,
				"annotationStatistics": commons.Pick(imageMonkeyDatabase.GetAnnotationStatistics("last-month"))[0],
				"validationStatistics": commons.Pick(imageMonkeyDatabase.GetValidationStatistics("last-month"))[0],
				"labeledObjectsStatistics": commons.Pick(imageMonkeyDatabase.GetLabeledObjectsStatistics("last-month"))[0],
				"annotationRefinementStatistics": commons.Pick(imageMonkeyDatabase.GetAnnotationRefinementStatistics("last-month"))[0],
				"imageDescriptionStatistics": commons.Pick(imageMonkeyDatabase.GetImageDescriptionStatistics("last-month"))[0],
				"donationsStatistics": commons.Pick(imageMonkeyDatabase.GetDonationsStatistics("last-month"))[0],
			})
		})
		router.GET("/donate", func(c *gin.Context) {
			c.HTML(http.StatusOK, "donate.html", gin.H{
				"title": "Donate Image",
				"randomWord": words[commons.Random(0, len(words) - 1)],
				"activeMenuNr": 2,
				"apiBaseUrl": apiBaseUrl,
				"words": words,
				"appIdentifier": webAppIdentifier,
				"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
			})
		})

		router.GET("/label", func(c *gin.Context) {
			sessionInformation := sessionCookieHandler.GetSessionInformation(c)

			mode := commons.GetParamFromUrlParams(c, "mode", "default")
			operationType := commons.GetParamFromUrlParams(c, "type", "object")

			imageId := ""
			if mode == "default" {
				imageId = commons.GetParamFromUrlParams(c, "image_id", "")
			}

			isModerator := false
			if sessionInformation.LoggedIn {
				userInfo, _ := imageMonkeyDatabase.GetUserInfo(sessionInformation.Username)
				if userInfo.IsModerator && userInfo.Permissions != nil && userInfo.Permissions.CanRemoveLabel {
					isModerator = true
				}
			}

			title := ""
			subtitle := ""
			activeMenuNr := 3
			if operationType == "object" {
				title = "Add Labels"
				subtitle = "Label all objects"
				activeMenuNr = 3
			} else {
				title = "Add Image Description"
				subtitle = "Describe the Image"
				activeMenuNr = 15
			}


			c.HTML(http.StatusOK, "label.html", gin.H{
				"title": title,
				"subtitle": subtitle,
				"imageId": imageId,
				"mode": mode,
				"type": operationType,
				"activeMenuNr": activeMenuNr,
				"apiBaseUrl": apiBaseUrl,
				"languages": languages.GetAllSupported(),
				"labelSuggestions": commons.Pick(imageMonkeyDatabase.GetLabelSuggestions())[0],
				"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
				"isModerator" : isModerator,
				"labelAccessors": commons.Pick(imageMonkeyDatabase.GetLabelAccessors())[0],
				"labelAccessorsLookup": commons.Pick(imageMonkeyDatabase.GetLabelAccessorsMapping())[0],
				"queryAttributes": commons.GetStaticQueryAttributes(),
			})
		})


		router.GET("/annotate", func(c *gin.Context) {
			params := c.Request.URL.Query()

			labelId, err := commons.GetLabelIdFromUrlParams(params)
			if err != nil {
				c.JSON(422, gin.H{"error": "label id needs to be an integer"})
				return
			}

			mode := commons.GetParamFromUrlParams(c, "mode", "default")
			onlyOnce := false
			var revision int64
			revision = -1
			showSkipAnnotationButtons := true
			validationId := ""
			annotationId := ""

			if mode == "default" {
				annotationId = commons.GetParamFromUrlParams(c, "annotation_id", "")
				if annotationId != "" {
					mode = "refine"
					onlyOnce = true
					showSkipAnnotationButtons = false //if there are already annotations, 
													 //then we do not need to show the blacklist annotation and unannotatable buttons

					revisionStr := commons.GetParamFromUrlParams(c, "rev", "-1")
					revision, err = strconv.ParseInt(revisionStr, 10, 32)
					if err != nil {
						ShowErrorPage(c)
						return
					}

				} else {
					validationId = commons.GetValidationIdFromUrlParams(params)
					if validationId != "" {
						//it doesn't make sene to use the validation id and the label id for querying - so we
						//give the validation id preference.
						labelId = ""
						onlyOnce = true
					}
				}
			}

			view := commons.GetParamFromUrlParams(c, "view", "default")
			
			c.HTML(http.StatusOK, "annotate.html", gin.H{
				"title": "Annotate",
				"activeMenuNr": 4,
				"apiBaseUrl": apiBaseUrl,
				"appIdentifier": webAppIdentifier,
				"playgroundBaseUrl": playgroundBaseUrl,
				"labelId": labelId,
				"annotationRevision": revision,
				"validationId": validationId,
				"annotationId": annotationId,
				"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
				"annotationMode": mode,
				"annotationView": view,
				"onlyOnce": onlyOnce,
				"sentryDsn": localSentryDsn,
				"showSkipAnnotationButtons": showSkipAnnotationButtons,
				"labelAccessors": commons.Pick(imageMonkeyDatabase.GetLabelAccessorDetails("normal"))[0],
				"queryAttributes": commons.GetStaticQueryAttributes(),
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

			showSettings := true
			if temp, ok := params["show_settings"]; ok {
				if temp[0] == "false" {
					showSettings = false
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

			validationId := commons.GetParamFromUrlParams(c, "validation_id", "")
			mode := commons.GetParamFromUrlParams(c, "mode", "default")

			c.HTML(http.StatusOK, "validate.html", gin.H{
				"title": "Validate Label",
				"activeMenuNr": 5,
				"showHeader": showHeader,
				"showFooter": showFooter,
				"showSettings": showSettings,
				"onlyOnce": onlyOnce,
				"apiBaseUrl": apiBaseUrl,
				"appIdentifier": appIdentifier,
				"callback": callback,
				"mode": mode,
				"validationId": validationId,
				"labelAccessors": commons.Pick(imageMonkeyDatabase.GetLabelAccessors())[0],
				"queryAttributes": commons.GetStaticQueryAttributes(),
				"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
			})
		})
		router.GET("/verify_annotation", func(c *gin.Context) {
			c.HTML(http.StatusOK, "validate_annotations.html", gin.H{
				"title": "Validate Annotations",
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
				"randomAnnotatedImage": commons.Pick(imageMonkeyDatabase.GetRandomAnnotationForQuizRefinement())[0],
				"activeMenuNr": 7,
				"apiBaseUrl": apiBaseUrl,
				"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
			})
		})	
		router.GET("/refine", func(c *gin.Context) {
			mode := commons.GetParamFromUrlParams(c, "mode", "default")

			c.HTML(http.StatusOK, "refinement.html", gin.H{
				"title": "Refinement",
				"activeMenuNr": 14,
				"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
				"apiBaseUrl": apiBaseUrl,
				"mode": mode,
				"labels": labelRefinementsMap,
				"labelAccessors": commons.Pick(imageMonkeyDatabase.GetLabelAccessors())[0],
				"labelCategories": commons.Pick(imageMonkeyDatabase.GetLabelCategories())[0],
			})
		})
		router.GET("/statistics", func(c *gin.Context) {
			c.HTML(http.StatusOK, "statistics.html", gin.H{
				"title": "Statistics",
				"words": words,
				"activeMenuNr": 8,
				"statistics": commons.Pick(imageMonkeyDatabase.Explore(words))[0],
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

			queryInfo.Query, queryInfo.AnnotationsOnly, _ = commons.GetExploreUrlParams(c)

			c.HTML(http.StatusOK, "explore.html", gin.H{
				"title": "Explore Dataset",
				"activeMenuNr": 9,
				"apiBaseUrl": apiBaseUrl,
				"labelAccessors": commons.Pick(imageMonkeyDatabase.GetLabelAccessors())[0],
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

			userInfo, _ := imageMonkeyDatabase.GetUserInfo(username)
			if userInfo.Name == "" {
				c.String(404, "404 page not found")
				return
			}

			sessionInformation := sessionCookieHandler.GetSessionInformation(c)

			var apiTokens []datastructures.APIToken
			if sessionInformation.Username == userInfo.Name { //only fetch API tokens in case it's our own profile
				apiTokens, err = imageMonkeyDatabase.GetApiTokens(username)
				if err != nil {
					c.String(500, "Internal server error - please try again later")
					return
				}
			}

			c.HTML(http.StatusOK, "profile.html", gin.H{
				"title": "Profile",
				"apiBaseUrl": apiBaseUrl,
				"activeMenuNr": -1,
				"statistics": commons.Pick(imageMonkeyDatabase.GetUserStatistics(username))[0],
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

		router.GET("/public_backup", func(c *gin.Context) {
			c.HTML(http.StatusOK, "public_backup.html", gin.H{
				"title": "Public Backup",
				"apiBaseUrl": apiBaseUrl,
				"activeMenuNr": -1,
				"publicBackups": publicBackups,
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
				"repository": "https://github.com/bbernhard/imagemonkey-core",
			})
		})

		router.GET("/labelrepository", func(c *gin.Context) {
			c.HTML(http.StatusOK, "labelrepository.html", gin.H{
				"title": "Label Repository",
				"apiBaseUrl": apiBaseUrl,
				"activeMenuNr": -1,
				"labelsRepositoryUrl": *labelsRepositoryUrl,
				"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
			})
		})

		router.GET("/moderation", func(c *gin.Context) {
			sessionInformation := sessionCookieHandler.GetSessionInformation(c)

			if !sessionInformation.IsModerator {
				ShowErrorPage(c)
				return
			} 

			c.HTML(http.StatusOK, "moderation.html", gin.H{
				"title": "Content Moderation",
				"apiBaseUrl": apiBaseUrl,
				"activeMenuNr": -1,
				//"activeMenuNr": 14,
				"sessionInformation": sessionInformation,
			})
		})

		router.GET("/image_unlock", func(c *gin.Context) {
			sessionInformation := sessionCookieHandler.GetSessionInformation(c)

			mode := commons.GetParamFromUrlParams(c, "mode", "default")

			isAuthenticated := false
			if sessionInformation.LoggedIn {
				userInfo, _ := imageMonkeyDatabase.GetUserInfo(sessionInformation.Username)
				if userInfo.IsModerator && userInfo.Permissions != nil && userInfo.Permissions.CanUnlockImage {
					isAuthenticated = true
				}
			}

			if !isAuthenticated {
				ShowErrorPage(c)
				return
			}

			c.HTML(http.StatusOK, "image_unlock.html", gin.H{
				"title": "Unlock Image",
				"clientSecret": clientSecret, 
				"clientId": clientId, 
				"apiBaseUrl": apiBaseUrl,
				"activeMenuNr": -1,
				"sessionInformation": sessionInformation,
				"mode": mode,
			})
		})

		router.GET("/models", func(c *gin.Context) {
			sessionInformation := sessionCookieHandler.GetSessionInformation(c)

			c.HTML(http.StatusOK, "models.html", gin.H{
				"title": "Models",
				"activeMenuNr": 16,
				"sessionInformation": sessionInformation,
				"models": availableModels,
			})
		})


		router.GET("/monitoring/", ReverseProxy(*netdataUrl, sessionCookieHandler, imageMonkeyDatabase))

		/*router.GET("/reset_password", func(c *gin.Context) {
			c.HTML(http.StatusOK, "reset_password.html", gin.H{
				"title": "Profile",
				"apiBaseUrl": apiBaseUrl,
				"activeMenuNr": -1,
				"sessionInformation": sessionCookieHandler.GetSessionInformation(c),
			})
		})*/

		router.POST("/api/sentry/store/", func(c *gin.Context) {
			//parse sentry dsn
			sentryDsnWithoutHttps := strings.Replace(sentryDsn, "https://", "", 1)
			sentryDsnParts := strings.Split(sentryDsnWithoutHttps, ":")
			if len(sentryDsnParts) != 2 { //invalid Sentry DSN
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			sentryKey := sentryDsnParts[0]

			sentryDsnSubParts := strings.Split(sentryDsnParts[1], "/")
			if len(sentryDsnSubParts) != 2 { //invalid Sentry DSN
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			entry := sentryDsnSubParts[1]


			u, err := url.ParseRequestURI(c.Request.RequestURI)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			baseUrl, err := url.Parse("https://sentry.io/api/" + entry + "/store/")
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			queryValues := u.Query()
			queryValues.Set("sentry_key", sentryKey) //replace Sentry Key with real key

			baseUrl.RawQuery = queryValues.Encode()

			var resp *http.Response
			var req *http.Request
			client := &http.Client{}

			req, err = http.NewRequest(c.Request.Method, baseUrl.String(), c.Request.Body)
			for name, value := range c.Request.Header {
				req.Header.Set(name, value[0])
			}
			resp, err = client.Do(req)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 && resp.StatusCode != 201  {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				c.JSON(500, gin.H{"error": "Couldn't process request - please try again later"})
				return
			}
    		c.JSON(resp.StatusCode, body)
		})

		router.NoRoute(func(c *gin.Context) {
			c.HTML(404, "404.html", gin.H{
				"title": "Page not found",
			})
		})
	}

	router.Run(":" + strconv.FormatInt(int64(*listenPort), 10))
}
