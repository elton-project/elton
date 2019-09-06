package simple

import . "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"

func newLocalNodeServer() *localNodeServer {
	return &localNodeServer{}
}

type localNodeServer struct {
	UnimplementedNodeServiceServer
}
