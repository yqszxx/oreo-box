package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli"
	"github.com/yqszxx/oreo-box/config"
	"github.com/yqszxx/oreo-box/internal"
	"github.com/yqszxx/oreo-box/internal/cgroup"
	"github.com/yqszxx/oreo-box/internal/cgroup/subsystems"
	"github.com/yqszxx/oreo-box/internal/fileSystem"
	"github.com/yqszxx/oreo-box/internal/network"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: "Create a box",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "i",
			Usage: "interactive mode",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpusetcpus",
			Usage: "cpusetcpus limit",
		},
		cli.StringFlag{
			Name:  "cpusetmems",
			Usage: "cpusetmems limit",
		},
		cli.StringFlag{
			Name:  "cpuquota",
			Usage: "cpuquota limit",
		},
		cli.StringFlag{
			Name:  "name",
			Usage: "box name",
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
		cli.StringSliceFlag{
			Name:  "e",
			Usage: "set environment",
		},
		cli.StringFlag{
			Name:  "net",
			Usage: "box network",
		},
		cli.StringSliceFlag{
			Name:  "p",
			Usage: "port mapping",
		},
	},
	Action: runHandler,
}

func runHandler(context *cli.Context) error {
	normalExit := false

	if len(context.Args()) < 1 {
		return fmt.Errorf("no enough arguments provided")
	}
	var cmdArray []string
	for _, arg := range context.Args() {
		cmdArray = append(cmdArray, arg)
	}

	//get image name
	//noinspection GoNilness
	imageName := cmdArray[0]
	//noinspection GoNilness
	cmdArray = cmdArray[1:]

	interactive := context.Bool("i")

	resConf := &subsystems.ResourceConfig{
		MemoryLimit: context.String("m"),
		CpuSetMems:  context.String("cpusetmems"),
		CpuSetCpus:  context.String("cpusetcpus"),
		CpuShare:    context.String("cpushare"),
		CpuQuotaUs:  context.String("cpuquota"),
	}

	boxName := context.String("name")
	volume := context.String("v")
	networkName := context.String("net")

	envSlice := context.StringSlice("e")
	portMapping := context.StringSlice("p")

	boxID := randStringBytes(10)
	if boxName == "" {
		boxName = boxID
	}

	// create pipe for sending command into box
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("cannot create new pipe: %v", err)
	}
	initCmd, err := os.Readlink("/proc/self/exe")
	if err != nil {
		return fmt.Errorf("cannot get the location of `self`: %v", err)
	}

	initProcess := exec.Command(initCmd, "init")
	initProcess.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}

	if interactive {
		initProcess.Stdin = os.Stdin
		initProcess.Stdout = os.Stdout
		initProcess.Stderr = os.Stderr
	} else {
		dataDir := path.Join(config.BoxDataPath, boxName)
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return fmt.Errorf("cannot create data dir `%s`: %v", dataDir, err)
		}
		logFilePath := path.Join(dataDir, config.LogFileName)
		logFile, err := os.Create(logFilePath)
		if err != nil {
			return fmt.Errorf("cannot create log file `%s`: %v", logFilePath, err)
		}
		initProcess.Stdout = logFile
	}

	initProcess.ExtraFiles = []*os.File{readPipe}
	initProcess.Env = append(os.Environ(), envSlice...)
	if err := fileSystem.NewWorkSpace(volume, imageName, boxName); err != nil {
		return fmt.Errorf("cannot create new workspace: %v", err)
	}
	initProcess.Dir = path.Join(config.BoxDataPath, boxName, config.MountPath)

	if err := initProcess.Start(); err != nil {
		return fmt.Errorf("cannot start init process: %v", err)
	}

	//record box info
	boxName, err = recordBoxInfo(initProcess.Process.Pid, cmdArray, boxName, boxID, volume)
	if err != nil {
		return fmt.Errorf("cannot record box info %v", err)
	}

	// use boxID as cgroup name
	log.Printf("creating cgroup for %v\n", boxID)
	cgroupManager := cgroup.NewCgroupManager(boxID)

	defer func() {
		if normalExit {
			return
		}
		log.Println("Removing cgroup...")
		if err := cgroupManager.Destroy(); err != nil {
			panic(err)
		}
	}()

	if err := cgroupManager.Set(resConf); err != nil {
		return fmt.Errorf("cgroup manager `set` failed with: %v", err)
	}

	defer func() {
		if normalExit {
			return
		}
		log.Println("Killing init process...")
		if err := syscall.Kill(initProcess.Process.Pid, syscall.SIGTERM); err != nil {
			// for init processes that don't return 0 as return value
			if err.Error() == "no such process" {
				log.Println("Init process is already dead!")
			} else {
				panic(err)
			}
		}
		log.Println("Removing workspace...")
		if err := fileSystem.DeleteWorkSpace(volume, boxName); err != nil {
			panic(err)
		}
	}()

	if err := cgroupManager.Apply(initProcess.Process.Pid); err != nil {
		return fmt.Errorf("cgroup manager `apply` failed with: %v", err)
	}

	if networkName != "" {
		// config box network
		if err := network.Init(); err != nil {
			return err
		}
		BoxInfo := &internal.BoxInfo{
			Id:          boxID,
			Pid:         strconv.Itoa(initProcess.Process.Pid),
			Name:        boxName,
			PortMapping: portMapping,
		}
		if err := network.Connect(networkName, BoxInfo); err != nil {
			return fmt.Errorf("cannot connect network: %v", err)
		}
	}

	if err := sendInitCommand(cmdArray, writePipe); err != nil {
		return err
	}

	if interactive {
		if err := initProcess.Wait(); err != nil {
			return fmt.Errorf("error waiting init process: %v", err)
		}

		if err := fileSystem.DeleteWorkSpace(volume, boxName); err != nil {
			return fmt.Errorf("cannot delete workspace: %v", err)
		}

		if err := deleteBoxInfo(boxName); err != nil {
			return fmt.Errorf("cannot delete box info dir: %v", err)
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

func recordBoxInfo(boxPID int, commandArray []string, boxName, id, volume string) (string, error) {
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")
	BoxInfo := &internal.BoxInfo{
		Id:          id,
		Pid:         strconv.Itoa(boxPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      internal.Running,
		Name:        boxName,
		Volume:      volume,
	}

	jsonBytes, err := json.Marshal(BoxInfo)
	if err != nil {
		return "", err
	}
	jsonStr := string(jsonBytes)

	dataDir := path.Join(config.BoxDataPath, boxName)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", err
	}
	infoFileName := path.Join(dataDir, config.InfoFileName)
	infoFile, err := os.Create(infoFileName)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := infoFile.Close(); err != nil {
			panic(err)
		}
	}()
	if _, err := infoFile.WriteString(jsonStr); err != nil {
		return "", err
	}

	return boxName, nil
}

func deleteBoxInfo(boxName string) error {
	dataDir := path.Join(config.BoxDataPath, boxName)
	if err := os.RemoveAll(dataDir); err != nil {
		return fmt.Errorf("cannot remove dir %s : %v", dataDir, err)
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
