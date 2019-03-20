package handler

import (
	"regexp"
	"strconv"
	"time"
	"net/http"
	"github.com/labstack/echo"
	"../model"
)

type albumResult struct {
	Album model.Album
}

func AlbumPage() echo.HandlerFunc {
	return func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		album := model.FindAlbum(id)

		go SyncAlbum(album)

		return  c.JSON(http.StatusOK, albumResult{Album: album})
	}
}

func SyncAlbum(album model.Album) {
	currentImages := album.Images
	files := loadFiles(conf.TargetDir, album.DirName)
	r := regexp.MustCompile(`\.(jpg|jpeg|png|gif)$`)

	c := make(chan bool, 10)
	for _, image := range currentImages {
		missing := true
		for _, file := range files {
			if file.Name() == image.Filename {
				missing = false
				break
			}
		}
		if missing {
			c <- true
			go removeImageHandler(c, image)
		}
	}


	for _,  file := range files {
		if r.MatchString(file.Name()) {
			missing := true
			for _, image := range currentImages {
				if image.Filename == file.Name() {
					missing = false
					mtime := file.ModTime()
					t, _ := time.Parse("2006-01-02 15:04:05", image.UpdatedAt)
					if mtime.Unix() >= t.Unix() {
						c <- true
						go updateImageHandler(c, image)
					}
					break
				}
			}
			if missing {
				c <- true
				go appendImageHandler(c, album, file.Name())
			}
		}
	}

	album.UpdateText()
	album.UpdateCount()

	return
}

func updateImageHandler(c chan bool, image model.Image) {
	defer func() { <-c }()
	image.Update()
}

func appendImageHandler(c chan bool, album model.Album, filename string) {
	defer func() { <-c }()
	album.AppendImage(filename)
}

func removeImageHandler(c chan bool, image model.Image) {
	defer func() { <-c }()
	image.Remove()
}
