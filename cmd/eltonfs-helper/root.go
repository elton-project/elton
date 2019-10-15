package main

import (
	"context"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var rootCmd = &cobra.Command{
	Use:           "eltonfs-helper --socket <SOCKET>",
	Short:         "User mode helper process for the eltonfs",
	RunE:          rootFn,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.Flags().String("socket", "", "Path to socket file")
}

func rootFn(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	go cancelWhenSignal(ctx, cancel)
	defer cancel()

	sock, err := cmd.Flags().GetString("socket")
	if err != nil {
		return xerrors.Errorf("socket args: %w", err)
	}

	conn, err := net.Dial("unix", sock)
	if err != nil {
		return xerrors.Errorf("connect: %w", err)
	}

	if err := handle(ctx, conn); err != nil {
		return xerrors.Errorf("handle: %w", err)
	}
	return nil
}

func cancelWhenSignal(ctx context.Context, cancel func()) {
	sig := make(chan os.Signal, 10)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sig:
		cancel()
	case <-ctx.Done():
	}
}
