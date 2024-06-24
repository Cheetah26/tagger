package main

import (
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"

	"github.com/skratchdot/open-golang/open"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// TaggerApp struct
type TaggerApp struct {
	ctx context.Context
	Tagger
	tempDir string
}

// startup is called at application startup
func (a *TaggerApp) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx

	temp, err := os.MkdirTemp("", "tagger-")
	if err != nil {
		panic("Could not create temp directory")
	}
	a.tempDir = temp
}

// domReady is called after front-end resources have been loaded
func (a TaggerApp) domReady(ctx context.Context) {
	// Add your action here
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *TaggerApp) beforeClose(ctx context.Context) (prevent bool) {
	os.RemoveAll(a.tempDir)
	a.db.Close()

	return false
}

// shutdown is called at application termination
func (a *TaggerApp) shutdown(ctx context.Context) {
	// Perform your teardown here
}

func (a *TaggerApp) OpenDBDialog() string {
	var defaultDirectory *string
	homedir, err := os.UserHomeDir()
	if err != nil {
		defaultDirectory = nil
	} else {
		desktop := path.Join(homedir, "Desktop")
		defaultDirectory = &desktop
	}

	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Choose Database",
		DefaultDirectory: *defaultDirectory,
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Database File",
				Pattern:     "*.db",
			},
		},
	})

	if err != nil {
		runtime.LogError(a.ctx, err.Error())
	}

	return path
}

func (a *TaggerApp) ImportFilesDialog() {
	paths, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Open File(s)",
		DefaultDirectory: "",
	})
	if err != nil {
		runtime.LogError(a.ctx, err.Error())
	}

	for _, path := range paths {
		if err := a.ImportFile(path); err != nil {
			runtime.LogError(a.ctx, err.Error())
		}
	}
}

func (a *TaggerApp) OpenFile(file File) error {
	if a.tempDir == "" {
		return errors.New("no temp dir")
	}

	path := filepath.Join(a.tempDir, file.Hash+file.Filetype)

	// check if the file exists before writing it
	if _, err := os.Stat(path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		// file doesn't exist, write it
		runtime.LogDebug(a.ctx, "WRITING FILE")
		if err := os.WriteFile(path, file.Data, os.ModePerm); err != nil {
			return err
		}
	}

	// open the file
	if err := open.Start(path); err != nil {
		return err
	}

	return nil
}
