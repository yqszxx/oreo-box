package cmd

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/yqszxx/oreo-box/internal/network"
)

var networkCommand = cli.Command{
	Name:  "network",
	Usage: "box network commands",
	Subcommands: []cli.Command{
		{
			Name:  "create",
			Usage: "create a box network",
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
					return fmt.Errorf("no enough arguments provided")
				}
				if err := network.Init(); err != nil {
					return fmt.Errorf("cannot init network controller: %v", err)
				}

				err := network.CreateNetwork(context.String("driver"), context.String("subnet"), context.Args()[0])
				if err != nil {
					return fmt.Errorf("cannot create network: %v", err)
				}
				fmt.Println(context.Args()[0])
				return nil
			},
		},
		{
			Name:  "list",
			Usage: "list box network",
			Action: func(context *cli.Context) error {
				if err := network.Init(); err != nil {
					return fmt.Errorf("cannot init network controller: %v", err)
				}
				if err := network.ListNetwork(); err != nil {
					return fmt.Errorf("cannot list networks: %v", err)
				}
				return nil
			},
		},
		{
			Name:  "remove",
			Usage: "remove box network",
			Action: func(context *cli.Context) error {
				if len(context.Args()) < 1 {
					return fmt.Errorf("no network name provided")
				}
				if err := network.Init(); err != nil {
					return fmt.Errorf("cannot init network controller: %v", err)
				}
				err := network.DeleteNetwork(context.Args()[0])
				if err != nil {
					return fmt.Errorf("cannot remove network: %v", err)
				}
				return nil
			},
		},
	},
}
