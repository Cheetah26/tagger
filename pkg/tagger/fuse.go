package tagger

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/winfsp/cgofuse/fuse"
)

type TaggerFS struct {
	t *Tagger
	h *fuse.FileSystemHost

	// Inherit default filesystem actions
	fuse.FileSystemBase
}

func CreateMount(path string, tagger *Tagger) *TaggerFS {
	tfs := &TaggerFS{
		t: tagger,
	}

	host := fuse.NewFileSystemHost(tfs)
	go host.Mount(path, []string{})

	tfs.h = host
	return tfs
}

// func (tfs *TaggerFS) Unmount() {
// 	tfs.h.Unmount()
// }

func pathToId(path string) int {
	_, fullname := filepath.Split(path)
	idString := strings.TrimSuffix(fullname, filepath.Ext(fullname))
	id, _ := strconv.Atoi(idString)
	return id
}

func (t *TaggerFS) Open(path string, flags int) (errc int, fh uint64) {
	return 0, 0
}

func (tfs *TaggerFS) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	id := pathToId(path)

	file := tfs.t.GetFile(id)

	// It's a directory
	if file == nil {
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	}

	info, _ := os.Stat(tfs.t.GetFilepath(*file))

	// It's a file
	stat.Mode = fuse.S_IFREG | 0444
	stat.Size = info.Size()
	return 0
}

func (tfs *TaggerFS) Read(path string, buff []byte, offset int64, fh uint64) (n int) {
	id := pathToId(path)

	file := tfs.t.GetFile(id)
	if file == nil {
		return fuse.EADDRNOTAVAIL
	}

	realPath := tfs.t.GetFilepath(*file)
	f, _ := os.Open(realPath)

	n, _ = f.ReadAt(buff, offset)
	return n
}

func (tfs *TaggerFS) Readdir(path string, fill func(name string, stat *fuse.Stat_t, offset int64) bool, offset int64, fh uint64) (errc int) {
	dir_stat := &fuse.Stat_t{Mode: fuse.S_IFDIR | 0555}

	// TODO: future behavior
	// Root level: get all tags
	// Subdirectory: get appropriate files, collect their tags, show folders for their tags minus already in path

	// Current directory
	fill(".", dir_stat, 0)
	fill("..", dir_stat, 0)

	// Get parts of the path
	parts := []string{}
	if strings.TrimSpace(path) != "/" {
		parts = strings.Split(path, "/")[1:]
	}

	// Files
	num_files := 0
	if len(parts) > 0 {
		tags := make([]Tag, len(parts))
		for i, v := range parts {
			tag, err := tfs.t.GetTag(v)
			if err != nil {
				fmt.Println(err)
				continue
			}
			tags[i] = *tag
		}

		files := tfs.t.GetFiles(tags)
		num_files = len(files)
		for _, file := range files {
			name := strconv.Itoa(file.Id) + file.Filetype
			file_stat := &fuse.Stat_t{
				Mode: fuse.S_IFREG | 0444,
				// Size: int64(len(file.Data)), TODO
			}
			fill(name, file_stat, 0)
		}
	}

	// Sub directories
	if num_files != 0 || len(parts) == 0 {
		alreadyInPath := func(tag string) bool {
			for _, part := range parts {
				if tag == part {
					return true
				}
			}
			return false
		}
		for _, tag := range tfs.t.GetAllTags() {
			if !alreadyInPath(tag.Name) {
				fill(tag.Name, dir_stat, 0)
			}
		}
	}

	return 0
}
