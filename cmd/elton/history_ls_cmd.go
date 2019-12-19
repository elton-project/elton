package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"golang.org/x/xerrors"
	"io"
)

func historyLsFn(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("")
	}

	volume := args[0]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := _historyLsFn(ctx, volume); err != nil {
		showError(err)
	}
	return nil
}
func _historyLsFn(ctx context.Context, volumeName string) error {
	// Get volume ID.
	cv, err := elton_v2.ApiClient{}.VolumeService()
	if err != nil {
		return xerrors.Errorf("api client: %w", err)
	}
	vRes, err := cv.InspectVolume(ctx, &elton_v2.InspectVolumeRequest{
		Name: volumeName,
	})
	if err != nil {
		return xerrors.Errorf("inspect volume: %w", err)
	}
	volID := vRes.GetId()

	// Get commit list.
	cc, err := elton_v2.ApiClient{}.CommitService()
	if err != nil {
		return xerrors.Errorf("api client: %w", err)
	}
	receiver, err := cc.ListCommits(ctx, &elton_v2.ListCommitsRequest{
		Id: volID,
	})
	if err != nil {
		return xerrors.Errorf("list commits: %w", err)
	}

	// Print commit ID list.
	for {
		cRes, err := receiver.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return xerrors.Errorf("api client: %w", err)
		}

		fmt.Println(cRes.GetId())
	}
	return nil
}
