package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/codegangsta/cli"
	"github.com/fukata/golang-stats-api-handler"

	elton "git.t-lab.cs.teu.ac.jp/nashio/elton/http"
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
		cli.IntFlag{
			Name:  "port, p",
			Value: 65432,
			Usage: "port number",
		},
		cli.StringFlag{
			Name:  "dir, d",
			Value: "./",
			Usage: "target directory",
		},
		cli.StringFlag{
			Name:  "proxy",
			Value: "localhost:56789",
			Usage: "proxy host",
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
			Usage: "db path",
		},
		cli.StringSliceFlag{
			Name:  "server, s",
			Value: &cli.StringSlice{},
			Usage: "server URL",
		},
	},
}

func doClient(c *cli.Context) {
	elton.InitClient(c.String("dir"), c.String("proxy"))

	mux := pat.New()
	mux.Get("/:dir/:key/:version", http.HandlerFunc(elton.ClientGetHandler))
	mux.Put("/:dir/:key", http.HandlerFunc(elton.ClientPutHandler))
	mux.Del("/:dir/:key", http.HandlerFunc(elton.ClientDeleteHandler))
	mux.Get("/api/stats", http.HandlerFunc(stats_api.Handler))
	http.Handle("/", mux)

	http.ListenAndServe(":"+c.String("port"), nil)
}

func doServer(c *cli.Context) {
	elton.InitServer(c.String("dir"))

	if c.Bool("migration") {
		elton.ServerMigration()
	}

	mux := pat.New()
	mux.Get("/:dir/:key/:version", http.HandlerFunc(elton.ServerGetHandler))
	mux.Put("/:dir/:key/:version", http.HandlerFunc(elton.ServerPutHandler))
	mux.Del("/:dir/:key", http.HandlerFunc(elton.ServerDeleteHandler))
	mux.Get("/api/stats", http.HandlerFunc(stats_api.Handler))
	mux.Post("/api/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	http.Handle("/", mux)

	http.ListenAndServe(":"+c.String("port"), nil)
}

func doProxy(c *cli.Context) {
	elton.InitProxy(c.String("dbpath"), c.StringSlice("server"))
	defer elton.DestoryProxy()

	mux := pat.New()
	mux.Get("/:dir/:key/:version", http.HandlerFunc(elton.ProxyGetHandler))
	mux.Post("/api/migration", http.HandlerFunc(elton.ProxyMigrationHandler))
	mux.Put("/:dir/:key", http.HandlerFunc(elton.ProxyPutHandler))
	mux.Del("/:dir/:key", http.HandlerFunc(elton.ProxyDeleteHandler))
	mux.Get("/api/stats", http.HandlerFunc(stats_api.Handler))
	http.Handle("/", mux)

	http.ListenAndServe(":"+c.String("port"), nil)
}
