package handler

import (
	"os"
	"time"
	"strings"
	"path/filepath"
	"io/ioutil"
	"crypto/md5"
	"encoding/hex"
	"github.com/jmoiron/sqlx"
	"../config"
	"../db"
)

var connection *sqlx.DB
var conf config.Config

func Initialize(c config.Config) {
	conf = c
	conf.Log.SetOutput(os.Stdout)
	connection = db.Connect(conf.DB)
}

func loadDir(target string) (dirs []os.FileInfo) {
	files, err := ioutil.ReadDir(target)
	for _, file := range files {
		if file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
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

func nowText() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func createThumbnail(path string) (thumbnail string) {
	resizer := Resizer{
		OrigPath: path,
		OutDir: conf.CacheDir,
		Width: 640,
		Height: 480,
	}
	thumbnail = resizeImage(resizer)
	return
}

func checkSum(path string) (checkSum string) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(file)
	h := md5.Sum(data)
	checkSum = hex.EncodeToString(h[:])
	return
}
