package fuse

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cheetah26/tagger/pkg/tagger"
	"github.com/winfsp/cgofuse/fuse"
)

func pathToId(path string) int64 {
	_, fullname := filepath.Split(path)
	idString := strings.TrimSuffix(fullname, filepath.Ext(fullname))
	id, _ := strconv.ParseInt(idString, 10, 64)
	return id
}

var ErrMountFailed = errors.New("mount failed")
var ErrUnmountFailed = errors.New("unmount failed")

type TaggerFS struct {
	t *tagger.Tagger

	errChan chan error
	// Inherit default filesystem actions
	fuse.FileSystemBase
}

// Mount a fuse filesystem at mountpoint to browse tags and files outside of Tagger.
//
// Tags are organized in a tree structure. The root level shows tags without any parents as folders. Opening one of these folders will show the tag's children, as well as the special folders '+' and '$'. The plus folder is used to add another root-level tag to the search (always an AND operation), while the '$' folder shows files which meet the current search criteria.
//
// The error channel sends errors for FUSE, as well as any Tagger errors encountered by the filesystem implementation.
//
// Incomplete example of tag structure, with special folders omitted:
//
//	mountpoint
//	 |- parent tag
//	 |   |- child tag
//	 |   \- +
//	 |       \- other parent
//	 \- other parent
func Mount(mountpoint string, tagger *tagger.Tagger) (func() error, <-chan error, error) {
	errChan := make(chan error, 1)

	tfs := &TaggerFS{
		t:       tagger,
		errChan: errChan,
	}

	host := fuse.NewFileSystemHost(tfs)
	host.SetCapReaddirPlus(true)

	// Mount / unmount code modified from rclone
	// https://github.com/rclone/rclone/blob/839eef0db269333870dc04cb79d0dd0c95e5a418/cmd/cmount/mount.go

	// Serve the mount point in the background returning error to errChan
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("fuse error: %v", r)
			}
		}()
		ok := host.Mount(mountpoint, nil)
		if !ok {
			errChan <- ErrMountFailed
		}
	}()

	unmount := func() error {
		ok := host.Unmount()
		if !ok {
			return ErrUnmountFailed
		}
		return nil
	}

	// wait 20ms for any errors, assume the mount is ok after that
	select {
	case err := <-errChan:
		return nil, nil, err
	case <-time.After(20 * time.Millisecond):
	}

	return unmount, errChan, nil
}

func (tfs *TaggerFS) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	id := pathToId(path)

	file, err := tfs.t.GetFile(id)
	if err != nil && !errors.Is(err, tagger.ErrFileNotExist) {
		tfs.errChan <- err
		return -fuse.ENOENT
	}

	// tag / directory
	if file == nil {
		stat.Mode = fuse.S_IFDIR
		return 0
	}

	// file
	info, _ := os.Stat(tfs.t.GetFilepath(file))

	stat.Mode = fuse.S_IFREG
	stat.Size = info.Size()
	stat.Mtim = fuse.NewTimespec(info.ModTime())
	return 0
}

func (tfs *TaggerFS) Read(path string, buff []byte, offset int64, fh uint64) (n int) {
	id := pathToId(path)

	file, err := tfs.t.GetFile(id)
	if err != nil {
		return -fuse.ENOENT
	}

	realPath := tfs.t.GetFilepath(file)
	f, _ := os.Open(realPath)

	n, _ = f.ReadAt(buff, offset)
	return n
}

func (tfs *TaggerFS) Readdir(path string, fill func(name string, stat *fuse.Stat_t, offset int64) bool, offset int64, fh uint64) (errc int) {
	dirStat := &fuse.Stat_t{Mode: fuse.S_IFDIR}
	fileStat := &fuse.Stat_t{Mode: fuse.S_IFREG}

	pathParts := strings.Split(path, "/")[1:]
	lastDir := pathParts[len(pathParts)-1]

	tags, err := tfs.t.GetAllTags()
	if err != nil {
		tfs.errChan <- err
		return -fuse.EIO
	}

	var selectedTagIds []int64
	var selectedTags []tagger.Tag
	for _, tag := range tags {
		if slices.Contains(pathParts, tag.Name) {
			selectedTagIds = append(selectedTagIds, tag.Id)
			selectedTags = append(selectedTags, tag)
		}
	}

	// show files matching the selection for the '$' folder
	if lastDir == "$" {
		// get files matching the current search, or all files if no tags are selected
		var files []tagger.File
		if len(selectedTags) > 0 {
			files, err = tfs.t.GetFilesByTag(selectedTags)
		} else {
			files, err = tfs.t.GetAllFiles()
		}
		if err != nil {
			tfs.errChan <- err
			return -fuse.EIO
		}

		for _, file := range files {
			name := fmt.Sprintf("%d.%s", file.Id, file.Filetype)
			fill(name, fileStat, 0)
		}
		return 0
	}

	// add the show files folder
	fill("$", dirStat, 0)

	// fill with unselected root-level tags on the root mountpoint or the '+' special folder
	if path == "/" || lastDir == "+" {
		for _, tag := range tags {
			files, err := tfs.t.GetFilesByTag(append(selectedTags, tag))
			if err != nil {
				tfs.errChan <- err
				continue
			}

			if len(files) > 0 && len(tag.Parents) == 0 && !slices.Contains(selectedTagIds, tag.Id) {
				fill(tag.Name, dirStat, 0)
			}
		}
		return 0
	}

	// otherwise, put the current dir's children as well as the '+' plus folder
	fill("+", dirStat, 0)

	var currentTag *tagger.Tag
	for _, tag := range tags {
		if tag.Name == lastDir {
			currentTag = &tag
			break
		}
	}
	if currentTag == nil {
		return -fuse.ENOENT
	}

	for _, tag := range tags {
		files, err := tfs.t.GetFilesByTag(append(selectedTags, tag))
		if err != nil {
			tfs.errChan <- err
			continue
		}

		if len(files) > 0 && slices.Contains(tag.Parents, currentTag.Id) {
			fill(tag.Name, dirStat, 0)
		}
	}

	return 0
}
