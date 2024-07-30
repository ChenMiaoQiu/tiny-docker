package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"syscall"

	"github.com/ChenMiaoQiu/tiny-docker/constant"
	"github.com/ChenMiaoQiu/tiny-docker/container"
	"github.com/sirupsen/logrus"
)

func stopContainer(containerId string) {
	// 1. 根据容器Id查询容器信息
	containerInfo, err := getInfoByContainerId(containerId)
	if err != nil {
		logrus.Errorf("Get container %s info error %v", containerId, err)
		return
	}
	pidInt, err := strconv.Atoi(containerInfo.Pid)
	if err != nil {
		logrus.Error("Conver pid form string to int error ", err)
		return
	}
	// 2. 发送SINGTERM信号
	err = syscall.Kill(pidInt, syscall.SIGTERM)
	if err != nil {
		logrus.Errorf("Stop container %s error %v", containerId, err)
		return
	}
	// 3. 修改容器信息，设置容器状态为stop, 清空pid
	containerInfo.Status = container.STOP
	containerInfo.Pid = ""
	newContent, err := json.Marshal(&containerInfo)
	if err != nil {
		logrus.Errorf("Json marshal %s error %v", containerId, err)
		return
	}
	// 4.重新写回存储容器信息的文件
	dirPath := fmt.Sprintf(container.InfoLocFormat, containerId)
	configFilePath := path.Join(dirPath, container.ConfigName)
	err = os.WriteFile(configFilePath, newContent, constant.Perm0622)
	if err != nil {
		logrus.Errorf("Write file %s error:%v", configFilePath, err)
	}
}

// getInfoByContainerId 获取容器运行PID
func getInfoByContainerId(containerId string) (container.Info, error) {
	dirPath := fmt.Sprintf(container.InfoLocFormat, containerId)
	configFilePath := path.Join(dirPath, container.ConfigName)
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return container.Info{}, err
	}
	var containerInfo container.Info
	err = json.Unmarshal(content, &containerInfo)
	if err != nil {
		return container.Info{}, err
	}
	return containerInfo, nil
}
