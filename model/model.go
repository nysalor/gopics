package model

import (
	"os"
	"path/filepath"
	"../config"
)

var conf config.Config

func Initialize(c config.Config) {
	conf = c
	conf.Log.SetOutput(os.Stdout)
}

type Album struct {
	Id           int64     `db:"id"`
	UpdatedAt    string    `db:"updated_at"`
	CreatedAt    string    `db:"created_at"`
	Name         string    `db:"name"`
	DirName      string    `db:"dirname"`
	Description  string    `db:"description"`
	ImagesCount  int       `db:"images_count"`
	Cover        string    `db:"cover"`
	CoverUrl     string
	Thumbnail    string    `db:"thumbnail"`
	ThumbnailUrl string
	Images       []Image
}

func (album *Album) SetCoverUrl() {
	album.CoverUrl = conf.BaseUrl + album.CoverPath()
	return
}

func (album *Album) CoverPath() (path string) {
	path = filepath.Join(conf.TargetDir, album.DirName, album.Cover)
	return
}

func (album *Album) SetThumbnailUrl() {
	album.ThumbnailUrl = conf.ThumbnailUrl + album.Thumbnail
	return
}

type Image struct {
	Id           int64     `db:"id"`
	UpdatedAt    string    `db:"updated_at"`
	CreatedAt    string    `db:"created_at"`
	AlbumId      int64     `db:"album_id"`
	Filename     string    `db:"filename"`
	Thumbnail    string    `db:"thumbnail"`
	Url          string
	ThumbnailUrl string
	Exif
}


func (image *Image) SetUrl(dirName string) {
	image.Url = conf.BaseUrl + dirName + "/" + image.Filename
	return
}

func (image *Image) FilePath(dirName string) (path string) {
	path = filepath.Join(conf.TargetDir, dirName, image.Filename)
	return
}

func (image *Image) SetThumbnailUrl() {
	image.ThumbnailUrl = conf.ThumbnailUrl + image.Thumbnail
	return
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
