package main

import (
	"os"

	"git.t-lab.cs.teu.ac.jp/nashio/elton/Godeps/_workspace/src/github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "elton"
	app.Usage = ""
	app.Author = "Taku MIZUNO"
	app.Email = "dev@nashio-lab.info"
	app.Commands = Commands
	app.Version = Version

	app.Run(os.Args)
}
