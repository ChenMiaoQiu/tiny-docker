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
	files, err := os.ReadDir(container.InfoLoc)
	if err != nil {
		logrus.Errorf("read dir %s error %v", container.InfoLoc, err)
		return
	}
	containers := make([]*container.Info, 0, len(files))
	for _, file := range files {
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			continue
		}
		containers = append(containers, tmpContainer)
	}

	// 打印
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, err = fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	if err != nil {
		logrus.Error("Fprint error ", err)
	}
	for _, item := range containers {
		_, err = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime)
		if err != nil {
			logrus.Error("Fprint info error ", err)
		}
	}

	if err = w.Flush(); err != nil {
		logrus.Error("Flush error ", err)
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
