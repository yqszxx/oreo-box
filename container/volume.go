package container

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

//Create a AUFS filesystem as container root workspace
func NewWorkSpace(volume, imageName, containerName string) error {
	if err := CreateReadOnlyLayer(imageName); err != nil {
		return err
	}
	if err := CreateWriteLayer(containerName); err != nil {
		return err
	}
	if err := CreateMountPoint(containerName, imageName); err != nil {
		return err
	}

	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			if err := MountVolume(volumeURLs, containerName); err != nil {
				return err
			}
			log.Printf("NewWorkSpace volume urls %q \n", volumeURLs)
		} else {
			log.Println("Volume parameter input is not correct.")
		}
	}
	return nil
}

//Decompression tar image
func CreateReadOnlyLayer(imageName string) error {
	unTarFolderUrl := RootUrl + "/" + imageName + "/"
	imageUrl := RootUrl + "/" + imageName + ".tar"
	exist, err := PathExists(unTarFolderUrl)
	if err != nil {
		return fmt.Errorf("fail to judge whether read only dir %s exist: %v", unTarFolderUrl, err)
	}
	if !exist {
		if err := os.MkdirAll(unTarFolderUrl, 0622); err != nil {
			return fmt.Errorf("fail to make read only dir %s : %v", unTarFolderUrl, err)
		}

		if _, err := exec.Command("tar", "-xvf", imageUrl, "-C", unTarFolderUrl).CombinedOutput(); err != nil {
			return fmt.Errorf("fail to untar image to read only dir %s : %v", unTarFolderUrl, err)
		}
	}
	return nil
}

func CreateWriteLayer(containerName string) error {
	writeURL := fmt.Sprintf(WriteLayerUrl, containerName)
	if err := os.MkdirAll(writeURL, 0777); err != nil {
		return fmt.Errorf("fail to make  write layer dir %s : %v", writeURL, err)
	}
	return nil
}

func MountVolume(volumeURLs []string, containerName string) error {
	parentUrl := volumeURLs[0]
	if err := os.Mkdir(parentUrl, 0777); err != nil {
		return fmt.Errorf("fail to make parent dir %s : %v", parentUrl, err)
	}
	containerUrl := volumeURLs[1]
	mntURL := fmt.Sprintf(MntUrl, containerName)
	containerVolumeURL := mntURL + "/" + containerUrl
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		return fmt.Errorf("fail to make container volume dir %s : %v", containerVolumeURL, err)
	}
	dirs := "dirs=" + parentUrl
	_, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL).CombinedOutput()
	if err != nil {
		return fmt.Errorf("fail to mount container volume : %v", err)
	}
	return nil
}

func CreateMountPoint(containerName, imageName string) error {
	mntUrl := fmt.Sprintf(MntUrl, containerName)
	if err := os.MkdirAll(mntUrl, 0777); err != nil {
		return fmt.Errorf("fail to make mountpoint dir %s : %v", mntUrl, err)
	}
	tmpWriteLayer := fmt.Sprintf(WriteLayerUrl, containerName)
	tmpImageLocation := RootUrl + "/" + imageName
	mntURL := fmt.Sprintf(MntUrl, containerName)
	dirs := "dirs=" + tmpWriteLayer + ":" + tmpImageLocation
	_, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL).CombinedOutput()
	if err != nil {
		return fmt.Errorf("fail to creat aufs mount point : %v", err)
	}
	return nil
}

//Delete the AUFS filesystem while container exit
func DeleteWorkSpace(volume, containerName string) error {
	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			if err := DeleteVolume(volumeURLs, containerName); err != nil {
				return err
			}
		}
	}
	if err := DeleteMountPoint(containerName); err != nil {
		return err
	}
	if err := DeleteWriteLayer(containerName); err != nil {
		return err
	}
	return nil
}

func DeleteMountPoint(containerName string) error {
	mntURL := fmt.Sprintf(MntUrl, containerName)
	_, err := exec.Command("umount", mntURL).CombinedOutput()
	if err != nil {
		return fmt.Errorf("fail to unmount dir %s : %v", mntURL, err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		return fmt.Errorf("fail to remove mountpoint dir %s : %v", mntURL, err)
	}
	return nil
}

func DeleteVolume(volumeURLs []string, containerName string) error {
	mntURL := fmt.Sprintf(MntUrl, containerName)
	containerUrl := mntURL + "/" + volumeURLs[1]
	if _, err := exec.Command("umount", containerUrl).CombinedOutput(); err != nil {
		return fmt.Errorf("fail to unmount conntainer volume %s : %v", containerUrl, err)
	}
	return nil
}

func DeleteWriteLayer(containerName string) error {
	writeURL := fmt.Sprintf(WriteLayerUrl, containerName)
	if err := os.RemoveAll(writeURL); err != nil {
		return fmt.Errorf("fail to remove writeLayer dir %s : %v", writeURL, err)
	}
	return nil
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
