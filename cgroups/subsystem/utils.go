package subsystem

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/ChenMiaoQiu/tiny-docker/constant"
	log "github.com/sirupsen/logrus"
)

const mountPointIndex = 4

// getCgroupPath 找到cgroup在文件系统中的绝对路径
func getCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := findCgroupMountpoint(subsystem)
	absPath := path.Join(cgroupRoot, cgroupPath)
	if !autoCreate {
		return absPath, nil
	}
	// 判断是否存在文件
	_, err := os.Stat(absPath)
	// 如果不存在，则创建
	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(absPath, constant.Perm0755)
		return absPath, err
	}

	if err == nil {
		err = fmt.Errorf("create cgroup")
	}
	return absPath, err
}

func findCgroupMountpoint(subsystem string) string {
	// 通过/proc/self/mountinfo 查看挂载信息
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return ""
	}
	defer f.Close()

	// 通过字符串处理来找到对应的subsystem文件夹位置
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// 47 36 0:41 / /sys/fs/cgroup/memory rw,nosuid,nodev,noexec,relatime shared:21 - cgroup cgroup rw,memory
		txt := scanner.Text()

		field := strings.Split(txt, " ")

		// 对最后的rw,memory 进行逗号分隔
		subsystems := strings.Split(field[len(field)-1], ",")
		for _, opt := range subsystems {
			if opt == subsystem {
				return field[mountPointIndex]
			}
		}
	}

	if err = scanner.Err(); err != nil {
		log.Error("read mountinfo err: ", err)
		return ""
	}
	return ""
}
