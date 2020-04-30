package main

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/yqszxx/oreo-box/cgroups/subsystems"
	"github.com/yqszxx/oreo-box/container"
	"github.com/yqszxx/oreo-box/network"
	"log"
	"os"
)

var runCommand = cli.Command{
	Name:  "run",
	Usage: `Create a container with namespace and cgroups limit ie: mydocker run -ti [image] [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "ti",
			Usage: "enable tty",
		},
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
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
			Usage: "container name",
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
			Usage: "container network",
		},
		cli.StringSliceFlag{
			Name:  "p",
			Usage: "port mapping",
		},
	},
	Action: func(context *cli.Context) error {
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

		createTty := context.Bool("ti")
		detach := context.Bool("d")

		if createTty && detach {
			return fmt.Errorf("ti and d paramter both provided")
		}
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSetMems:  context.String("cpusetmems"),
			CpuSetCpus:  context.String("cpusetcpus"),
			CpuShare:    context.String("cpushare"),
			CpuQuotaUs:  context.String("cpuquota"),
		}

		containerName := context.String("name")
		volume := context.String("v")
		network := context.String("net")

		envSlice := context.StringSlice("e")
		portmapping := context.StringSlice("p")

		return Run(createTty, cmdArray, resConf, containerName, volume, imageName, envSlice, network, portmapping)
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside",
	Action: func(context *cli.Context) error {
		log.Println("Starting init process...")
		return container.RunContainerInitProcess()
	},
}

var listCommand = cli.Command{
	Name:  "ps",
	Usage: "list all the containers",
	Action: func(context *cli.Context) error {
		if err := ListContainers(); err != nil {
			return fmt.Errorf("list failed because: %v", err)
		}
		return nil
	},
}

var logCommand = cli.Command{
	Name:  "logs",
	Usage: "print logs of a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("cannot find container name")
		}
		containerName := context.Args().Get(0)
		if err := logContainer(containerName); err != nil {
			return fmt.Errorf("log failed because: %v", err)
		}
		return nil
	},
}

var execCommand = cli.Command{
	Name:  "exec",
	Usage: "exec a command into container",
	Action: func(context *cli.Context) error {
		//This is for callback
		if os.Getenv(ENV_EXEC_PID) != "" && os.Getenv(ENV_EXEC_CMD) != "" {
			log.Println("Cgo code executed")
			return nil
		}

		if len(context.Args()) < 2 {
			return fmt.Errorf("cannot find container name or command")
		}
		containerName := context.Args().Get(0)
		var commandArray []string
		for _, arg := range context.Args().Tail() {
			commandArray = append(commandArray, arg)
		}
		if err := ExecContainer(containerName, commandArray); err != nil {
			return fmt.Errorf("exec failed because: %v", err)
		}
		return nil
	},
}

var stopCommand = cli.Command{
	Name:  "stop",
	Usage: "stop a container",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("cannot find container name")
		}
		containerName := context.Args().Get(0)
		if err := stopContainer(containerName); err != nil {
			return err
		}
		return nil
	},
}

var removeCommand = cli.Command{
	Name:  "rm",
	Usage: "remove unused containers",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("cannot find container name")
		}
		containerName := context.Args().Get(0)
		if err := removeContainer(containerName); err != nil {
			return err
		}
		return nil
	},
}

var commitCommand = cli.Command{
	Name:  "commit",
	Usage: "commit a container into image",
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 2 {
			return fmt.Errorf("cannot find container name and image name")
		}
		containerName := context.Args().Get(0)
		imageName := context.Args().Get(1)
		if err := commitContainer(containerName, imageName); err != nil {
			return fmt.Errorf("exec commit failed becuase: %v", err)
		}
		return nil
	},
}

var networkCommand = cli.Command{
	Name:  "network",
	Usage: "container network commands",
	Subcommands: []cli.Command{
		{
			Name:  "create",
			Usage: "create a container network",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "driver",
					Usage: "network driver",
				},
				cli.StringFlag{
					Name:  "subnet",
					Usage: "subnet cidr",
				},
			},
			Action: func(context *cli.Context) error {
				if len(context.Args()) < 1 {
					return fmt.Errorf("cannot find network name")
				}
				if err := network.Init(); err != nil {
					return fmt.Errorf("init network failed becuase: %v", err)
				}

				err := network.CreateNetwork(context.String("driver"), context.String("subnet"), context.Args()[0])
				if err != nil {
					return fmt.Errorf("create network error: %+v", err)
				}
				return nil
			},
		},
		{
			Name:  "list",
			Usage: "list container network",
			Action: func(context *cli.Context) error {
				if err := network.Init(); err != nil {
					return fmt.Errorf("init network failed becuase: %v", err)
				}
				network.ListNetwork()
				return nil
			},
		},
		{
			Name:  "remove",
			Usage: "remove container network",
			Action: func(context *cli.Context) error {
				if len(context.Args()) < 1 {
					return fmt.Errorf("cannot find network name")
				}
				if err := network.Init(); err != nil {
					return fmt.Errorf("init network failed becuase: %v", err)
				}
				err := network.DeleteNetwork(context.Args()[0])
				if err != nil {
					return fmt.Errorf("remove network error: %+v", err)
				}
				return nil
			},
		},
	},
}
