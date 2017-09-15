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
)

func main() {
	fmt.Printf("Starting Web Service...\n")

	log.SetLevel(log.DebugLevel)

	fmt.Printf("Setting environment variable for sentry\n")
	os.Setenv("SENTRY_DSN", WEB_SENTRY_DSN)

	releaseMode := flag.Bool("release", false, "Run in release mode")
	wordlistDir := flag.String("wordlist", "../wordlists/en/misc.txt", "Path to wordlist")
	donationsDir := flag.String("donations_dir", "../donations/", "Location of the uploaded donations")
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
	words, err := getWordLists(*wordlistDir)
	if(err != nil){
		fmt.Printf("Couldn't read wordlists...terminating!")
		log.Fatal(err)
	}


	tmpl := template.Must(template.New("main").Funcs(funcMap).ParseGlob(*htmlDir + "*"))


	router := gin.Default()
	router.SetHTMLTemplate(tmpl)
	router.Static("./js", "../js") //serve javascript files
	router.Static("./css", "../css") //serve css files
	router.Static("./donations", *donationsDir) //serve images
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "ImageMonkey",
			"activeMenuNr": 1,
		})
	})
	router.GET("/donate", func(c *gin.Context) {
		c.HTML(http.StatusOK, "donate.html", gin.H{
			"title": "Donate",
			"randomWord": words[random(0, len(words) - 1)],
			"activeMenuNr": 2,
		})
	})
	router.GET("/verify", func(c *gin.Context) {
		c.HTML(http.StatusOK, "validate.html", gin.H{
			"title": "Validate",
			"randomImage": getRandomImage(),
			"activeMenuNr": 3,
		})
	})

	router.GET("/export", func(c *gin.Context) {
		c.HTML(http.StatusOK, "export.html", gin.H{
			"title": "Export",
			"words": words,
			"activeMenuNr": 5,
		})
	})
	router.GET("/explore", func(c *gin.Context) {
		c.HTML(http.StatusOK, "explore.html", gin.H{
			"title": "Explore",
			"words": words,
			"activeMenuNr": 4,
			"graphData": explore(),
		})
	})

	router.POST("/validate/:imageid/:param", func(c *gin.Context) {
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

	router.GET("/validate", func(c *gin.Context) {
		randomImage := getRandomImage()
		c.JSON(http.StatusOK, gin.H{"uuid": randomImage.Id, "url": randomImage.Url, "label": randomImage.Label, "provider": randomImage.Provider})
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

		uuid := uuid.NewV4().String()
		err = c.SaveUploadedFile(header, (*donationsDir + uuid))
		if(err != nil){
			c.String(500, "Couldn't add photo - please try again later")
        	return
		}

		err = addDonatedPhoto(uuid, label)
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

	router.Run(":8080")
}
