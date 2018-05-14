package handler

import (
	"regexp"
	"io/ioutil"
	"net/http"
	"github.com/labstack/echo"
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
		files, err := ioutil.ReadDir(target + "/" + dir.Name())
		if err != nil {
			continue
		}
		var images []Image
		for _, file := range files {
			if r.MatchString(file.Name()) {
				images = append(images, Image{Filename: file.Name()})
			}
		}
		albums = append(albums, Album{Name: dir.Name(), Images: images})
	}

	return albums
}
