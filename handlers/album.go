package handler

import (
	"regexp"
	"strconv"
	"os"
	"bufio"
	"time"
	"path/filepath"
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
		album, images := loadAlbum(id)

		reload := updateAlbum(album, images)
		if reload {
			album, images = loadAlbum(id)
		}

		album.Images = images

		return  c.JSON(http.StatusOK, albumResult{Album: album})
	}
}

func loadAlbum(albumId int) (album model.Album, images []model.Image) {
	album = loadAlbumCache(albumId)
	images = loadImageCache(album)
	return
}

func loadAlbumCache(albumId int) (album model.Album) {
	sql := "select id, name, dirname, images_count, updated_at, created_at from albums where id = ?"
	err := connection.Get(&album, sql, albumId)
	if err != nil {
		return
	}
	album.SetCoverUrl()
	album.SetThumbnailUrl()

	return
}

func loadImageCache(album model.Album) (images []model.Image) {
	sql := "select id, album_id, filename, maker, model, lens_maker, lens_model, took_at, f_number, focal_length, iso, latitude, longitude, updated_at, created_at from images where album_id = ? order by took_at asc"
	rows, err := connection.Queryx(sql, album.Id)
	if err != nil {
		return
	}

	for rows.Next() {
		image := model.Image{}
		err := rows.StructScan(&image)
		if err != nil {
			return
		}
		image.SetUrl(conf.BaseUrl, album.DirName)
		images = append(images, image)
	}
	return
}

func updateAlbum(album model.Album, images []model.Image) bool {
	newFiles := []string{}
	missingImages := []model.Image{}
	updateImages := []model.Image{}
	files := loadFile(conf.TargetDir, album.DirName)
	r := regexp.MustCompile(`\.(jpg|jpeg|png|gif)$`)

	for _,  file := range files {
		missing := true
		if r.MatchString(file.Name()) {
			for _, image := range images {
				if image.Filename == file.Name() {
					missing = false
					mtime := file.ModTime()
					t, _ := time.Parse("2006-01-02 15:04:05", image.UpdatedAt)
					if mtime.Unix() >= t.Unix() {
						updateImages = append(updateImages, image)
					}
					break
				}
			}
			if missing {
				newFiles = append(newFiles, file.Name())
			}
		}
	}

	for _, image := range images {
		missing := true
		for _, file := range files {
			if file.Name() == image.Filename {
				missing = false
				break
			}
		}
		if missing {
			missingImages = append(missingImages, image)
		}
	}

	resAppend := appendImage(newFiles, album)
	resRemove := removeImage(missingImages)
	resText := updateText(album)
	_ = updateImage(updateImages)

	if resAppend || resRemove {
		updateImageCount(album)
		if album.Cover == "" {
			initializeCover(album)
		}
		return true
	}

	if resText {
		return true
	}

	return false
}

func appendImage(files []string, album model.Album) (result bool) {
	result = false
	now :=  nowText()
	for _, file := range files {
		image := mergeExif(model.Image{Filename: file, AlbumId: album.Id}, album)
		thumbnail := createThumbnail(image.FilePath(album.DirName))
		res, err := connection.NamedExec(`INSERT INTO images (album_id, filename, thumbnail, maker, model, lens_maker, lens_model, f_number, focal_length, iso, latitude, longitude, took_at, updated_at, created_at) VALUES (:album_id, :filename, :thumbnail, :maker, :model, :lens_maker, :lens_model, :f_number, :focal_length, :iso, :latitude, :longitude, :took_at, :updated_at, :created_at)`,
			map[string]interface{} {
				"album_id": image.AlbumId,
				"filename": image.Filename,
				"thumbnail": thumbnail,
				"maker": image.Maker,
				"model": image.Model,
				"lens_maker": image.LensMaker,
				"lens_model": image.LensModel,
				"f_number": image.FNumber,
				"focal_length": image.FocalLength,
				"iso": image.Iso,
				"latitude": image.Latitude,
				"longitude": image.Longitude,
				"took_at": image.TookAt,
				"updated_at": now,
				"created_at": now,
			})
		rows, _ := res.RowsAffected()
		if err == nil &&  rows > 0 {
			result = true
		}
	}
	return
}

