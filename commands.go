package main

import (
	"log"

	"github.com/codegangsta/cli"

	"git.t-lab.cs.teu.ac.jp/nashio/elton/config"
	elton "git.t-lab.cs.teu.ac.jp/nashio/elton/http"
)

var Commands = []cli.Command{
	commandProxy,
	commandServer,
}

var commandProxy = cli.Command{
	Name:        "proxy",
	Usage:       "",
	Description: ``,
	Action:      doProxy,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "file, f",
			Value: "config.tml",
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
	},
}

func doProxy(c *cli.Context) {
	conf, err := config.Load(c.String("file"))
	if err != nil {
		log.Fatal(err)
	}

	proxy := elton.NewProxy(conf)
	proxy.Serve()
}

func doServer(c *cli.Context) {
	conf, err := config.Load(c.String("file"))
	if err != nil {
		log.Fatal(err)
	}

	server := elton.NewServer(conf)
	server.Serve()
}
