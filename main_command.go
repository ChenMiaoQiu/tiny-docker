package main

import (
	"fmt"
	"os"

	"github.com/ChenMiaoQiu/tiny-docker/cgroups/subsystem"
	"github.com/ChenMiaoQiu/tiny-docker/container"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit
			mydocker run -it [imageName] [command]`,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "it",
			Usage: "enable tty",
		},
		&cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		&cli.StringFlag{
			Name:  "mem",
			Usage: "memory limit,e.g.: -mem 100m", // 限制内存使用率
		},
		&cli.IntFlag{
			Name:  "cpu",
			Usage: "cpu quota,e.g.: -cpu 100", // 限制进程 cpu 使用率
		},
		&cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit,e.g.: -cpuset 2,4", // 限制进程 cpu 使用率
		},
		&cli.StringFlag{
			Name:  "v",
			Usage: "volume,e.g.: -v /etc/conf:/etc/conf",
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "container name, e.g.: -name containerName",
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() < 2 {
			return fmt.Errorf("missing imageName or container command")
		}
		imageName := ctx.Args().Get(0)
		cmd := ctx.Args().Slice()[1:]
		tty := ctx.Bool("it")
		detach := ctx.Bool("d")

		// tty 和 detach只能生效一个
		if tty && detach {
			return fmt.Errorf("it and d paramter can not both provided")
		}

		// 构建资源控制器
		memoryLimit := ctx.String("mem")
		cpuLimit := ctx.Int("cpu")
		cpusetLimit := ctx.String("cpuset")

		limitConfig := &subsystem.ResourceConfig{
			MemoryLimit: memoryLimit,
			CpuSet:      cpusetLimit,
			CpuCfsQuota: cpuLimit,
		}

		volume := ctx.String("v")
		containerName := ctx.String("name")
		if tty || detach {
			Run(tty, cmd, limitConfig, volume, containerName, imageName)
		}
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(ctx *cli.Context) error {
		log.Infof("init container")
		err := container.RunContainerInitProcess()
		return err
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit container to image. e.g. tiny-docker commit [containerId] [imageName]",
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() < 2 {
			return fmt.Errorf("missing container id or image name")
		}
		containerId := ctx.Args().Get(0)
		imageName := ctx.Args().Get(1)

		return commitContainer(containerId, imageName)
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all the containers",
	Action: func(ctx *cli.Context) error {
		ListContainerInfos()
		return nil
	},
}

var logCommand = cli.Command{
	Name:  "logs",
	Usage: "print logs of a container",
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() < 1 {
			return fmt.Errorf("please input your container name")
		}
		containerName := ctx.Args().Get(0)
		logContainer(containerName)
		return nil
	},
}

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command into container",
	Action: func(ctx *cli.Context) error {
		// 如果环境变量存在，说明C代码已经运行过了，即setns系统调用已经执行了，这里就直接返回，避免重复执行
		if os.Getenv(EnvExecPid) != "" {
			log.Infof("pid callback pid %v", os.Getgid())
			return nil
		}
		// tiny-docker [container name] [cmd]
		if ctx.Args().Len() < 2 {
			return fmt.Errorf("missing container name or command")
		}
		containerName := ctx.Args().Get(0)
		// 第0位为容器名
		cmdArray := ctx.Args().Slice()[1:]
		ExecContainer(containerName, cmdArray)
		return nil
	},
}

var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop container,e.g. tiny-docker stop [containerId]",
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() < 1 {
			return fmt.Errorf("missing container id")
		}
		containerId := ctx.Args().Get(0)
		stopContainer(containerId)
		return nil
	},
}

var removeCommand = cli.Command{
	Name:  "rm",
	Usage: "remove container,e.g. tiny-docker rm [containerId]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "f",
			Usage: "enforce remove container",
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() < 1 {
			return fmt.Errorf("missing container id")
		}
		force := ctx.Bool("f")
		containerId := ctx.Args().Get(0)
		removeContainer(containerId, force)
		return nil
	},
}
