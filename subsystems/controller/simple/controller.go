package simple

import (
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
)

func NewController() *Controller {
	v := newLocalVolumeServer()
	return &Controller{
		MetaServiceServer:   newLocalMetaServer(),
		NodeServiceServer:   newLocalNodeServer(),
		VolumeServiceServer: v,
		CommitServiceServer: v,
	}
}

type Controller struct {
	MetaServiceServer
	NodeServiceServer
	VolumeServiceServer
	CommitServiceServer
}
