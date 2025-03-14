package tagger_test

import (
	"testing"

	"github.com/cheetah26/tagger/pkg/tagger"
)

func compareTags(a, b *tagger.Tag) bool {
	parentsMatch := true
	for i := range len(a.Parents) {
		if !compareTags(&a.Parents[i], &b.Parents[i]) {
			parentsMatch = false
			break
		}
	}
	return a == b || (a.Name == b.Name && parentsMatch)
}

func TestInsertAndRemoveTags(t *testing.T) {
	tr := openTestDB(t.TempDir())

	name := "tag1"
	actual := tr.AddTag(name)
	if !compareTags(actual, &tagger.Tag{Id: 0, Name: name, Parents: []tagger.Tag{}}) {
		t.Error("Insert fail: Inserted tag does not match")
		return
	}

	parent := "parent"
	child := "child"
	notChild := "not child"

	tr.AddTag(parent)
	tr.AddTag(child)
	tr.AddTag(notChild)

	if len(tr.GetAllTags()) != 3 {
		t.Error("Insert Fail: Incorrect number of tags inserted")
		return
	}
}

// tag tag
// update tag
// get parent tags
// get child tags
// don't allow child to become parent of its parent (infinite recursion)
