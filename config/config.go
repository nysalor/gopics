package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

type Config struct {
	TargetDir    string
	CacheDir     string
	Port         int
	Host         string
	BaseUrl      string
	ThumbnailUrl string
	Log          *logrus.Logger
	DB           Database
	Verbose      bool
}

type Database struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

func (db Database) Url() (url string) {
	url = fmt.Sprintf("%s:%s@(%s:%s)/%s", db.User, db.Password, db.Host, db.Port, db.Name)
	return
}
