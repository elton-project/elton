package main

import (
	"fmt"
	"os"
)

func Main() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("%+v\n", err)
		return 1
	}
	return 0
}
func main() {
	os.Exit(Main())
}
