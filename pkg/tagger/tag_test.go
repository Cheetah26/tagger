package tagger_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/cheetah26/tagger/pkg/tagger"
)

func compareTagRelatives(tr *tagger.Tagger, a, b *tagger.Tag) bool {
	if (len(a.Parents) != len(b.Parents)) ||
		(len(a.Children) != len(b.Children)) {
		return false
	}

	for i := range len(a.Parents) {
		parentA, err := tr.GetTagById(a.Parents[i])
		if err != nil {
			return false
		}

		parentB, err := tr.GetTagById(b.Parents[i])
		if err != nil {
			return false
		}

		if parentA.Id != parentB.Id || !compareTagRelatives(tr, parentA, parentB) {
			return false
		}
	}

	for i := range len(a.Children) {
		childA, err := tr.GetTagById(a.Children[i])
		if err != nil {
			return false
		}

		childB, err := tr.GetTagById(b.Children[i])
		if err != nil {
			return false
		}

		if childA.Id != childB.Id || !compareTagRelatives(tr, childA, childB) {
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

		tag, err := tr.GetTagByName(name)
		if tag == nil || err != nil {
			t.Error("Get fail: Error getting tag by name")
			continue
		}
		if len(tag.Parents) != 0 {
			t.Error("Get fail: Retrieved tag has parents")
		}

		err = tr.RemoveTag(tag)
		if err != nil {
			t.Error("Remove fail: Error removing tag")
		}
		got, err := tr.GetTagByName(name)
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
	err := tr.UpdateTag(tag)
	if err != nil {
		t.Error("Update fail: Error while updating tag")
	}

	if got, err := tr.GetTagByName("initial"); got != nil && err == nil {
		t.Error("Update fail: Name not changed")
	}

	if got, err := tr.GetTagByName("updated"); got == nil || err != nil {
		t.Error("Update fail: Unable to retrieve by new name")
	}

	if got, err := tr.GetTagById(tag.Id); got == nil || err != nil {
		t.Error("Update fail: Unable to retrieve by id")
	}
}

func TestParents(t *testing.T) {
	// setup
	tr := openTestDB(t.TempDir())

	parent, _ := tr.AddTag("parent")
	child, _ := tr.AddTag("child")
	grandchild, _ := tr.AddTag("grandchild")

	// add parents
	child.Parents = append(child.Parents, parent.Id)
	tr.UpdateTag(child)
	grandchild.Parents = append(grandchild.Parents, child.Id)
	tr.UpdateTag(grandchild)

	updated, err := tr.GetTagById(child.Id)
	if err != nil {
		t.Error("Add parent: Unable to get child")
	}
	if !compareTagRelatives(tr, updated, child) {
		t.Error("Add parent: Child's parent tags are incorrect")
	}
	updated, err = tr.GetTagById(grandchild.Id)
	if err != nil {
		t.Error("Add parent: Unable to get grandchild")
	}
	if !compareTagRelatives(tr, updated, grandchild) {
		t.Error("Add parent: Grandchild's parents are incorrect")
	}

	// remove parent from child
	child.Parents = []tagger.TagID{}
	tr.UpdateTag(child)

	updated, err = tr.GetTagById(child.Id)
	if err != nil {
		t.Error("Remove parent: Unable to get child")
	}
	if !compareTagRelatives(tr, updated, child) {
		t.Error("Remove parent: Child's parent tags are incorrect")
	}
}

func TestNonexistent(t *testing.T) {
	tr := openTestDB(t.TempDir())

	tag, err := tr.GetTagById(35)
	if tag != nil {
		t.Error("Get: Results returned for nonexistent tag")
	}
	if !errors.Is(err, tagger.ErrTagNotExist) {
		t.Error("Get: No error returned on nonexistent tag")
	}

	tag, err = tr.GetTagByName("fake")
	if tag != nil {
		t.Error("Get: Results returned for nonexistent tag")
	}
	if !errors.Is(err, tagger.ErrTagNotExist) {
		t.Error("Get: No error returned on nonexistent tag")
	}

	allTags, err := tr.GetAllTags()
	if len(allTags) != 0 {
		t.Error("Get all: Results returned when no tags in database")
	}
	if err != nil {
		t.Error("Get all: Error when no tags in database")
	}

	fake := &tagger.Tag{
		Id:      22,
		Name:    "fake tag",
		Parents: []tagger.TagID{},
	}

	err = tr.UpdateTag(fake)
	if !errors.Is(err, tagger.ErrTagNotExist) {
		t.Error("Update: No error returned when updating nonexistent tag")
	}

	err = tr.RemoveTag(fake)
	if !errors.Is(err, tagger.ErrTagNotExist) {
		t.Error("Remove: No error returned when removing nonexistent tag")
	}
}

// TODO tests that should fail:
// don't allow child to become parent of its parent (infinite recursion)
// update with bad values
