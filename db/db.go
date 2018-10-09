package db

import (
	"log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"../config"
)

func Connect(conf config.Database) *sqlx.DB {
	db, err := sqlx.Connect("mysql", conf.Url())
	if err != nil {
		log.Fatalln(err)
	}
	return db
}
