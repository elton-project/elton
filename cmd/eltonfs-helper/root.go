package main

import (
	"context"
	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	RpcServerRetryInterval = 100 * time.Millisecond
	RpcServerRetryTimeout  = 3 * time.Second
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

	conn, err := tryConnect("unix", sock, RpcServerRetryInterval, RpcServerRetryTimeout, ctx)
	if err != nil {
		return xerrors.Errorf("connect: %w", err)
	}

	if err := handle(ctx, conn); err != nil {
		return xerrors.Errorf("handle: %w", err)
	}
	return nil
}

// tryConnect creates a connection to eltonfs in-kernel RPC server.
// The net.Dial() may be fail depending on the initialization timing.  It retries net.Dial() to create a connection certainly.
func tryConnect(network, addr string, interval time.Duration, timeout time.Duration, ctx context.Context) (net.Conn, error) {
	var lastErr error
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	timeoutCtx, _ := context.WithTimeout(ctx, timeout)

	for i := 0; ; i++ {
		conn, err := net.Dial(network, addr)
		if err == nil {
			return conn, nil
		}
		lastErr = xerrors.Errorf("tryConnect: %w", err)
		log.Printf("failed to connect RPC server.  Retry after %s", interval.String())

		select {
		case <-ticker.C:
		case <-timeoutCtx.Done():
			return nil, lastErr
		}
	}
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
