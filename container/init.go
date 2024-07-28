package container

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ChenMiaoQiu/tiny-docker/constant"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// index0：标准输入
// index1：标准输出
// index2：标准错误
// index3：带过来的第一个FD，也就是readPipe
const fdIndex = 3

// RunContainerInitProcess 启动容器的init进程
/*
这里的init函数是在容器内部执行的，也就是说，代码执行到这里后，容器所在的进程其实就已经创建出来了，
这是本容器执行的第一一个进程。
使用mount先去挂载proc文件系统，以便后面通过ps等系统命令去查看当前进程资源的情况。
*/
func RunContainerInitProcess() error {
	command := readUserCommand()
	if command == nil {
		return errors.New("run command in container err, command is nil")
	}

	// 挂载文件系统
	setUpMount()

	path, err := exec.LookPath(command[0])
	if err != nil {
		logrus.Errorf("Exec loop path error %v", err)
		return err
	}

	logrus.Info("Find path: ", path)
	logrus.Info("All command is: ", command)

	if err := syscall.Exec(path, command, os.Environ()); err != nil {
		logrus.Errorf(err.Error())
	}
	return nil
}

// readUserCommand 从管道中获取要执行的命令
func readUserCommand() []string {
	pipe := os.NewFile(uintptr(fdIndex), "pipe")
	defer pipe.Close()
	msg, err := io.ReadAll(pipe)
	if err != nil {
		logrus.Error("Fail to read pipe msg: ", msg)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

// 初始化挂载点
func setUpMount() {
	pwd, err := os.Getwd()
	if err != nil {
		logrus.Errorf("Get current location error %v", err)
		return
	}
	logrus.Infof("Current location is %s", pwd)

	// mount proc 之前先把所有挂载点的传播类型改为 private，避免本 namespace 中的挂载事件外泄。
	err = syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	if err != nil {
		logrus.Error("set private mount fail: ", err)
		return
	}

	err = pivoteRoot(pwd)
	if err != nil {
		logrus.Errorf("pivotRoot failed: %v", err)
		return
	}

	// 如果不先做 private mount，会导致挂载事件外泄，后续再执行 mydocker 命令时 /proc 文件系统异常
	// 执行 mount -t proc proc /proc 命令重新挂载来解决
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	_ = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")

	// 由于前面 pivotRoot 切换了 rootfs，因此这里重新 mount 一下 /dev 目录
	// tmpfs 是基于 件系 使用 RAM、swap 分区来存储。
	// 不挂载 /dev，会导致容器内部无法访问和使用许多设备，这可能导致系统无法正常工作
	err = syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	if err != nil {
		logrus.Error("mount tmpfs error: ", err)
	}
}

// pivoteRoot 原先的根文件系统会被移到指定的目录，而新的根文件系统会变为指定的目录
// PivotRoot调用有限制，newRoot和oldRoot不能在同一个文件系统下。
// 因此，为了使当前root的老root和新root不在同一个文件系统下，这里把root重新mount了一次。
func pivoteRoot(root string) error {
	// 重复挂载root目录，创建一个镜像副本
	err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, "")
	if err != nil {
		return errors.Wrap(err, "mount rootfs to itself")
	}
	// 创建临时目录挂载旧root
	pivotDir := filepath.Join(root, ".pivot_root")
	// 创建一个文件夹，开放所有读写执行权限
	err = os.Mkdir(pivotDir, constant.Perm0777)
	if err != nil {
		return err
	}
	// 执行pivot_root调用,将系统rootfs切换到新的rootfs,
	// PivotRoot调用会把 old_root挂载到pivotDir,也就是rootfs/.pivot_root,挂载点现在依然可以在mount命令中看到
	err = syscall.PivotRoot(root, pivotDir)
	if err != nil {
		return errors.WithMessagef(err, "pivotRoot failed,new_root:%v old_root:%v", root, pivotDir)
	}
	// 修改当前工作目录至新的根目录
	err = syscall.Chdir("/")
	if err != nil {
		return errors.WithMessage(err, "chdir to / failed")
	}

	// 最后再把old_root umount了，即 umount rootfs/.pivot_root
	// 由于当前已经是在 rootfs 下了，就不能再用上面的rootfs/.pivot_root这个路径了,现在直接用/.pivot_root这个路径即可
	pivotDir = filepath.Join("/", ".pivot_root")
	err = syscall.Unmount(pivotDir, syscall.MNT_DETACH)
	if err != nil {
		return errors.WithMessage(err, "unmount pivote_root dir fail")
	}
	// 删除临时挂载文件夹
	return os.Remove(pivotDir)
}
