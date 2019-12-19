package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:              "elton",
	TraverseChildren: true,
}
var volumeCmd = &cobra.Command{
	Use: "volume",
}
var volumeLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all volumes",
	RunE:  volumeLsFn,
}
var volumeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a volume",
	RunE:  volumeCreateFn,
}

func init() {
	volumeCmd.AddCommand(volumeLsCmd)
	rootCmd.AddCommand(volumeCmd)
}
func main() {
	os.Exit(Main())
}
func Main() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		return 1
	}
	return 0
}
func showError(err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %+v\n", err.Error())
}
