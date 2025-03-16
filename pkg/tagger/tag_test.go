package tagger_test

import (
	"errors"
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

	allTags, err := tr.GetAllTags()
	if err != nil {
		t.Error("Unable to get all tags")
	}
	if len(allTags) != 0 {
		t.Error("Initial state: Incorrect number of tags in database")
	}

	// add tags
	for i := range COUNT {
		name := getTagName(i)

		tag, err := tr.AddTag(name)
		if tag == nil || err != nil {
			t.Error("Insert fail: Error while inserting tag")
		} else {
			if tag.Name != name {
				t.Error("Insert fail: Inserted tag has incorrect name")
			}
			if len(tag.Parents) != 0 {
				t.Error("Insert fail: Inserted tag has parents")
			}

			allTags, err := tr.GetAllTags()
			if err != nil {
				t.Error("Unable to get all tags")
			}
			if len(allTags) != i+1 {
				t.Error("Insert fail: Incorrect number of tags in database")
			}
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
		got, err := tr.GetTag(name)
		if got != nil {
			t.Error("Remove fail: Still able to retrieve tag")
		}
		if !errors.Is(err, tagger.ErrTagNotExist) {
			t.Error("Remove fail: No error returned")
		}

		allTags, err := tr.GetAllTags()
		if err != nil {
			t.Error("Unable to get all tags")
		}
		if len(allTags) != COUNT-i-1 {
			t.Error("Remove fail: Incorrect number of tags in database")
		}
	}
}

func TestUpdate(t *testing.T) {
	tr := openTestDB(t.TempDir())

	tag, _ := tr.AddTag("initial")

	tag.Name = "updated"
	err := tr.UpdateTag(*tag)
	if err != nil {
		t.Error("Update fail: Error while updating tag")
	}

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

	parent, _ := tr.AddTag("parent")
	child, _ := tr.AddTag("child")
	grandchild, _ := tr.AddTag("grandchild")

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

func TestNonexistent(t *testing.T) {
	tr := openTestDB(t.TempDir())

	tag, err := tr.GetTag("fake")
	if tag != nil {
		t.Error("Get: Results returned for nonexistent tag")
	}
	if !errors.Is(err, tagger.ErrTagNotExist) {
		t.Error("Get: No error returned on nonexistent tag")
	}

	allTags, err := tr.GetAllTags()
	if allTags != nil {
		t.Error("Get all: Results returned when no tags in database")
	}
	if err != nil {
		t.Error("Get all: Error when no tags in database")
	}

	fake := tagger.Tag{
		Id:      22,
		Name:    "fake tag",
		Parents: []tagger.Tag{},
	}

	err = tr.UpdateTag(fake)
	if !errors.Is(err, tagger.ErrTagNotExist) {
		t.Error("Update: No error returned when updating nonexistent tag")
	}

	err = tr.RemoveTag(fake)
	if !errors.Is(err, tagger.ErrTagNotExist) {
		t.Error("Remove: No error returned when removing nonexistent tag")
	}

	tags, err := tr.GetChildTags(fake)
	if err != nil {
		t.Error("Get children: Error returned")
	}
	if tags != nil {
		t.Error("Get children: Value returned while children of nonexistent tag")
	}

	tags, err = tr.GetParentTags(fake)
	if err != nil {
		t.Error("Get parents: Error returned")
	}
	if tags != nil {
		t.Error("Get parents: Value returned while parents of nonexistent tag")
	}
}

// TODO tests that should fail:
// don't allow child to become parent of its parent (infinite recursion)
// update with bad values
