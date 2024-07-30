package main

import (
	"os/exec"

	"github.com/ChenMiaoQiu/tiny-docker/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func commitContainer(containerId string, imageName string) error {
	mntPath := utils.GetMerged(containerId)
	imageTar := utils.GetImage(imageName)
	exist, err := utils.PathExists(imageTar)
	if err != nil {
		return errors.WithMessagef(err, "check is image [%s/%s] exist failed", imageName, imageTar)
	}

	if exist {
		return errors.New("Image Already Exist")
	}
	logrus.Infof("commitContainer imageTar:%s", imageTar)
	_, err = exec.Command("tar", "-czf", imageTar, "-C", mntPath, ".").CombinedOutput()
	if err != nil {
		logrus.Errorf("tar folder %s error %v", mntPath, err)
	}
	return err
}
