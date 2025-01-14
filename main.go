package main

import (
	"os"

	"github.com/cheetah26/tagger/internal/app"
	"github.com/cheetah26/tagger/internal/cli"
)

func main() {
	// start as a service
	if len(os.Args) > 1 {
		cli.Cli()
		return
	}

	// otherwise, start the gui
	app.Run()
}
