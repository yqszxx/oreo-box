package container

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func RunContainerInitProcess() error {
	cmdArray, err := readUserCommand()
	if err != nil {
		return err
	}
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("run container get user command error, cmdArray is nil")
	}
	//ls, err := exec.Command("/bin/ls", "-al", "/proc/self/ns").CombinedOutput()
	//fmt.Println(string(ls))
	//os.Exit(0)
	if err := setUpMount(); err != nil {
		return err
	}
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		return fmt.Errorf("fail to search for an executable named file in the dir path : %v", err)
	}
	log.Printf("Find path %s", path)
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}

func readUserCommand() ([]string, error) {
	pipe := os.NewFile(uintptr(3), "pipe")
	defer pipe.Close()
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		return nil, fmt.Errorf("init read pipe error : %v", err)
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " "), nil
}

/**
Init 挂载点
*/
func setUpMount() error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("fail to get current location : %v", err)
	}
	log.Printf("Current location is %s \n", pwd)
	//pivot root
	if err := pivotRoot(pwd); err != nil {
		return err
	}
	//mount proc and dev
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	if err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), ""); err != nil {
		return fmt.Errorf(err.Error())
	}
	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755"); err != nil {
		return fmt.Errorf(err.Error())
	}
	return nil
}

func pivotRoot(root string) error {
	/**
	  为了使当前root的老 root 和新 root 不在同一个文件系统下，我们把root重新mount了一次
	  bind mount是把相同的内容换了一个挂载点的挂载方法
	*/
	//creat direction rootfs/.pivotDir to store old root
	pivotDir := filepath.Join(root, ".old_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return fmt.Errorf("fail to make old root dir : %v", err)
	}

	//mount root to make sure the new rootfs is not in the same location as the old rootfs
	if err := syscall.Mount("/", "/", "private", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("fail to mount rootfs to itself : %v", err)
	}

	// switch the filesystem to the new root
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("fail to pivot root : %v", err)
	}

	// change the current working directory to the new root
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("fail to change dir to / : %v", err)
	}

	//the direction of old root in the new root is changed
	pivotDir = filepath.Join("/", ".old_root")
	//unmount old root from new root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("fail to unmount old root : %v", err)
	}
	//delete the temporary direction
	if err := os.Remove(pivotDir); err != nil {
		return fmt.Errorf("fail to remove old root : %v", err)
	}
	return nil
}
