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
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug utilities",
}
var debugDumpObjCmd = &cobra.Command{
	Use:   "dump-obj",
	Short: "Dump objects with human-readable string",
	RunE:  debugDumpObjFn,
}

func init() {
	volumeCmd.AddCommand(volumeLsCmd, volumeCreateCmd)
	debugCmd.AddCommand(debugDumpObjCmd)
	rootCmd.AddCommand(volumeCmd, debugCmd)
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
