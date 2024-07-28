package main

import (
	"fmt"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func commitContainer(imageName string) error {
	mntPath := "/root/merged"
	imageTar := "/root/" + imageName + ".tar"
	fmt.Println("commitContainer imageTar:", imageTar)
	_, err := exec.Command("tar", "-czf", imageTar, "-C", mntPath, ".").CombinedOutput()
	if err != nil {
		logrus.Errorf("tar folder %s error %v", mntPath, err)
	}
	return err
}
