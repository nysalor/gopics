package model

import (
	"path/filepath"
	"../config"
)

var Config config.Config

type UrlString struct {
	string
}

type Album struct {
	Id          int64     `db:"id"`
	UpdatedAt   string    `db:"updated_at"`
	CreatedAt   string    `db:"created_at"`
	Name        string    `db:"name"`
	DirName     string    `db:"dirname"`
	Description string    `db:"description"`
	ImagesCount int       `db:"images_count"`
	Cover       string    `db:"cover"`
	CoverUrl    string
	Images      []Image
}

func (album *Album) SetCoverUrl(baseUrl string) {
	album.CoverUrl = filepath.Join(baseUrl, album.DirName, album.Cover)
	return
}

type Image struct {
	Id          int64     `db:"id"`
	UpdatedAt   string    `db:"updated_at"`
	CreatedAt   string    `db:"created_at"`
	AlbumId     int64     `db:"album_id"`
	Filename    string    `db:"filename"`
	Url         string
	Exif
}


func (image *Image) SetUrl(baseUrl string, dirName string) {
	image.Url = filepath.Join(baseUrl, dirName, image.Filename)
}

type Exif struct {
	Maker       string    `db:"maker"`
	Model       string    `db:"model"`
	LensMaker   string    `db:"lens_maker"`
	LensModel   string    `db:"lens_model"`
	TookAt      string    `db:"took_at"`
	FNumber     string    `db:"f_number"`
	FocalLength string    `db:"focal_length"`
	Iso         string    `db:"iso"`
	Latitude    float64   `db:"latitude"`
	Longitude   float64   `db:"longitude"`
}
