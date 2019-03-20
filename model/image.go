package model

import (
	"path/filepath"
)

type Image struct {
	Id           int64     `db:"id"`
	UpdatedAt    string    `db:"updated_at"`
	CreatedAt    string    `db:"created_at"`
	AlbumId      int64     `db:"album_id"`
	Filename     string    `db:"filename"`
	Thumbnail    string    `db:"thumbnail"`
	Url          string
	ThumbnailUrl string
	Album        Album
	Exif
}

func (image *Image) SetUrl() {
	image.Url = conf.BaseUrl + "/" + filepath.Join(image.Album.DirName, image.Filename)
	return
}

func (image *Image) FilePath() (path string) {
	path = filepath.Join(conf.TargetDir, image.Album.DirName, image.Filename)
	return
}

func (image *Image) SetThumbnailUrl() {
	if image.Thumbnail != "" {
		image.ThumbnailUrl = conf.ThumbnailUrl + "/" + image.Thumbnail
	}
	return
}

func (image *Image) MergeExif() Image {
	exif := DecodeExif(image.FilePath())

	image.Maker = exif.Maker
	image.Model = exif.Model
	image.LensMaker = exif.LensMaker
	image.LensModel = exif.LensModel
	image.TookAt = exif.TookAt
	image.FNumber = exif.FNumber
	image.FocalLength = exif.FocalLength
	image.Iso = exif.Iso
	image.ExposureTime = exif.ExposureTime
	image.Latitude = exif.Latitude
	image.Longitude = exif.Longitude

	return *image
}

func (image *Image) Update() (result bool) {
	result = false

	now :=  nowText()
	image.MergeExif()
	thumbnail := createThumbnail(image.FilePath())
	res, err := connection.NamedExec(`UPDATE images SET thumbnail = :thumbnail, maker = :maker, model = :model, lens_maker = :lens_maker, lens_model = :lens_model, f_number = :f_number, focal_length = :focal_length, iso = :iso, exposure = :exposure, latitude = :latitude, longitude = :longitude, took_at = :took_at, updated_at = :updated_at WHERE id = :id`,
		map[string]interface{} {
			"id": image.Id,
			"thumbnail": thumbnail,
			"maker": image.Maker,
			"model": image.Model,
			"lens_maker": image.LensMaker,
			"lens_model": image.LensModel,
			"f_number": image.FNumber,
			"focal_length": image.FocalLength,
			"iso": image.Iso,
			"exposure": image.ExposureTime,
			"latitude": image.Latitude,
			"longitude": image.Longitude,
			"took_at": image.TookAt,
			"updated_at": now,
		})
	rows, _ := res.RowsAffected()

	if err == nil && rows > 0 {
		result = true
	}
	return
}

func (image *Image) Remove() (result bool) {
	result = false

	res, err := connection.NamedExec("DELETE FROM images WHERE id = :image_id", image.Id)
	rows, _ := res.RowsAffected()
	if err == nil && rows > 0 {
		result = true
	}
	return
}
