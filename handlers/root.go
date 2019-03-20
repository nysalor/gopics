package handler

import (
	_ "log"
	_ "fmt"
	"net/http"
	"github.com/labstack/echo"
	"../model"
)

type indexResult struct {
	Albums []model.Album
}

type updateResult struct {
	Updated bool
}

func IndexPage() echo.HandlerFunc {
	return func(c echo.Context) error {
		albums := model.Albums()
		go syncAlbums(albums)

		return  c.JSON(http.StatusOK, indexResult{Albums: albums})
	}
}

func UpdatePage() echo.HandlerFunc {
	return func(c echo.Context) error {
		albums := model.Albums()
		go syncAlbums(albums)

		return  c.JSON(http.StatusOK, indexResult{Albums: albums})
	}
}

func syncAlbums(albums []model.Album) (result bool) {
	result = false
	dirs := loadDir(conf.TargetDir)
	c := make(chan bool, 10)
	for _, dir := range dirs {
		missing := true
		for _, album := range albums {
			if album.DirName == dir.Name() {
				missing = false
				c <- true
				go syncAlbumHandler(c, album)
				break
			}
		}
		if missing {
			c <- true
			go appendAlbumHandler(c, dir.Name())
			result = true
		}
	}

	for _, album := range albums {
		missing := true
		for _, dir := range dirs {
			if dir.Name() == album.DirName {
				missing = false
				break
			}
		}
		if missing {
			album.Remove()
			result = true
		}
	}

	return result
}

func syncAlbumHandler(c chan bool, album model.Album) {
	defer func() { <-c }()
	SyncAlbum(album)
	return
}

func appendAlbumHandler(c chan bool, dirname string) {
	defer func() { <-c }()
	album := model.AppendAlbum(dirname)
	if album.Id > 0 {
		album.InitializeImages()
		album.InitializeCover()
		album.UpdateText()
	}
	return
}
