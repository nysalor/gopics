package model

import (
	"os"
	"strconv"
	"strings"
	_ "fmt"
	"github.com/rwcarlsen/goexif/exif"
)

type Exif struct {
	Maker         string    `db:"maker"`
	Model         string    `db:"model"`
	LensMaker     string    `db:"lens_maker"`
	LensModel     string    `db:"lens_model"`
	TookAt        string    `db:"took_at"`
	FNumber       string    `db:"f_number"`
	FocalLength   string    `db:"focal_length"`
	Iso           string    `db:"iso"`
	ExposureTime  string    `db:"exposure"`
	Latitude      float64   `db:"latitude"`
	Longitude     float64   `db:"longitude"`
}

type LatLong struct {
	Latitude float64
	Longitude float64
}

func DecodeExif(path string) (ex Exif) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	x, err := exif.Decode(f)
	if err != nil {
		panic(err)
	}

	datetime, _ := x.DateTime()
	latlong := getLatLong(x)

	ex = Exif{
		Maker: getExifTag(x, exif.Make),
		Model: getExifTag(x, exif.Model),
		LensMaker:getExifTag(x, exif.LensMake),
		LensModel: getExifTag(x, exif.LensModel),
		TookAt: datetime.Format("2006-01-02 15:04:05"),
		FNumber: getFNumber(x),
		FocalLength: getFocal(x),
		Iso: getExifTag(x, exif.ISOSpeedRatings),
		ExposureTime: getExifTag(x, exif.ExposureTime),
		Latitude: latlong.Latitude,
		Longitude: latlong.Longitude,
	}

	return ex
}

func getExifTag(x *exif.Exif, fn exif.FieldName) string {
	tag, err := x.Get(fn)
	if err != nil {
		return ""
	}
	return strings.Replace(tag.String(), "\"", "", -1)
}

func getLatLong(x *exif.Exif) LatLong {
	lat, long, err := x.LatLong()
	if err != nil {
		return LatLong{}
	} else {
		return LatLong{Latitude: lat, Longitude: long}
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
