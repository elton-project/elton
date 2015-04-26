package main

import (
	"log"

	"github.com/codegangsta/cli"

	"./elton"
	http "./http"
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
	},
}

func doProxy(c *cli.Context) {
	log.SetPrefix("[elton proxy] ")
	conf, err := elton.Load(c.String("file"))
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
	log.SetPrefix("[elton server] ")
	server := http.NewServer(c.String("port"), c.String("dir"))
	server.Serve()
}
