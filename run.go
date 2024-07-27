package main

import (
	"os"
	"strings"

	"github.com/ChenMiaoQiu/tiny-docker/cgroups"
	"github.com/ChenMiaoQiu/tiny-docker/cgroups/subsystem"
	"github.com/ChenMiaoQiu/tiny-docker/container"
	log "github.com/sirupsen/logrus"
)

func Run(tty bool, cmdArr []string, resourcesConfig *subsystem.ResourceConfig) {
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Errorf("New parent process error")
		return
	}
	err := parent.Start()
	if err != nil {
		log.Error(err)
	}

	cgroupManager := cgroups.NewCgroupManager("tiny-docker")
	// 进程结束时自动删除对应cgroup资源限制
	_ = cgroupManager.Set(resourcesConfig)
	_ = cgroupManager.Apply(parent.Process.Pid, resourcesConfig)

	// 创建完子进程后发送参数
	sendInitCommand(cmdArr, writePipe)
	_ = parent.Wait()
	cgroupManager.Destroy()
	os.Exit(-1)
}

func sendInitCommand(comArr []string, writePipe *os.File) {
	command := strings.Join(comArr, " ")
	log.Info("command is: ", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
