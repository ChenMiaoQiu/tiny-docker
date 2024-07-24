package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const usage = `mydocker is a simple container runtime implementation.
			   The purpose of this project is to learn how docker works and how to write a docker by ourselves
			   Enjoy it, just for fun.`

func main() {
	app := &cli.App{
		Name:  "tinydocker",
		Usage: usage,
		Before: func(ctx *cli.Context) error {
			log.SetFormatter(&log.JSONFormatter{})

			log.SetOutput(os.Stdout)
			return nil
		},
		Commands: []*cli.Command{&initCommand, &runCommand},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
