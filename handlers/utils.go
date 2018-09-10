package handler

import (
	"os"
	"regexp"
	"path/filepath"
	"io/ioutil"
	"../config"
	"../db"
	"../model"
)

var connection = db.Connection
var Config config.Config

func Initialize(conf config.Config) {
	Config = conf
	Config.Log.SetOutput(os.Stdout)
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

func loadFile(target string, dir string) (files []os.FileInfo) {
	entries, err := ioutil.ReadDir(filepath.Join(target, dir))
	for _, file := range entries {
		if !file.IsDir() {
			files = append(files, file)
		}
	}
		
	if err != nil {
		return nil
	}
	return
}

func inspectImages(target string) (images []model.Image) {
	r := regexp.MustCompile(`\.(jpg|jpeg|png|gif)$`)
	files := loadDir(target)
	for _, file := range files {
		if r.MatchString(file.Name()) {
			exif := decodeExif(filepath.Join(target, file.Name()))
			images = append(images, model.Image{Filename: file.Name(), Exif: exif})
		}
	}
	return
}

/*
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
*/

