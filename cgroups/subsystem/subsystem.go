package subsystem

// ResourceConfig 用于传递资源限制配置的结构体，包含内存限制，CPU 时间片权重，CPU核心数
type ResourceConfig struct {
	MemoryLimit string
	CpuCfsQuota int
	CpuShare    string
	CpuSet      string
}

type Subsystem interface {
	// Name 返回当前Subsystem的名称,比如cpu、memory
	Name() string
	// Set 设置某个cgroup在这个Subsystem中的资源限制
	Set(path string, res *ResourceConfig) error
	// Apply 将进程添加到某个cgroup中
	Apply(path string, pid int, res *ResourceConfig) error
	// Remove 移除某个cgroup
	Remove(path string) error
}

var SubsystemsIns = []Subsystem{
	&MemorySubsystem{},
	&CpuSubsystem{},
	&CpusetSubsystem{},
}
