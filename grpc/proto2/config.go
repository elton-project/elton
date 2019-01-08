package proto2

import (
	"net"
	"time"
)

type Config struct {
	Controller      *ControllerConfig
	RPCTimeout      time.Duration
	ShutdownTimeout time.Duration
}

type ControllerConfig struct {
	// controllerがlistenするアドレス。
	// controllerを起動しないノードでは、nilになる。
	ListenAddr net.Addr
	// 初期ノードのアドレス。
	// 他のノードのアドレスは、初期ノードから取得してくるr。
	Servers []net.Addr
}
