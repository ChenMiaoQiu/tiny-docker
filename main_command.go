package main

import (
	"fmt"

	"github.com/ChenMiaoQiu/tiny-docker/cgroups/subsystem"
	"github.com/ChenMiaoQiu/tiny-docker/container"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit
			mydocker run -it [command]`,
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
		if ctx.Args().Len() < 1 {
			return fmt.Errorf("missing container command")
		}
		cmd := ctx.Args().Slice()
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
			Run(tty, cmd, limitConfig, volume, containerName)
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
	Usage: "commit container to image",
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() < 1 {
			return fmt.Errorf("missing image name")
		}
		imageName := ctx.Args().Get(0)
		return commitContainer(imageName)
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
