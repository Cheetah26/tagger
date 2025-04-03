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

func pathToFileId(path string) int64 {
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

// Statfs gets file system statistics.
func (tfs *TaggerFS) Statfs(path string, stat *fuse.Statfs_t) (errc int) {
	// TODO: show the total size of the tagger files
	stat.Frsize = 1024
	stat.Blocks = 1024
	stat.Bfree = 626
	stat.Bavail = stat.Bfree
	stat.Namemax = 255
	return 0
}

// Mknod creates a file node.
func (tfs *TaggerFS) Mknod(path string, mode uint32, dev uint64) (errc int) {
	return -fuse.ENOSYS
}

// Mkdir creates a directory.
func (tfs *TaggerFS) Mkdir(path string, mode uint32) (errc int) {
	return -fuse.ENOSYS
}

// Rmdir removes a directory.
func (tfs *TaggerFS) Rmdir(path string) (errc int) {
	return -fuse.ENOSYS
}

// Rename renames a file.
func (tfs *TaggerFS) Rename(oldpath string, newpath string) (errc int) {
	return -fuse.ENOSYS
}

// Chmod changes the permission bits of a file.
func (tfs *TaggerFS) Chmod(path string, mode uint32) (errc int) {
	return -fuse.ENOSYS
}

// Chown changes the owner and group of a file.
func (tfs *TaggerFS) Chown(path string, uid uint32, gid uint32) (errc int) {
	return -fuse.ENOSYS
}

// Utimens changes the access and modification times of a file.
func (tfs *TaggerFS) Utimens(path string, tmsp []fuse.Timespec) (errc int) {
	return -fuse.ENOSYS
}

// Create creates and opens a file.
// The flags are a combination of the fuse.O_* constants.
func (tfs *TaggerFS) Create(path string, flags int, mode uint32) (errc int, fh uint64) {
	return -fuse.ENOSYS, ^uint64(0)
}

// Open opens a file.
// The flags are a combination of the fuse.O_* constants.
func (tfs *TaggerFS) Open(path string, flags int) (errc int, fh uint64) {
	id := pathToFileId(path)
	_, err := tfs.t.GetFile(id)
	if err != nil {
		if errors.Is(err, tagger.ErrFileNotExist) {
			return -fuse.ENOSYS, ^uint64(0)
		} else {
			tfs.errChan <- err
			return -fuse.EIO, ^uint64(0)
		}
	}

	return 0, ^uint64(0)
}

// Getattr gets file attributes.
func (tfs *TaggerFS) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	id := pathToFileId(path)

	file, err := tfs.t.GetFile(id)
	if err != nil && !errors.Is(err, tagger.ErrFileNotExist) {
		tfs.errChan <- err
		return -fuse.ENOENT
	}

	// tag / directory
	if file == nil {
		stat.Mode = fuse.S_IFDIR | 0777
		return 0
	}

	// file
	info, err := os.Stat(tfs.t.GetFilepath(file))
	if err != nil {
		tfs.errChan <- err
		return -fuse.ENOENT
	}

	stat.Mode = fuse.S_IFREG | 0777
	stat.Size = info.Size()
	stat.Mtim = fuse.NewTimespec(info.ModTime())

	return 0
}

// Read reads data from a file.
func (tfs *TaggerFS) Read(path string, buff []byte, offset int64, fh uint64) (errc int) {
	id := pathToFileId(path)

	file, err := tfs.t.GetFile(id)
	if err != nil {
		tfs.errChan <- err
		return -fuse.ENOENT
	}

	realPath := tfs.t.GetFilepath(file)
	f, _ := os.Open(realPath)

	bytes, _ := f.ReadAt(buff, offset)
	return bytes
}

// Write writes data to a file.
func (tfs *TaggerFS) Write(path string, buff []byte, ofst int64, fh uint64) (errc int) {
	return -fuse.ENOSYS
}

// Release closes an open file.
func (tfs *TaggerFS) Release(path string, fh uint64) (errc int) {
	return -fuse.ENOSYS
}

// Fsync synchronizes file contents.
func (tfs *TaggerFS) Fsync(path string, datasync bool, fh uint64) (errc int) {
	return -fuse.ENOSYS
}

// Opendir opens a directory.
func (tfs *TaggerFS) Opendir(path string) (errc int, fh uint64) {
	pathParts := strings.Split(path, "/")[1:]

	for _, part := range pathParts {
		switch part {
		case "":
			continue
		case "+":
			continue
		case "$":
			continue

		default:
			_, err := tfs.t.GetTagByName(part)
			if err != nil {
				if errors.Is(err, tagger.ErrTagNotExist) {
					// a path is invalid if any of its parts are neither a tag or any of the special folders
					return -fuse.ENOTDIR, ^uint64(0)
				} else {
					tfs.errChan <- err
					return -fuse.EIO, ^uint64(0)
				}
			}
			continue
		}
	}

	return 0, ^uint64(0)
}

// Readdir reads a directory.
func (tfs *TaggerFS) Readdir(
	path string,
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64,
) (errc int) {
	dirStat := &fuse.Stat_t{Mode: fuse.S_IFDIR | 0777}
	fileStat := &fuse.Stat_t{Mode: fuse.S_IFREG | 0777}

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

	fill(".", nil, 0)
	fill("..", nil, 0)

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

// Releasedir closes an open directory.
func (tfs *TaggerFS) Releasedir(path string, fh uint64) (errc int) {
	return -fuse.ENOSYS
}
