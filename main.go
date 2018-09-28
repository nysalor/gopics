package main

import (
	"flag"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sirupsen/logrus"
	"./config"
	"./handlers"
)

var dir = flag.String("d", "./images", "target dir")
var port = flag.Int("p", 5000, "server port number")
var host = flag.String("h", "localhost", "server port number")
var baseUrl = flag.String("u", "http://localhost", "base url for images")

func main() {
	flag.Parse()
	conf := config.Config{TargetDir: *dir, Port:*port, Host: *host, BaseUrl: *baseUrl, Log: logrus.New()}
	handler.Initialize(conf)

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
