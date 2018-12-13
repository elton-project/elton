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
	DelivererServiceType
	EventManagerServiceType
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
	// コンテキストが終了するまで、受信したリクエストは全て処理するべきである。
	// コンテキストが終了前に、能動的にreturnすることも可能だが、やるべきではない。
	Serve(info *ServerInfo) error

	// Created()は、初期化処理をする。
	// このメソッドは、Serve()が実行される前に呼び出される。
	// 既にlistenしているのでアドレスやポート番号は決まっているが、いかなる接続もacceptされない。
	Created(info *ServerInfo) error
	// Running()は、サーバの起動直後に行なうべき処理をする。
	// Running()とServe()は並行処理されるが、同時に実行されることは保証しない。
	// タイミング次第では、まだServe()が実行開始されていない状態でRunning()が実行される可能性がある。
	Running(info *ServerInfo) error
	// Prestop()は、サーバの終了直前に行うべき処理をする。
	// このメソッドの実行終了後にServe()が終了する。
	// 通常はPrestop()とServe()は並行処理される。しかし、Serve()が先に終了する可能性もあることに注意。
	Prestop(info *ServerInfo) error
	// Stopped()は、サーバの終了後に行うべき処理をする。
	// Serve()の実行終了後に呼び出される。
	// まだlistenしているが、いかなる接続もacceptされない。
	Stopped(info *ServerInfo) error
}
