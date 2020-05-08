package fileSystem

import (
	"fmt"
	"github.com/yqszxx/oreo-box/config"
	"github.com/yqszxx/oreo-box/internal"
	"log"
	"os"
	"path"
	"strings"
	"syscall"
)

//Create a AUFS filesystem as box root workspace
func NewWorkSpace(volume, imageName, boxName string) error {
	if err := CreateWriteLayer(boxName); err != nil {
		return err
	}
	if err := MountImage(boxName, imageName); err != nil {
		return err
	}

	if volume != "" {
		volumePaths := strings.Split(volume, ":")
		length := len(volumePaths)
		if length == 2 && volumePaths[0] != "" && volumePaths[1] != "" {
			if err := MountVolume(volumePaths, boxName); err != nil {
				return err
			}
			log.Printf("NewWorkSpace volume Paths %q \n", volumePaths)
		} else {
			log.Println("Volume parameter input is not correct.")
		}
	}
	return nil
}

func CreateWriteLayer(boxName string) error {
	writePath := path.Join(config.WritableLayerPath, boxName)
	if err := os.MkdirAll(writePath, 0755); err != nil {
		return fmt.Errorf("fail to make  write layer dir %s : %v", writePath, err)
	}
	return nil
}

func MountVolume(volumePaths []string, boxName string) error {
	parentPath := volumePaths[0]
	if err := os.Mkdir(parentPath, 0755); err != nil {
		return fmt.Errorf("fail to make parent dir %s : %v", parentPath, err)
	}
	boxPath := volumePaths[1]
	mountPath := path.Join(config.BoxDataPath, boxName, config.MountPath)
	boxVolumePath := path.Join(mountPath, boxPath)
	if err := os.Mkdir(boxVolumePath, 0755); err != nil {
		return fmt.Errorf("fail to make box volume dir %s : %v", boxVolumePath, err)
	}
	dirs := "dirs=" + parentPath
	if err := syscall.Mount("none", mountPath, "aufs", 0, dirs); err != nil {
		return fmt.Errorf("fail to mount box volume : %v", err)
	}
	return nil
}

func MountImage(boxName, imageName string) error {
	mountPath := path.Join(config.BoxDataPath, boxName, config.MountPath)
	if err := os.MkdirAll(mountPath, 0755); err != nil {
		return fmt.Errorf("fail to make mountpoint dir %s : %v", mountPath, err)
	}
	writableLayerPath := path.Join(config.WritableLayerPath, boxName)
	imagePath := path.Join(config.ImagePath, imageName)
	if !internal.Exist(imagePath, true) {
		return fmt.Errorf("cannot find image `%s` at `%s`", imageName, imagePath)
	}

	dirs := "dirs=" + writableLayerPath + ":" + imagePath
	if err := syscall.Mount("none", mountPath, "aufs", 0, dirs); err != nil {
		return fmt.Errorf("fail to mount aufs `%s` -> `%s`: %v", dirs, mountPath, err)
	}
	return nil
}

//Delete the AUFS filesystem when box exits
func DeleteWorkSpace(volume, boxName string) error {
	if volume != "" {
		volumePaths := strings.Split(volume, ":")
		length := len(volumePaths)
		if length == 2 && volumePaths[0] != "" && volumePaths[1] != "" {
			if err := DeleteVolume(volumePaths, boxName); err != nil {
				return err
			}
		}
	}
	if err := DeleteMountPoint(boxName); err != nil {
		return err
	}
	if err := DeleteWriteLayer(boxName); err != nil {
		return err
	}
	return nil
}

func DeleteMountPoint(boxName string) error {
	mountPath := path.Join(config.BoxDataPath, boxName, config.MountPath)
	if err := syscall.Unmount(mountPath, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("fail to unmount dir %s : %v", mountPath, err)
	}
	if err := os.RemoveAll(mountPath); err != nil {
		return fmt.Errorf("fail to remove mountpoint dir %s : %v", mountPath, err)
	}
	return nil
}

func DeleteVolume(volumePaths []string, boxName string) error {
	mountPath := path.Join(config.BoxDataPath, boxName, config.MountPath)
	volumeMountPath := path.Join(mountPath, volumePaths[1])
	if err := syscall.Unmount(volumeMountPath, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("fail to unmount box volume %s : %v", volumeMountPath, err)
	}
	return nil
}

func DeleteWriteLayer(boxName string) error {
	writePath := path.Join(config.WritableLayerPath, boxName)
	if err := os.RemoveAll(writePath); err != nil {
		return fmt.Errorf("fail to remove writeLayer dir %s : %v", writePath, err)
	}
	return nil
}
