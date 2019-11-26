package main

import (
	"log"
	"os"
)

func Main() (exitCode int) {
	log.Println("starting eltonfs-helper")
	defer log.Printf("stopping eltonfs-helper with exit-code(%d)", exitCode)

	if err := rootCmd.Execute(); err != nil {
		log.Printf("%+v\n", err)
		return 1
	}
	return 0
}
func main() {
	os.Exit(Main())
}
