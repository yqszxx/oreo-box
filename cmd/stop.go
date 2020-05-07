package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"github.com/yqszxx/oreo-box/config"
	"github.com/yqszxx/oreo-box/internal"
	"io/ioutil"
	"path"
	"strconv"
	"syscall"
)

var stopCommand = cli.Command{
	Name:   "stop",
	Usage:  "Stop a box",
	Action: stopHandler,
}

func stopHandler(context *cli.Context) error {
	if len(context.Args()) < 1 {
		return fmt.Errorf("no box name provided")
	}
	boxName := context.Args().Get(0)

	boxInfo, err := internal.GetBoxInfoByName(boxName)
	if err != nil {
		return fmt.Errorf("fail to get box %s info : %v", boxName, err)
	}

	pid, err := strconv.Atoi(boxInfo.Pid)
	if err != nil {
		return fmt.Errorf("fail to conver pid from string to int: %v", err)
	}

	if internal.IsAlive(pid) {
		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			return fmt.Errorf("fail to stop box `%s`: %v", boxName, err)
		}
	}

	boxInfo.Status = internal.Stopped
	boxInfo.Pid = ""
	newContentBytes, err := json.Marshal(boxInfo)
	if err != nil {
		return fmt.Errorf("fail to serilize box info for `%s`: %v", boxName, err)
	}
	configFilePath := path.Join(config.BoxDataPath, boxName, config.InfoFileName)
	if err := ioutil.WriteFile(configFilePath, newContentBytes, 0644); err != nil {
		return fmt.Errorf("fail to write data to file `%s`: %v", configFilePath, err)
	}

	return nil
}
