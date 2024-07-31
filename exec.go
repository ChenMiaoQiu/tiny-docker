package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/ChenMiaoQiu/tiny-docker/container"
	_ "github.com/ChenMiaoQiu/tiny-docker/setns"
	"github.com/sirupsen/logrus"
)

// nsenter里的C代码里已经出现tiny_docker_pid和tiny_docker_cmd这两个Key,主要是为了控制是否执行C代码里面的setns.
const (
	EnvExecPid = "tiny_docker_pid"
	EnvExecCmd = "tiny_docker_cmd"
)

func ExecContainer(containerId string, comArray []string) {
	pid, err := getPidByContainerId(containerId)
	if err != nil {
		logrus.Errorf("Exec container getContainerPidByName %s error %v", containerId, err)
		return
	}

	if pid == "" {
		logrus.Error("Not found container pid ", err)
		return
	}

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 传递命令
	cmdStr := strings.Join(comArray, " ")
	logrus.Infof("container pid: %s command: %s", pid, cmdStr)
	_ = os.Setenv(EnvExecPid, pid)
	_ = os.Setenv(EnvExecCmd, cmdStr)
	containerEnv := getEnvsByPid(pid)
	cmd.Env = append(os.Environ(), containerEnv...)

	if err = cmd.Run(); err != nil {
		logrus.Errorf("Exec container %s error %v", containerId, err)
	}
}

// getPidByContainerId 获取容器在运行中的pid
func getPidByContainerId(containerId string) (string, error) {
	// 拼接出记录容器信息的文件路径
	dirPath := fmt.Sprintf(container.InfoLocFormat, containerId)
	configFilePath := path.Join(dirPath, container.ConfigName)
	// 解析内容获取pid
	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var containerInfo container.Info
	err = json.Unmarshal(content, &containerInfo)
	if err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}

// getEnvsByPid 获取指定pid进程的环境变量
func getEnvsByPid(pid string) []string {
	// 环境变量存放路径
	path := fmt.Sprintf("/proc/%s/environ", pid)
	content, err := os.ReadFile(path)
	if err != nil {
		logrus.Errorf("Read file %s error %v", path, err)
		return nil
	}
	// env split by \u0000
	envs := strings.Split(string(content), "\u0000")
	return envs
}
