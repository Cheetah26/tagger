package main

import (
	"database/sql"
	"errors"
	"fmt"
)

type Tag struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Parents []Tag  `json:"parents"`
}

func (t *Tagger) AddTag(tagName string) *Tag {
	res, err := t.db.Exec("INSERT INTO Tags(name) VALUES(?)", tagName)
	if err != nil {
		fmt.Println(err.Error())
	}

	id, _ := res.LastInsertId()

	return &Tag{
		Id:   int(id),
		Name: tagName,
	}
}

func (t *Tagger) GetTag(name string) (*Tag, error) {
	row := t.db.QueryRow("SELECT * FROM Tags WHERE Name = ?", name)

	var tag Tag
	if err := row.Scan(&tag.Id, &tag.Name); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			fmt.Println(err.Error())
		}
		return nil, err
	}

	return &tag, nil
}

func (t *Tagger) GetAllTags() []Tag {
	rows, err := t.db.Query("SELECT * FROM Tags")
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			fmt.Println(err.Error())
		}
	}
	defer rows.Close()

	var tags []Tag

	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.Id, &tag.Name); err != nil {
			fmt.Println(err.Error())
		}
		tags = append(tags, tag)
	}

	return tags
}

func (t *Tagger) RemoveTag(tag Tag) error {
	_, err := t.db.Exec("DELETE FROM Tags WHERE Id = ?", tag.Id)

	if err != nil {
		return err
	}
	return nil
}