func removeImage(images []model.Image) (result bool) {
	result = false
	for _, image := range images {
		res, err := connection.NamedExec("DELETE FROM images WHERE id = :image_id", image.Id)
		rows, _ := res.RowsAffected()
		if err == nil && rows > 0 {
			result = true
		}
	}
	return
}

func updateImage(images []model.Image) (result bool) {
	result = false
	now :=  nowText()
	for _, image := range images {
		res, err := connection.NamedExec(`UPDATE images SET maker = :maker, model = :model, lens_maker = :lens_maker, lens_model = :lens_model, f_number = :f_number, focal_length = :focal_length, iso = :iso, latitude = :latitude, longitude = :longitude, took_at = :took_at, updated_at = :updated_at WHERE id = :id`,
			map[string]interface{} {
				"id": image.Id,
				"maker": image.Maker,
				"model": image.Model,
				"lens_maker": image.LensMaker,
				"lens_model": image.LensModel,
				"f_number": image.FNumber,
				"focal_length": image.FocalLength,
				"iso": image.Iso,
				"latitude": image.Latitude,
				"longitude": image.Longitude,
				"took_at": image.TookAt,
				"updated_at": now,
			})
		rows, _ := res.RowsAffected()
		if err == nil && rows > 0 {
			result = true
		}
	}
	return
}

func updateImageCount(album model.Album) bool {
	var count int
	err := connection.Get(&count, "SELECT count(*) FROM images WHERE album_id = ?", album.Id)
	if err != nil {
		return false
	}
	if album.ImagesCount != count {
		album.ImagesCount = count
		_, err := connection.Exec("UPDATE albums SET images_count = ? WHERE id = ?", count, album.Id)
		if err != nil {
			return false
		}
		return true
	}
	return false
}

func initializeCover(album model.Album) bool {
	sql := "select filename from images where album_id = ? order by took_at asc limit 1"
	rows := connection.QueryRow(sql, album.Id)
	var filename string
	rows.Scan(&filename)
	album.Cover = filename
	thumbnail := createThumbnail(album.CoverPath())

	if filename != "" {
		now :=  nowText()
		_, err := connection.Exec("update albums set cover = ?, thumbnail = ?, updated_at = ? WHERE id = ?", filename, thumbnail, now, album.Id)
		if err != nil {
			return false
		}

		return true
	}
	return false
}

func mergeExif(image model.Image, album model.Album) model.Image {
	exif := decodeExif(filepath.Join(conf.TargetDir, album.DirName, image.Filename))

	image.Maker = exif.Maker
	image.Model = exif.Model
	image.LensMaker = exif.LensMaker
	image.LensModel = exif.LensModel
	image.TookAt = exif.TookAt
	image.FNumber = exif.FNumber
	image.FocalLength = exif.FocalLength
	image.Iso = exif.Iso
	image.Latitude = exif.Latitude
	image.Longitude = exif.Longitude

	return image
}

func updateText(album model.Album) bool {
	path := filepath.Join(conf.TargetDir, album.DirName, "album.txt")

	f, err := os.Open(path)
	if err != nil {
		return false
	}

	lines := make([]string, 0, 5)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	name := lines[0]
	description := lines[1]
	cover := lines[2]

	if cover != "" {
		_, err := os.Stat(filepath.Join(conf.TargetDir, album.DirName, cover))
		if os.IsNotExist(err) {
			cover = ""
		}
	}

	if ((album.Name != name) || (album.Description != description) || (album.Cover != cover)) {
		album.Name = name
		album.Description = description
		album.Cover = cover
		thumbnail := createThumbnail(album.CoverPath())
		now :=  nowText()
		_, err = connection.Exec("update albums set name = ?, description = ?, cover = ?, thumbnail = ?, updated_at = ? WHERE id = ?", name, description, cover, thumbnail, now, album.Id)
		if err != nil {
			panic(err)
		}
	}

	return true
}
