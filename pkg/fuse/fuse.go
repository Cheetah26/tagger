// Some functions modified from https://github.com/winfsp/cgofuse/blob/master/examples/passthrough/passthrough.go

package fuse

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
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

func fileToFilename(file tagger.File) string {
	return fmt.Sprintf("%d.%s", file.Id, file.Filetype)
}

func pathToDirs(path string) []string {
	path = filepath.Clean(path)

	return strings.Split(path, string(filepath.Separator))
}

// A null file handle
const nullFH = ^uint64(0)

var (
	dirStat  = &fuse.Stat_t{Mode: fuse.S_IFDIR | 0755}
	fileStat = &fuse.Stat_t{Mode: fuse.S_IFREG | 0755}
)

type TaggerFS struct {
	t *tagger.Tagger

	errChan chan error

	openFiles      map[uint64]*os.File
	openFilesMutex sync.Mutex

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
		t:         tagger,
		errChan:   errChan,
		openFiles: make(map[uint64]*os.File),
	}

	host := fuse.NewFileSystemHost(tfs)

	// Mount / unmount code modified from rclone
	// https://github.com/rclone/rclone/blob/839eef0db269333870dc04cb79d0dd0c95e5a418/cmd/cmount/mount.go

	// Serve the mount point in the background returning error to errChan
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("fuse panic: %v", r)
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

// Attempt to get a file from its handle, or path if the handle is nil. Returns nil file if no such file is open
func (tfs *TaggerFS) getFileEntry(path string, fh uint64) *os.File {
	if fh == nullFH {
		for _, f := range tfs.openFiles {
			if f.Name() == path {
				return f
			}
		}
		return nil
	}

	return tfs.openFiles[fh]
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

// Utimens changes the access and modification times of a file.
func (tfs *TaggerFS) Utimens(path string, tmsp []fuse.Timespec) (errc int) {
	return -fuse.ENOSYS
}

// Create creates and opens a file.
// The flags are a combination of the fuse.O_* constants.
func (tfs *TaggerFS) Create(path string, flags int, mode uint32) (errc int, fh uint64) {
	return -fuse.ENOSYS, nullFH
}

// Open opens a file.
// The flags are a combination of the fuse.O_* constants.
func (tfs *TaggerFS) Open(path string, flags int) (errc int, fh uint64) {
	id := pathToFileId(path)

	file, err := tfs.t.GetFile(id)
	if err != nil {
		if errors.Is(err, tagger.ErrFileNotExist) {
			return -fuse.ENOSYS, nullFH
		} else {
			tfs.errChan <- NewTFSError("Open", err)
			return -fuse.EIO, nullFH
		}
	}

	sourceFile, err := os.OpenFile(
		tfs.t.GetFilepath(file),
		flags,
		os.FileMode(0),
	)

	if err != nil {
		tfs.errChan <- NewTFSError("Open", err)
		return -fuse.EIO, nullFH
	}

	tfs.openFilesMutex.Lock()
	defer tfs.openFilesMutex.Unlock()

	handle := uint64(0)
	for {
		if tfs.openFiles[handle] == nil {
			break
		}
		handle += 1
	}

	tfs.openFiles[handle] = sourceFile

	return 0, handle
}

// Getattr gets file attributes.
// This appears to frequently be called with fh == nullFH, so the handle is ignored in favor of the file path
func (tfs *TaggerFS) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	id := pathToFileId(path)

	file, err := tfs.t.GetFile(id)
	if err != nil && !errors.Is(err, tagger.ErrFileNotExist) {
		tfs.errChan <- NewTFSError("Getattr", err)
		return -fuse.ENOENT
	}

	// tag / directory
	if file == nil {
		stat.Mode = dirStat.Mode
		return 0
	}

	// file
	info, err := os.Stat(tfs.t.GetFilepath(file))
	if err != nil {
		tfs.errChan <- NewTFSError("Getattr", err)
		return -fuse.ENOENT
	}

	stat.Mode = fileStat.Mode
	stat.Size = info.Size()
	stat.Mtim = fuse.NewTimespec(info.ModTime())

	UID, GID, _ := fuse.Getcontext()
	stat.Uid = UID
	stat.Gid = GID

	return 0
}

// Truncate changes the size of a file.
func (tfs *TaggerFS) Truncate(path string, size int64, fh uint64) int {
	file := tfs.getFileEntry(path, fh)
	if file == nil {
		return -fuse.ENOENT
	}

	err := file.Truncate(size)
	if err != nil {
		tfs.errChan <- NewTFSError("Truncate", err)
		return -fuse.EIO
	}

	return 0
}

