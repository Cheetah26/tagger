package service

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cheetah26/tagger/pkg/tagger"
	svc "github.com/kardianos/service"
)

type TaggerService struct {
	mountpoint string
	tfs        *tagger.TaggerFS
}

func (ts *TaggerService) Start(s svc.Service) error {
	go ts.connectAndMount(dbPath)

	return nil
}

func (ts *TaggerService) connectAndMount(dbpath string) {
	dir, fullname := filepath.Split(dbpath)
	name := strings.TrimSuffix(fullname, filepath.Ext(fullname))
	mountpoint := filepath.Join(dir, name)

	ts.mountpoint = mountpoint

	t := tagger.Tagger{}
	t.Open(dbpath)

	if _, err := os.Stat(mountpoint); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			os.Mkdir(mountpoint, os.FileMode(0777))
		}
		// else {
		// 	log.Println(err)
		// 	log.Fatal(err.Error())
		// }
	}

	ts.tfs = tagger.CreateMount(mountpoint, &t)
}

func (ts *TaggerService) Stop(s svc.Service) error {
	ts.tfs.Unmount()
	os.RemoveAll(ts.mountpoint)
	return nil
}

var dbPath string

func Run() {
	if len(os.Args) < 3 {
		println("Database path is required\nUsage: `tagger --service <path to db>`")
		os.Exit(1)
	}

	dbPath = os.Args[2]

	service_config := &svc.Config{
		Name:        "Tagger",
		DisplayName: "Tagger",
		Description: "Organize files with tags",
	}

	ts := &TaggerService{}

	s, err := svc.New(ts, service_config)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}
