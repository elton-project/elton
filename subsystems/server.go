package subsystems

import "context"

type Server interface {
	Name() string
	Configure() error
	Listen() error
	Serve(ctx context.Context) error
}
