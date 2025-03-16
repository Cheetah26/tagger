package tagger_test

import (
	"strconv"
	"testing"

	"github.com/cheetah26/tagger/pkg/tagger"
)

func compareTagParents(a, b *tagger.Tag) bool {
	if len(a.Parents) != len(b.Parents) {
		return false
	}

	for i := range len(a.Parents) {
		if a.Id != b.Id || !compareTagParents(&a.Parents[i], &b.Parents[i]) {
			return false
		}
	}

	return true
}

func TestInsertAndRemove(t *testing.T) {
	// setup
	const COUNT = 200

	tr := openTestDB(t.TempDir())

	getTagName := func(num int) string {
		return "tag" + strconv.Itoa(num)
	}

	// add tags
	for i := range COUNT {
		name := getTagName(i)

		tag := tr.AddTag(name)
		if tag.Name != name {
			t.Error("Insert fail: Inserted tag has incorrect name")
		}
		if len(tag.Parents) != 0 {
			t.Error("Insert fail: Inserted tag has parents")
		}

		if len(tr.GetAllTags()) != i+1 {
			t.Error("Insert fail: Incorrect number of tags in database")
		}
	}

	// get and remove tags
	for i := range COUNT {
		name := getTagName(i)

		tag, err := tr.GetTag(name)
		if tag == nil || err != nil {
			t.Error("Get fail: Error getting tag")
			continue
		}
		if tag.Name != name {
			t.Error("Get fail: Retrieved tag has incorrect name")
		}
		if len(tag.Parents) != 0 {
			t.Error("Get fail: Retrieved tag has parents")
		}

		err = tr.RemoveTag(*tag)
		if err != nil {
			t.Error("Remove fail: Error removing tag")
		}
		if got, err := tr.GetTag(name); got != nil || err != nil {
			t.Error("Remove fail: Still able to retrieve tag")
		}

		if len(tr.GetAllTags()) != COUNT-i-1 {
			t.Error("Remove fail: Incorrect number of tags in database")
		}
	}
}

func TestUpdate(t *testing.T) {
	tr := openTestDB(t.TempDir())

	tag := tr.AddTag("initial")

	tag.Name = "updated"
	tr.UpdateTag(*tag)

	if got, err := tr.GetTag("child"); got != nil && err == nil {
		t.Error("Update fail: Name not changed")
	}

	if got, err := tr.GetTag("updated"); got == nil || err != nil {
		t.Error("Update fail: Unable to retrieve by new name")
	}
}

func TestParents(t *testing.T) {
	// setup
	tr := openTestDB(t.TempDir())

	parent := tr.AddTag("parent")
	child := tr.AddTag("child")
	grandchild := tr.AddTag("grandchild")

	// add parents
	child.Parents = append(child.Parents, *parent)
	tr.UpdateTag(*child)
	grandchild.Parents = append(grandchild.Parents, *child)
	tr.UpdateTag(*grandchild)

	updated, err := tr.GetTag("grandchild")
	if err != nil {
		t.Error("Add parent fail: Unable to get grandchild")
	}
	if !compareTagParents(updated, grandchild) {
		t.Error("Add parent fail: Parent tag tree is incorrect")
	}

	// get child from parent
	got, err := tr.GetChildTags(*parent)
	if got == nil || err != nil {
		t.Error("Get child failed: Error retrieving child tag")
	}
	if len(got) != 1 {
		t.Error("Get child failed: Incorrect number of results")
	}
	if got[0].Id != child.Id {
		t.Error("Get child failed: Retrieved tag is incorrect")
	}

	// remove parent from child
	child.Parents = []tagger.Tag{}
	tr.UpdateTag(*child)

	updated, err = tr.GetTag("child")
	if err != nil {
		t.Error("Remove parent fail: Unable to get child")
	}
	if !compareTagParents(updated, child) {
		t.Error("Remove parent fail: Parent tag not removed")
	}
}

// TODO tests that should fail:
// don't allow child to become parent of its parent (infinite recursion)
// get non-existant tag
// update non-existant tag
// remove non-existant tag
// update with bad values
