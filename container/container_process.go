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
func NewParentProcess(tty bool, volume string) (*exec.Cmd, *os.File) {
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
	}

	// 将读取方转入子进程
	cmd.ExtraFiles = []*os.File{readPipe}
	rootPath := "/root"
	NewWorkSpace(rootPath, volume)
	cmd.Dir = path.Join(rootPath, "merged")
	return cmd, writePipe
}

func NewWorkSpace(rootPath string, volume string) {
	createLower(rootPath)
	createDirs(rootPath)
	mountOverlayFS(rootPath)

	// 判断是否指定的数据卷，如果指定了数据卷则挂载数据卷
	if volume != "" {
		mntPath := path.Join(rootPath, "merged")
		hostPath, containerPath, err := utils.VolumeExtract(volume)
		if err != nil {
			logrus.Error("extract volume failed,maybe volume parameter input is not correct, detail: ", err)
			return
		}
		mountVolume(mntPath, hostPath, containerPath)
	}
}

// createLower 将busybox作为overlayfs的lower层
func createLower(rootURL string) {
	busyboxURL := path.Join(rootURL, "busybox")
	busyboxTarURL := path.Join(rootURL, "busybox.tar")
	// 检查是否存在busybox 文件夹
	exist, err := utils.PathExists(busyboxURL)
	if err != nil {
		logrus.Infof("Fail to check busybox url %v exists: %v", busyboxURL, err)
	}

	// 如果不存在则创建目录并解压到busybox文件夹
	if !exist {
		err = os.Mkdir(busyboxURL, constant.Perm0777)
		if err != nil {
			logrus.Errorf("Mkdir dir %s error: %v", busyboxURL, err)
		}
		_, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput()
		if err != nil {
			logrus.Errorf("Untar dir %s error: %v", busyboxTarURL, err)
		}
	}
}

// createDirs 创建overlayfs需要的的upper、worker目录
func createDirs(rootPath string) {
	dirs := []string{
		path.Join(rootPath, "merged"),
		path.Join(rootPath, "upper"),
		path.Join(rootPath, "work"),
	}

	for _, dir := range dirs {
		if err := os.Mkdir(dir, constant.Perm0777); err != nil {
			logrus.Errorf("mkdir dir %s error. %v", dir, err)
		}
	}
}

// mountOverlayFS 挂载overlayfs
func mountOverlayFS(rootPath string) {
	// 拼接参数
	// e.g. lowerdir=/root/busybox,upperdir=/root/upper,workdir=/root/work
	dirs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", path.Join(rootPath, "busybox"),
		path.Join(rootPath, "upper"), path.Join(rootPath, "work"))

	// 完整命令：mount -t overlay overlay -o lowerdir=/root/busybox,upperdir=/root/upper,workdir=/root/work /root/merged
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, path.Join(rootPath, "merged"))
	logrus.Infof("mount overlayfs: [%s]", cmd.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorf("%v", err)
	}
}

// DeleteWorkSpace Delete the AUFS filesystem while container exit
func DeleteWorkSpace(rootPath string, volume string) {
	mntPath := path.Join(rootPath, "merged")
	// 判断是否存在数据卷，存在则取消挂载
	if volume != "" {
		_, containerPath, err := utils.VolumeExtract(volume)
		if err != nil {
			logrus.Errorf("extract volume failed, maybe volume parameter input is not correct, detail:%v", err)
			return
		}
		umountVolume(mntPath, containerPath)
	}

	// 如果先删除再取消挂载，数据卷无法保存数据
	umountOverlayFS(mntPath)
	deleteDirs(rootPath)
}

func umountOverlayFS(mntPath string) {
	cmd := exec.Command("umount", mntPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Error(err)
	}
}

func deleteDirs(rootPath string) {
	dirs := []string{
		path.Join(rootPath, "merged"),
		path.Join(rootPath, "upper"),
		path.Join(rootPath, "work"),
	}

	for _, dir := range dirs {
		err := os.RemoveAll(dir)
		if err != nil {
			logrus.Errorf("Remove dir %s error: %v", dir, err)
		}
	}
}
