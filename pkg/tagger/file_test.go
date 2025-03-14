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

func createTestFile(dir string, name string) (fullPath string, expected *tagger.File) {
	const FILETYPE = "txt"
	fullName := fmt.Sprintf("%s.%s", name, FILETYPE)
	fullPath = path.Join(dir, fullName)
	contents := []byte(fmt.Sprintf("Contents of: %s", fullName))

	expected = &tagger.File{
		Id:          0,
		Hash:        tagger.Hash(contents),
		Filetype:    FILETYPE,
		Description: fullName,
		Tags:        []tagger.Tag{},
	}

	err := os.WriteFile(
		fullPath,
		contents,
		0644,
	)
	if err != nil {
		panic(err)
	}

	return
}

// Compare tagger.File structs, ignoring their IDs
func compareFiles(a, b *tagger.File) bool {
	return a == b ||
		(a.Hash == b.Hash &&
			a.Filetype == b.Filetype &&
			a.Description == b.Description)
}

func TestInsertAndRemoveFiles(t *testing.T) {
	// setup
	const COUNT = 200

	tr := openTestDB(t.TempDir())

	// create and import files
	importDir := t.TempDir()
	for i := range COUNT {
		i_string := strconv.Itoa(i)
		fullPath, expected := createTestFile(importDir, i_string)
		err := tr.ImportFile(fullPath)
		if err != nil {
			t.Errorf("Import failed: %s", err.Error())
		}

		// check that each is imported as expected
		insertedFiles := tr.GetAllFiles()
		if len(insertedFiles) != i+1 {
			t.Error("Imported incorrectly: File added to database")
			return
		}

		actual := &insertedFiles[i]

		if !compareFiles(expected, actual) {
			t.Errorf("Imported incorrectly: Mismatch in database\n\tExpected: %+v\n\tGot: %+v", expected, actual)
			return
		}

		info, err := os.Stat(tr.GetFilepath(*actual))
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
			t.Errorf("File not removed: Inconsistent number of files in database")
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
	tr := openTestDB(t.TempDir())

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
// search by tag directly (single, multiple)
// search by tag parent (single, multiple)
