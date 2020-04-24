package main

import (
	"encoding/json"
	"fmt"
	"github.com/yqszxx/oreo-box/container"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
)

func stopContainer(containerName string) error {
	pid, err := GetContainerPidByName(containerName)
	if err != nil {
		return fmt.Errorf("fail to get contaienr pid by name %s : %v", containerName, err)
	}
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return fmt.Errorf("fail to conver pid from string to int : %v", err)
	}
	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		return fmt.Errorf("fail to stop container %s : %v", containerName, err)
	}
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		return fmt.Errorf("fail to get container %s info : %v", containerName, err)
	}
	containerInfo.Status = container.STOP
	containerInfo.Pid = " "
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		return fmt.Errorf("fail to exec Json marshal %s : %v", containerName, err)
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	if err := ioutil.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		return fmt.Errorf("fail to write data to file %s : %v", configFilePath, err)
	}
	return nil
}

func getContainerInfoByName(containerName string) (*container.ContainerInfo, error) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("fail to read file %s : %v", configFilePath, err)
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return nil, fmt.Errorf("fail to get ContainerInfoByName unmarshal : %v", err)

	}
	return &containerInfo, nil
}

func removeContainer(containerName string) error {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		return fmt.Errorf("fail to get container %s info : %v", containerName, err)
	}
	if containerInfo.Status != container.STOP {
		return fmt.Errorf("couldn't remove running container")
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		return fmt.Errorf("fail to remove file %s : %v", dirURL, err)
	}
	if err := container.DeleteWorkSpace(containerInfo.Volume, containerName); err != nil {
		return err
	}
	return nil
}
