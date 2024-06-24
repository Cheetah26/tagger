package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/skratchdot/open-golang/open"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// TaggerApp struct
type TaggerApp struct {
	ctx context.Context
	Tagger
}

// startup is called at application startup
func (a *TaggerApp) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx
}

// domReady is called after front-end resources have been loaded
func (a TaggerApp) domReady(ctx context.Context) {
	// Add your action here
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *TaggerApp) beforeClose(ctx context.Context) (prevent bool) {
	a.db.Close()

	return false
}

// shutdown is called at application termination
func (a *TaggerApp) shutdown(ctx context.Context) {
	// Perform your teardown here
}

func (a *TaggerApp) OpenDBDialog() string {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose Database",
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
	path := filepath.Join(a.dir, strconv.Itoa(file.Id)+file.Filetype)
	err := open.Start(path)
	if err != nil {
		return err
	}

	return nil
}

// App middleware responds to HTTP requests for files in the database
// and serves them to the client
func (a *TaggerApp) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Handle requests starting with /file/
			match, err := regexp.Match("/file/.+", []byte(r.URL.Path))
			if err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if !match {
				next.ServeHTTP(w, r)
				return
			}

			fileIdString := strings.TrimPrefix(r.URL.Path, "/file/")
			fileId, err := strconv.Atoi(fileIdString)
			if err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusNotFound)
				return
			}

			file := a.GetFile(fileId)
			if file == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			filePath := a.GetFilepath(*file)

			data, err := os.ReadFile(filePath)
			if err != nil {
				w.Write([]byte(err.Error()))
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.Write(data)
		})
}
