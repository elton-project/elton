package main

import (
	"context"
	"fmt"
	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"net"
)

type SubsystemType int64
type ServiceType int64

const (
	UnknownSubsystemType = SubsystemType(iota)
	ControllerSubsystemType
	StorageSubsystemType
	ObjectSubsystemType
	SchedulerSubsystemType

	UnknownServiceType = ServiceType(iota)
	ListenerServiceType
	ManagerServiceType
	DelivererServiceType
)

type Subsystem interface {
	fmt.Stringer
	Name() string
	SubsystemType() SubsystemType

	Setup(ctx context.Context, manager *ServiceManager) []error
	Serve(ctx context.Context, manager *ServiceManager) []error
}

type ServerInfo struct {
	proto2.ServerInfo
	Ctx      context.Context
	Listener net.Listener
}

type Service interface {
	fmt.Stringer
	Name() string
	SubsystemType() SubsystemType
	ServiceType() ServiceType

	// Serve()は、サービスを提供するメソッドである。
	Serve(info *ServerInfo) error

	// Created()は、サーバ起動直前の初期化処理をする。
	// このメソッドは、Serve()が実行される前に呼び出される。
	Created(info *ServerInfo) error
	// Running()は、サーバの起動直後に行なうべき処理をする。
	// Running()とServe()は、それぞれgoroutineで並行処理される。
	Running(info *ServerInfo) error
	// Prestop()は、サーバの終了直前に行うべき処理をする。
	// Prestop()とServe()は、それぞれgoroutineで並行処理される。
	Prestop(info *ServerInfo) error
	// Stopped()は、サーバの終了後に行うべき処理をする。
	// Serve()の実行終了後に呼び出される。
	Stopped(info *ServerInfo) error
}
