package tagger

import (
	"database/sql"
	"errors"
)

type Tag struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Parents []Tag  `json:"parents"`
}

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
		Id:   int(id),
		Name: tagName,
	}, nil
}

func (t *Tagger) GetTag(name string) (*Tag, error) {
	var tag Tag

	row := t.db.QueryRow("SELECT * FROM Tags WHERE Name = ?", name)

	if err := row.Scan(&tag.Id, &tag.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTagNotExist
		}
		return nil, NewDatabaseError(err)
	}

	parents, err := t.GetParentTags(tag)
	if err != nil {
		return nil, err
	}
	tag.Parents = parents

	return &tag, nil
}

func (t *Tagger) GetAllTags() ([]Tag, error) {
	var tags []Tag

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

		parents, err := t.GetParentTags(tag)
		if err != nil {
			return nil, err
		}

		tag.Parents = parents

		tags = append(tags, tag)
	}

	return tags, nil
}

// Recursively get a tag's parents
func (t *Tagger) GetParentTags(tag Tag) ([]Tag, error) {
	var tags []Tag

	rows, err := t.db.Query(
		"SELECT Id, Name FROM Tags INNER JOIN TagTag ON Id = ParentTagId WHERE ChildTagId = ?", tag.Id)
	if err != nil {
		return nil, NewDatabaseError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.Id, &tag.Name); err != nil {
			return nil, NewDatabaseError(err)
		}

		// TODO should this be recursive or not?
		tag.Parents, err = t.GetParentTags(tag)
		if err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

// Get the immediate children of a tag
func (t *Tagger) GetChildTags(tag Tag) ([]Tag, error) {
	var tags []Tag

	rows, err := t.db.Query(
		"SELECT Id, Name FROM Tags INNER JOIN (SELECT * FROM TagTag) ON Id == ChildTagId WHERE ParentTagId == ?", tag.Id)
	if err != nil {
		return nil, NewDatabaseError(err)
	}

	defer rows.Close()

	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.Id, &tag.Name); err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

func (t *Tagger) RemoveTag(tag Tag) error {
	res, err := t.db.Exec("DELETE FROM Tags WHERE Id = ?", tag.Id)

	affected, _ := res.RowsAffected()
	if err == nil && affected < 1 {
		err = ErrTagNotExist
	}

	return err
}

func (t *Tagger) UpdateTag(tag Tag) error {
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
		_, err = tx.Exec("INSERT INTO TagTag(ParentTagId, ChildTagId) VALUES(?, ?)", parent.Id, tag.Id)
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
