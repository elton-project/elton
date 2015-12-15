package main

import (
	"os"
	"runtime"

	"github.com/codegangsta/cli"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	app := cli.NewApp()
	app.Name = "elton"
	app.Usage = ""
	app.Author = "Taku MIZUNO"
	app.Email = "dev@nashio-lab.info"
	app.Commands = Commands
	app.Version = Version

	app.Run(os.Args)
}
