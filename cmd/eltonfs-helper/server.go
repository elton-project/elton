package main

import (
	"context"
	"encoding/binary"
	"net"
)

func handle(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	body := []byte("hello")
	size := make([]byte, 8)
	binary.BigEndian.PutUint64(size, uint64(len(body)))

	conn.Write(size)
	conn.Write(body)
	// TODO: send/recv data
	return nil
}
