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

type updateResult struct {
	Updated bool
}

func IndexPage() echo.HandlerFunc {
	return func(c echo.Context) error {
		albums := loadCache()
		res := UpdateDirs(albums)
		if res {
			albums = loadCache()
		}

		return  c.JSON(http.StatusOK, indexResult{Albums: albums})
	}
}

func UpdatePage() echo.HandlerFunc {
	return func(c echo.Context) error {
		albums := loadCache()
		res := UpdateDirs(albums)
		if res {
			albums = loadCache()
		}

		for _, album := range albums {
			images := loadImageCache(album)
			updateAlbum(album, images)
		}

		return  c.JSON(http.StatusOK, updateResult{Updated: res})
	}
}

func UpdateDirs(albums []model.Album) bool {
	dirs := loadDir(conf.TargetDir)
	res := updateCache(dirs, albums)
	return res
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

	resAppend := appendCache(missingDirs)
	resRemove := removeCache(missingAlbums)
	resText := updateAlbumTexts(albums)

	if resAppend || resRemove || resText {
		return true
	} else {
		return false
	}
}

func appendCache(dirs []string) (result bool) {
	result = false
	for _, dir := range dirs {
		now :=  time.Now().Format("2006-01-02 15:04:05")
		album := model.Album{Name: dir, DirName: dir, UpdatedAt: now, CreatedAt: now}

		res, err := connection.NamedExec(`INSERT INTO albums (name, dirname, updated_at, created_at, images_count) VALUES (:name, :dirname, :updated_at, :created_at, :images_count)`,
			map[string]interface{} {
				"name": album.Name,
				"dirname": album.DirName,
				"updated_at": album.UpdatedAt,
				"created_at": album.CreatedAt,
				"images_count": 0,
			})
		albumId, err := res.LastInsertId()
		if err != nil {
			return
		}
		album.Id = albumId

		rows, _ := res.RowsAffected()
		if err == nil && rows > 0 {
			updateAlbum(album, []model.Image{})
			result = true
		}
	}
	return
}

func removeCache(albums []model.Album) (result bool) {
	result = false
	for _, album := range albums {
		resAlbum := connection.MustExec("DELETE FROM albums WHERE id = ?", album.Id)
		rowsAlbum, _ := resAlbum.RowsAffected()

		resImage := connection.MustExec("DELETE FROM images WHERE album_id = ?", album.Id)
		rowsImage, _ := resImage.RowsAffected()

		if rowsAlbum > 0 || rowsImage > 0 {
			result = true
		}
	}
	return
}

func loadCache() (albums []model.Album) {
	sql := "select id, name, description, dirname, images_count, cover, thumbnail, updated_at, created_at from albums"

	rows, err := connection.Queryx(sql)
	if err != nil {
		return
	}

	for rows.Next() {
		album := model.Album{}
		err := rows.StructScan(&album)
		if err != nil {
			return
		}
		album.SetCoverUrl()
		album.SetThumbnailUrl()
		albums = append(albums, album)
	}

	return albums
}

func updateAlbumTexts(albums []model.Album) (result bool) {
	result = false
	for _, album := range albums {
		res := updateText(album)
		result = result || res
	}
	return
}
