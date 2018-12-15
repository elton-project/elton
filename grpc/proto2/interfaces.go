package proto2

import (
	"context"
	"fmt"
	"net"
)

type EventSender interface {
	Send(eventType EventType)
}

type Subsystem interface {
	fmt.Stringer
	Name() string
	SubsystemType() SubsystemType

	Setup(ctx context.Context, manager *ServiceManager) []error
	Serve(ctx context.Context, manager *ServiceManager) []error
}

type ServerConfig struct {
	ServerInfo
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
	Serve(config *ServerConfig) error

	// Created()は、初期化処理をする。
	// このメソッドは、Serve()が実行される前に呼び出される。
	// 既にlistenしているのでアドレスやポート番号は決まっているが、いかなる接続もacceptされない。
	Created(config *ServerConfig) error
	// Running()は、サーバの起動直後に行なうべき処理をする。
	// Running()とServe()は並行処理されるが、同時に実行されることは保証しない。
	// タイミング次第では、まだServe()が実行開始されていない状態でRunning()が実行される可能性がある。
	Running(config *ServerConfig) error
	// Prestop()は、サーバの終了直前に行うべき処理をする。
	// このメソッドの実行終了後にServe()が終了する。
	// 通常はPrestop()とServe()は並行処理される。しかし、Serve()が先に終了する可能性もあることに注意。
	Prestop(config *ServerConfig) error
	// Stopped()は、サーバの終了後に行うべき処理をする。
	// Serve()の実行終了後に呼び出される。
	// まだlistenしているが、いかなる接続もacceptされない。
	Stopped(config *ServerConfig) error
}
