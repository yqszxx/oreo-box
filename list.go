package main

import (
	"encoding/json"
	"fmt"
	"github.com/yqszxx/oreo-box/container"
	"io/ioutil"
	"os"
	"text/tabwriter"
)

func ListContainers() error {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, "")
	dirURL = dirURL[:len(dirURL)-1]
	files, err := ioutil.ReadDir(dirURL)
	if err != nil {
		return fmt.Errorf("cannot read dir %s : %v", dirURL, err)
	}

	var containers []*container.ContainerInfo
	for _, file := range files {
		if file.Name() == "network" {
			continue
		}
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			return fmt.Errorf("cannot get container info : %v", err)
		}
		containers = append(containers, tmpContainer)
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	if _, err := fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n"); err != nil {
		return fmt.Errorf("fail to exec fmt.Fprint : %v", err)
	}
	for _, item := range containers {
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

func getContainerInfo(file os.FileInfo) (*container.ContainerInfo, error) {
	containerName := file.Name()
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFileDir = configFileDir + container.ConfigName
	content, err := ioutil.ReadFile(configFileDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s : %v", configFileDir, err)
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		return nil, fmt.Errorf("fail to exec json unmarshal : %v", err)
	}

	return &containerInfo, nil
}
