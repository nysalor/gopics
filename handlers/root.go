package handler

import (
	_ "log"
	_ "fmt"
	"sync"
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

	var missingAlbums []model.Album
	var missingDirs []string
	var syncAlbums []model.Album

	for _, dir := range dirs {
		missing := true
		for _, album := range albums {
			if album.DirName == dir.Name() {
				missing = false
				syncAlbums = append(syncAlbums, album)
				break
			}
		}
		if missing {
			missingDirs = append(missingDirs, dir.Name())
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
			missingAlbums = append(missingAlbums, album)
		}
	}

	wg := &sync.WaitGroup{}
	c := make(chan bool, 10)

	for _, missingAlbum := range missingAlbums {
		if missingAlbum.Locked == 0 {
			c <- true
			wg.Add(1)
			go func(a model.Album) {
				defer func() { <-c }()
				DebugLog("removing: " + a.Name)
				a.Remove()
				wg.Done()
			}(missingAlbum)
		}
	}

	for _, missingDir := range missingDirs {
		c <- true
		wg.Add(1)
		go func(d string) {
			defer func() { <-c }()
			DebugLog("appending: " + d)
			album := model.AppendAlbum(d)
			if album.Id > 0 {
				DebugLog("created: " + album.Name)
				album.InitializeImages()
				DebugLog("initialized: " + album.Name)
				album.InitializeCover()
				album.UpdateText()
				album.Unlock()
				wg.Done()
			}
		}(missingDir)
	}

	wg.Wait()

	for _, syncAlbum := range syncAlbums {
		if syncAlbum.Locked == 0 {
			SyncAlbum(syncAlbum)
		}
	}


	return result
}
