package tagger

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	// File
	Id          int64  `json:"id"`
	Hash        string `json:"hash"`
	Filetype    string `json:"filetype"`
	Description string `json:"description"`

	// FileTags
	Tags []Tag `json:"tags"`
}

func Hash(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

func rowsToFiles(rows *sql.Rows) []File {
	var files []File

	for rows.Next() {
		var file File
		var name sql.NullString
		var description sql.NullString

		err := rows.Scan(&file.Id, &file.Hash, &file.Filetype, &name, &description)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				fmt.Println(err.Error())
			}
		}

		if description.Valid {
			file.Description = description.String
		}

		files = append(files, file)
	}

	return files
}

func (t *Tagger) GetFilepath(file File) string {
	idStr := fmt.Sprintf("%d", file.Id)

	var level1 string
	var level2 string
	if len(idStr) < 2 {
		level1 = "0"
		level2 = string(idStr[0])
	} else {
		level1 = string(idStr[0])
		level2 = string(idStr[1])
	}

	return filepath.Join(t.dir, level1, level2, idStr+"."+file.Filetype)
}

func (t *Tagger) ImportFile(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	_, filename := filepath.Split(path)

	file := File{
		Hash:        Hash(data),
		Filetype:    filepath.Ext(filename)[1:],
		Description: filename,
	}

	// Insert using a transaction in case copying the file fails
	tx, err := t.db.Begin()
	if err != nil {
		return nil, err
	}
	result, err := tx.Exec(
		"INSERT INTO Files(Hash, Filetype, Description) VALUES(?, ?, ?)", file.Hash, file.Filetype, file.Description)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	file.Id = id
	newPath := t.GetFilepath(file)

	err = os.MkdirAll(filepath.Dir(newPath), 0777)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = os.WriteFile(newPath, data, 0777)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return &file, nil
}

// Remove a file from the database and disk
// If the file is not found on disk, assume that it was deleted and proceed normally
func (t *Tagger) RemoveFile(file File) error {
	err := os.Remove(t.GetFilepath(file))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	res, err := t.db.Exec("DELETE FROM Files WHERE Id = ?", file.Id)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected < 1 {
		return ErrFileNotExist
	}

	return nil
}

func (t *Tagger) GetFile(id int64) *File {
	// Get the file
	row := t.db.QueryRow("SELECT Id, Hash, Filetype, Name, Description FROM Files WHERE Id = ?", id)

	file := &File{}
	var name sql.NullString
	var description sql.NullString

	err := row.Scan(&file.Id, &file.Hash, &file.Filetype, &name, &description)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			panic(err) // TODO handle this properly
		}
		return nil
	}

	if description.Valid {
		file.Description = description.String
	}

	// Get it's tags
	rows, err := t.db.Query("SELECT Id, Name FROM Tags LEFT JOIN FileTag on Id = TagId WHERE FileId = ?", file.Id)
	if err != nil {
		fmt.Println(err)
	}

	var tags []Tag

	for rows.Next() {
		var tag Tag
		err := rows.Scan(&tag.Id, &tag.Name)
		if err != nil {
			fmt.Println(err)
		}

		tag.Parents, err = t.GetParentTags(tag)
		if err != nil {
			fmt.Println(err)
		}

		tags = append(tags, tag)
	}

	file.Tags = tags

	return file
}

func (t *Tagger) GetAllFiles() []File {
	rows, err := t.db.Query("SELECT Id, Hash, Filetype, Name, Description FROM Files")
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			fmt.Println(err.Error())
		}
	}
	defer rows.Close()

	return rowsToFiles(rows)
}

func (t *Tagger) GetFiles(tags []Tag) []File {
	if len(tags) == 0 {
		return nil
	}

	placeholders := make([]string, len(tags))
	tagIds := make([]any, len(tags))
	for i, v := range tags {
		placeholders[i] = "(?)"
		tagIds[i] = v.Id
	}

	rows, err := t.db.Query(`
WITH RECURSIVE
selection(id) AS (
	VALUES `+strings.Join(placeholders, ",")+`
),
-- collect child tags of the selection recursively
child_tags(id, ancestor) AS (
    SELECT id, id AS ancestor FROM selection
    UNION ALL
    SELECT ChildTagId AS id, ancestor FROM child_tags JOIN TagTag ON id = ParentTagId
),
-- files tagged with a child and all other tags, OR the selected tags directly
file_ids AS (
	SELECT FileId AS file_id FROM (
			-- children tags of the selection
			--   (ancestor + distinct prevents multiple child tags
			--     of the same parent from increasing the tag count)
			SELECT DISTINCT FileId, ancestor AS TagId FROM FileTag
			JOIN child_tags ON id = TagId
			WHERE ancestor IN selection
			UNION ALL
			-- non-children tags
			SELECT FileId, TagId FROM FileTag
			WHERE TagId NOT IN (SELECT ancestor FROM child_tags)
					AND TagId IN selection
	)
	GROUP BY FileId
	HAVING count(TagId) >= (select count(*) FROM selection)
)
SELECT Id, Hash, Filetype, Name, Description FROM Files
JOIN file_ids ON Id = file_id
`, tagIds...)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			fmt.Println(err.Error())
		}
	}
	defer rows.Close()

	return rowsToFiles(rows)
}

func (t *Tagger) GetUntaggedFiles() ([]File, error) {
	rows, err := t.db.Query(`
		SELECT Id, Hash, Filetype, Name, Description FROM Files
		LEFT JOIN FileTag ON FileId = Id
		GROUP BY Id
		HAVING COUNT(TagId) = 0
	`)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	defer rows.Close()

	return rowsToFiles(rows), nil
}

func (t *Tagger) TagFile(file File, tag Tag) error {
	_, err := t.db.Exec("INSERT INTO FileTag(FileId, TagId) VALUES(?, ?)", file.Id, tag.Id)
	if err != nil {
		return err
	}
	return nil
}

func (t *Tagger) UntagFile(file File, tag Tag) error {
	_, err := t.db.Exec("DELETE FROM FileTag WHERE FileId = ? AND TagId = ?", file.Id, tag.Id)
	if err != nil {
		return err
	}
	return nil
}
