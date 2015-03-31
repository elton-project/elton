package main

import (
	"log"

	"github.com/codegangsta/cli"

	"git.t-lab.cs.teu.ac.jp/nashio/elton/config"
	"git.t-lab.cs.teu.ac.jp/nashio/elton/http"
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
		cli.IntFlag{
			Name:  "port, p",
			Value: 24680,
			Usage: "port number",
		},
		cli.StringFlag{
			Name:  "dir, d",
			Value: "./",
			Usage: "target directory",
		},
		cli.IntFlag{
			Name:  "weight",
			Value: 1,
			Usage: "weight",
		},
	},
}

func doProxy(c *cli.Context) {
	conf, err := config.Load(c.String("file"))
	if err != nil {
		log.Fatal(err)
	}

	proxy, err := http.NewProxy(conf)
	if err != nil {
		log.Fatal(err)
	}

	proxy.Serve()
}

func doServer(c *cli.Context) {
	server := http.NewServer(c.String("port"), c.String("dir"), c.Int("weight"))
	server.Serve()
}
