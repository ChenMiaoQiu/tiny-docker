package container

import (
	"os"
	"os/exec"

	"github.com/ChenMiaoQiu/tiny-docker/constant"
	"github.com/ChenMiaoQiu/tiny-docker/utils"
	"github.com/sirupsen/logrus"
)

func NewWorkSpace(containerID string, imageName string, volume string) {
	createLower(containerID, imageName)
	createDirs(containerID)
	mountOverlayFS(containerID)

	// 判断是否指定的数据卷，如果指定了数据卷则挂载数据卷
	if volume != "" {
		mntPath := utils.GetMerged(containerID)
		hostPath, containerPath, err := utils.VolumeExtract(volume)
		if err != nil {
			logrus.Error("extract volume failed,maybe volume parameter input is not correct, detail: ", err)
			return
		}
		mountVolume(mntPath, hostPath, containerPath)
	}
}

// createLower 将busybox作为overlayfs的lower层
func createLower(containerId string, imageName string) {
	// 获取lower目录位置
	lowerPath := utils.GetLower(containerId)
	// 获取对应镜像位置
	imagePath := utils.GetImage(imageName)
	// 检查是否存在lower目录
	exist, err := utils.PathExists(lowerPath)
	if err != nil {
		logrus.Infof("Fail to check lowerPath url %v exists: %v", lowerPath, err)
	}

	// 如果不存在则创建目录并解压镜像到对应lower文件夹
	if !exist {
		err = os.MkdirAll(lowerPath, constant.Perm0777)
		if err != nil {
			logrus.Errorf("Mkdir lower dir %s error: %v", lowerPath, err)
		}
		_, err := exec.Command("tar", "-xvf", imagePath, "-C", lowerPath).CombinedOutput()
		if err != nil {
			logrus.Errorf("Untar dir %s error: %v", lowerPath, err)
		}
	}
}

// createDirs 创建overlayfs需要的的upper、worker目录
func createDirs(containerId string) {
	dirs := []string{
		utils.GetUpper(containerId),
		utils.GetWorker(containerId),
		utils.GetMerged(containerId),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, constant.Perm0777); err != nil {
			logrus.Errorf("mkdir dir %s error. %v", dir, err)
		}
	}
}

// mountOverlayFS 挂载overlayfs
func mountOverlayFS(containerId string) {
	// 拼接参数
	// e.g. lowerdir=/root/{containerID}/lower,upperdir=/root/{containerID}/upper,workdir=/root/{containerID}/work
	lowerDir := utils.GetLower(containerId)
	upperDir := utils.GetUpper(containerId)
	workDir := utils.GetWorker(containerId)
	mergedDir := utils.GetMerged(containerId)
	dirs := utils.GetOverlayFSDirs(lowerDir, upperDir, workDir)

	// 完整命令：mount -t overlay overlay -o lowerdir=/root/{containerID}/lower,upperdir=/root/{containerID}/upper,workdir=/root/{containerID}/work /root/{containerID}/merged
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mergedDir)
	logrus.Infof("mount overlayfs: [%s]", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("%v", err)
	}
}

// DeleteWorkSpace Delete the AUFS filesystem while container exit
func DeleteWorkSpace(containerId string, volume string) {
	// 判断是否存在数据卷，存在则取消挂载
	if volume != "" {
		mntPath := utils.GetMerged(containerId)
		_, containerPath, err := utils.VolumeExtract(volume)
		if err != nil {
			logrus.Errorf("extract volume failed, maybe volume parameter input is not correct, detail:%v", err)
			return
		}
		umountVolume(mntPath, containerPath)
	}

	// 如果先删除再取消挂载，数据卷无法保存数据
	umountOverlayFS(containerId)
	deleteDirs(containerId)
}

func umountOverlayFS(containerId string) {
	mntPath := utils.GetMerged(containerId)
	cmd := exec.Command("umount", mntPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	logrus.Info("umountOverlayFS, cmd:", cmd.String())
	if err := cmd.Run(); err != nil {
		logrus.Error(err)
	}
}

func deleteDirs(containerId string) {
	dirs := []string{
		utils.GetMerged(containerId),
		utils.GetLower(containerId),
		utils.GetUpper(containerId),
		utils.GetWorker(containerId),
		utils.GetRoot(containerId),
	}

	for _, dir := range dirs {
		err := os.RemoveAll(dir)
		if err != nil {
			logrus.Errorf("Remove dir %s error: %v", dir, err)
		}
	}
}
