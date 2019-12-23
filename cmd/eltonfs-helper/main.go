package main

import (
	"golang.org/x/xerrors"
	"log"
	"log/syslog"
	"os"
)

func Main() (exitCode int) {
	// Change log output direction to syslog.
	w, err := syslog.New(syslog.LOG_USER|syslog.LOG_INFO, "eltonfs-helper")
	if err != nil {
		// Write error log to stderr.
		log.Printf("%+v\n", xerrors.Errorf("initializing syslog writer: %w", err))
		return 2
	}
	defer w.Close()
	log.SetOutput(w)
	log.SetFlags(log.Llongfile)

	log.Println("starting eltonfs-helper")
	defer func() {
		log.Printf("stopping eltonfs-helper with exit-code(%d)", exitCode)
	}()

	if err := rootCmd.Execute(); err != nil {
		log.Printf("%+v\n", err)
		return 1
	}
	return 0
}
func main() {
	os.Exit(Main())
}
