package proto2

import (
	"github.com/rs/xid"
	"net"
)

func NewServerInfo(addr net.Addr) *ServerInfo {
	return &ServerInfo{
		Guid:    xid.New().String(),
		Group:   GetGroups(),
		Address: addr.String(),
	}
}
