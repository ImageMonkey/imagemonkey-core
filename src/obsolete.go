package main
/*// Common return format.
type Common struct {
	Stat    string `json:"stat"`
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

// Photo in flickr.photos.search
type Photo struct {
	Id       string `json:"id"`
	Owner    string `json:"owner"`
	Title    string `json:"title"`
	Secret   string `json:"secret"`
	Server   string `json:"server"`
	Farm     int64  `json:"farm"`
	Ispublic int64  `json:"ispublic"`
	Isfriend int64  `json:"isfriend"`
	Isfamily int64  `json:"isfamily"`
}

// PhotosSearch in flickr.photos.search
type PhotosSearch struct {
	Photos struct {
		Page    int         `json:"page"`
		Pages   int         `json:"pages"`
		Perpage int         `json:"perpage"`
		Total   interface{} `json:"total"`
		Photo   []Photo     `json:"photo"`
	} `json:"photos"`
	//Stat   string `json:"stat"`
	Common
}*/

/*func searchFlickrByTag(tag string) (PhotosSearch, error) {
	var data PhotosSearch
	url := "https://api.flickr.com/services/rest/?method=flickr.photos.search&api_key=3131b9e69b232359ee2ebaca98565d28&tags=" + tag + "&is_commons=&format=json&nojsoncallback=1"
	fmt.Printf("url = %s", url)

	res, err := http.Get(url)
	if err != nil {
	  log.Fatal(err)
	  return data, err
	}

	// read body
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
  		log.Fatal(err)
  		return data, err
	}

	if res.StatusCode != 200 {
		fmt.Printf("Unexpected status code\n")
  		log.Fatal("Unexpected status code", res.StatusCode)
  		return data, err
	}

	//log.Printf("Body: %s\n", body)
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		fmt.Printf("Couldn't Unmarshal\n")
		log.Fatal(err)
		return data, err
	}

	//fmt.Printf("contents of decoded json is: %#v\r\n", data)
	//fmt.Printf("Done: %d\n", len(data.Photos.Photo))
	return data, nil
}*/


/*func nextFlickrPhotoUrl() Image {
	var image Image
	image.Url = ""
	image.Id = ""
	image.Provider = "flickr"

	res, err := searchFlickrByTag("flower")
	if(err == nil){
		//fmt.Printf("Is nil\n")
		numberOfPhotos := len(res.Photos.Photo)
		//fmt.Printf("%+v", res.Photos)
		if(numberOfPhotos != 0){
			//fmt.Printf("numberOfPhotos is not 0\n")
			randomNumber := random(0, numberOfPhotos - 1)
	    	if(randomNumber < numberOfPhotos){ //jut a safety check
	    		photo := res.Photos.Photo[randomNumber]
	    		image.Url = "https://farm" + strconv.FormatInt(photo.Farm, 10) + ".staticflickr.com/" + photo.Server + "/" + photo.Id + "_" + photo.Secret + "_b.jpg"
	    		image.Id = photo.Id
	    		fmt.Printf("url = %s\n", image.Url)
	    	}
	    }
	}
	fmt.Printf("UUUUrl = %s\n", image.Url)
	return image
}*/

/*func labelImage(label string, imageIdentifier string, imageProvider string){
	//open database
	db, err := sql.Open("postgres", dbConnectionString)
	if err != nil {
		log.Fatal(err)
		return
	}
		

	//get or insert label
	rows, err := db.Query("SELECT id FROM label WHERE name = $1", label)
	if(err != nil){
		log.Fatal(err)
		return
	}

	labelId := 0
	if(rows.Next()){
		err := rows.Scan(&labelId)
		if(err != nil){
			log.Fatal(err)
			return
		}
	} else{ //insert
		err := db.QueryRow("INSERT INTO label(name) VALUES($1) RETURNING id", label).Scan(&labelId)
		if(err != nil){
			log.Fatal(err)
			return
		}
	}

	//get or insert image
	rows, err = db.Query("SELECT id FROM image WHERE key = $1", label)
	if(err != nil){
		log.Fatal(err)
		return
	}

	imageId := 0
	if(rows.Next()){
		err := rows.Scan(&imageId)
		if(err != nil){
			log.Fatal(err)
			return
		}
	} else{ //insert
		err := db.QueryRow("INSERT INTO image(key, image_provider_id) SELECT $1, p.id FROM image_provider p WHERE p.name = $2 RETURNING id", 
							imageIdentifier, imageProvider).Scan(&imageId)
		if(err != nil){
			log.Fatal(err)
			return
		}
	}

	imageClassificationId := 0
	//insert image classification
	err = db.QueryRow("INSERT INTO image_classification(image_id, label_id) VALUES($1, $2) RETURNING id", imageId, labelId).Scan(&imageClassificationId)
	if(err != nil){
		log.Fatal(err)
		return
	}
}*/