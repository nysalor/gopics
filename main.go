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

func main() {
	flag.Parse()
	conf := config.Config{TargetDir: *dir, Port: *port, Host: *host, Log: logrus.New()}
	handler.Initialize(conf)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", handler.IndexPage())
	e.GET("/albums/:id", handler.AlbumPage())

	e.Debug = true

	e.Start(fmt.Sprintf("%s:%d", *host, *port))
}
