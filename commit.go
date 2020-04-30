package main

import (
	"fmt"
	"github.com/yqszxx/oreo-box/container"
	"os/exec"
)

func commitContainer(containerName, imageName string) error {
	mntURL := fmt.Sprintf(container.MntUrl, containerName)
	mntURL += "/"

	imageTar := container.RootUrl + "/" + imageName + ".tar"

	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		return fmt.Errorf("fail to tar folder %s : %v", mntURL, err)
	}
	return nil
}
