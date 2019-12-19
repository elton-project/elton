package eltonfs_rpc

import (
	"context"
	"golang.org/x/xerrors"
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
	case CreateObjectRequestStructID:
		handleCreateObject(ns)
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

		// Get commit info from meta node.
		c, err := ApiClient{}.CommitService()
		if err != nil {
			return nil, xerrors.Errorf("api client: %w", err)
		}
		res, err := c.GetCommit(context.Background(), req.ToGRPC())
		if err != nil {
			return nil, xerrors.Errorf("call api: %w", err)
		}

		return GetCommitInfoResponse{}.FromGRPC(res), nil
	})
}

func handleGetObjectRequest(ns ClientNS) {
	rpcHandlerHelper(ns, &GetObjectRequest{}, func(rawReq interface{}) (interface{}, error) {
		req := rawReq.(*GetObjectRequest)

		// Get object from storage.
		c, err := ApiClient{}.StorageService()
		if err != nil {
			return nil, xerrors.Errorf("api client: %w", err)
		}
		res, err := c.GetObject(context.Background(), req.ToGRPC())
		if err != nil {
			return nil, xerrors.Errorf("call api: %w", err)
		}

		return GetObjectResponse{}.FromGRPC(res), nil
	})
}

func handleCreateObject(ns ClientNS) {
	rpcHandlerHelper(ns, &CreateObjectRequest{}, func(rawReq interface{}) (i interface{}, e error) {
		req := rawReq.(*CreateObjectRequest)

		// Send create object request.
		c, err := ApiClient{}.StorageService()
		if err != nil {
			return nil, xerrors.Errorf("api client: %w", err)
		}
		res, err := c.CreateObject(context.Background(), req.ToGRPC())
		if err != nil {
			return nil, xerrors.Errorf("call api: %w", err)
		}
		return CreateObjectResponse{}.FromGRPC(res), nil
	})
}

func handleCreateCommitRequest(ns ClientNS) {
	rpcHandlerHelper(ns, &CreateCommitRequest{}, func(rawReq interface{}) (interface{}, error) {
		req := rawReq.(*CreateCommitRequest)

		// Send commit request.
		c, err := ApiClient{}.CommitService()
		if err != nil {
			return nil, xerrors.Errorf("api client: %w", err)
		}
		res, err := c.Commit(context.Background(), req.ToGRPC())
		if err != nil {
			return nil, xerrors.Errorf("call api: %w", err)
		}

		return CreateCommitResponse{}.FromGRPC(res), nil
	})
}
