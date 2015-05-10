package main

import (
	"log"

	"github.com/codegangsta/cli"

	"./elton"
	http "./http"
)

var Commands = []cli.Command{
	commandClient,
	commandServer,
}

var commandClient = cli.Command{
	Name:        "client",
	Usage:       "",
	Description: ``,
	Action:      doClient,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Value: "proxy_config.tml",
			Usage: "config file",
		},
	},
}

var commandServer = cli.Command{
	Name:        "server",
	Usage:       "",
	Description: ``,
	Action:      doServer,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Value: "config.tml",
			Usage: "config file",
		},
		cli.BoolFlag{
			Name:  "backup",
			Usage: "Backup flag",
		},
	},
}

func doClient(c *cli.Context) {
}

func doServer(c *cli.Context) {
	log.SetPrefix("[elton server] ")
	conf, err := elton.Load(c.String("file"))
	backup := c.Bool("backup")
	if err != nil {
		log.Fatal(err)
	}

	server, err := http.NewEltonServer(conf, backup)
	if err != nil {
		log.Fatal(err)
	}
	server.Serve()
}
