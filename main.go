package main

import (
	"flag"
	"fmt"
	"strings"
	"os"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"./config"
	"./handlers"
	"./model"
)

var dir = flag.String("d", "./images", "target dir")
var cacheDir = flag.String("c", "./.cache", "cache dir")
var port = flag.Int("p", 5000, "server port number")
var host = flag.String("h", "localhost", "server port number")  
var baseUrl = flag.String("u", "http://localhost", "base url for images")
var thumbnailUrl = flag.String("t", "http://localhost/cache", "base url for thumbnail")

func main() {
	flag.Parse()
	baseUrlTrim := strings.Trim(*baseUrl, "/")
	thumbnailUrlTrim := strings.Trim(*thumbnailUrl, "/")

	dbConf := config.Database{
		User: getEnv("DB_USER", "gopics"),
		Password: getEnv("DB_PASSWORD", ""),
		Host: getEnv("DB_HOST", "localhost"),
		Port: getEnv("DB_PORT", "3306"),
		Name: getEnv("DB_NAME", "gopics"),
	}

	conf := config.Config{
		TargetDir: *dir,
		CacheDir: *cacheDir,
		Port: *port,
		Host: *host,
		BaseUrl: baseUrlTrim,
		ThumbnailUrl: thumbnailUrlTrim,
		Log: logrus.New(),
		DB: dbConf,
	}

	handler.Initialize(conf)
	model.Initialize(conf)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/", handler.IndexPage())
	e.GET("/albums/:id", handler.AlbumPage())
	e.POST("/update", handler.UpdatePage())

	e.Debug = true

	e.Start(fmt.Sprintf("%s:%d", *host, *port))
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}
