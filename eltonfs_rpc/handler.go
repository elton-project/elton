package eltonfs_rpc

import (
	"context"
	elton_v2 "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"log"
)

type RpcHandler func(ClientNS, StructID, PacketFlag)

func defaultHandler(ns ClientNS, sid StructID, flags PacketFlag) {
	switch sid {
	case PingStructID:
		handlePing(ns)
	case GetCommitInfoRequestStructID:
		handleGetCommitInfoRequest(ns)
	case GetObjectRequestStructID:
		handleGetObjectRequest(ns)
	case CreateCommitRequestStructID:
		handleCreateCommitRequest(ns)
	default:
		log.Println(xerrors.Errorf("not implemented handler: struct_id=%d", sid))
	}
}

func rpcHandlerHelper(ns ClientNS, reqType interface{}, fn func(rawReq interface{}) (interface{}, error)) {
	defer ns.Close()

	rawReq, err := ns.Recv(reqType)
	if err != nil {
		log.Println(xerrors.Errorf("recv request: %w", err))
		return
	}

	res, err := fn(rawReq)
	if err != nil {
		se := &SessionError{
			ErrorID: Internal,
			Reason:  err.Error(),
		}
		if e := ns.SendErr(se); e != nil {
			log.Println(xerrors.Errorf("send error: %w", e))
		}
		return
	}
	if err := ns.Send(res); err != nil {
		log.Println(xerrors.Errorf("send response: %w", err))
		return
	}
}

func handlePing(ns ClientNS) {
	rpcHandlerHelper(ns, &Ping{}, func(rawReq interface{}) (interface{}, error) {
		return &Ping{}, nil
	})
}

func handleGetCommitInfoRequest(ns ClientNS) {
	rpcHandlerHelper(ns, &GetCommitInfoRequest{}, func(rawReq interface{}) (interface{}, error) {
		req := rawReq.(*GetCommitInfoRequest)

		cc, err := grpc.Dial("", nil) // todo: get meta server address
		if err != nil {
			return nil, xerrors.Errorf("dial: %w", err)
		}
		c := elton_v2.NewCommitServiceClient(cc)
		res, err := c.GetCommit(context.Background(), req.ToGRPC())
		if err != nil {
			return nil, xerrors.Errorf("call api: %w", err)
		}

		// todo: get commit info from meta node.

		return GetCommitInfoResponse{}.FromGRPC(res), nil
	})
}

func handleGetObjectRequest(ns ClientNS) {
	rpcHandlerHelper(ns, &GetObjectRequest{}, func(rawReq interface{}) (interface{}, error) {
		req := rawReq.(*GetObjectRequest)
		// todo: get object from storage.
		return &GetObjectResponse{
			ID:     req.ID,            // todo
			Offset: req.Offset,        // todo
			Body:   EltonObjectBody{}, // todo
		}, nil
	})
}

func handleCreateCommitRequest(ns ClientNS) {
	rpcHandlerHelper(ns, &CreateCommitRequest{}, func(rawReq interface{}) (interface{}, error) {
		req := rawReq.(*CreateCommitRequest)
		_ = req
		// todo: call commit api.
		return &CreateCommitResponse{
			ID: "", // todo
		}, nil
	})
}
