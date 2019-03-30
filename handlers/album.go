package handler

import (
	"regexp"
	"strconv"
	"time"
	"sync"
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

		if album.Locked == 0 {
			go SyncAlbum(album)
		} else {
			DebugLog("album locked:" + album.Name)
		}

		return  c.JSON(http.StatusOK, albumResult{Album: album})
	}
}

func SyncAlbum(album model.Album) {
	DebugLog("syncing: " + album.Name)
	currentImages := album.Images
	files := loadFiles(conf.TargetDir, album.DirName)
	r := regexp.MustCompile(`\.(jpg|jpeg|png|gif)$`)

	var missingImages []model.Image
	var missingFiles []string
	var updateImages []model.Image

	album.Lock()

	for _, image := range currentImages {
		missing := true
		for _, file := range files {
			if file.Name() == image.Filename {
				missing = false
				break
			}
		}
		if missing {
			DebugLog("missing file: " + image.Filename)
			missingImages = append(missingImages, image)
		}
	}


	for _,  file := range files {
		if r.MatchString(file.Name()) {
			missing := true
			for _, image := range currentImages {
				if image.Filename == file.Name() {
					DebugLog("found: " + image.Filename)
					missing = false
					mtime := file.ModTime()
					t, _ := time.Parse("2006-01-02 15:04:05", image.UpdatedAt)
					if mtime.Unix() >= t.Unix() {
						DebugLog("updated image: " + image.Filename)
						updateImages = append(updateImages, image)
					}
					break
				}
			}
			if missing {
				DebugLog("missing image: " + file.Name())
				missingFiles = append(missingFiles, file.Name())
			}
		}
	}


	c := make(chan bool, 10)
	wg := &sync.WaitGroup{}

	for _, missingImage := range missingImages {
		c <- true
		wg.Add(1)
		go func(i model.Image) {
			defer func() { <-c }()
			DebugLog("removing: " + i.Filename)
			i.Remove()
			wg.Done()
		}(missingImage)
	}

	for _, missingFile := range missingFiles {
		c <- true
		wg.Add(1)
		go func(f string) {
			defer func() { <-c }()
			image := album.AppendImage(f)
			DebugLog("appended: " + image.Filename)
			wg.Done()
		}(missingFile)
	}

	for _, updateImage := range updateImages {
		c <- true
		wg.Add(1)
		go func(i model.Image) {
			defer func() { <-c }()
			i.Update()
			DebugLog("upodated: " + i.Filename)
			wg.Done()
		}(updateImage)
	}

	wg.Wait()

	album.UpdateText()
	album.UpdateCount()
	album.Unlock()

	return
}
