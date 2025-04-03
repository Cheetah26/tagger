package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cheetah26/tagger/pkg/fuse"
	"github.com/cheetah26/tagger/pkg/tagger"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func usage() {
	fmt.Println(` Usage:
	- open the gui	"tagger"
	- mount a db	"tagger <path-to-database> <path-to-mountpoint>"`)
}

func main() {
	switch len(os.Args) {
	case 1:
		gui()
	case 3:
		cli()
	default:
		usage()
		os.Exit(1)
	}
}

func gui() {
	app := application.New(application.Options{
		Name:        "Tagger",
		Description: "Tag your files",
		Services: []application.Service{
			application.NewService(&TaggerService{}, application.ServiceOptions{
				Route: "/file/",
			}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.NewWebviewWindowWithOptions(application.WebviewWindowOptions{
		Title:            "Tagger",
		BackgroundColour: application.NewRGB(255, 255, 255),
		URL:              "/",
	})

	err := app.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func cli() {
	// otherwise try to mount
	databaseFile := os.Args[1]
	mountpoint := os.Args[2]
	if databaseFile == "" || mountpoint == "" {
		usage()
		os.Exit(1)
	}

	fmt.Printf("Mounting tags from %s at %s\n", databaseFile, mountpoint)

	tr, err := tagger.Open(databaseFile)
	if err != nil {
		log.Fatal(err)
	}

	unmount, errChan, err := fuse.Mount(mountpoint, tr)
	if err != nil {
		log.Fatal(err.Error())
	}
	go func() {
		if err = <-errChan; err != nil {
			log.Println(err.Error())
		}
	}()

	defer unmount()

	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	<-interrupt
}
