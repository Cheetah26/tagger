package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type File struct {
	Hash        string `json:"hash"`
	Data        []byte `json:"data"`
	Filetype    string `json:"filetype"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Tags        []Tag  `json:"tags"`
}

func hash(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

func rowsToFiles(rows *sql.Rows) []File {
	var files []File

	for rows.Next() {
		var file File
		var name sql.NullString
		var description sql.NullString

		err := rows.Scan(&file.Hash, &file.Filetype, &name, &description)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				fmt.Println(err.Error())
			}
		}

		if name.Valid {
			file.Name = name.String
		}
		if description.Valid {
			file.Description = description.String
		}

		files = append(files, file)
	}

	return files
}

func (t *Tagger) ImportFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		panic("Cannot find file")
	}

	_, filename := filepath.Split(path)

	file := File{
		Hash:        hash(data),
		Data:        data,
		Filetype:    strings.ToLower(filepath.Ext(filename)),
		Description: filename,
	}

	if err := t.AddFile(file); err != nil {
		return err
	}

	return nil
}

func (t *Tagger) AddFile(file File) error {
	_, err := t.db.Exec(
		"INSERT INTO Files(Hash, Data, Filetype, Description) VALUES(?, ?, ?, ?)",
		file.Hash, file.Data, file.Filetype, file.Description)

	if err != nil {
		return err
	}

	return nil
}

func (t *Tagger) RemoveFile(file File) error {
	_, err := t.db.Exec("DELETE FROM Files WHERE Hash = ?", file.Hash)

	if err != nil {
		return err
	}

	return nil
}

func (t *Tagger) GetFile(hash string) *File {
	// Get the file
	row := t.db.QueryRow("SELECT Hash, Data, Filetype, Name, Description FROM Files WHERE Hash = ?", hash)

	file := &File{}
	var name sql.NullString
	var description sql.NullString

	err := row.Scan(&file.Hash, &file.Data, &file.Filetype, &name, &description)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			fmt.Println(err.Error())
		}
		return nil
	}

	if name.Valid {
		file.Name = name.String
	}
	if description.Valid {
		file.Description = description.String
	}

	// Get it's tags
	rows, err := t.db.Query("SELECT Id, Name FROM Tags INNER JOIN FileTag on Id = TagId WHERE FileHash = ?", file.Hash)
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

		tags = append(tags, tag)
	}

	file.Tags = tags

	return file
}

func (t *Tagger) GetAllFiles() []File {
	rows, err := t.db.Query("SELECT Hash, Filetype, Name FROM Files ORDER BY RANDOM()")
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

	tag_ids := make([]string, len(tags))
	for i, v := range tags {
		tag_ids[i] = strconv.Itoa(v.Id)
	}

	// I don't like this but doing it properly didn't work
	query := fmt.Sprintf(`
	SELECT Hash, Filetype, Name, Description
	FROM Files
	INNER JOIN FileTag
		ON Hash = FileHash
		WHERE TagId IN (%s)
		GROUP BY FileHash
		HAVING COUNT (*) >= %d
		`,
		strings.Join(tag_ids, ","),
		len(tag_ids))

	rows, err := t.db.Query(query)

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
		SELECT Hash, Filetype, Name, Description FROM Files
		LEFT JOIN FileTag ON FileHash = Hash
		GROUP BY Hash
		HAVING COUNT(TagId) = 0
	`)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	defer rows.Close()

	return rowsToFiles(rows), nil
}

func (t *Tagger) ReadFile(file File) *File {
	// Get the file
	row := t.db.QueryRow("SELECT Data FROM Files WHERE Hash = ?", file.Hash)

	err := row.Scan(&file.Data)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	return &file
}

func (t *Tagger) TagFile(file File, tag Tag) error {
	_, err := t.db.Exec("INSERT INTO FileTag(FileHash, TagId) VALUES(?, ?)", file.Hash, tag.Id)
	if err != nil {
		return err
	}
	return nil
}

func (t *Tagger) UntagFile(file File, tag Tag) error {
	_, err := t.db.Exec("DELETE FROM FileTag WHERE FileHash = ? AND TagId = ?", file.Hash, tag.Id)
	if err != nil {
		return err
	}
	return nil
}
