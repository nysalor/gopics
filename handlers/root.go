package handler

import (
	"time"
	"os"
	_ "log"
	_ "fmt"
	"net/http"
	"github.com/labstack/echo"
	"../model"
)

type indexResult struct {
	Albums []model.Album
}

func IndexPage() echo.HandlerFunc {
	return func(c echo.Context) error {
		dirs := loadDir(Config.TargetDir)
		albums := loadCache()

		reload := updateCache(dirs, albums)
		if reload {
			albums = loadCache()
		}

		return  c.JSON(http.StatusOK, indexResult{Albums: albums})
	}
}

func updateCache(dirs []os.FileInfo, albums []model.Album) bool {
	missingDirs := []string{}
	missingAlbums := []model.Album{}

	for _, dir := range dirs {
		missing := true
		for _, album := range albums {
			if album.DirName == dir.Name() {
				missing = false
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

	appendCache(missingDirs)
	removeCache(missingAlbums)

	if len(missingDirs) > 0 || len(missingAlbums) > 0 {
		return true
	} else {
		return false
	}
}

func appendCache(dirs []string) {
	for _, dir := range dirs {
		now :=  time.Now().Format("2006-01-02 15:04:05")
		album := model.Album{Name: dir, DirName: dir, UpdatedAt: now, CreatedAt: now}

		result, err := connection.NamedExec(`INSERT INTO albums (name, dirname, updated_at, created_at, images_count) VALUES (:name, :dirname, :updated_at, :created_at, :images_count)`,
			map[string]interface{} {
				"name": album.Name,
				"dirname": album.DirName,
				"updated_at": album.UpdatedAt,
				"created_at": album.CreatedAt,
				"images_count": 0,
			})
		if err != nil {
			return
		}
		albumId, err := result.LastInsertId()
		if err != nil {
			return
		}
		album.Id = albumId

		updateAlbum(album, []model.Image{})
	}
	return
}

func removeCache(albums []model.Album) {
	for _, album := range albums {
		_, err := connection.NamedExec("DELETE FROM albums WHERE id = :album_id", album.Id)
		_, err = connection.NamedExec("DELETE FROM images WHERE album_id = :album_id", album.Id)
		if err != nil {
			return
		}
	}
}

func loadCache() (albums []model.Album) {
	sql := "select id, name, dirname, images_count, updated_at, created_at from albums"
	err := connection.Select(&albums, sql)
	if err != nil {
		return
	}
	return
}
