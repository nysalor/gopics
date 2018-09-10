package handler

import (
	"os"
	"strconv"
	"path/filepath"
	"net/http"
	"github.com/labstack/echo"
	"github.com/rwcarlsen/goexif/exif"
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
		Config.Log.Fatal(err)
		return
	}
	return
}

func loadImageCache(albumId int) (images []model.Image) {
	sql := "select id, album_id, filename, maker, model, lens_maker, lens_model, took_at, f_number, focal_length, iso, latitude, longitude from images where album_id = ?"
	err := connection.Select(&images, sql, albumId)
	if err != nil {
		Config.Log.Fatal(err)
		return
	}
	return
}

func updateAlbum(album model.Album, images []model.Image) bool {
	additionalImages := []model.Image{}
	missingImages := []model.Image{}
	files := loadFile(Config.TargetDir, album.DirName)

	for _,  file := range files {
		missing := true
		for _, image := range images {
			if image.Filename == file.Name() {
				missing = false
				break
			}
		}
		if missing {
			exif := decodeExif(filepath.Join(Config.TargetDir, album.DirName, file.Name()))
			additionalImages = append(additionalImages, model.Image{Filename: file.Name(), Exif: exif})
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
		return false
	}
	
}

func appendImage(images []model.Image) {
	for _, image := range images {
		_, err := connection.NamedExec(`INSERT INTO images (album_id, filename, maker, model, lens_maker, lens_model, f_number, focal_length, iso, latitude, longitude, took_at) VALUES (:album_id, :filename, :maker, :model, :lens_maker, :lens_model, :f_number, :focal_length, :iso, :latitude, :longitude, :took_at)`,
			map[string]interface{} {
				"album_id": image.AlbumId,
				"filename": image.Filename,
				"maker": image.Exif.Maker,
				"model": image.Exif.Model,
				"lens_maker": image.Exif.LensMaker,
				"lens_model": image.Exif.LensModel,
				"f_number": image.Exif.FNumber,
				"focal_length": image.Exif.FocalLength,
				"iso": image.Exif.Iso,
				"latitude": image.Exif.LatLong.Latitude,
				"longitude": image.Exif.LatLong.Longitude,
				"took_at": image.Exif.DateTime,
			})
		if err != nil {
			Config.Log.Fatal(err)
			return
		}
	}
	return
}

func removeImage(images []model.Image) {
	for _, image := range images {
		_, err := connection.NamedExec("DELETE FROM images WHERE id = :image_id", image.Id)
		if err != nil {
			Config.Log.Fatal(err)
			return
		}
	}
}

func decodeExif(path string) model.Exif {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	x, err := exif.Decode(f)
	if err != nil {
		panic(err)
	}

	latlong := getLatLong(x)
	datetime, _ := x.DateTime()
	flength := getFocal(x)
	fnumber := getFNumber(x)

	ex := model.Exif{
		Model: getExifTag(x, exif.Model),
		LensModel: getExifTag(x, exif.LensModel),
		DateTime: datetime,
		FNumber: fnumber,
		FocalLength: flength,
		Iso: getExifTag(x, exif.ISOSpeedRatings),
		LatLong: latlong,
	}

	return ex
}

func getExifTag(x *exif.Exif, fn exif.FieldName) string {
	tag, err := x.Get(fn)
	if err != nil {
		return ""
	}

	return tag.String()
}

func getLatLong(x *exif.Exif) model.LatLong {
	lat, long, err := x.LatLong()
	if err != nil {
		return model.LatLong{}
	} else {
		return model.LatLong{Latitude: lat, Longitude: long}
	}
}

func getFocal(x *exif.Exif) (flength string) {
	focal, _ := x.Get(exif.FocalLength)
	n, d, err := focal.Rat2(0)
	number := n / d
	if err != nil {
		flength = ""
	} else {

		flength = strconv.FormatInt(number, 10)
	}

	return flength
}

func getFNumber(x *exif.Exif) (string) {
	fn, err := x.Get(exif.FNumber)
	if err != nil {
		return ""
	}

	n, d, err := fn.Rat2(0)
	if err != nil {
		return ""
	}

	if err != nil {
		return ""
	}

	return strconv.FormatFloat(float64(n) / float64(d), 'f', 1, 64)
}
