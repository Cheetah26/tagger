// Some functions modified from https://github.com/winfsp/cgofuse/blob/master/examples/passthrough/passthrough.go

package fuse

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
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

var (
	ErrMountFailed   = errors.New("mount failed")
	ErrUnmountFailed = errors.New("unmount failed")
)

// A null file handle
const nullFH = ^uint64(0)

var (
	dirStat  = &fuse.Stat_t{Mode: fuse.S_IFDIR | 0700}
	fileStat = &fuse.Stat_t{Mode: fuse.S_IFREG | 0700}
)

type TaggerFS struct {
	t *tagger.Tagger

	errChan chan error // TODO: add a custom wrapper for errors. currently impossible to tell which function the error originates from

	tempFiles []tempFile

	// Inherit default filesystem actions
	fuse.FileSystemBase
}

type tempFile struct {
	requestedPath string
	tempPath      string
	fh            uint64
}

// Mount a fuse filesystem at mountpoint to browse tags and files outside of Tagger.
//
// Tags are organized in a tree structure. The root level shows tags without any parents as folders. Opening one of these folders will show the tag's children, as well as the special folders '+' and '$'. The plus folder is used to add another root-level tag to the search (always an AND operation), while the '$' folder shows files which meet the current search criteria.
//
// The error channel sends errors for FUSE, as well as any Tagger errors encountered by the filesystem implementation. When this happens, an I/O error is returned to the OS.
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
func (tfs *TaggerFS) Utimens(path string, timespec []fuse.Timespec) (errc int) {
	// newly created files
	for _, tempFile := range tfs.tempFiles {
		if path == tempFile.requestedPath {
			ts := make([]syscall.Timespec, 2)
			ts[0].Sec, ts[0].Nsec = timespec[0].Sec, timespec[0].Nsec
			ts[1].Sec, ts[1].Nsec = timespec[1].Sec, timespec[1].Nsec

			err := syscall.UtimesNano(tempFile.tempPath, ts)
			if err != nil {
				tfs.errChan <- err
				return -fuse.EIO
			}
			return 0
		}
	}

	// existing files
	id := pathToFileId(path)

	file, err := tfs.t.GetFile(id)
	if err != nil {
		if errors.Is(err, tagger.ErrFileNotExist) {
			return -fuse.ENOENT
		} else {
			tfs.errChan <- err
			return -fuse.EIO
		}
	}

	ts := make([]syscall.Timespec, 2)
	ts[0].Sec, ts[0].Nsec = timespec[0].Sec, timespec[0].Nsec
	ts[1].Sec, ts[1].Nsec = timespec[1].Sec, timespec[1].Nsec

	err = syscall.UtimesNano(tfs.t.GetFilepath(file), ts)
	if err != nil {
		tfs.errChan <- err
		return -fuse.EIO
	}

	return 0
}

// Create creates and opens a file.
// The flags are a combination of the fuse.O_* constants.
func (tfs *TaggerFS) Create(path string, flags int, mode uint32) (errc int, fh uint64) {
	_, requestedName := filepath.Split(path)

	// create a temporary file
	tempPath := filepath.Join(os.TempDir(), requestedName)
	tempFH, err := syscall.Open(tempPath, flags, mode)
	if err != nil {
		tfs.errChan <- err
		return -fuse.EIO, nullFH
	}

	tfs.tempFiles = append(tfs.tempFiles, tempFile{
		requestedPath: path,
		tempPath:      tempPath,
		fh:            uint64(tempFH),
	})

	return 0, uint64(tempFH)
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
			tfs.errChan <- err
			return -fuse.EIO, nullFH
		}
	}

	handle, err := syscall.Open(
		tfs.t.GetFilepath(file),
		flags,
		0,
	)

	if err != nil {
		tfs.errChan <- err
		return -fuse.EIO, nullFH
	}

	return 0, uint64(handle)

}

