package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/yqszxx/oreo-box/container"
	"io/ioutil"
	"os"
)

func logContainer(containerName string) error {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	logFileLocation := dirURL + container.ContainerLogFile
	file, err := os.Open(logFileLocation)
	defer func() {
		log.Println("closing LogFileLocation ...")
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	if err != nil {
		return fmt.Errorf("cannot open Log container file %s : %v", logFileLocation, err)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("cannot read Log container file %s : %v", logFileLocation, err)
	}
	if _, err := fmt.Fprint(os.Stdout, string(content)); err != nil {
		return err
	}
	return nil
}
