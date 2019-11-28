package eltonfs_rpc

import (
	"golang.org/x/xerrors"
	"log"
)

type RpcHandler func(ClientNS, StructID, PacketFlag)

func defaultHandler(ns ClientNS, sid StructID, flags PacketFlag) {
	switch sid {
	case PingStructID:
		handlePing(ns)
	default:
		log.Println(xerrors.Errorf("not implemented handler: struct_id=%d", sid))
	}
}

func handlePing(ns ClientNS) {
	_, err := ns.Recv(&Ping{})
	if err != nil {
		log.Println(xerrors.Errorf("handlePing: recv ping: %w", err))
		return
	}

	if ns.IsSendable() {
		if err := ns.Send(&Ping{}); err != nil {
			log.Println(xerrors.Errorf("handlePing: send reply: %w", err))
			return
		}
	}

	if err := ns.Close(); err != nil {
		log.Println(xerrors.Errorf("handlePing: close: %w", err))
		return
	}
}
