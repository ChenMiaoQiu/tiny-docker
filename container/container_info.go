package container

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ChenMiaoQiu/tiny-docker/constant"
	"github.com/ChenMiaoQiu/tiny-docker/utils"
	"github.com/pkg/errors"
)

const (
	RUNNING       = "running"
	STOP          = "stopped"
	Exit          = "exited"
	InfoLoc       = "/var/lib/tiny-docker/containers/"
	InfoLocFormat = InfoLoc + "%s/"
	ConfigName    = "config.json"
	IDLength      = 10
	LogFile       = "%s-json.log"
)

type Info struct {
	Pid         string   `json:"pid"`         // 容器的init进程在宿主机上的 PID
	Id          string   `json:"id"`          // 容器Id
	Name        string   `json:"name"`        // 容器名
	Command     string   `json:"command"`     // 容器内init运行命令
	CreatedTime string   `json:"createTime"`  // 创建时间
	Status      string   `json:"status"`      // 容器的状态
	Volume      string   `json:"volume"`      // 容器数据卷
	NetworkName string   `json:"networkName"` // 容器所在的网络
	PortMapping []string `json:"portMapping"` // 端口映射
	IP          string   `json:"ip"`          // 容器IP
}

// RecordContainerInfo 记录容器信息
func RecordContainerInfo(containerPID int, commandArray []string, containerName string, containerId string, volume string, network string, portMapping []string, ip string) (*Info, error) {
	// 如果未指定容器名，则使用随机生成的containerID
	if containerName == "" {
		containerName = containerId
	}
	command := strings.Join(commandArray, "")
	containerInfo := &Info{
		Pid:         strconv.Itoa(containerPID),
		Id:          containerId,
		Name:        containerName,
		Command:     command,
		CreatedTime: time.Now().Format("2006-01-02 15:04:05"),
		Status:      RUNNING,
		Volume:      volume,
		NetworkName: network,
		PortMapping: portMapping,
		IP:          ip,
	}

	jsonByte, err := json.Marshal(containerInfo)
	if err != nil {
		return nil, errors.WithMessage(err, "container info marshal failed")
	}
	jsonStr := string(jsonByte)
	// 拼接出存储容器信息文件的路径，如果目录不存在则级联创建
	dirPath := fmt.Sprintf(InfoLocFormat, containerId)
	if err := os.MkdirAll(dirPath, constant.Perm0622); err != nil {
		return nil, errors.WithMessagef(err, "mkdir %s failed", dirPath)
	}

	// 写入文件信息
	fileName := path.Join(dirPath, ConfigName)
	file, err := os.Create(fileName)
	defer func() {
		_ = file.Close()
	}()
	if err != nil {
		return nil, errors.WithMessagef(err, "create file %s failed", fileName)
	}
	_, err = file.WriteString(jsonStr)
	if err != nil {
		return nil, errors.WithMessagef(err, "write container info to config file %s failed", fileName)
	}

	return containerInfo, nil
}

// DeleteContainerInfo 删除容器日志
func DeleteContainerInfo(containerID string) error {
	dirPath := fmt.Sprintf(InfoLocFormat, containerID)
	if err := os.RemoveAll(dirPath); err != nil {
		return errors.WithMessagef(err, "remove dir %s failed", dirPath)
	}
	return nil
}

// GenerateContainerID 生成容器id
func GenerateContainerID() string {
	return utils.RandStringBytes(IDLength)
}

// GetLogFile 获取日志文件名称
func GetLogFile(containerId string) string {
	return fmt.Sprintf(LogFile, containerId)
}
