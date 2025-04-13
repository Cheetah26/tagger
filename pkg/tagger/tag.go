package tagger

import (
	"database/sql"
	"errors"
)

type TagID int64

type Tag struct {
	Id       TagID   `json:"id"`
	Name     string  `json:"name"`
	Parents  []TagID `json:"parents"`
	Children []TagID `json:"children"`
}

type TagMap map[TagID]Tag

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
		Id:   TagID(id),
		Name: tagName,
	}, nil
}

func (t *Tagger) GetTagById(id TagID) (*Tag, error) {
	var tag Tag

	row := t.db.QueryRow("SELECT * FROM Tags WHERE Id = ?", id)

	if err := row.Scan(&tag.Id, &tag.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTagNotExist
		}
		return nil, NewDatabaseError(err)
	}

	if err := t.getRelativeTags(&tag); err != nil {
		return nil, NewDatabaseError(err)
	}

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

	if err := t.getRelativeTags(&tag); err != nil {
		return nil, NewDatabaseError(err)
	}

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

		if err := t.getRelativeTags(&tag); err != nil {
			return nil, NewDatabaseError(err)
		}

		tags[tag.Id] = tag
	}

	return tags, nil
}

func (t *Tagger) getRelativeTags(tag *Tag) error {
	parents, err := t.getParentTagIds(tag)
	if err != nil {
		return err
	}
	tag.Parents = parents

	children, err := t.getChildTagIds(tag)
	if err != nil {
		return err
	}
	tag.Children = children

	return nil
}

func (t *Tagger) getParentTagIds(tag *Tag) ([]TagID, error) {
	var parents []TagID

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

		parents = append(parents, TagID(parent))
	}

	return parents, nil
}

func (t *Tagger) getChildTagIds(tag *Tag) ([]TagID, error) {
	var children []TagID

	rows, err := t.db.Query("SELECT ChildTagId FROM TagTag WHERE ParentTagId = ?", tag.Id)
	if err != nil {
		return nil, NewDatabaseError(err)
	}
	for rows.Next() {
		var child int64
		err = rows.Scan(&child)
		if err != nil {
			return nil, NewDatabaseError(err)
		}

		children = append(children, TagID(child))
	}

	return children, nil
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
	_, err = tx.Exec("DELETE FROM TagTag WHERE ParentTagId = ? OR ChildTagId = ?", tag.Id, tag.Id)
	allErrors = append(allErrors, NewDatabaseError(err))

	for _, parent := range tag.Parents {
		_, err = tx.Exec("INSERT INTO TagTag(ParentTagId, ChildTagId) VALUES(?, ?)", parent, tag.Id)
		allErrors = append(allErrors, NewDatabaseError(err))
	}

	for _, child := range tag.Children {
		_, err = tx.Exec("INSERT INTO TagTag(ParentTagId, ChildTagId) VALUES(?, ?)", tag.Id, child)
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

type TagOrdering int

const (
	Id        TagOrdering = iota // Ascending
	Name                         // Ascending
	FileCount                    // Descending
)

func (t *Tagger) GetTagIdsOrdered(ordering TagOrdering) ([]TagID, error) {
	var ids []TagID

	var rows *sql.Rows
	var err error
	switch ordering {
	case Id:
		rows, err = t.db.Query("SELECT Id FROM Tags ORDER BY Id")
	case Name:
		rows, err = t.db.Query("SELECT Id FROM Tags ORDER BY Name")
	case FileCount:
		rows, err = t.db.Query(`
			SELECT Id
			FROM (
				SELECT Id, Name, FileId
				FROM Tags
				LEFT JOIN FileTag ON Id = TagId
				UNION
				SELECT ParentTagId AS Id, "", FileId
				FROM TagTag
				INNER JOIN FileTag ON ChildTagId = TagId
			)
			GROUP BY Id
			ORDER BY Count(DISTINCT FileId) DESC
		`)
	}

	if err != nil {
		return nil, NewDatabaseError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id uint64
		rows.Scan(&id)
		ids = append(ids, TagID(id))
	}

	return ids, nil
}
