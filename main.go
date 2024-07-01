package main

import (
	"os"

	"github.com/cheetah26/tagger/internal/app"
	"github.com/cheetah26/tagger/internal/service"
)

func main() {
	// start as a service
	if len(os.Args) > 1 && os.Args[1] == "--service" {
		service.Run()
		return
	}

	// otherwise, start the gui
	app.Run()
}
