package cmd

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/yqszxx/oreo-box/internal/image"
)

var imageCommand = cli.Command{
	Name:  "image",
	Usage: "Manage images",
	Subcommands: []cli.Command{
		{
			Name:  "import",
			Usage: "Import an image",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "f",
					Usage: "force import",
				},
			},
			Action: func(context *cli.Context) error {
				if len(context.Args()) < 2 {
					return fmt.Errorf("no enough arguments provided")
				}

				return image.Import(context.Args().Get(0), context.Args().Get(1))
			},
		},
		{
			Name:  "list",
			Usage: "List images",
			Action: func(context *cli.Context) error {
				return image.List()
			},
		},
		{
			Name:  "remove",
			Usage: "Remove an image",
			Action: func(context *cli.Context) error {
				if len(context.Args()) < 1 {
					return fmt.Errorf("no image name provided")
				}

				//TODO: need to check whether image is being used

				return fmt.Errorf("not implemented")
			},
		},
	},
}
