package db

import (
	"log"
	"os"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var Connection = DBConnect()

func DBConnect() *sqlx.DB {
	db, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@(%s:%s)/%s", os.Getenv("DB_USER"),  os.Getenv("DB_PASSWORD"),  os.Getenv("DB_HOST"),  os.Getenv("DB_PORT"),  os.Getenv("DB_NAME")))
	if err != nil {
		log.Fatalln(err)
	}
	return db
}