// Read reads data from a file.
func (tfs *TaggerFS) Read(path string, buff []byte, offset int64, fh uint64) (numBytes int) {
	file := tfs.getFileEntry(path, fh)
	if file == nil {
		return -fuse.ENOENT
	}

	numBytes, err := file.ReadAt(buff, offset)
	if err != nil && !errors.Is(err, io.EOF) {
		tfs.errChan <- NewTFSError("Read", err)
		return -fuse.EIO
	}

	return numBytes
}

// Write writes data to a file.
func (tfs *TaggerFS) Write(path string, buff []byte, offset int64, fh uint64) (numBytes int) {
	file := tfs.getFileEntry(path, fh)
	if file == nil {
		return -fuse.ENOENT
	}

	numBytes, err := file.WriteAt(buff, offset)
	if err != nil {
		tfs.errChan <- NewTFSError("Write", err)
		return -fuse.EIO
	}

	return numBytes
}

// Release closes an open file.
func (tfs *TaggerFS) Release(path string, fh uint64) (errc int) {
	tfs.openFilesMutex.Lock()
	defer tfs.openFilesMutex.Unlock()

	file := tfs.getFileEntry(path, fh)
	if file == nil {
		return -fuse.ENOENT
	}

	err := file.Close()
	if err != nil {
		tfs.errChan <- NewTFSError("Release", err)
		return -fuse.EIO
	}

	for h, f := range tfs.openFiles {
		if f == file {
			tfs.openFiles[h] = nil
		}
	}

	return 0
}

// Fsync synchronizes file contents.
func (tfs *TaggerFS) Fsync(path string, datasync bool, fh uint64) (errc int) {
	return -fuse.ENOSYS
}

// Opendir opens a directory.
// Since none of the directories are "real", this only checks if the path is valid and always returns a null handle
func (tfs *TaggerFS) Opendir(path string) (errc int, fh uint64) {
	for _, dir := range pathToDirs(path) {
		// skip valid parts of the path which aren't a tag
		if dir == "" || dir == "+" || dir == "$" {
			continue
		}

		_, err := tfs.t.GetTagByName(dir)
		if err != nil {
			if errors.Is(err, tagger.ErrTagNotExist) {
				// this dir doesn't exist if its path includes a non-tag
				return -fuse.ENOTDIR, nullFH
			} else {
				tfs.errChan <- NewTFSError("Opendir", err)
				return -fuse.EIO, nullFH
			}
		}
	}

	return 0, nullFH
}

// Readdir reads a directory.
func (tfs *TaggerFS) Readdir(
	path string,
	fill func(name string, stat *fuse.Stat_t, offset int64) bool,
	offset int64,
	fh uint64,
) (errc int) {
	dirs := pathToDirs(path)
	lastDir := dirs[len(dirs)-1]

	tags, err := tfs.t.GetAllTags()
	if err != nil {
		tfs.errChan <- NewTFSError("Readdir", err)
		return -fuse.EIO
	}

	var selectedTagIds []tagger.TagID
	var selectedTags []tagger.Tag
	for _, tag := range tags {
		if slices.Contains(dirs, tag.Name) {
			selectedTagIds = append(selectedTagIds, tag.Id)
			selectedTags = append(selectedTags, tag)
		}
	}

	fill(".", dirStat, 0)
	fill("..", dirStat, 0)

	switch lastDir {
	case "$": // show files matching the current tag selection, or all tags if no selection
		var files []tagger.File
		if len(selectedTags) > 0 {
			files, err = tfs.t.GetFilesByTag(selectedTags)
		} else {
			files, err = tfs.t.GetAllFiles()
		}
		if err != nil {
			tfs.errChan <- NewTFSError("Readdir", err)
			return -fuse.EIO
		}

		for _, file := range files {
			name := fileToFilename(file)
			fill(name, fileStat, 0)
		}

	case "": // fill with unselected root-level tags on the root or the '+' special folder
		fill("$", dirStat, 0) // add the show files folder
		fallthrough
	case "+":
		for _, tag := range tags {
			files, err := tfs.t.GetFilesByTag(append(selectedTags, tag))
			if err != nil {
				tfs.errChan <- NewTFSError("Readdir", err)
				continue
			}

			if len(files) > 0 && len(tag.Parents) == 0 && !slices.Contains(selectedTagIds, tag.Id) {
				fill(tag.Name, dirStat, 0)
			}
		}
	default: // show folders for the current dir's child tags and both special folders
		fill("+", dirStat, 0)
		fill("$", dirStat, 0)

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
				tfs.errChan <- NewTFSError("Readdir", err)
				continue
			}

			if len(files) > 0 && slices.Contains(tag.Parents, currentTag.Id) {
				fill(tag.Name, dirStat, 0)
			}
		}
	}

	return 0
}

// Releasedir closes an open directory.
func (tfs *TaggerFS) Releasedir(path string, fh uint64) (errc int) {
	return 0
}