// Getattr gets file attributes.
// This appears to frequently be called with fh == nullFH, so the handle is ignored in favor of the file path
func (tfs *TaggerFS) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	dirs := pathToDirs(path)
	lastDir := dirs[len(dirs)-1]

	// check that the leading path is valid
	for _, dir := range dirs[:1] {
		// skip valid parts of the path which aren't a tag
		if dir == "" || dir == "+" || dir == "$" {
			continue
		}

		_, err := tfs.t.GetTagByName(dir)
		if err != nil {
			if errors.Is(err, tagger.ErrTagNotExist) {
				// this path doesn't exist if it includes a non-tag
				return -fuse.ENOENT
			} else {
				tfs.errChan <- err
				return -fuse.EIO
			}
		}
	}

	// is the final entry a special folder
	if lastDir == "" || lastDir == "+" || lastDir == "$" {
		stat.Mode = dirStat.Mode
		return 0
	}

	// is the final entry a tag
	tag, err := tfs.t.GetTagByName(lastDir)
	if err != nil && !errors.Is(err, tagger.ErrTagNotExist) {
		tfs.errChan <- err
		return -fuse.EIO
	}
	if tag != nil {
		stat.Mode = dirStat.Mode
		return 0
	}

	// is it a newly created file
	for _, tempFile := range tfs.tempFiles {
		if path == tempFile.requestedPath {
			info, err := os.Stat(tempFile.tempPath)
			if err != nil {
				tfs.errChan <- err
				return -fuse.ENOENT
			}
			stat.Mode = fileStat.Mode
			stat.Size = info.Size()
			stat.Mtim = fuse.NewTimespec(info.ModTime())
			return 0
		}
	}

	// is it a file
	id := pathToFileId(path)

	file, err := tfs.t.GetFile(id)
	if err != nil && !errors.Is(err, tagger.ErrFileNotExist) {
		tfs.errChan <- err
		return -fuse.ENOENT
	}
	// last entry is not a known dir or file
	if file == nil {
		return -fuse.ENOENT
	}

	info, err := os.Stat(tfs.t.GetFilepath(file))
	if err != nil {
		tfs.errChan <- err
		return -fuse.ENOENT
	}

	stat.Mode = fileStat.Mode
	stat.Size = info.Size()
	stat.Mtim = fuse.NewTimespec(info.ModTime())

	return 0
}

// Truncate changes the size of a file.
func (tfs *TaggerFS) Truncate(path string, size int64, fh uint64) int {
	if fh == nullFH {
		id := pathToFileId(path)

		file, err := tfs.t.GetFile(id)
		if err != nil {
			if errors.Is(err, tagger.ErrFileNotExist) {
				return -fuse.ENOSYS
			} else {
				tfs.errChan <- err
				return -fuse.EIO
			}
		}

		err = syscall.Truncate(tfs.t.GetFilepath(file), size)
		if err != nil {
			tfs.errChan <- err
			return -fuse.EIO
		}
	} else {
		err := syscall.Ftruncate(int(fh), size)
		if err != nil {
			tfs.errChan <- err
			return -fuse.EIO
		}
	}

	return 0
}

// Read reads data from a file.
func (tfs *TaggerFS) Read(path string, buff []byte, offset int64, fh uint64) (numBytes int) {
	numBytes, err := syscall.Pread(int(fh), buff, offset)
	if err != nil {
		tfs.errChan <- err
		return -fuse.EIO
	}

	return numBytes
}

// Write writes data to a file.
func (tfs *TaggerFS) Write(path string, buff []byte, offset int64, fh uint64) (numBytes int) {
	numBytes, err := syscall.Pwrite(int(fh), buff, offset)
	if err != nil {
		tfs.errChan <- err
		return -fuse.EIO
	}

	return numBytes
}

// Release closes an open file.
func (tfs *TaggerFS) Release(path string, fh uint64) (errc int) {
	err := syscall.Close(int(fh))
	if err != nil {
		tfs.errChan <- err
		return -fuse.EIO
	}

	// if it was a file created through the FUSE layer, import it and remove the temp file
	for i, tempFile := range tfs.tempFiles {
		if fh == tempFile.fh {
			tfs.tempFiles = slices.Delete(tfs.tempFiles, i, i+1)
			file, err := tfs.t.ImportFile(tempFile.tempPath)
			if err != nil {
				tfs.errChan <- err
				return -fuse.EIO
			}
			err = os.Remove(tempFile.tempPath)
			if err != nil {
				tfs.errChan <- err
				return -fuse.EIO
			}

			// tag the new file based on it's path
			// only the most-specific tags are included, i.e. those with no children in the path
			// for example, the path "a/b/c/+/x/y/" will add the tags c and y
			var selectedTags = make(tagger.TagMap)
			var parentTagIds = make(map[int64]bool)
			for _, dir := range pathToDirs(tempFile.requestedPath) {
				tag, err := tfs.t.GetTagByName(dir)
				if errors.Is(err, tagger.ErrTagNotExist) {
					continue
				}
				if err != nil {
					tfs.errChan <- err
					return -fuse.EIO
				}
				selectedTags[tag.Id] = *tag
				for _, id := range tag.Parents {
					parentTagIds[id] = true
				}
			}

			parentTagIdsList := make([]int64, 0, len(parentTagIds))
			for k := range parentTagIds {
				parentTagIdsList = append(parentTagIdsList, k)
			}

			for id, tag := range selectedTags {
				if !slices.Contains(parentTagIdsList, id) {
					err = tfs.t.TagFile(file, &tag)
					if err != nil {
						tfs.errChan <- err
						return -fuse.EIO
					}

				}
			}

			return 0
		}
	}

	return 0
}

// Unlink removes a file.
func (tfs *TaggerFS) Unlink(path string) (errc int) {
	return -fuse.ENOSYS
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
				tfs.errChan <- err
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
		tfs.errChan <- err
		return -fuse.EIO
	}

	var selectedTagIds []int64
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
			tfs.errChan <- err
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
				tfs.errChan <- err
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
				tfs.errChan <- err
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
