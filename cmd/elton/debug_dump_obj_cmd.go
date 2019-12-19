package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	localStorage "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems/storage/local"
	"golang.org/x/xerrors"
	"os"
)

func debugDumpObjFn(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if len(args) == 0 {
		args = []string{"-"}
	}

	if err := _debugDumpObjFn(ctx, args); err != nil {
		showError(err)
	}
	return nil
}

func _debugDumpObjFn(ctx context.Context, files []string) error {
	for i, file := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		s, err := dumpObj(file)
		if err != nil {
			return err
		}

		if len(files) > 1 {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("==> %s <==", file)
		}
		fmt.Print(s)
	}
	return nil
}
func dumpObj(file string) (string, error) {
	var f *os.File
	if file == "-" {
		f = os.Stdin
	} else {
		var err error
		f, err = os.Open(file)
		if err != nil {
			return "", xerrors.Errorf("open: %w", err)
		}
		defer f.Close()
	}

	s, err := localStorage.DumpHeader(f)
	if err != nil {
		return "", xerrors.Errorf("dump obj: %w", err)
	}
	return s, nil
}
