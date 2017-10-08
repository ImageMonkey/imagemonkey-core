package main

	
import (
	"net/http"
	"html/template"
	"github.com/gin-gonic/gin"
	"time"
	"fmt"
	"github.com/satori/go.uuid"
	"strings"
	"os"
	"gopkg.in/h2non/filetype.v1"
	log "github.com/Sirupsen/logrus"
	"flag"
	"database/sql"
)

var db *sql.DB

func main() {
	fmt.Printf("Starting Web Service...\n")

	log.SetLevel(log.DebugLevel)

	fmt.Printf("Setting environment variable for sentry\n")
	os.Setenv("SENTRY_DSN", WEB_SENTRY_DSN)

	releaseMode := flag.Bool("release", false, "Run in release mode")
	wordlistDir := flag.String("wordlist", "../wordlists/en/misc.txt", "Path to wordlist")
	donationsDir := flag.String("donations_dir", "../donations/", "Location of the uploaded and verified donations")
	unverifiedDonationsDir := flag.String("unverified_donations_dir", "../unverified_donations/", "Location of the uploaded but unverified donations")
	htmlDir := flag.String("html_dir", "../html/templates/", "Location of the html directory")

	flag.Parse()
	if(*releaseMode){
		fmt.Printf("Starting gin in release mode!\n")
		gin.SetMode(gin.ReleaseMode)
	}

	funcMap := template.FuncMap{
	    "formatTime": func(raw int64) string {
	        t := time.Unix(raw, 0)

	        return t.Format("Jan 2 15:04:05 2006")
	    },
	}

	fmt.Printf("Reading wordlists...")
	words, err := getStrWordLists(*wordlistDir)
	if(err != nil){
		fmt.Printf("Couldn't read wordlists...terminating!")
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
		})
	})
	router.GET("/donate", func(c *gin.Context) {
		c.HTML(http.StatusOK, "donate.html", gin.H{
			"title": "Donate Image",
			"randomWord": words[random(0, len(words) - 1)],
			"activeMenuNr": 2,
		})
	})
	router.GET("/annotate", func(c *gin.Context) {
		c.HTML(http.StatusOK, "annotate.html", gin.H{
			"title": "Annotate",
			"randomImage": getRandomUnannotatedImage(),
			"activeMenuNr": 3,
		})
	})
	router.GET("/verify", func(c *gin.Context) {
		params := c.Request.URL.Query()
		
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
		c.HTML(http.StatusOK, "validate.html", gin.H{
			"title": "Validate Label",
			"randomImage": getRandomImage(),
			"activeMenuNr": 4,
			"showHeader": showHeader,
			"showFooter": showFooter,
		})
	})
	router.GET("/verify_annotation", func(c *gin.Context) {
		c.HTML(http.StatusOK, "validate_annotations.html", gin.H{
			"title": "Validate Annotations",
			"randomImage": getRandomAnnotatedImage(),
			"activeMenuNr": 5,
		})
	})
	router.GET("/explore", func(c *gin.Context) {
		c.HTML(http.StatusOK, "explore.html", gin.H{
			"title": "Explore Dataset",
			"words": words,
			"activeMenuNr": 6,
			"graphData": explore(),
		})
	})
	router.GET("/export", func(c *gin.Context) {
		c.HTML(http.StatusOK, "export.html", gin.H{
			"title": "Export Dataset",
			"words": words,
			"activeMenuNr": 7,
		})
	})
	router.GET("/mobile", func(c *gin.Context) {
		c.HTML(http.StatusOK, "mobile.html", gin.H{
			"title": "Mobile App",
			"activeMenuNr": 8,
		})
	})
	router.GET("/playground", func(c *gin.Context) {
		c.HTML(http.StatusOK, "playground.html", gin.H{
			"title": "Playground",
			"activeMenuNr": 9,
			"playgroundPredictBaseUrl": "https://playground.imagemonkey.io",
			//"playgroundPredictBaseUrl": "http://127.0.0.1:8081",
		})
	})

	router.GET("/validate", func(c *gin.Context) {
		randomImage := getRandomImage()
		c.JSON(http.StatusOK, gin.H{"uuid": randomImage.Id, "label": randomImage.Label, "provider": randomImage.Provider})
	})

	router.GET("/annotate/data", func(c *gin.Context) {
		randomImage := getRandomUnannotatedImage()
		c.JSON(http.StatusOK, gin.H{"uuid": randomImage.Id, "label": randomImage.Label, "provider": randomImage.Provider})
	})

	router.GET("/annotation/data", func(c *gin.Context) {
		randomImage := getRandomAnnotatedImage()
		c.JSON(http.StatusOK, gin.H{"uuid": randomImage.Id, "label": randomImage.Label, "provider": randomImage.Provider, "annotations": randomImage.Annotations})
	})

	router.POST("/donate", func(c *gin.Context) {
		label := c.PostForm("label")

		file, header, err := c.Request.FormFile("file")
		if(err != nil){
			c.String(422, "Picture is missing")
			return
		}

		// Create a buffer to store the header of the file
        fileHeader := make([]byte, 512)

        // Copy the file header into the buffer
		if _, err := file.Read(fileHeader); err != nil {
			c.String(422, "Unable to detect MIME type")
			return
		}

		// set position back to start.
		if _, err := file.Seek(0, 0); err != nil {
			c.String(422, "Unable to detect MIME type")
			return
		}
		
		if(!filetype.IsImage(fileHeader)){
			c.String(422, "Unsopported MIME type detected")
			return
		}

		//check if image already exists by using an image hash
		hash, err := hashImage(file)
        if(err != nil){
        	fmt.Printf("%s\n", err.Error())
        	c.String(500, "Couldn't add photo - please try again later")
        	return 
        }
        exists, err := imageExists(hash)
        if(err != nil){
        	c.String(500, "Couldn't add photo - please try again later")
        	return
        }
        if(exists){
        	c.String(409, "Couldn't add photo - image already exists")
        	return
        }


        //image doesn't already exist, so save it and add it to the database
		uuid := uuid.NewV4().String()
		err = c.SaveUploadedFile(header, (*unverifiedDonationsDir + uuid))
		if(err != nil){
			c.String(500, "Couldn't add photo - please try again later")
        	return
		}

		err = addDonatedPhoto(uuid, hash, label)
		if(err != nil){
			c.String(500, "Couldn't add photo - please try again later")
			return	
		}

		c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", header.Filename))
	})

	router.GET("/data", func(c *gin.Context) {
		tags := ""
		params := c.Request.URL.Query()
		if temp, ok := params["tags"]; ok {
			tags = temp[0]
			jsonData, err := export(strings.Split(tags, ","))
			if(err == nil){
				c.JSON(http.StatusOK, jsonData)
				return
			} else{
				c.JSON(http.StatusInternalServerError, gin.H{"Error": "Couldn't export data"})
				return
			}
		} else {
			c.JSON(422, gin.H{"error": "No tags specified"})
			return
		}
	})

	router.POST("/annotate/:imageid", func(c *gin.Context) {
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

	router.POST("/donation/:imageid/validate/:param", func(c *gin.Context) {
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
		} else{
			c.JSON(http.StatusOK, nil)
			return
		}
	})

	router.POST("/annotation/:imageid/validate/:param", func(c *gin.Context) {
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
		} else{
			c.JSON(http.StatusOK, nil)
			return
		}
	})

	router.Run(":8080")
}
