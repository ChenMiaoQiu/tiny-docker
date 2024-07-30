package main

import (
	"os"
	"strings"

	"github.com/ChenMiaoQiu/tiny-docker/cgroups"
	"github.com/ChenMiaoQiu/tiny-docker/cgroups/subsystem"
	"github.com/ChenMiaoQiu/tiny-docker/container"
	"github.com/sirupsen/logrus"
)

func Run(tty bool, cmdArr []string, resourcesConfig *subsystem.ResourceConfig, volume string, containerName string, imageName string) {
	containerId := container.GenerateContainerID()

	parent, writePipe := container.NewParentProcess(tty, volume, containerId, imageName)
	if parent == nil {
		logrus.Error("New parent process error")
		return
	}
	err := parent.Start()
	if err != nil {
		logrus.Error(err)
	}

	// 记录容器信息
	err = container.RecordContainerInfo(parent.Process.Pid, cmdArr, containerName, containerId, volume)
	if err != nil {
		logrus.Error("Record container info error ", err)
		return
	}

	cgroupManager := cgroups.NewCgroupManager("tiny-docker")
	// 配置cgroup资源限制
	_ = cgroupManager.Set(resourcesConfig)
	_ = cgroupManager.Apply(parent.Process.Pid, resourcesConfig)
	// 进程结束时自动删除对应cgroup资源限制
	defer cgroupManager.Destroy()

	// 创建完子进程后发送参数
	sendInitCommand(cmdArr, writePipe)
	// 如果是tty，那么父进程等待，就是前台运行，否则就是跳过，实现后台运行
	if tty {
		_ = parent.Wait()
		// 解绑并删除overlayFS 使用的upper work mount 文件夹
		container.DeleteWorkSpace(containerId, volume)
		container.DeleteContainerInfo(containerId)
	}
}

func sendInitCommand(comArr []string, writePipe *os.File) {
	command := strings.Join(comArr, " ")
	logrus.Info("command is: ", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
