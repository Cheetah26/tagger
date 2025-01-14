package app

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/skratchdot/open-golang/open"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/cheetah26/tagger/pkg/tagger"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

type TaggerApp struct {
	ctx context.Context
	tagger.Tagger
}

func Run() {
	// Create an instance of the app structure
	app := &TaggerApp{}

	// Create application with options
	err := wails.Run(&options.App{
		Title:             "tagger",
		Width:             800,
		Height:            600,
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,
		BackgroundColour:  &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		AssetServer: &assetserver.Options{
			Assets:     assets,
			Middleware: app.Middleware,
		},
		Menu:             nil,
		Logger:           nil,
		LogLevel:         logger.DEBUG,
		OnStartup:        app.startup,
		OnDomReady:       app.domReady,
		OnBeforeClose:    app.beforeClose,
		OnShutdown:       app.shutdown,
		WindowStartState: options.Normal,
		Bind: []interface{}{
			app,
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop: true,
		},
		// Windows platform specific options
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			// DisableFramelessWindowDecorations: false,
			WebviewUserDataPath: "",
			ZoomFactor:          1.0,
		},
		// Mac platform specific options
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			About: &mac.AboutInfo{
				Title:   "tagger",
				Message: "",
				Icon:    icon,
			},
		},
	})

	if err != nil {
		log.Fatal(err)
	}
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
	a.Close()

	return false
}

// shutdown is called at application termination
func (a *TaggerApp) shutdown(ctx context.Context) {
	// Perform your teardown here
}

// App middleware responds to HTTP requests for files in the database
// and serves them to the client
func (a *TaggerApp) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Handle requests starting with /file/
			fileIdString, found := strings.CutPrefix(r.URL.Path, "/file/")
			if !found {
				next.ServeHTTP(w, r)
				return
			}

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

// -- Additional GUI-specific functions to be called from the frontend --

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

func (a *TaggerApp) OpenFile(file tagger.File) error {
	path := a.GetFilepath(file)
	fmt.Println(path)
	err := open.Start(path)
	if err != nil {
		return err
	}

	return nil
}
