package handler

import (
	"regexp"
	"strconv"
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
	images = loadImageCache(albumId)
	return
}

func loadAlbumCache(albumId int) (album model.Album) {
	sql := "select id, name, dirname, images_count, updated_at, created_at from albums where id = ?"
	err := connection.Get(&album, sql, albumId)
	if err != nil {
		return
	}
	return
}

func loadImageCache(albumId int) (images []model.Image) {
	sql := "select id, album_id, filename, maker, model, lens_maker, lens_model, took_at, f_number, focal_length, iso, latitude, longitude from images where album_id = ?"
	err := connection.Select(&images, sql, albumId)
	if err != nil {
		return
	}
	return
}

func updateAlbum(album model.Album, images []model.Image) bool {
	additionalImages := []model.Image{}
	missingImages := []model.Image{}
	files := loadFile(Config.TargetDir, album.DirName)
	r := regexp.MustCompile(`\.(jpg|jpeg|png|gif)$`)

	for _,  file := range files {
		missing := true
		if r.MatchString(file.Name()) {
			for _, image := range images {
				if image.Filename == file.Name() {
					missing = false
					break
				}
			}
			if missing {
				newImage := mergeExif(model.Image{Filename: file.Name(), AlbumId: album.Id}, album)
				additionalImages = append(additionalImages, newImage)
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

	appendImage(additionalImages)
	removeImage(missingImages)

	if len(additionalImages) > 0 || len(missingImages) > 0 {
		return true
	} else {
		updateImageCount(album)
		return false
	}

}

func appendImage(images []model.Image) {
	now :=  time.Now().Format("2006-01-02 15:04:05")
	for _, image := range images {
		_, err := connection.NamedExec(`INSERT INTO images (album_id, filename, maker, model, lens_maker, lens_model, f_number, focal_length, iso, latitude, longitude, took_at, updated_at, created_at) VALUES (:album_id, :filename, :maker, :model, :lens_maker, :lens_model, :f_number, :focal_length, :iso, :latitude, :longitude, :took_at, :updated_at, :created_at)`,
			map[string]interface{} {
				"album_id": image.AlbumId,
				"filename": image.Filename,
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
		if err != nil {
			return
		}
	}
	return
}

func removeImage(images []model.Image) {
	for _, image := range images {
		_, err := connection.NamedExec("DELETE FROM images WHERE id = :image_id", image.Id)
		if err != nil {
			return
		}
	}
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

func mergeExif(image model.Image, album model.Album) model.Image {
	exif := decodeExif(filepath.Join(Config.TargetDir, album.DirName, image.Filename))

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
