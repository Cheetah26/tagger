package tagger

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

func Open(path string) (*Tagger, error) {
	t := &Tagger{}

	dir, _ := filepath.Split(path)
	t.dir = dir

	db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	t.db = db

	_, err = db.Exec(CREATE_SCHEMA)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Tagger) Close() {
	t.db.Close()
}

func (t *Tagger) GetDir() string {
	return t.dir
}
