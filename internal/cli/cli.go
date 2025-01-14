package cli

import (
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/cheetah26/tagger/pkg/tagger"
)

type TaggerCli struct {
	tfs *tagger.TaggerFS
}

func (tc *TaggerCli) mount() {
	if len(os.Args) < 4 {
		println("Database path is required\nUsage: `tagger --mount <path to db> <mountpoint>`")
		os.Exit(1)
	}

	dbPath := os.Args[2]
	mountpoint := os.Args[3]

	// dir, fullname := filepath.Split(dbPath)
	// name := strings.TrimSuffix(fullname, filepath.Ext(fullname))
	// mountpoint := filepath.Join(dir, name)

	t := tagger.Tagger{}
	t.Open(dbPath)

	if _, err := os.Stat(mountpoint); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			os.Mkdir(mountpoint, os.FileMode(0777))
		}
		// else {
		// 	log.Println(err)
		// 	log.Fatal(err.Error())
		// }
	}

	tc.tfs = tagger.CreateMount(mountpoint, &t)

	handleInterrupt(func() {
		os.RemoveAll(mountpoint)
	})
}

func handleInterrupt(handler func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	func() {
		for sig := range c {
			println(fmt.Sprintf("\nReceived %s, cleaning up...", sig))
			handler()
			os.Exit(0)
		}
	}()
}

func Cli() {
	tc := &TaggerCli{}

	switch os.Args[1] {
	case "--mount":
		tc.mount()
	default:
		println("Unknown option, please use one of:\n\t--mount\n")
		os.Exit(1)
	}
}
