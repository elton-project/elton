package main

import (
	"log"

	"github.com/codegangsta/cli"

	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc"
	elton "gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/server"
)

var Commands = []cli.Command{
	commandClient,
	commandServer,
}

var commandClient = cli.Command{
	Name:        "master",
	Usage:       "",
	Description: ``,
	Action:      doMaster,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Value: "config.tml",
			Usage: "config file",
		},
	},
}

var commandServer = cli.Command{
	Name:        "slave",
	Usage:       "",
	Description: ``,
	Action:      doSlave,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Value: "config.tml",
			Usage: "config file",
		},
		cli.BoolFlag{
			Name:  "backup",
			Usage: "backup flag",
		},
	},
}

func doMaster(c *cli.Context) {
	log.SetPrefix("[elton master] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	conf, err := elton.Load(c.String("file"))
	if err != nil {
		log.Fatal(err)
	}

	server, err := grpc.NewEltonMaster(conf)
	if err != nil {
		log.Fatal(err)
	}

	server.Serve()
}

func doSlave(c *cli.Context) {
	log.SetPrefix("[elton slave] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	conf, err := elton.Load(c.String("file"))
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewEltonSlave(conf, c.Bool("backup"))
	defer server.FS.PurgeTimer.Stop()

	server.Serve()
}
