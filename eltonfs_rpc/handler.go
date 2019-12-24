package eltonfs_rpc

import (
	"context"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	"golang.org/x/xerrors"
	"io"
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
	case NotifyLatestCommitRequestID:
		handleNotifyLatestCommitRequest(ns)
	case GetVolumeIDRequestID:
		handleGetVolumeIDRequest(ns)
	default:
		err := xerrors.Errorf("not implemented handler: struct_id=%d", sid)
		log.Println(err)
		ns.CloseWithError(SessionError{
			ErrorID: UnsupportedStruct,
			Reason:  err.Error(),
		})
	}
}

func rpcHandlerHelper(ns ClientNS, reqType interface{}, fn func(rawReq interface{}) (interface{}, error)) {
	// TODO: ns leak when an error occurred.
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
	// WORKAROUND: prevent dead-lock if panics during executing handler.
	//
	// panicや何らかのエラーが生じた後にns.Close()すると、closeパケットを送信してから返答を待機する。
	// しかし、現在の実装では返答がkmodから返されないため、デッドロック状態に陥ってしまう。
	// このため、処理中にpanicしても該当goroutineがロック状態になるだけであり、意図せずプロセスが動き続けてしまう。
	// https://gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/issues/153#note_3754
	if err := ns.Close(); err != nil {
		log.Println(xerrors.Errorf("close error: %w", err))
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
		c, err := elton_v2.ApiClient{}.CommitService()
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
		c, err := elton_v2.ApiClient{}.StorageService()
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
		c, err := elton_v2.ApiClient{}.StorageService()
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
		c, err := elton_v2.ApiClient{}.CommitService()
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

func handleNotifyLatestCommitRequest(ns ClientNS) {
	rpcHandlerHelper(ns, &NotifyLatestCommitRequest{}, func(rawReq interface{}) (i interface{}, err error) {
		req := rawReq.(*NotifyLatestCommitRequest)

		c, err := elton_v2.ApiClient{}.CommitService()
		if err != nil {
			return nil, xerrors.Errorf("api client: %w", err)
		}
		receiver, err := c.ListCommits(context.Background(), &elton_v2.ListCommitsRequest{
			Id:    req.VolumeID.ToGRC(),
			Limit: 1,
		})
		res, err := receiver.Recv()
		if err != nil {
			return nil, xerrors.Errorf("receiver: unexpected error: %w", err)
		}
		out := NotifyLatestCommit{}.FromGRPC(res.GetId())
		if _, err := receiver.Recv(); err != io.EOF {
			return nil, xerrors.Errorf("receiver: not EOF: %w", err)
		}
		return out, nil
	})
}

func handleGetVolumeIDRequest(ns ClientNS) {
	rpcHandlerHelper(ns, &GetVolumeIDRequest{}, func(rawReq interface{}) (i interface{}, err error) {
		req := rawReq.(*GetVolumeIDRequest)

		c, err := elton_v2.ApiClient{}.VolumeService()
		if err != nil {
			return nil, xerrors.Errorf("api client: %w", err)
		}
		res, err := c.InspectVolume(context.Background(), &elton_v2.InspectVolumeRequest{
			Name: req.VolumeName,
		})
		if err != nil {
			return nil, xerrors.Errorf("call api: %w", err)
		}
		return res.GetId(), nil
	})
}
