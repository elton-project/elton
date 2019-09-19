package simple

import (
	. "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/api/v2"
	controller_db "gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems/controller/db"
)

func NewController(databasePath string) (*Controller, func() error) {
	stores, closer, err := controller_db.CreateLocalDB(databasePath)
	if err != nil {
		// TODO: change return type.
		panic(err)
	}

	v := newLocalVolumeServer(stores.VolumeStore(), stores.CommitStore())
	return &Controller{
		MetaServiceServer:   newLocalMetaServer(),
		NodeServiceServer:   newLocalNodeServer(),
		VolumeServiceServer: v,
		CommitServiceServer: v,
	}, closer
}

type Controller struct {
	MetaServiceServer
	NodeServiceServer
	VolumeServiceServer
	CommitServiceServer
}
