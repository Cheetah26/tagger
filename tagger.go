package main

import (
	"database/sql"
	"path/filepath"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var CREATE_SCHEMA string

type Tagger struct {
	db  *sql.DB
	dir string
}

func (t *Tagger) Open(path string) {
	dir, _ := filepath.Split(path)
	t.dir = dir

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		panic(err.Error())
	}

	_, err = db.Exec(CREATE_SCHEMA)
	if err != nil {
		panic(err.Error())
	}

	t.db = db
}
