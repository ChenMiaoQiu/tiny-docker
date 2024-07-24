package main

import (
	"os"
	"strings"

	"github.com/ChenMiaoQiu/tiny-docker/container"
	log "github.com/sirupsen/logrus"
)

func Run(tty bool, cmdArr []string) {
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	err := parent.Start()
	if err != nil {
		log.Error(err)
	}

	// 创建完子进程后发送参数
	sendInitCommand(cmdArr, writePipe)
	_ = parent.Wait()
	os.Exit(-1)
}

func sendInitCommand(comArr []string, writePipe *os.File) {
	command := strings.Join(comArr, " ")
	log.Info("command is: ", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
