package main

import (
	"os"

	"github.com/ChenMiaoQiu/tiny-docker/container"
	log "github.com/sirupsen/logrus"
)

func Run(tty bool, cmd string) {
	parent := container.NewParentProcess(tty, cmd)
	err := parent.Start()
	if err != nil {
		log.Error(err)
	}
	_ = parent.Wait()
	os.Exit(-1)
}
