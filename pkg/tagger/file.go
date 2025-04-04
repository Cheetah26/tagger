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

// Convert sql.Rows to a slice of files.
// Query MUST select the following columns in order:
// Id, Hash, Filetype, Description
func (t *Tagger) rowsToFiles(rows *sql.Rows) ([]File, error) {
	var files []File

	for rows.Next() {
		var file File
		var description sql.NullString

		err := rows.Scan(&file.Id, &file.Hash, &file.Filetype, &description)
		if err != nil {
			return nil, NewDatabaseError(err)
		}

		if description.Valid {
			file.Description = description.String
		}

		tags, err := t.getFileTags(&file)
		if err != nil {
			return nil, err
		}

		file.Tags = tags

		files = append(files, file)
	}

	return files, nil
}

// Query for a file's tags
func (t *Tagger) getFileTags(file *File) ([]Tag, error) {
	rows, err := t.db.Query("SELECT Id, Name FROM Tags LEFT JOIN FileTag on Id = TagId WHERE FileId = ?", file.Id)
	if err != nil {
		return nil, NewDatabaseError(err)
	}

	var tags []Tag

	for rows.Next() {
		var tag Tag
		err := rows.Scan(&tag.Id, &tag.Name)
		if err != nil {
			return nil, NewDatabaseError(err)
		}

		parents, err := t.getParentTagIds(&tag)
		if err != nil {
			return nil, err
		}

		tag.Parents = parents

		tags = append(tags, tag)
	}

	return tags, nil
}

// Import a file into Tagger, copying it from the source directory and adding it to the database.
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

	var errs []error

	// use a transaction in case copying the file fails
	tx, err := t.db.Begin()
	errs = append(errs, NewDatabaseError(err))

	result, err := tx.Exec(
		"INSERT INTO Files(Hash, Filetype, Description) VALUES(?, ?, ?)", file.Hash, file.Filetype, file.Description)
	errs = append(errs, NewDatabaseError(err))

	id, err := result.LastInsertId()
	errs = append(errs, NewDatabaseError(err))

	// write to disk
	file.Id = id
	newPath := t.GetFilepath(&file)

	err = os.MkdirAll(filepath.Dir(newPath), 0777)
	errs = append(errs, NewFilesystemError(err))

	err = os.WriteFile(newPath, data, 0777)
	errs = append(errs, NewFilesystemError(err))

	collectedErrors := errors.Join(errs...)
	if collectedErrors != nil {
		tx.Rollback()
		return nil, collectedErrors
	}

	tx.Commit()
	return &file, nil
}

func (t *Tagger) GetFile(fileId int64) (*File, error) {
	rows, err := t.db.Query("SELECT Id, Hash, Filetype, Description FROM Files WHERE Id = ?", fileId)
	if err != nil {
		return nil, NewDatabaseError(err)
	}

	files, err := t.rowsToFiles(rows)
	if err != nil {
		return nil, NewDatabaseError(err)
	}

	if len(files) != 1 {
		return nil, ErrFileNotExist
	}

	return &files[0], nil
}

// Get the path to a file in Tagger's directory.
// Does not verify that the file exists on disk, but only construts the path where the file is expected to be.
func (t *Tagger) GetFilepath(file *File) string {
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

func (t *Tagger) GetAllFiles() ([]File, error) {
	rows, err := t.db.Query("SELECT Id, Hash, Filetype, Description FROM Files")
	if err != nil {
		return nil, NewDatabaseError(err)
	}
	defer rows.Close()

	return t.rowsToFiles(rows)
}

func (t *Tagger) GetFilesByTag(tags []Tag) ([]File, error) {
	if len(tags) == 0 {
		return []File{}, nil
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
		SELECT Id, Hash, Filetype, Description FROM Files
		JOIN file_ids ON Id = file_id
	`, tagIds...)

	if err != nil {
		return nil, NewDatabaseError(err)
	}

	defer rows.Close()

	return t.rowsToFiles(rows)
}

// Get all files without any tags
func (t *Tagger) GetUntaggedFiles() ([]File, error) {
	rows, err := t.db.Query(`
		SELECT Id, Hash, Filetype, Description FROM Files
		LEFT JOIN FileTag ON FileId = Id
		GROUP BY Id
		HAVING COUNT(TagId) = 0
	`)

	if err != nil {
		return nil, NewDatabaseError(err)
	}

	defer rows.Close()

	return t.rowsToFiles(rows)
}

// Add a tag to a file.
func (t *Tagger) TagFile(file *File, tag *Tag) error {
	// TODO should this check that the file & tag actaully exist?
	_, err := t.db.Exec("INSERT INTO FileTag(FileId, TagId) VALUES(?, ?)", file.Id, tag.Id)

	return NewDatabaseError(err)
}

// Remove a specific tag from a file. Will not produce an error if either value is invalid.
func (t *Tagger) UntagFile(file *File, tag *Tag) error {
	_, err := t.db.Exec("DELETE FROM FileTag WHERE FileId = ? AND TagId = ?", file.Id, tag.Id)

	return NewDatabaseError(err)
}

// Remove a file from the database and disk
// If the file is not found on disk, assume that it was deleted and proceed normally
func (t *Tagger) RemoveFile(file *File) error {
	err := os.Remove(t.GetFilepath(file))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return NewFilesystemError(err)
	}

	res, err := t.db.Exec("DELETE FROM Files WHERE Id = ?", file.Id)
	if err != nil {
		return NewDatabaseError(err)
	}
	if affected, _ := res.RowsAffected(); affected < 1 {
		return ErrFileNotExist
	}

	_, err = t.db.Exec("DELETE FROM FileTag WHERE FileId = ?", file.Id)
	if err != nil {
		return NewDatabaseError(err)
	}

	return nil
}
