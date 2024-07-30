package container

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"syscall"

	"github.com/ChenMiaoQiu/tiny-docker/constant"
	"github.com/ChenMiaoQiu/tiny-docker/utils"
	"github.com/sirupsen/logrus"
)

// NewParentProcess 构建 command 用于启动一个新进程
/*
这里是父进程，也就是当前进程执行的内容。
1.这里的/proc/se1f/exe调用中，/proc/self/ 指的是当前运行进程自己的环境，exec 其实就是自己调用了自己，使用这种方式对创建出来的进程进行初始化
2.后面的args是参数，其中init是传递给本进程的第一个参数，在本例中，其实就是会去调用initCommand去初始化进程的一些环境和资源
3.下面的clone参数就是去fork出来一个新进程，并且使用了namespace隔离新创建的进程和外部环境。
4.如果用户指定了-it参数，就需要把当前进程的输入输出导入到标准输入输出上
*/
func NewParentProcess(tty bool, volume string, containerId string, imageName string) (*exec.Cmd, *os.File) {
	// 创建匿名管道用于传递参数
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		logrus.Errorf("Create pipe error: %v", err)
		return nil, nil
	}
	// 这里的 init 指令就用用来在子进程中调用 initCommand
	cmd := exec.Command("/proc/self/exe", "init")
	// 设置隔离模式
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	// 将输入输出绑定至终端
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// 对于后台运行的容器，将stdout、stderr 重定向到日志文件中
		dirPath := fmt.Sprintf(InfoLocFormat, containerId)
		if err = os.MkdirAll(dirPath, constant.Perm0622); err != nil {
			logrus.Errorf("NewParentProcess mkdir %s error %v", dirPath, err)
			return nil, nil
		}
		stdLogFilePath := path.Join(dirPath, GetLogFile(containerId))
		stdLogFile, err := os.Create(stdLogFilePath)
		if err != nil {
			logrus.Errorf("NewParentProcess create file %s error %v", stdLogFilePath, err)
			return nil, nil
		}
		cmd.Stdout = stdLogFile
		cmd.Stderr = stdLogFile
	}

	// 将读取方转入子进程
	cmd.ExtraFiles = []*os.File{readPipe}
	NewWorkSpace(containerId, imageName, volume)
	cmd.Dir = utils.GetMerged(containerId)
	return cmd, writePipe
}
