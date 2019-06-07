package main

import(
	"database/sql"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"image"
    "os"
    _ "image/jpeg"
	_ "image/png"
)

type ImageRow struct {
	Id int64
	Uuid string
	Unlocked bool
}

func getImageDimension(imagePath string) (int, int) {
    file, err := os.Open(imagePath)
    if err != nil {
    	log.Fatal("Couldn't open file ", imagePath, " : ", err.Error())
    }
    defer file.Close()

    image, _, err := image.DecodeConfig(file)
    if err != nil {
    	log.Fatal("Couldn't decode file ", imagePath, " : ", err.Error())
    }
    return image.Width, image.Height
}

func addDimensions(tx *sql.Tx) error {
	rows, err := tx.Query(`SELECT id, key, unlocked FROM image`)
	if err != nil {
		return err
	}

	var imageRows []ImageRow
	for rows.Next() {
		var imageRow ImageRow
		err = rows.Scan(&imageRow.Id, &imageRow.Uuid, &imageRow.Unlocked)
		if err != nil {
			return err
		}

		imageRows = append(imageRows, imageRow)
	}

	rows.Close()

	for i, imageRow := range imageRows {
		path := ""
		if imageRow.Unlocked {
			path = "../donations/" + imageRow.Uuid
		} else {
			path = "../unverified_donations/" + imageRow.Uuid
		}
		imageWidth, imageHeight := getImageDimension(path)
		_, err = tx.Exec(`UPDATE image SET width = $1, height = $2 WHERE id = $3`, imageWidth, imageHeight, imageRow.Id)
		if err != nil {
			return err
		}
		log.Info("#", i, " Updated dimensions for image ", imageRow.Uuid, "(height:", imageHeight, ", width:", imageWidth, ")")
	}

	return nil
}

func main() {
	db, err := sql.Open("postgres", IMAGE_DB_CONNECTION_STRING)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}


	tx, err := db.Begin()
    if err != nil {
    	log.Fatal("Couldn't start transaction: ", err.Error())
    }

    err = addDimensions(tx)
    if err != nil {
    	tx.Rollback()
    	log.Fatal("Couldn't add dimensions: ", err.Error())
    }

    err = tx.Commit()
	if err != nil {
		log.Fatal("Couldn't commit changes: ", err.Error())
	}

}
