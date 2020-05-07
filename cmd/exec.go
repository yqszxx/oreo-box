package cmd

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/yqszxx/oreo-box/config"
	"github.com/yqszxx/oreo-box/internal"
	"log"
	"os"
	"os/exec"
	"strings"
)

var execCommand = cli.Command{
	Name:   "exec",
	Usage:  "Execute a command inside specified box",
	Action: execHandler,
}

func execHandler(context *cli.Context) error {
	//This is for callback
	if os.Getenv(config.EnvExecPid) != "" && os.Getenv(config.EnvExecCmd) != "" {
		log.Println("Cgo code executed, no further actions needed, exiting...")
		return nil
	}

	if len(context.Args()) < 2 {
		return fmt.Errorf("no enough arguments provided")
	}
	boxName := context.Args().Get(0)
	var commandArray []string
	for _, arg := range context.Args().Tail() {
		commandArray = append(commandArray, arg)
	}

	pid, err := internal.GetBoxPidByName(boxName)
	if err != nil {
		return fmt.Errorf("getBoxPidByName(%s) failed with error: %v", boxName, err)
	}

	cmdStr := strings.Join(commandArray, " ")
	log.Printf("exec in box pid %s with command '%s'\n", pid, cmdStr)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := os.Setenv(config.EnvExecPid, pid); err != nil {
		return err
	}
	if err := os.Setenv(config.EnvExecPid, cmdStr); err != nil {
		return err
	}

	boxEnvs, err := internal.GetEnvsByPid(pid)
	if err != nil {
		return err
	}
	cmd.Env = append(os.Environ(), boxEnvs...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot exec in box %s: %v", boxName, err)
	}

	return nil
}
