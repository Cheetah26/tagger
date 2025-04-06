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

		file, err := tr.ImportFile(fullPath)
		if !compareFiles(expected, file) {
			t.Error("Import: Returned file does not match expectation")
		}
		if err != nil {
			t.Errorf("Import: %s", err.Error())
		}

		// check that each is imported as expected
		allFiles, err := tr.GetAllFiles()
		if err != nil {
			t.Error("Import: Unable to get all files")
		}
		if len(allFiles) != i+1 {
			t.Error("Import: File added to database")
		}

		actual, err := tr.GetFile(file.Id)
		if err != nil {
			t.Error("Import: Unable to get file after insert")
		}

		if !compareFiles(expected, actual) {
			t.Errorf("Import: Mismatch in database\n\tExpected: %+v\n\tGot: %+v", expected, actual)
		}

		info, err := os.Stat(tr.GetFilepath(actual))
		if err != nil || info.Size() <= 0 {
			t.Error("Import: File not written")
		}
	}

	// remove each file
	allFiles, err := tr.GetAllFiles()
	if err != nil {
		t.Error("Import: Unable to get all files")
	}

	for i := range COUNT {
		err := tr.RemoveFile(&allFiles[i])
		if err != nil {
			t.Error(err)
		}

		newAllFiles, err := tr.GetAllFiles()
		if err != nil {
			t.Error("Import: Unable to get all files")
		}
		if len(newAllFiles) != COUNT-i-1 {
			t.Error("Remove: Inconsistent number of files in database")
		}

		_, err = os.Stat(tr.GetFilepath(&allFiles[i]))
		if !errors.Is(err, os.ErrNotExist) {
			t.Error("Remove: Still on disk")
		}
	}
}

func TestTagFile(t *testing.T) {
	// setup
	tr := openTestDB(t.TempDir())

	filePath, _ := createTestFile(t.TempDir(), "cat")

	catFile, err := tr.ImportFile(filePath)
	if err != nil {
		t.Fatal("Setup: Error importing file")
	}

	animalTag, err := tr.AddTag("animal")
	if err != nil {
		t.Fatal("Setup: Error importing file")
	}
	plantTag, err := tr.AddTag("plant")
	if err != nil {
		t.Fatal("Setup: Error importing file")
	}

	checkUntaggedFiles := func(when string, expected int) {
		untagged, err := tr.GetUntaggedFiles()
		if err != nil {
			t.Errorf("%s: Error getting untagged files", when)
		}
		if len(untagged) != expected {
			t.Errorf("%s: Incorrect number of untagged files", when)
		}
	}

	checkUntaggedFiles("Setup", 1)

	got, err := tr.GetFilesByTag([]tagger.Tag{})
	if err != nil {
		t.Error("Get empty tag selection: Error getting files")
	}
	if len(got) != 0 {
		t.Error("Get empty tag selection: Incorrect number of files returned")
	}

	// tag file
	err = tr.TagFile(catFile, animalTag)
	if err != nil {
		t.Error("Tag file: Error tagging file")
	}

	checkUntaggedFiles("Tag file", 0)

	got, err = tr.GetFilesByTag([]tagger.Tag{*animalTag})
	if err != nil {
		t.Error("Get tagged file: Error getting file by tag")
	}
	if len(got) != 1 {
		t.Error("Get tagged file: Incorrect number of files returned")
	} else if !compareFiles(&got[0], catFile) {
		t.Error("Get tagged file: File not in results")
	}

	got, err = tr.GetFilesByTag([]tagger.Tag{*plantTag})
	if err != nil {
		t.Error("Get tagged file: Error getting file by tag")
	}
	if len(got) != 0 {
		t.Error("Get tagged file: Incorrect number of files returned")
	}

	// untag file
	err = tr.UntagFile(catFile, animalTag)
	if err != nil {
		t.Error("Untag file: Error untagging")
	}
	got, err = tr.GetFilesByTag([]tagger.Tag{*animalTag})
	if err != nil {
		t.Error("Get tagged file: Error getting file by tag")
	}
	if len(got) != 0 {
		t.Error("Get tagged file: Incorrect number of files returned")
	}

	checkUntaggedFiles("Tag file", 1)
}

