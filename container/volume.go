package container

import (
	"os"
	"os/exec"
	"path"

	"github.com/ChenMiaoQiu/tiny-docker/constant"
	"github.com/sirupsen/logrus"
)

// mountVolume 挂载数据卷
func mountVolume(mntPath string, hostPath string, containerPath string) {
	err := os.Mkdir(hostPath, constant.Perm0777)
	if err != nil {
		logrus.Infof("mkdir host volume dir %s error: %v", hostPath, err)
	}

	containerPathInHost := path.Join(mntPath, containerPath)
	err = os.Mkdir(containerPathInHost, constant.Perm0777)
	if err != nil {
		logrus.Infof("mkdir container volume dir %s error: %v", hostPath, err)
	}

	// 使用bind mount 挂载目录
	cmd := exec.Command("mount", "-o", "bind", hostPath, containerPathInHost)
	logrus.Infof("bind mount source volume path: %s, dest path: %s", hostPath, containerPathInHost)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		logrus.Error("mount volume failed, error: ", err)
	}
}

func umountVolume(mntPath, containerPath string) {
	containerPathInHost := path.Join(mntPath, containerPath)
	cmd := exec.Command("umount", containerPathInHost)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Error("umount volume failed, error: ", err)
	}
}
