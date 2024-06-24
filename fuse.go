package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/winfsp/cgofuse/fuse"
)

type TaggerFS struct {
	t *Tagger
	fuse.FileSystemBase
}

func (t *TaggerFS) Open(path string, flags int) (errc int, fh uint64) {
	return 0, 0
}

func (tfs *TaggerFS) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	_, fullname := filepath.Split(path)
	name := strings.TrimSuffix(fullname, filepath.Ext(fullname))
	file := tfs.t.GetFile(name)

	// It's a directory
	if file == nil {
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	}

	// It's a file
	// fmt.Println(name)
	stat.Mode = fuse.S_IFREG | 0444
	stat.Size = int64(len(file.Data))
	return 0
}

func (tfs *TaggerFS) Read(path string, buff []byte, offset int64, fh uint64) (n int) {
	_, fullname := filepath.Split(path)
	name := strings.TrimSuffix(fullname, filepath.Ext(fullname))
	file := tfs.t.GetFile(name)
	if file == nil {
		return 0
	}

	end_offset := offset + int64(len(buff))
	if end_offset > int64(len(file.Data)) {
		end_offset = int64(len(file.Data))
	}
	if end_offset < offset {
		return 0
	}

	return copy(buff, file.Data[offset:end_offset])
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
		fmt.Println(path)
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
			name := file.Hash + file.Filetype
			file_stat := &fuse.Stat_t{
				Mode: fuse.S_IFREG | 0444,
				Size: int64(len(file.Data)),
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

func MountFS(path string, tagger *Tagger) {
	tfs := &TaggerFS{
		t: tagger,
	}
	host := fuse.NewFileSystemHost(tfs)
	host.Mount(path, []string{})
}
