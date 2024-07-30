package utils

import "fmt"

// 容器相关目录
const (
	ImagePath       = "/var/lib/tiny-docker/image/"
	RootPath        = "/var/lib/tiny-docker/overlay2/"
	lowerDirFormat  = RootPath + "%s/lower"
	upperDirFormat  = RootPath + "%s/upper"
	workDirFormat   = RootPath + "%s/work"
	mergedDirFormat = RootPath + "%s/merged"
	overlayFSFormat = "lowerdir=%s,upperdir=%s,workdir=%s"
)

// 获取容器文件夹
func GetRoot(containerID string) string { return RootPath + containerID }

// 获取容器镜像存储位置
func GetImage(imageName string) string { return fmt.Sprintf("%s%s.tar", ImagePath, imageName) }

// 获取lower文件夹位置
func GetLower(containerID string) string {
	return fmt.Sprintf(lowerDirFormat, containerID)
}

// 获取upper文件夹位置
func GetUpper(containerID string) string {
	return fmt.Sprintf(upperDirFormat, containerID)
}

// 获取work文件夹位置
func GetWorker(containerID string) string {
	return fmt.Sprintf(workDirFormat, containerID)
}

// 获取merged文件夹位置
func GetMerged(containerID string) string { return fmt.Sprintf(mergedDirFormat, containerID) }

// 获取OverlayFS参数
func GetOverlayFSDirs(lower, upper, worker string) string {
	return fmt.Sprintf(overlayFSFormat, lower, upper, worker)
}
