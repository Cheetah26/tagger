package fuse_test

import (
	"errors"
	"os"
	"path"
	"slices"
	"testing"

	"github.com/cheetah26/tagger/pkg/fuse"
	"github.com/cheetah26/tagger/pkg/tagger"
)

func setup(t *testing.T) *tagger.Tagger {
	// Tags:       |  Files:
	//     a    b  |   f1 [a1, b]
	//    / \      |   f2 [a2]
	//  a1   a2    |

	dbFile := path.Join(t.TempDir(), "test.db")
	tr, err := tagger.Open(dbFile)
	if err != nil {
		t.Fatal("Setup: Unable to initialize Tagger")
	}

	var errs []error

	a, err := tr.AddTag("a")
	errs = append(errs, err)

	a1, err := tr.AddTag("a1")
	errs = append(errs, err)
	a1.Parents = []int64{a.Id}
	errs = append(errs, tr.UpdateTag(a1))

	a2, err := tr.AddTag("a2")
	errs = append(errs, err)
	a2.Parents = []int64{a.Id}
	errs = append(errs, tr.UpdateTag(a2))

	b, err := tr.AddTag("b")
	errs = append(errs, err)

	importDir := t.TempDir()

	f1Path := path.Join(importDir, "f1.txt")
	err = os.WriteFile(f1Path, []byte("f1"), 0644)
	errs = append(errs, err)
	f1, err := tr.ImportFile(f1Path)
	errs = append(errs, err)

	f2Path := path.Join(importDir, "f2.txt")
	err = os.WriteFile(f2Path, []byte("f2"), 0644)
	errs = append(errs, err)
	f2, err := tr.ImportFile(f2Path)
	errs = append(errs, err)

	errs = append(errs, tr.TagFile(f1, a1))
	errs = append(errs, tr.TagFile(f1, b))

	errs = append(errs, tr.TagFile(f2, a2))

	if errors.Join(errs...) != nil {
		t.Fatal("Setup: Error creating tags or files")
	}

	return tr
}

func checkDirContents(t *testing.T, path string, shouldHaveDirs []string, shouldHaveFiles []string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		t.Error("Error reading directory, ", err.Error())
	}

	var dirNames []string
	var fileNames []string
	for _, e := range entries {
		if e.IsDir() {
			dirNames = append(dirNames, e.Name())
		} else {
			fileNames = append(fileNames, e.Name())
		}
	}

	t.Logf("Contents of %s:\n\tDirs: %v\n\tFiles: %v", path, dirNames, fileNames)

	if len(dirNames) != len(shouldHaveDirs) {
		t.Error("Incorrect number of directories")
	}
	for _, dir := range shouldHaveDirs {
		if !slices.Contains(dirNames, dir) {
			t.Error("Missing dir:", dir)
		}
	}

	if len(fileNames) != len(shouldHaveFiles) {
		t.Error("Incorrect number of files")
	}
	for _, file := range shouldHaveFiles {
		if !slices.Contains(fileNames, file) {
			t.Error("Missing file:", file)
		}
	}
}

func TestBasicFunctionality(t *testing.T) {
	tr := setup(t)

	mountDir := path.Join(t.TempDir(), "/mnt/")
	err := os.MkdirAll(mountDir, os.ModeDir|os.ModePerm)
	if err != nil {
		t.Fatal("Error creating mount directory")
	}

	unmount, errChan, err := fuse.Mount(mountDir, tr)
	if err != nil {
		t.Fatal(err.Error())
	}
	go func() {
		if err = <-errChan; err != nil {
			t.Error(err.Error())
		}
	}()

	defer unmount()

	// mount root should have top-level tags
	checkDirContents(t, mountDir, []string{"$", "a", "b"}, nil)
	checkDirContents(t, path.Join(mountDir, "/$"), nil, []string{"1.txt", "2.txt"})

	// folder "a" should have tag a's children, and "a/$" all files tagged a, a1, or a2
	checkDirContents(t, path.Join(mountDir, "/a/"), []string{"$", "+", "a1", "a2"}, nil)
	checkDirContents(t, path.Join(mountDir, "/a/$"), nil, []string{"1.txt", "2.txt"})

	// folder "b" should have no dirs b/c it has no child tags
	checkDirContents(t, path.Join(mountDir, "/b/"), []string{"$", "+"}, nil)
	checkDirContents(t, path.Join(mountDir, "/b/$"), nil, []string{"1.txt"})

	// read a file
	contents, err := os.ReadFile(path.Join(mountDir, "/$/1.txt"))
	if err != nil {
		t.Error(err.Error())
	}
	if string(contents) != "f1" {
		t.Error("Read: Incorrect file contents")
	}

	// write to a file
	err = os.WriteFile(path.Join(mountDir, "/$/1.txt"), []byte("new contents"), 0)
	if err != nil {
		t.Error(err.Error())
	}
	contents, err = os.ReadFile(path.Join(mountDir, "/$/1.txt"))
	if err != nil {
		t.Error(err.Error())
	}
	if string(contents) != "new contents" {
		t.Error("Read: Incorrect file contents")
	}

	// create a new file
	err = os.WriteFile(path.Join(mountDir, "/a/a1/+/b/new.txt"), []byte("new file"), 0)
	if err != nil {
		t.Error(err.Error())
	}
	files, _ := os.ReadDir(path.Join(mountDir, "/$/"))
	t.Log(files)
	contents, err = os.ReadFile(path.Join(mountDir, "/$/3.txt"))
	if err != nil {
		t.Error(err.Error())
	}
	if string(contents) != "new file" {
		t.Error("Read: Incorrect file contents")
	}

}
