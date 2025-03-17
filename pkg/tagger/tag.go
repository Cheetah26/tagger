package tagger

import (
	"database/sql"
	"errors"
)

type Tag struct {
	Id      int64   `json:"id"`
	Name    string  `json:"name"`
	Parents []int64 `json:"parents"`
}

type TagMap map[int64]Tag

func (t *Tagger) AddTag(tagName string) (*Tag, error) {
	res, err := t.db.Exec("INSERT INTO Tags(name) VALUES(?)", tagName)
	if err != nil {
		return nil, NewDatabaseError(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, NewDatabaseError(err)
	}

	return &Tag{
		Id:   id,
		Name: tagName,
	}, nil
}

func (t *Tagger) GetTagById(id int64) (*Tag, error) {
	var tag Tag

	row := t.db.QueryRow("SELECT * FROM Tags WHERE Id = ?", id)

	if err := row.Scan(&tag.Id, &tag.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTagNotExist
		}
		return nil, NewDatabaseError(err)
	}

	parents, err := t.getParentTagIds(&tag)
	if err != nil {
		return nil, err
	}
	tag.Parents = parents

	return &tag, nil
}

func (t *Tagger) GetTagByName(name string) (*Tag, error) {
	var tag Tag

	row := t.db.QueryRow("SELECT * FROM Tags WHERE Name = ?", name)

	if err := row.Scan(&tag.Id, &tag.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTagNotExist
		}
		return nil, NewDatabaseError(err)
	}

	parents, err := t.getParentTagIds(&tag)
	if err != nil {
		return nil, err
	}
	tag.Parents = parents

	return &tag, nil
}

func (t *Tagger) GetAllTags() (TagMap, error) {
	tags := make(TagMap)

	rows, err := t.db.Query("SELECT * FROM Tags")
	if err != nil {
		return nil, NewDatabaseError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.Id, &tag.Name); err != nil {
			return nil, NewDatabaseError(err)
		}

		parents, err := t.getParentTagIds(&tag)
		if err != nil {
			return nil, err
		}
		tag.Parents = parents

		tags[tag.Id] = tag
	}

	return tags, nil
}

func (t *Tagger) getParentTagIds(tag *Tag) ([]int64, error) {
	var parents []int64

	rows, err := t.db.Query("SELECT ParentTagId FROM TagTag WHERE ChildTagId = ?", tag.Id)
	if err != nil {
		return nil, NewDatabaseError(err)
	}
	for rows.Next() {
		var parent int64
		err = rows.Scan(&parent)
		if err != nil {
			return nil, NewDatabaseError(err)
		}

		parents = append(parents, parent)
	}

	return parents, nil
}

func (t *Tagger) RemoveTag(tag *Tag) error {
	res, err := t.db.Exec("DELETE FROM Tags WHERE Id = ?", tag.Id)

	affected, _ := res.RowsAffected()
	if err == nil && affected < 1 {
		err = ErrTagNotExist
	}

	return err
}

func (t *Tagger) UpdateTag(tag *Tag) error {
	var allErrors []error

	tx, err := t.db.Begin()
	allErrors = append(allErrors, NewDatabaseError(err))

	res, err := tx.Exec("UPDATE Tags SET Name = ? WHERE Id = ?", tag.Name, tag.Id)
	if err != nil {
		allErrors = append(allErrors, NewDatabaseError(err))
	} else {
		if affected, _ := res.RowsAffected(); affected < 1 {
			allErrors = append(allErrors, ErrTagNotExist)
		}
	}

	// Remove all parent tags, then re-add them using the updated list
	_, err = tx.Exec("DELETE FROM TagTag WHERE ChildTagId = ?", tag.Id)
	allErrors = append(allErrors, NewDatabaseError(err))

	for _, parent := range tag.Parents {
		_, err = tx.Exec("INSERT INTO TagTag(ParentTagId, ChildTagId) VALUES(?, ?)", parent, tag.Id)
		allErrors = append(allErrors, NewDatabaseError(err))
	}

	allErrorsJoined := errors.Join(allErrors...)
	if allErrorsJoined != nil {
		tx.Rollback()
	} else {
		tx.Commit()
	}

	return allErrorsJoined
}
