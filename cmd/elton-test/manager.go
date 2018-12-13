package main

import (
	"context"
	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"go.uber.org/zap"
	"math/rand"
	"net"
	"sync"
	"time"
)

type SubsystemManager struct {
	ControllerServers []net.Addr
	ShutdownTimeout   time.Duration

	isConfigured bool
	subsystems   []Subsystem
	manager      *ServiceManager
}

func (m *SubsystemManager) Add(subsystem Subsystem) {
	if m.isConfigured {
		zap.S().Panic("SubsystemManager",
			"error", "Add() method was called after Setup()")
		panic("Add() method was called after Setup()")
	}
	m.subsystems = append(m.subsystems, subsystem)
}
func (m *SubsystemManager) Setup(ctx context.Context) (errors []error) {
	if m.isConfigured {
		zap.S().Panic("SubsystemManager",
			"error", "Setup() method was called two times")
		panic("Setup() method was called two times")
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	var lock sync.Mutex
	m.manager = &ServiceManager{
		ControllerServers: m.ControllerServers,
	}
	handleErrors := func(errs []error) {
		if len(errs) > 0 {
			lock.Lock()
			errors = append(errors, errs...)
			lock.Unlock()
		}
	}

	for _, s := range m.subsystems {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handleErrors(s.Setup(ctx, m.manager))
		}()
	}
	m.isConfigured = true
	return
}
func (m *SubsystemManager) Serve(parentCtx context.Context) (errors []error) {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(parentCtx)

	var lock sync.Mutex
	handleErrors := func(errs []error) {
		if len(errs) > 0 {
			lock.Lock()
			errors = append(errors, errs...)
			lock.Unlock()

			// 1つでもサブシステムの起動に失敗したら、全てをキャンセルする。
			cancel()
		}
	}

	for _, s := range m.subsystems {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handleErrors(s.Serve(ctx, m.manager))
		}()
	}
	return
}

// 1つのプロセス内で動作しているサービスを管理する。
// ServiceManager は、 Subsystem ごとに1つ利用する。
type ServiceManager struct {
	ControllerServers []net.Addr
	ShutdownTimeout   time.Duration

	isConfigured bool
	services     []Service
	sockets      []net.Listener
}

func (m *ServiceManager) Add(service Service) {
	if m.isConfigured {
		zap.S().Panic("ServiceManager",
			"error", "Add() method was called after Setup()")
		panic("Add() method was called after Setup()")
	}
	m.services = append(m.services, service)
}
func (m *ServiceManager) Setup(ctx context.Context) (errors []error) {
	if m.isConfigured {
		zap.S().Panic("ServiceManager",
			"error", "Setup() method was called two times")
		panic("Setup() method was called two times")
	}

	if err := m.allocateListeners(); err != nil {
		// listenできなかったら、即座に中断
		return []error{err}
	}

	for i := range m.services {
		sock := m.sockets[i]
		srv := m.services[i]
		info := &ServerInfo{
			ServerInfo: *proto2.NewServerInfo(sock.Addr()),
			Ctx:        ctx,
			Listener:   sock,
		}

		// Created eventを出す。
		zap.S().Debugw("SM.Serve", "service", srv, "status", "created")
		if err := srv.Created(info); err != nil {
			errors = append(errors, err)
		}
	}
	m.isConfigured = true
	return
}
func (m *ServiceManager) Serve(parentCtx context.Context) (errors []error) {
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(parentCtx)

	var lock sync.Mutex
	handleError := func(err error) bool {
		if err != nil {
			lock.Lock()
			errors = append(errors, err)
			lock.Unlock()

			// 1つでもサービスの起動に失敗したら、全てをキャンセルする。
			cancel()
			return true
		}
		return false
	}

	// Serve all services.
	for i := range m.services {
		sock := m.sockets[i]
		srv := m.services[i]
		info := &ServerInfo{
			ServerInfo: *proto2.NewServerInfo(sock.Addr()),
			Ctx:        ctx,
			Listener:   sock,
		}
		wg.Add(1)
		go func() {
			defer wg.Done()

			var innerWg sync.WaitGroup
			defer innerWg.Wait()

			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				zap.S().Debugw("SM.Serve", "service", srv, "status", "running")
				handleError(srv.Running(info))
			}()

			zap.S().Debugw("SM.Serve", "service", srv, "status", "serve")
			if handleError(srv.Serve(info)) {
				return
			}
			// Running()が終了するまで待つ。
			innerWg.Wait()

			// TODO: 外部からの終了要求を受け付ける前に、prestop()を実行する。
			zap.S().Debugw("SM.Serve", "service", srv, "status", "prestop")
			if handleError(srv.Prestop(info)) {
				return
			}

			zap.S().Debugw("SM.Serve", "service", srv, "status", "stopped")
			if handleError(srv.Stopped(info)) {
				return
			}
		}()
	}
	return
}

// Addrs returns service addresses reachable from other nodes.
func (m *ServiceManager) Addrs() (addrs []net.Addr) {
	for _, s := range m.sockets {
		addr, ok := s.Addr().(*net.TCPAddr)
		if !ok {
			zap.S().Panicw("SM.Addrs",
				"reason", "unsupported protocol",
				"network", s.Addr().Network(),
				"addr", s.Addr().String())
			panic("unsupported protocol")
		}

		newAddr := &net.TCPAddr{
			IP:   GetPreferredIP(nil),
			Port: addr.Port,
		}
		addrs = append(addrs, newAddr)
	}
	return
}
func (m *ServiceManager) ControllerServer() net.Addr {
	if len(m.ControllerServers) > 0 {
		// 候補の中からランダムに選ぶ
		length := len(m.ControllerServers)
		idx := rand.Intn(length)
		return m.ControllerServers[idx]
	}
	panic("Not found controller server")
}
func (m *ServiceManager) socket() (net.Listener, error) {
	return net.Listen("tcp", "0.0.0.0:0")
}
func (m *ServiceManager) allocateListeners() (err error) {
	defer func() {
		if err != nil {
			for _, sock := range m.sockets {
				sock.Close()
			}
			m.sockets = nil
		}
	}()

	for range m.services {
		var sock net.Listener
		sock, err = m.socket()
		if err != nil {
			return
		}
		m.sockets = append(m.sockets, sock)
	}
	return
}
func (m *ServiceManager) unallocateListeners() {
	m.sockets = nil
}
