package model

import (
	"os"
	"time"
	"bufio"
	"regexp"
	"io/ioutil"
	"path/filepath"
)

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
	Locked       int       `db:"locked"`
	Images       []Image
}

func Albums() (albums []Album) {
	sql := "select id, name, description, dirname, images_count, cover, thumbnail, locked, updated_at, created_at from albums"

	rows, err := connection.Queryx(sql)
	if err != nil {
		return
	}

	for rows.Next() {
		album := Album{}
		err := rows.StructScan(&album)
		if err != nil {
			return
		}
		album.SetCoverUrl()
		album.SetThumbnailUrl()
		album.LoadImages()
		albums = append(albums, album)
	}

	return albums
}

func FindAlbum(id int) (album Album) {
	sql := "select id, name, dirname, images_count, updated_at, created_at from albums where id = ?"
	err := connection.Get(&album, sql, id)
	if err != nil {
		return
	}
	album.SetCoverUrl()
	album.SetThumbnailUrl()
	album.LoadImages()

	return
}

func FindAlbumByName(dirName string) (album Album) {
	sql := "select id, name, description, dirname, images_count, cover, thumbnail, locked, updated_at, created_at from albums where name = ?"
	err := connection.Get(&album, sql, dirName)
	if err != nil {
		return
	}
	album.SetCoverUrl()
	album.SetThumbnailUrl()
	album.LoadImages()

	return
}

func SearchAlbums(str string) (albums []Album) {
	sql := "select id, name, description, dirname, images_count, cover, thumbnail, locked, updated_at, created_at from albums where name like ? or description like ?"

	rows, err := connection.Queryx(sql, "%" + str + "%", "%" + str + "%")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		album := Album{}
		err := rows.StructScan(&album)
		if err != nil {
			return
		}
		album.SetCoverUrl()
		album.SetThumbnailUrl()
		album.LoadImages()
		albums = append(albums, album)
	}

	return albums
}

func AppendAlbum(dir string) (album Album) {
	now :=  time.Now().Format("2006-01-02 15:04:05")
	album = Album{Name: dir, DirName: dir, UpdatedAt: now, CreatedAt: now}

	res, err := connection.NamedExec(`INSERT INTO albums (name, dirname, updated_at, created_at, images_count, locked) VALUES (:name, :dirname, :updated_at, :created_at, :images_count, 1)`,
		map[string]interface{} {
			"name": album.Name,
			"dirname": album.DirName,
			"updated_at": album.UpdatedAt,
			"created_at": album.CreatedAt,
			"images_count": 0,
		})
	albumId, err := res.LastInsertId()
	if err != nil {
		return
	}
	album.Id = albumId
	return
}

func (album *Album) SetCoverUrl() {
	if album.Cover != "" {
		album.CoverUrl = conf.BaseUrl + "/" + filepath.Join(album.DirName, album.Cover)
	}
	return
}

func (album *Album) CoverPath() (path string) {
	if album.Cover != "" {
		path = filepath.Join(conf.TargetDir, album.DirName, album.Cover)
	}
	return
}

func (album *Album) SetThumbnailUrl() {
	if album.Thumbnail != "" {
		album.ThumbnailUrl = conf.ThumbnailUrl + "/" + album.Thumbnail
	}
	return
}

func (album *Album) LoadImages() {
	sql := "select id, album_id, filename, thumbnail, maker, model, lens_maker, lens_model, took_at, f_number, focal_length, iso, exposure, latitude, longitude, updated_at, created_at from images where album_id = ? order by took_at asc"
	rows, err := connection.Queryx(sql, album.Id)
	if err != nil {
		return
	}

	var images []Image

	for rows.Next() {
		image := Image{}
		err := rows.StructScan(&image)
		if err != nil {
			return
		}
		image.Album = *album
		image.SetUrl()
		image.SetThumbnailUrl()
		images = append(images, image)
	}

	album.Images = images

	return
}

func (album *Album) AppendImage(file string) (image Image) {
	now :=  nowText()
	image = Image{Filename: file, AlbumId: album.Id, Album: *album}
	image.MergeExif()
	thumbnail := createThumbnail(image.FilePath())
	res, err := connection.NamedExec(`INSERT INTO images (album_id, filename, thumbnail, maker, model, lens_maker, lens_model, f_number, focal_length, iso, exposure, latitude, longitude, took_at, updated_at, created_at) VALUES (:album_id, :filename, :thumbnail, :maker, :model, :lens_maker, :lens_model, :f_number, :focal_length, :iso, :exposure, :latitude, :longitude, :took_at, :updated_at, :created_at)`,
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
			"exposure": image.ExposureTime,
			"latitude": image.Latitude,
			"longitude": image.Longitude,
			"took_at": image.TookAt,
			"updated_at": now,
			"created_at": now,
		})
	rows, _ := res.RowsAffected()

	if err != nil &&  rows > 0 {
		return
	}

	DebugLog("append: " + image.Filename + " -> " + album.Name)
	return
}

func (album *Album) UpdateText() (result bool) {
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

func (album *Album) UpdateCount() (result bool) {
	result = false

	var count int
	err := connection.Get(&count, "SELECT count(*) FROM images WHERE album_id = ?", album.Id)
	if err != nil {
		return
	}

	if album.ImagesCount != count {
		album.ImagesCount = count
		_, err := connection.Exec("UPDATE albums SET images_count = ? WHERE id = ?", count, album.Id)
		if err != nil {
			return
		}
		result = true
	}
	return
}

func (album *Album) InitializeImages() bool {
	r := regexp.MustCompile(`\.(jpg|jpeg|png|gif)$`)

	files, err := ioutil.ReadDir(filepath.Join(conf.TargetDir, album.DirName))
	for _, file := range files {
		if !file.IsDir() {
			if r.MatchString(file.Name()) {
				album.AppendImage(file.Name())
			}
		}
	}

	if err != nil {
		return false
	}
	album.UpdateCount()

	return true
}

func (album *Album) InitializeCover() bool {
	sql := "select filename from images where album_id = ? order by took_at asc limit 1"
	rows := connection.QueryRow(sql, album.Id)
	var filename string
	rows.Scan(&filename)
	if filename != "" {
		album.Cover = filename
		thumbnail := createThumbnail(album.CoverPath())

		now :=  nowText()
		_, err := connection.Exec("update albums set cover = ?, thumbnail = ?, updated_at = ? WHERE id = ?", filename, thumbnail, now, album.Id)
		if err != nil {
			return false
		}
		return true
	}
	return false
}

func (album *Album) Remove() (result bool) {
	resAlbum := connection.MustExec("DELETE FROM albums WHERE id = ?", album.Id)
	rowsAlbum, _ := resAlbum.RowsAffected()
	resImage := connection.MustExec("DELETE FROM images WHERE album_id = ?", album.Id)
	rowsImage, _ := resImage.RowsAffected()

	if rowsAlbum > 0 || rowsImage > 0 {
		result = true
	}

	return
}

func (album *Album) Lock() {
	DebugLog("locked: " + album.Name)
	album.Locked = 1
	_, err := connection.Exec("update albums set locked = 1 WHERE id = ?", album.Id)
	if err != nil {
		panic(err)
	}
}

func (album *Album) Unlock() {
	DebugLog("unlocked: " + album.Name)
	album.Locked = 0
	_, err := connection.Exec("update albums set locked = 0 WHERE id = ?", album.Id)
	if err != nil {
		panic(err)
	}
}
