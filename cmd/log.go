package cmd

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/yqszxx/oreo-box/config"
	"io/ioutil"
	"log"
	"os"
	"path"
)

var logCommand = cli.Command{
	Name:   "logs",
	Usage:  "print logs of a box",
	Action: logHandler,
}

func logHandler(context *cli.Context) error {
	if len(context.Args()) < 1 {
		return fmt.Errorf("no box name provided")
	}
	boxName := context.Args().Get(0)

	logFileLocation := path.Join(config.BoxDataPath, boxName, config.LogFileName)

	file, err := os.Open(logFileLocation)
	defer func() {
		log.Println("closing LogFileLocation ...")
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	if err != nil {
		return fmt.Errorf("cannot open Log box file %s : %v", logFileLocation, err)
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("cannot read Log box file %s : %v", logFileLocation, err)
	}

	if _, err := fmt.Fprint(os.Stdout, string(content)); err != nil {
		return err
	}
	return nil
}
