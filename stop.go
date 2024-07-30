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

func stopContainer(containerId string) error {
	// 1. 根据容器Id查询容器信息
	containerInfo, err := getInfoByContainerId(containerId)
	if err != nil {
		logrus.Errorf("Get container %s info error %v", containerId, err)
		return err
	}
	pidInt, err := strconv.Atoi(containerInfo.Pid)
	if err != nil {
		logrus.Error("Conver pid form string to int error ", err)
		return err
	}
	// 2. 发送SINGTERM信号
	err = syscall.Kill(pidInt, syscall.SIGTERM)
	if err != nil {
		logrus.Errorf("Stop container %s error %v", containerId, err)
	}
	// 3. 修改容器信息，设置容器状态为stop, 清空pid
	containerInfo.Status = container.STOP
	containerInfo.Pid = ""
	newContent, err := json.Marshal(&containerInfo)
	if err != nil {
		logrus.Errorf("Json marshal %s error %v", containerId, err)
		return err
	}
	// 4.重新写回存储容器信息的文件
	dirPath := fmt.Sprintf(container.InfoLocFormat, containerId)
	configFilePath := path.Join(dirPath, container.ConfigName)
	err = os.WriteFile(configFilePath, newContent, constant.Perm0622)
	if err != nil {
		logrus.Errorf("Write file %s error:%v", configFilePath, err)
	}
	return nil
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

func removeContainer(containerId string, force bool) {
	// 查询对应容器信息
	containerInfo, err := getInfoByContainerId(containerId)
	if err != nil {
		logrus.Errorf("Get container %s info error %v", containerId, err)
		return
	}
	logrus.Info(containerInfo.Status)
	switch containerInfo.Status {
	case container.STOP: // STOP状态直接删除
		// 先删除目录
		err = container.DeleteContainerInfo(containerId)
		if err != nil {
			logrus.Errorf("Remove container [%s]'s config failed, detail: %v", containerId, err)
			return
		}
		// 删除工作文件夹
		container.DeleteWorkSpace(containerId, containerInfo.Volume)
	case container.RUNNING: // 如果状态为运行中，判断是否强制删除，如果强制删除则先暂停再删除
		if !force {
			logrus.Errorf(`Couldn't remove running container [%s], Stop the container before attempting removal or force remove`, containerId)
			return
		}
		logrus.Infof("force delete running container [%s]", containerId)
		err = stopContainer(containerId)
		if err != nil {
			return
		}
		removeContainer(containerId, force)
	default:
		logrus.Errorf("Couldn't remove container,invalid status %s", containerInfo.Status)
		return
	}
}
