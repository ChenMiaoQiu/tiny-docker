package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"text/tabwriter"

	"github.com/ChenMiaoQiu/tiny-docker/container"
	"github.com/sirupsen/logrus"
)

// ListContainerInfos 打印容器日志信息
func ListContainerInfos() {
	// 读取存放容器信息目录下的所有文件
	files, err := os.ReadDir(container.InfoLoc)
	if err != nil {
		logrus.Errorf("read dir %s error %v", container.InfoLoc, err)
		return
	}
	containers := make([]*container.Info, 0, len(files))
	for _, file := range files {
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			logrus.Errorf("get container info error %v", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}
	// 使用tabwriter.NewWriter在控制台打印出容器信息
	// tabwriter 是引用的text/tabwriter类库，用于在控制台打印对齐的表格
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, err = fmt.Fprint(w, "ID\tNAME\tPID\tIP\tSTATUS\tCOMMAND\tCREATED\n")
	if err != nil {
		logrus.Errorf("Fprint error %v", err)
	}
	for _, item := range containers {
		_, err = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.IP,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime)
		if err != nil {
			logrus.Errorf("Fprint error %v", err)
		}
	}
	if err = w.Flush(); err != nil {
		logrus.Errorf("Flush error %v", err)
	}
}

func getContainerInfo(file os.DirEntry) (*container.Info, error) {
	// 根据文件名拼接出完整路径
	configFileDir := fmt.Sprintf(container.InfoLocFormat, file.Name())
	configFileDir = path.Join(configFileDir, container.ConfigName)
	// 读取文件
	content, err := os.ReadFile(configFileDir)
	if err != nil {
		logrus.Errorf("read file %s error %v", configFileDir, err)
		return nil, err
	}
	info := new(container.Info)
	err = json.Unmarshal(content, info)
	if err != nil {
		logrus.Error("info json unmarshal error ", err)
		return nil, err
	}
	return info, nil
}
