package main

import (
	"fmt"
	"os"
)

func Main() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}
func main() {
	os.Exit(Main())
}
