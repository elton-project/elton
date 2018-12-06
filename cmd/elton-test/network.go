package main

import (
	"go.uber.org/zap"
	"net"
)

// Get preferred ip of this machine.
// The target argument is used to select preferred network interfaces.
// If target argument is nil, use outbound IP address (8.8.8.8) instead.
func GetPreferredIP(target net.IP) net.IP {
	if target == nil {
		target = net.IPv4(8, 8, 8, 8)
	}

	// This is the best solution to get preferred ip when the machine has multiple ip interfaces.
	// See https://stackoverflow.com/a/37382208
	conn, err := net.Dial("udp", target.String()+":80")
	if err != nil {
		zap.S().Panicw("GetPreferredIP", "error", err)
		panic(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}
