package handler

import (
	"os"
	"strings"
	"path/filepath"
	"io/ioutil"
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

func loadFiles(target string, dir string) (files []os.FileInfo) {
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
