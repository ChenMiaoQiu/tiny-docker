package subsystem

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/ChenMiaoQiu/tiny-docker/constant"
)

type CpusetSubsystem struct {
}

func (s *CpusetSubsystem) Name() string {
	return "cpuset"
}

// Set 设置cgroupPath 对应的cgroup cpuset限制
func (s *CpusetSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if res.CpuSet == "" {
		return nil
	}

	subsysCgroupPath, err := getCgroupPath(s.Name(), cgroupPath, true)
	if err != nil {
		return err
	}
	// 设置这个cgroup的cpuset限制，即将限制写入到cgroup对应目录的cpuset.cpus 文件中。
	err = os.WriteFile(path.Join(subsysCgroupPath, "cpuset.cpus"), []byte(res.CpuSet), constant.Perm0644)
	if err != nil {
		return fmt.Errorf("set cgroup cpuset fail %v", err)
	}
	return nil
}

// Apply 将pid加入到对应cgroupPath对应的cgroup中
func (s *CpusetSubsystem) Apply(cgroupPath string, pid int, res *ResourceConfig) error {
	if res.CpuSet == "" {
		return nil
	}
	subsysCgroupPath, err := getCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return fmt.Errorf("%v fail get cgroup: %s", err, cgroupPath)
	}

	err = os.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), constant.Perm0644)
	if err != nil {
		return fmt.Errorf("set cgroup proc fail: %v", err)
	}
	return nil
}

// Remove 删除cgroupPath对应的cgroup
func (s *CpusetSubsystem) Remove(cgroupPath string) error {
	subsysCgroupPath, err := getCgroupPath(s.Name(), cgroupPath, false)
	if err != nil {
		return err
	}
	return os.RemoveAll(subsysCgroupPath)
}
