// Package litestore ....
// НЕ РЕАЛИЗОВАНО!!!
package litestore

import (
	"database/sql"
	"log"
	//_ "github.com/mattn/go-sqlite3"
	//_ "github.com/mxk/go-sqlite/sqlite3"
)

// Store ...
type Store struct {
	//db   *sql.DB
	path string
}

// New создаем новое хранилище sqlite
func New(path string) *Store {

	return &Store{
		//db:   initDB(path),
		path: path,
	}
}

func initDB(path string) *sql.DB {

	//Открываем БД, если не существует то создаем
	db, err := sql.Open("sqlite3", path+"?cache=shared&mode=rwc")

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS SMS (" +
		"ID INTEGER PRIMARY KEY, " +
		"Type TEXT, " +
		"Name TEXT, " +
		"Time INTEGER)")

	if err != nil {
		log.Fatal(err)
	}

	return db
}
