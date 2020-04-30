package main

import (
	"encoding/json"
	"fmt"
	"github.com/yqszxx/oreo-box/container"
	_ "github.com/yqszxx/oreo-box/nsenter"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

const ENV_EXEC_PID = "mydocker_pid"
const ENV_EXEC_CMD = "mydocker_cmd"

func ExecContainer(containerName string, comArray []string) error {
	pid, err := GetContainerPidByName(containerName)
	if err != nil {
		return fmt.Errorf("getContainerPidByName(%s) failed with error: %v", containerName, err)
	}

	cmdStr := strings.Join(comArray, " ")
	log.Printf("exec in container pid %s with command '%s'\n", pid, cmdStr)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := os.Setenv(ENV_EXEC_PID, pid); err != nil {
		return err
	}
	if err := os.Setenv(ENV_EXEC_CMD, cmdStr); err != nil {
		return err
	}

	containerEnvs, err := getEnvsByPid(pid)
	if err != nil {
		return err
	}
	cmd.Env = append(os.Environ(), containerEnvs...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cannot exec in container %s: %v", containerName, err)
	}

	return nil
}

func GetContainerPidByName(containerName string) (string, error) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}

func getEnvsByPid(pid string) ([]string, error) {
	path := fmt.Sprintf("/proc/%s/environ", pid)
	contentBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s : %v", path, err)
	}
	//env split by \u0000
	envs := strings.Split(string(contentBytes), "\u0000")
	return envs, nil
}
