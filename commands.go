package main

import (
	"github.com/bmizerany/pat"
	"github.com/codegangsta/cli"
	"github.com/fukata/golang-stats-api-handler"
	"nashio-lab.info/elton/fs"
	"net/http"
)

var Commands = []cli.Command{
	commandClient,
	commandServer,
	commandProxy,
}

var commandClient = cli.Command{
	Name:  "client",
	Usage: "",
	Description: `
`,
	Action: doClient,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "lang, l",
			Value:  "english",
			Usage:  "language for the greeting",
			EnvVar: "LEGACY_COMPAT_LANG,APP_LANG,LANG",
		},
	},
}

var commandServer = cli.Command{
	Name:  "server",
	Usage: "",
	Description: `
`,
	Action: doServer,
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "port, p",
			Value: 12345,
			Usage: "port number",
		},
		cli.StringFlag{
			Name:  "dir, d",
			Value: "./",
			Usage: "target directory",
		},
		cli.StringFlag{
			Name:  "proxy",
			Value: "http://localhost:56789",
			Usage: "proxy host",
		},
		cli.BoolFlag{
			Name:  "migration",
			Usage: "migration",
		},
	},
}

var commandProxy = cli.Command{
	Name:  "proxy",
	Usage: "",
	Description: `
`,
	Action: doProxy,
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "port, p",
			Value: 56789,
			Usage: "port number",
		},
		cli.StringFlag{
			Name:  "dbpath",
			Value: "./elton.db",
			Usage: "db paht",
		},
	},
}

func doClient(c *cli.Context) {
}

func doServer(c *cli.Context) {
	fs.ServerInitialize(c.String("dir"), c.String("proxy"))

	if c.Bool("migration") {
		fs.ServerMigration()
	}

	mux := pat.New()
	mux.Get("/:dir/:key/:version", http.HandlerFunc(fs.ServerGet))
	mux.Put("/:dir/:key/:version", http.HandlerFunc(fs.ServerPut))
	mux.Get("/api/stats", http.HandlerFunc(stats_api.Handler))
	http.Handle("/", mux)

	http.ListenAndServe(":"+c.String("port"), nil)
}

func doProxy(c *cli.Context) {
	fs.ProxyInitialize(c.String("dbpath"))
	defer fs.ProxyDestory()

	mux := pat.New()
	mux.Get("/:dir/:key/:version", http.HandlerFunc(fs.ProxyGet))
	mux.Post("/api/migration", http.HandlerFunc(fs.ProxyMigration))
	mux.Put("/:dir/:key", http.HandlerFunc(fs.ProxyPut))
	mux.Get("/api/stats", http.HandlerFunc(stats_api.Handler))
	http.Handle("/", mux)

	http.ListenAndServe(":"+c.String("port"), nil)
}
