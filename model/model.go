package model

import (
	"os"
	"time"
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
	thumbnail = resizer.ResizeImage()
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

func DebugLog(str string) {
	if conf.Verbose {
		conf.Log.Info("[" + nowText() + "] " + str)
	}
	return
}
