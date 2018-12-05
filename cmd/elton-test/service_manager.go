package main

import (
	"context"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net"
	"sync"
	"time"
)

// 1つのプロセス内で動作しているサービスを管理する。
// ServiceManager は、 Subsystem ごとに1つ利用する。
type ServiceManager struct {
	Services        []Service
	ShutdownTimeout time.Duration

	sockets []net.Listener
}

func (m *ServiceManager) Setup(ctx context.Context) (err error) {
	if err = m.allocateListeners(); err != nil {
		return err
	}

	m.bindListeners()
	if err = m.register(ctx); err != nil {
		return err
	}
	return nil
}
func (m *ServiceManager) Serve(parentCtx context.Context) (errors []error) {
	var wgServe, wgEch sync.WaitGroup
	ech := make(chan error)

	ctx, cancel := context.WithCancel(parentCtx)

	// Serve all services.
	for i := range m.Services {
		srv := m.Services[i]
		wgServe.Add(1)
		go func() {
			defer wgServe.Done()
			zap.S().Debugw("SM.Serve server", "service", srv, "status", "serving")
			ech <- srv.Serve(ctx, m.sockets[i])
			zap.S().Debugw("SM.Serve server", "service", srv, "status", "served")
			shutdownCtx, _ := context.WithTimeout(context.Background(), m.ShutdownTimeout)
			ech <- srv.Unregister(shutdownCtx)
		}()
	}

	// Collect all errors from ech.
	// どれか1つのサービスでエラー終了した場合、contextをキャンセルして、全てのサービスを停止する。
	wgEch.Add(1)
	go func() {
		defer wgEch.Done()
		zap.S().Debugw("SM.Serve errorCollector", "status", "waiting")
		for err := range ech {
			if err != nil {
				cancel()
				errors = append(errors, err)
			}
		}
		zap.S().Debugw("SM.Serve errorCollector", "status", "finished")
	}()

	// Wait for all services.
	zap.S().Debugw("SM.Serve", "status", "server waiting")
	wgServe.Wait()
	close(ech)
	// Wait for the error collector.
	zap.S().Debugw("SM.Serve", "status", "errorCollector waiting")
	wgEch.Wait()
	zap.S().Debugw("SM.Serve", "status", "finished")
	return
}
func (m *ServiceManager) Addrs() (addrs []net.Addr) {
	for _, s := range m.sockets {
		addrs = append(addrs, s.Addr())
	}
	return
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

	for range m.Services {
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
func (m *ServiceManager) bindListeners() {
	for i := range m.Services {
		addr := m.sockets[i].Addr()
		m.Services[i].SetAddr(addr)
	}
}
func (m *ServiceManager) register(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			m.unregister()
		}
	}()

	var eg errgroup.Group
	for _, _srv := range m.Services {
		srv := _srv
		eg.Go(func() error {
			return srv.Register(ctx)
		})
	}
	err = eg.Wait()
	return
}
func (m *ServiceManager) unregister() error {
	var eg errgroup.Group
	ctx, _ := context.WithTimeout(context.Background(), m.ShutdownTimeout)

	for _, _srv := range m.Services {
		srv := _srv
		eg.Go(func() error {
			return srv.Unregister(ctx)
		})
	}
	return eg.Wait()
}
