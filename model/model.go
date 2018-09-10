package model

import (
	"time"
)

type Album struct {
	Id          int64     `db:"id"`
	Name        string    `db:"name"`
	DirName     string    `db:"dirname"`
	UpdatedAt   string    `db:"updated_at"`
	CreatedAt   string    `db:"created_at"`
	ImagesCount int       `db:"images_count"`
	Images      []Image
}

type Image struct {
	Id          int64     `db:"id"`
	AlbumId     int64     `db:"album_id"`
	Filename    string    `db:"filename"`
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
	Exif        Exif
}

type Exif struct {
	Maker string
	Model string
	LensMaker string
	LensModel string
	DateTime time.Time
	FNumber string
	FocalLength string
	Iso string
	LatLong LatLong
}

type LatLong struct {
	Latitude float64
	Longitude float64
}
