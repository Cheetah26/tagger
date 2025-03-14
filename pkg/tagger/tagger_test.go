package tagger_test

import (
	"os"
	"path"

	"github.com/cheetah26/tagger/pkg/tagger"
)

func openTestDB(dir string) *tagger.Tagger {
	fullPath := path.Join(dir, "test.db")

	_, err := os.Create(fullPath)
	if err != nil {
		panic("Failed to create test database file")
	}

	tr, err := tagger.Open(fullPath)
	if err != nil {
		panic("Failed to initialize test database")
	}

	return tr
}
