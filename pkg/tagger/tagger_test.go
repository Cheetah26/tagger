package tagger_test

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/cheetah26/tagger/pkg/tagger"
)

func CreateEmptyDB(dir string) string {
	fullPath := path.Join(dir, "test.db")

	_, err := os.Create(fullPath)
	if err != nil {
		os.Exit(1)
	}

	return fullPath
}

func CreateFile(dir string, name string, contents string) string {
	fullPath := path.Join(dir, fmt.Sprintf("%s.txt", name))

	err := os.WriteFile(
		fullPath,
		[]byte(contents),
		0644,
	)
	if err != nil {
		os.Exit(1)
	}

	return fullPath
}

func CompareFiles(a, b *tagger.File) bool {
	return a == b ||
		(a.Hash == b.Hash &&
			a.Filetype == b.Filetype &&
			a.Description == b.Description)
}

func TestInsertAndRemoveFiles(t *testing.T) {
	// setup
	const COUNT = 200

	taggerDir := t.TempDir()
	dbFile := CreateEmptyDB(taggerDir)

	tr := &tagger.Tagger{}
	tr.Open(dbFile)

	// create and import files
	importDir := t.TempDir()
	for i := range COUNT {
		i_string := strconv.Itoa(i)
		fileName := CreateFile(importDir, i_string, i_string)
		err := tr.ImportFile(fileName)
		if err != nil {
			t.Error(err)
		}

		// check that each is imported as expected
		insertedFiles := tr.GetAllFiles()
		if len(insertedFiles) != i+1 {
			t.Error("Imported incorrectly: File added to database")
			return
		}

		actual := insertedFiles[i]

		expected := tagger.File{
			Id:          0,
			Filetype:    "txt",
			Hash:        tagger.Hash([]byte(i_string)),
			Description: fmt.Sprintf("%s.txt", i_string),
		}
		if !CompareFiles(&expected, &actual) {
			t.Errorf("Imported incorrectly: Mismatch in database\n\tExpected: %+v\n\tGot: %+v", expected, actual)
			return
		}

		info, err := os.Stat(tr.GetFilepath(actual))
		if err != nil || info.Size() <= 0 {
			t.Error("Imported incorrectly: File not written")
		}
	}

	// remove each file
	allFiles := tr.GetAllFiles()
	for i := range COUNT {
		err := tr.RemoveFile(allFiles[i])
		if err != nil {
			t.Error(err)
		}

		if len(tr.GetAllFiles()) != COUNT-i-1 {
			t.Errorf("File not removed: Still in database")
			return
		}

		if tr.GetFile(allFiles[i].Id) != nil {
			t.Errorf("File not removed: Still in database")
			return
		}

		_, err = os.Stat(tr.GetFilepath(allFiles[i]))
		if !errors.Is(err, os.ErrNotExist) {
			t.Error("File not removed: Still on disk")
		}
	}
}

func TestInsertFail(t *testing.T) {
	taggerDir := t.TempDir()
	dbFile := CreateEmptyDB(taggerDir)

	tr := &tagger.Tagger{}
	tr.Open(dbFile)

	fileName := path.Join(t.TempDir(), "fake.txt")
	err := tr.ImportFile(fileName)
	if err == nil {
		t.Error("Expected ErrNotExist, got nil")
		return
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected ErrNotExist, got %s", err.Error())
	}
}

// TODO More tests:
// get / remove a file after deletion
// add & remove file tag
// tag tag
// get parent tags
// get child tags
// search by tag directly (single, multiple)
// search by tag parent (single, multiple)
