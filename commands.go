package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"

	"./elton"
	elhttp "./http"
)

var Commands = []cli.Command{
	commandClient,
	commandServer,
}

var commandClient = cli.Command{
	Name:        "client",
	Usage:       "",
	Description: ``,
	Action:      doClient,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "R, r",
			Usage: "recursive",
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
		cli.BoolFlag{
			Name:  "backup",
			Usage: "Backup flag",
		},
	},
}

func doClient(c *cli.Context) {
	log.SetPrefix("[elton client] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	r := c.Bool("R")
	args := c.Args()

	if r {
		if err := filepath.Walk(args[0], func(p string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			return uploadFile(p, path.Join(args[1], p))
		}); err != nil {
			log.Println(err)
		}
		return
	}

	info, _ := os.Stat(args[0])
	if info.IsDir() {
		log.Printf("expected -R option")
		return
	}

	if err := uploadFile(args[0], args[1]); err != nil {
		log.Println(err)
	}
}

func doServer(c *cli.Context) {
	log.SetPrefix("[elton server] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	conf, err := elton.Load(c.String("file"))
	if err != nil {
		log.Fatal(err)
	}

	server, err := elhttp.NewEltonServer(conf)
	if err != nil {
		log.Fatal(err)
	}
	server.Serve()
}

func uploadFile(name string, url string) error {
	// file, err := os.Open(name)
	// if err != nil {
	// 	return err
	// }
	// defer file.Close()

	// body := new(bytes.Buffer)
	// writer := multipart.NewWriter(body)

	// part, err := writer.CreateFormFile("file", name)
	// if err != nil {
	// 	writer.Close()
	// 	return err
	// }

	// if _, err = io.Copy(part, file); err != nil {
	// 	writer.Close()
	// 	return err
	// }
	// writer.Close()

	// args := strings.Split(url, "/")
	// req, _ := http.NewRequest("PUT", "http://"+path.Join(args[0], "elton", strings.Join(args[1:], "/")), body)
	// req.Header.Add("Content-Type", writer.FormDataContentType())

	// client := new(http.Client)
	// res, err := client.Do(req)
	// if err != nil {
	// 	return err
	// } else if res.StatusCode != http.StatusOK {
	// 	return fmt.Errorf("http status: %v", res.StatusCode)
	// }
	// defer res.Body.Close()

	// return nil
	args := strings.Split(url, "/")
	return fmt.Errorf("%s: %s", name, "http://"+path.Join(args[0], "elton", strings.Join(args[1:], "/")))
}
