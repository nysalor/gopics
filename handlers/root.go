package handler

import (
	"regexp"
	"time"
	"os"
	"strconv"
	_ "log"
	_ "fmt"
	"path/filepath"
	"io/ioutil"
	"net/http"
	"github.com/labstack/echo"
	"github.com/rwcarlsen/goexif/exif"
	"../db"
)

type Result struct {
	Albums []Album
}

type Album struct {
	Id          int64     `db:"id"`
	Name        string    `db:"name"`
	DirName     string    `db:"dirname"`
	UpdatedAt   time.Time `db:"updated_at"`
	CreatedAt   time.Time `db:"created_at"`
	ImagesCount int       `db:"images_count"`
}

type Image struct {
	Id          int64     `db:"id"`
	AlbumId     int64     `db:"album_id"`
	Filename    string    `db:"filename"`
	Model       string    `db:"model"`
	LensModel   string    `db:"lens_model"`
	TookAt      time.Time `db:"took_at"`
	FNumber     string    `db:"f_number"`
	FocalLength string    `db:"focal_length"`
	Iso         string    `db:"iso"`
	Latitude    float64   `db:"latitude"`
	Longitude   float64   `db:"longitude"`
	Exif        Exif
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

var connection = db.Connection

func MainPage(targetDir string) echo.HandlerFunc {
	dirs := loadDir(targetDir)
	albums := loadCache()

	reload := updateCache(dirs, albums)
	if reload {
		albums = loadCache()
	}

	return func(c echo.Context) error {
		return  c.JSON(http.StatusOK, Result{Albums: albums})
	}
}

func loadDir(target string) (dirs []os.FileInfo) {
	files, err := ioutil.ReadDir(target)
	for _, file := range files {
		if file.IsDir() {
			dirs = append(dirs, file)
		}
	}
		
	if err != nil {
		return nil
	}
	return
}

func loadFile(target string) (filess []os.FileInfo) {
	files, err := ioutil.ReadDir(target)
	for _, file := range files {
		if !file.IsDir() {
			files = append(files, file)
		}
	}
		
	if err != nil {
		return nil
	}
	return
}

func updateCache(dirs []os.FileInfo, albums []Album) bool {
	missingDirs := []string{}
	missingAlbums := []Album{}

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

	appendCache(missingDirs)
	removeCache(missingAlbums)

	return false
}

func appendCache(dirs []string) {
	for _, dir := range dirs {
		album := Album{Name: dir}
		images := inspectImages(dir)

		result, err := connection.NamedExec(`INSERT INTO albums (name, dirname, updated_at, created_at, images_count) VALUES (:name, :dirname, :updated_at, :created_at, :images_count)`,
			map[string]interface{} {
				"name": album.Name,
				"dirname": album.DirName,
				"updated_at": album.UpdatedAt,
				"created_at": album.CreatedAt,
				"images_count": len(images),
			})
		albumId, err := result.LastInsertId()
		for _, image := range images {
			image.AlbumId = albumId
			appendImage(image)
		}
		if err != nil {
			return
		}
	}
	return
}

func appendImage(image Image) (imageId int) {
	_, err := connection.NamedExec(`INSERT INTO images (album_id, filename, model, lens_model, datetime, f_number, focal_length, iso, lat_long, updated_at, created_at) VALUES (:album_id, :filename, :model, :lens_model, :datetime, :f_number, :focal_length, :iso, :latitude, :longitude, :took_at, :updated_at, :created_at)`,
		map[string]interface{} {
			"album_id": image.AlbumId,
			"filename": image.Filename,
			"model": image.Exif.Model,
			"lens_model": image.Exif.LensModel,
			"f_number": image.Exif.FNumber,
			"focal_length": image.Exif.FocalLength,
			"iso": image.Exif.Iso,
			"latitude": image.Exif.LatLong.Latitude,
			"longitude": image.Exif.LatLong.Longitude,
			"took_at": image.Exif.DateTime,
		})
	if err != nil {
		return
	}
	return
}

func removeCache(albums []Album) {
	for _, album := range albums {
		_, err := connection.NamedExec("DELETE FROM albums WHERE id = :album_id", album.Id)
		_, err = connection.NamedExec("DELETE FROM images WHERE album_id = :album_id", album.Id)
		if err != nil {
			return
		}
	}
}

func loadCache() (albums []Album) {
	sql := "select id, name, dirname, images_count, updated_at from albums"
	err := connection.Select(&albums, sql)
	if err != nil {
		return
	}
	return
}

func inspectTree(target string) (albums []Album) {
	dirs, err := ioutil.ReadDir(target)
	if err != nil {
		return nil
	}
	for _, dir := range dirs {
		path := filepath.Join(target, dir.Name())
		finfo, _ := os.Stat(path)
		if finfo.IsDir() {
			albums = append(albums, Album{Name: dir.Name()})
		}
	}
	return albums
}

func inspectImages(target string) (images []Image) {
	r := regexp.MustCompile(`\.(jpg|jpeg|png|gif)$`)
	files := loadDir(target)
	for _, file := range files {
		if r.MatchString(file.Name()) {
			exif := decodeExif(filepath.Join(target, file.Name()))
			images = append(images, Image{Filename: file.Name(), Exif: exif})
		}
	}
	return
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
