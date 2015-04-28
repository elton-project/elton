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
			Value: "server_config.tml",
			Usage: "config file",
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
	conf, err := elton.Load(c.String("file"))
	if err != nil {
		log.Fatal(err)
	}

	server := http.NewServer(conf)
	server.Serve()
}
