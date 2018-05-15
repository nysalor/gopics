package handler

import (
	"regexp"
	"time"
	"os"
	"strconv"
	"path/filepath"
	"io/ioutil"
	"net/http"
	"github.com/labstack/echo"
	"github.com/rwcarlsen/goexif/exif"
)

type Result struct {
	Albums []Album
}

type Album struct {
	Name string
	Images []Image
}

type Image struct {
	Filename string
	Exif Exif
}

type Exif struct {
	Model string
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

func MainPage() echo.HandlerFunc {
	targetDir := "./images"
	albums := inspect(targetDir)

	return func(c echo.Context) error {
		return  c.JSON(http.StatusOK, Result{Albums: albums})
	}
}

func inspect(target string) []Album {
	var albums []Album
	r := regexp.MustCompile(`\.(jpg|jpeg|png|gif)$`)

	dirs, err := ioutil.ReadDir(target)
	if err != nil {
		return albums
	}

	for _, dir := range dirs {
		files, err := ioutil.ReadDir(filepath.Join(target, dir.Name()))
		if err != nil {
			continue
		}
		var images []Image
		for _, file := range files {
			if r.MatchString(file.Name()) {
				exif := decodeExif(filepath.Join(target, dir.Name(), file.Name()))
				images = append(images, Image{Filename: file.Name(), Exif: exif})

			}
		}
		albums = append(albums, Album{Name: dir.Name(), Images: images})
	}

	return albums
}

func decodeExif(path string) Exif {
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

	ex := Exif{
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
