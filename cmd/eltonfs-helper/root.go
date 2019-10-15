package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "eltonfs-helper --socket <SOCKET>",
	Short: "User mode helper process for the eltonfs",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("ok")
		return nil
	},
}

func init() {
	rootCmd.Flags().String("socket", "", "Path to socket file")
}
