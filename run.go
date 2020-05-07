package main

import (
	"encoding/json"
	"fmt"
	"github.com/yqszxx/oreo-box/cgroups"
	"github.com/yqszxx/oreo-box/cgroups/subsystems"
	"github.com/yqszxx/oreo-box/container"
	"github.com/yqszxx/oreo-box/network"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func Run(tty bool, comArray []string, res *subsystems.ResourceConfig, containerName, volume, imageName string,
	envSlice []string, nw string, portmapping []string) error {
	normalExit := false

	containerID := randStringBytes(10)
	if containerName == "" {
		containerName = containerID
	}

	parent, writePipe, err := container.NewParentProcess(tty, containerName, volume, imageName, envSlice)
	if err != nil {
		return fmt.Errorf("cannot create parent process: %v", err)
	}

	if err := parent.Start(); err != nil {
		return fmt.Errorf("cannot start init process: %v", err)
	}

	//record container info
	containerName, err = recordContainerInfo(parent.Process.Pid, comArray, containerName, containerID, volume)
	if err != nil {
		return fmt.Errorf("cannot record container info %v", err)
	}

	// use containerID as cgroup name
	log.Printf("creating cgroup for %v\n", containerID)
	cgroupManager := cgroups.NewCgroupManager(containerID)
	//defer cgroupManager.Destroy()
	defer func() {
		if normalExit {
			return
		}
		log.Println("Removing cgroup...")
		if err := cgroupManager.Destroy(); err != nil {
			panic(err)
		}
	}()

	if err := cgroupManager.Set(res); err != nil {
		return fmt.Errorf("cgroup manager `set` failed with: %v", err)
	}

	defer func() {
		if normalExit {
			return
		}
		log.Println("Killing init process...")
		if err := syscall.Kill(parent.Process.Pid, syscall.SIGTERM); err != nil {
			// for init processes that don't return 0 as return value
			if err.Error() == "no such process" {
				log.Println("Init process is already dead!")
			} else {
				panic(err)
			}
		}
		log.Println("Removing workspace...")
		if err := container.DeleteWorkSpace(volume, containerName); err != nil {
			panic(err)
		}
	}()

	if err := cgroupManager.Apply(parent.Process.Pid); err != nil {
		return fmt.Errorf("cgroup manager `apply` failed with: %v", err)
	}

	if nw != "" {
		// config container network
		if err := network.Init(); err != nil {
			return err
		}
		containerInfo := &container.ContainerInfo{
			Id:          containerID,
			Pid:         strconv.Itoa(parent.Process.Pid),
			Name:        containerName,
			PortMapping: portmapping,
		}
		if err := network.Connect(nw, containerInfo); err != nil {
			return fmt.Errorf("cannot connect network: %v", err)
		}
	}

	if err := sendInitCommand(comArray, writePipe); err != nil {
		return err
	}

	if tty {
		if err := parent.Wait(); err != nil {
			return fmt.Errorf("error waiting init process: %v", err)
		}
		if err := deleteContainerInfo(containerName); err != nil {
			return fmt.Errorf("cannot delete container info dir: %v", err)
		}

		if err := container.DeleteWorkSpace(volume, containerName); err != nil {
			return fmt.Errorf("cannot delete workspace: %v", err)
		}
		log.Println("Interactive mode terminated successfully")
	}

	normalExit = true
	return nil
}

func sendInitCommand(comArray []string, writePipe *os.File) error {
	command := strings.Join(comArray, " ")
	log.Printf("command all is %s", command)
	if _, err := writePipe.WriteString(command); err != nil {
		return err
	}
	if err := writePipe.Close(); err != nil {
		return err
	}
	return nil
}

func recordContainerInfo(containerPID int, commandArray []string, containerName, id, volume string) (string, error) {
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")
	containerInfo := &container.ContainerInfo{
		Id:          id,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      container.RUNNING,
		Name:        containerName,
		Volume:      volume,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		return "", err
	}
	jsonStr := string(jsonBytes)

	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		return "", err
	}
	fileName := dirUrl + "/" + container.ConfigName
	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()
	if _, err := file.WriteString(jsonStr); err != nil {
		return "", err
	}

	return containerName, nil
}

func deleteContainerInfo(containerId string) error {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirURL); err != nil {
		return fmt.Errorf("cannot remove dir %s : %v", dirURL, err)
	}

	return nil
}

func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
