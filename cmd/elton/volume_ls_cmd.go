package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"golang.org/x/xerrors"
	"io"
	"sort"
)

func volumeLsFn(cmd *cobra.Command, args []string) error {
	if err := _volumeLsFn(); err != nil {
		showError(err)
	}
	return nil
}
func _volumeLsFn() error {
	c, err := elton_v2.VolumeService()
	if err != nil {
		return xerrors.Errorf("api client: %w", err)
	}
	defer elton_v2.Close(c)

	// Call api and store results to names slice.
	var names []string
	req := &elton_v2.ListVolumesRequest{Limit: 1000}
	receiver, err := c.ListVolumes(context.Background(), req)
	if err != nil {
		return xerrors.Errorf("api client: %w", err)
	}
	for {
		res, err := receiver.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return xerrors.Errorf("api client: %w", err)
		}
		names = append(names, res.GetInfo().GetName())
	}

	// Print volume names to stdout.
	sort.Strings(names)
	for _, name := range names {
		fmt.Println(name)
	}
	return nil
}
