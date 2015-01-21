package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "elton"
	app.Version = Version
	app.Usage = ""
	app.Author = "Taku MIZUNO"
	app.Email = "dev@nashio-lab.info"
	app.Commands = Commands

	app.Run(os.Args)
}
