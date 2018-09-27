package handler

import (
	"os"
	"time"
	"strings"
	"path/filepath"
	"io/ioutil"
	"../config"
	"../db"
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
