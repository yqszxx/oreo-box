package internal

import (
	"encoding/json"
	"fmt"
	"github.com/yqszxx/oreo-box/config"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"
)

func IsAlive(pid int) bool {
	process, _ := os.FindProcess(pid)
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return false
	}
	return true
}

func GetBoxPidByName(boxName string) (string, error) {
	infoFilePath := path.Join(config.BoxDataPath, boxName, config.InfoFileName)
	contentBytes, err := ioutil.ReadFile(infoFilePath)
	if err != nil {
		return "", err
	}
	var BoxInfo BoxInfo
	if err := json.Unmarshal(contentBytes, &BoxInfo); err != nil {
		return "", err
	}
	return BoxInfo.Pid, nil
}

func GetEnvsByPid(pid string) ([]string, error) {
	environPath := fmt.Sprintf("/proc/%s/environ", pid)
	contentBytes, err := ioutil.ReadFile(environPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s : %v", environPath, err)
	}
	//env split by \u0000
	envs := strings.Split(string(contentBytes), "\u0000")
	return envs, nil
}

func GetBoxInfoByName(boxName string) (*BoxInfo, error) {
	configFilePath := path.Join(config.BoxDataPath, boxName, config.InfoFileName)
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("fail to read file %s : %v", configFilePath, err)
	}
	var BoxInfo BoxInfo
	if err := json.Unmarshal(contentBytes, &BoxInfo); err != nil {
		return nil, fmt.Errorf("fail to get BoxInfoByName unmarshal : %v", err)

	}
	return &BoxInfo, nil
}
