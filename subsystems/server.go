package subsystems

import (
	"context"
	"net"
)

type Server interface {
	Name() string
	Configure() error
	Listen() error
	SetListener(l net.Listener)
	Serve(ctx context.Context) error
}
