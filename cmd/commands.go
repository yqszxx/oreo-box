package cmd

import (
	"github.com/urfave/cli"
)

var Commands = []cli.Command{
	initCommand,
	runCommand,
	listCommand,
	logCommand,
	execCommand,
	stopCommand,
	removeCommand,
	networkCommand,
	imageCommand,
}
