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

		parents, err := t.GetParentTags(tag)
		if err != nil {
			fmt.Println(err.Error())
		}

		tag.Parents = parents

		tags = append(tags, tag)
	}

	return tags
}

// TODO do this with a recursive CTE in the database, not like this
func (t *Tagger) GetParentTags(tag Tag) ([]Tag, error) {
	var tags []Tag

	rows, err := t.db.Query(
		"SELECT Id, Name FROM Tags INNER JOIN TagTag ON Id = ParentTagId WHERE ChildTagId = ?", tag.Id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return tags, err
		}
	}
	defer rows.Close()

	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.Id, &tag.Name); err != nil {
			fmt.Println(err.Error())
		}

		tag.Parents, err = t.GetParentTags(tag)
		if err != nil {
			fmt.Println(err.Error())
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

func (t *Tagger) RemoveTag(tag Tag) error {
	_, err := t.db.Exec("DELETE FROM Tags WHERE Id = ?", tag.Id)

	if err != nil {
		return err
	}
	return nil
}

func (t *Tagger) UpdateTag(tag Tag) error {
	tx, err := t.db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("UPDATE Tags SET Name = ? WHERE Id = ?", tag.Name, tag.Id)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Remove all parent tags, then re-add them using the updated list
	_, err = tx.Exec("DELETE FROM TagTag WHERE ChildTagId = ?", tag.Id)
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, parent := range tag.Parents {
		_, err = tx.Exec("INSERT INTO TagTag(ParentTagId, ChildTagId) VALUES(?, ?)", parent.Id, tag.Id)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}