func TestComplexTagSearch(t *testing.T) {
	// Tags:       |  Files:
	//     a    b  |   f1 [a1, b]
	//    / \      |   f2 [a2]
	//  a1   a2    |

	// setup
	tr := openTestDB(t.TempDir())

	var errs []error

	a, err := tr.AddTag("A")
	errs = append(errs, err)

	a1, err := tr.AddTag("A1")
	errs = append(errs, err)
	a1.Parents = []tagger.TagID{a.Id}
	errs = append(errs, tr.UpdateTag(a1))

	a2, err := tr.AddTag("A2")
	errs = append(errs, err)
	a2.Parents = []tagger.TagID{a.Id}
	errs = append(errs, tr.UpdateTag(a2))

	b, err := tr.AddTag("B")
	errs = append(errs, err)

	importDir := t.TempDir()
	filePath, _ := createTestFile(importDir, "f1")
	f1, err := tr.ImportFile(filePath)
	errs = append(errs, err)

	filePath, _ = createTestFile(importDir, "f2")
	f2, err := tr.ImportFile(filePath)
	errs = append(errs, err)

	errs = append(errs, tr.TagFile(f1, a1))
	errs = append(errs, tr.TagFile(f1, b))

	errs = append(errs, tr.TagFile(f2, a2))

	if errors.Join(errs...) != nil {
		t.Fatal("Setup: Error creating tags or files")
	}

	matches := func(actual []tagger.File, expected []tagger.File) bool {
		if len(actual) != len(expected) {
			return false
		}

		for i := range actual {
			if actual[i].Id != expected[i].Id {
				return false
			}
		}

		return true
	}

	// test
	actual, err := tr.GetFilesByTag([]tagger.Tag{*a})
	if err != nil {
		t.Error("Get: Error getting file by tag")
	}
	if !matches(actual, []tagger.File{*f1, *f2}) {
		t.Error("Get: Incorrect results")
	}

	actual, err = tr.GetFilesByTag([]tagger.Tag{*b})
	if err != nil {
		t.Error("Get: Error getting file by tag")
	}
	if !matches(actual, []tagger.File{*f1}) {
		t.Error("Get: Incorrect results")
	}

	actual, err = tr.GetFilesByTag([]tagger.Tag{*a, *b})
	if err != nil {
		t.Error("Get: Error getting file by tag")
	}
	if !matches(actual, []tagger.File{*f1}) {
		t.Error("Get: Incorrect results")
	}

	actual, err = tr.GetFilesByTag([]tagger.Tag{*a1})
	if err != nil {
		t.Error("Get: Error getting file by tag")
	}
	if !matches(actual, []tagger.File{*f1}) {
		t.Error("Get: Incorrect results")
	}

	actual, err = tr.GetFilesByTag([]tagger.Tag{*a2})
	if err != nil {
		t.Error("Get: Error getting file by tag")
	}
	if !matches(actual, []tagger.File{*f2}) {
		t.Error("Get: Incorrect results")
	}

	actual, err = tr.GetFilesByTag([]tagger.Tag{*a, *a1})
	if err != nil {
		t.Error("Get: Error getting file by tag")
	}
	if !matches(actual, []tagger.File{*f1}) {
		t.Error("Get: Incorrect results")
	}

	actual, err = tr.GetFilesByTag([]tagger.Tag{*a, *a2})
	if err != nil {
		t.Error("Get: Error getting file by tag")
	}
	if !matches(actual, []tagger.File{*f2}) {
		t.Error("Get: Incorrect results")
	}

	actual, err = tr.GetFilesByTag([]tagger.Tag{*a1, *b})
	if err != nil {
		t.Error("Get: Error getting file by tag")
	}
	if !matches(actual, []tagger.File{*f1}) {
		t.Error("Get: Incorrect results")
	}

	actual, err = tr.GetFilesByTag([]tagger.Tag{*a2, *b})
	if err != nil {
		t.Error("Get: Error getting file by tag")
	}
	if !matches(actual, []tagger.File{}) {
		t.Error("Get: Incorrect results")
	}

	actual, err = tr.GetFilesByTag([]tagger.Tag{*a1, *a2})
	if err != nil {
		t.Error("Get: Error getting file by tag")
	}
	if !matches(actual, []tagger.File{}) {
		t.Error("Get: Incorrect results")
	}
}

func TestNonexistentFile(t *testing.T) {
	tr := openTestDB(t.TempDir())

	fakePath := path.Join(t.TempDir(), "fake.txt")
	fakeFile := tagger.File{
		Id:          45,
		Hash:        tagger.Hash([]byte("some data")),
		Filetype:    "txt",
		Description: "a fake file",
		Tags:        []tagger.Tag{},
	}

	// try insert
	file, err := tr.ImportFile(fakePath)
	if file != nil {
		t.Error("Insert: Returned nonexistent file on insert")
	}
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		t.Error("Insert: Expected filesystem error ErrNotExist")
	}

	// try get
	got, err := tr.GetFile(fakeFile.Id)
	if !errors.Is(err, tagger.ErrFileNotExist) {
		t.Error("Get: No error when attempting to retrieve after failed insert")
	}
	if got != nil {
		t.Error("Get: Got nonexistent file")
	}

	// try remove
	err = tr.RemoveFile(&fakeFile)
	if err == nil {
		t.Error("Remove: No error when removing nonexistent file")
	}

	// try tag
	tag, err := tr.AddTag("tag")
	if err != nil {
		t.Fatal("Tag: Unable to create tag")
	}

	err = tr.TagFile(&fakeFile, tag)
	if err == nil {
		t.Error("Tag: No error when tagging nonexistent file")
	}
}

func TestDeletedFile(t *testing.T) {
	tr := openTestDB(t.TempDir())

	// setup
	sourcePath, _ := createTestFile(t.TempDir(), "a_file")

	file, err := tr.ImportFile(sourcePath)
	if err != nil {
		t.Fatal("Setup: Error creating test file")
	}

	taggerPath := tr.GetFilepath(file)

	// delete it outside of Tagger
	err = os.Remove(taggerPath)
	if err != nil {
		t.Fatal("Setup: Failed to remove file")
	}

	// remove in Tagger
	err = tr.RemoveFile(file)
	if err != nil {
		t.Error("Remove: Error removing file already deleted on disk")
	}
}
