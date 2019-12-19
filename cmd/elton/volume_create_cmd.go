package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"golang.org/x/xerrors"
)

func volumeCreateFn(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := _volumeCreateFn(ctx, args); err != nil {
		showError(err)
	}
	return nil
}

func _volumeCreateFn(ctx context.Context, names []string) error {
	c, err := elton_v2.ApiClient{}.VolumeService()
	if err != nil {
		return xerrors.Errorf("api client: %w", err)
	}

	for _, name := range names {
		// Create a volume with specified name.
		req := &elton_v2.CreateVolumeRequest{
			Info: &elton_v2.VolumeInfo{
				Name: name,
			},
		}
		res, err := c.CreateVolume(ctx, req)
		if err != nil {
			return xerrors.Errorf("api client: %w", err)
		}
		// Show volume ID.
		fmt.Println(res.GetId())
	}
	return nil
}
