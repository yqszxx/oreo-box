package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"github.com/yqszxx/oreo-box/config"
	"github.com/yqszxx/oreo-box/internal"
	"io/ioutil"
	"os"
	"path"
	"text/tabwriter"
)

var listCommand = cli.Command{
	Name:   "ps",
	Usage:  "list all the boxes",
	Action: listHandler,
}

func listHandler(*cli.Context) error {
	files, err := ioutil.ReadDir(config.BoxDataPath)
	if err != nil {
		return fmt.Errorf("cannot read dir %s : %v", config.BoxDataPath, err)
	}

	var boxes []*internal.BoxInfo
	for _, file := range files {
		tmpBox, err := getBoxInfo(file)
		if err != nil {
			return fmt.Errorf("cannot get box info : %v", err)
		}
		boxes = append(boxes, tmpBox)
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	if _, err := fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n"); err != nil {
		return fmt.Errorf("fail to exec fmt.Fprint : %v", err)
	}
	for _, item := range boxes {
		_, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime)
		if err != nil {
			return fmt.Errorf("fail to exec fmt.Fprintf %v", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("cannot flush : %v", err)
	}
	return nil
}

func getBoxInfo(file os.FileInfo) (*internal.BoxInfo, error) {
	boxName := file.Name()
	configFileDir := path.Join(config.BoxDataPath, boxName, config.InfoFileName)
	content, err := ioutil.ReadFile(configFileDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s : %v", configFileDir, err)
	}
	var BoxInfo internal.BoxInfo
	if err := json.Unmarshal(content, &BoxInfo); err != nil {
		return nil, fmt.Errorf("fail to exec json unmarshal : %v", err)
	}

	return &BoxInfo, nil
}
