package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"./handlers"
)

var targetDir = "./images"

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", handler.MainPage(targetDir))

	e.Debug = true
	e.Start(":5000")
}
