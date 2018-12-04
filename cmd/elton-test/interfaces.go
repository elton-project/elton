package main

import (
	"context"
	"net"
)

type Subsystem interface {
	Name() string
	Setup(ctx context.Context) error
	Serve(ctx context.Context) []error
}

type Service interface {
	Name() string
	SetAddr(addr string)
	Register(ctx context.Context) error
	Unregister(ctx context.Context) error
	Serve(ctx context.Context, listener net.Listener) error
}
