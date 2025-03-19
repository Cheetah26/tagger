package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"os"

	"github.com/cheetah26/tagger/pkg/tagger"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

type TaggerService struct {
	tagger.Tagger
}

func (s *TaggerService) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	application.Get().RegisterApplicationEventHook(
		events.ApplicationEventType(events.Common.WindowFilesDropped),
		func(event *application.ApplicationEvent) {
			application.Get().Logger.Debug(event.Context().Filename())
		})

	return nil
}

// App middleware responds to HTTP requests for files in the database
// and serves them to the client
func (a *TaggerService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fileId, err := strconv.ParseInt(r.URL.Path, 10, 64)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	file, err := a.GetFile(fileId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}

	filePath := a.GetFilepath(file)

	data, err := os.ReadFile(filePath)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "")
	w.Write(data)
}

func (a *TaggerService) Open() error {
	dialog := application.OpenFileDialog()
	dialog.SetOptions(&application.OpenFileDialogOptions{
		Title: "Choose Database",
		Filters: []application.FileFilter{
			{
				DisplayName: "Database File",
				Pattern:     "*.db",
			},
		},
	})

	path, err := dialog.PromptForSingleSelection()
	if err != nil {
		return err
	}

	t, err := tagger.Open(path)
	if err != nil {
		return err
	}

	a.Tagger = *t

	return nil
}

func (a *TaggerService) ImportFilesDialog() {
	dialog := application.OpenFileDialog()
	dialog.SetOptions(&application.OpenFileDialogOptions{
		Title:     "Open File(s)",
		Directory: "",
	})

	paths, err := dialog.PromptForMultipleSelection()

	if err != nil {
		application.Get().Logger.Error(err.Error())
	}

	for _, path := range paths {
		if _, err := a.ImportFile(path); err != nil {
			application.Get().Logger.Error(err.Error())
		}
	}
}

func (a *TaggerService) OpenFile(file tagger.File) error {
	path := a.GetFilepath(&file)
	fmt.Println(path)
	err := application.Get().BrowserOpenFile(path)
	if err != nil {
		return err
	}

	return nil
}

func (a *TaggerService) Reveal(file tagger.File) error {
	path := a.GetFilepath(&file)
	err := application.Get().OpenFileManager(path, false)
	if err != nil {
		return err
	}

	return nil
}
