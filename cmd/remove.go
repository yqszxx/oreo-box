package cmd

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/yqszxx/oreo-box/config"
	"github.com/yqszxx/oreo-box/internal"
	"github.com/yqszxx/oreo-box/internal/fileSystem"
	"os"
	"path"
	"strconv"
)

var removeCommand = cli.Command{
	Name:   "remove",
	Usage:  "Remove a box",
	Action: removeHandler,
}

func removeHandler(context *cli.Context) error {
	if len(context.Args()) < 1 {
		return fmt.Errorf("cannot find box name")
	}
	boxName := context.Args().Get(0)

	boxInfo, err := internal.GetBoxInfoByName(boxName)
	if err != nil {
		return fmt.Errorf("fail to get box %s info : %v", boxName, err)
	}
	pid, err := strconv.Atoi(boxInfo.Pid)
	if err != nil {
		return fmt.Errorf("fail to conver pid from string to int : %v", err)
	}
	if boxInfo.Status == internal.Running && internal.IsAlive(pid) {
		return fmt.Errorf("couldn't remove running box")
	}
	if err := fileSystem.DeleteWorkSpace(boxInfo.Volume, boxName); err != nil {
		return fmt.Errorf("cannot delete workspace of box `%s`: %v", boxName, err)
	}
	dataDir := path.Join(config.BoxDataPath, boxName)
	if err := os.RemoveAll(dataDir); err != nil {
		return fmt.Errorf("fail to remove file %s : %v", dataDir, err)
	}

	return nil
}
