package main

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/yqszxx/oreo-box/cmd"
	"io/ioutil"
	"log"
	"os"
)

const usage = `A lightweight yet secure container runtime.`

func main() {

	app := cli.NewApp()
	app.Name = "Oreo Box"
	app.Version = "20.05-rc1"
	app.Usage = usage

	app.Commands = cmd.Commands

	app.Before = func(context *cli.Context) error {
		// Log as JSON instead of the default ASCII formatter.
		log.SetFlags(log.Llongfile | log.LstdFlags)

		log.SetOutput(os.Stdout)
		return nil
	}

	cli.OsExiter = func(code int) {}
	cli.ErrWriter = ioutil.Discard

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
