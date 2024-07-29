package main

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/ChenMiaoQiu/tiny-docker/container"
	"github.com/sirupsen/logrus"
)

func logContainer(containerId string) {
	logFileLocation := path.Join(fmt.Sprintf(container.InfoLocFormat, containerId), container.GetLogFile(containerId))
	file, err := os.Open(logFileLocation)
	defer func() {
		file.Close()
	}()
	if err != nil {
		logrus.Errorf("Log container open file %s error %v", logFileLocation, err)
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		logrus.Errorf("Log container read file %s error %v", logFileLocation, err)
		return
	}
	_, err = fmt.Fprint(os.Stdout, string(content))
	if err != nil {
		logrus.Errorf("Log container Fprint  error %v", err)
		return
	}
}
