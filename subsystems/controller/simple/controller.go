package simple

import (
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
)

func NewController() *Controller {
	return &Controller{
		MetaServiceServer:   newLocalMetaServer(),
		NodeServiceServer:   newLocalNodeServer(),
		VolumeServiceServer: newLocalVolumeServer(),
		CommitServiceServer: &UnimplementedCommitServiceServer{},
	}
}

type Controller struct {
	MetaServiceServer
	NodeServiceServer
	VolumeServiceServer
	CommitServiceServer
}
