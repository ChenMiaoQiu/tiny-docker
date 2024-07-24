package container

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
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
	mountProc()

	command := readUserCommand()
	if command == nil {
		return errors.New("run command in container err, command is nil")
	}
	path, err := exec.LookPath(command[0])
	if err != nil {
		log.Errorf("Exec loop path error %v", err)
		return err
	}

	log.Info("Find path: ", path)
	log.Info("All command is: ", command)
	if err := syscall.Exec(path, command[1:], os.Environ()); err != nil {
		log.Errorf(err.Error())
	}
	return nil
}

// 从管道中获取要执行的命令
func readUserCommand() []string {
	pipe := os.NewFile(uintptr(fdIndex), "pipe")
	defer pipe.Close()
	msg, err := io.ReadAll(pipe)
	if err != nil {
		log.Error("Fail to read pipe msg: ", msg)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

func mountProc() {
	// systemd 加入linux之后, mount namespace 就变成 shared by default, 所以你必须显示声明你要这个新的mount namespace独立。
	// 即 mount proc 之前先把所有挂载点的传播类型改为 private，避免本 namespace 中的挂载事件外泄。
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	// 如果不先做 private mount，会导致挂载事件外泄，后续再执行 mydocker 命令时 /proc 文件系统异常
	// 执行 mount -t proc proc /proc 命令重新挂载来解决
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	_ = syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
}
