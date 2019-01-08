package proto2

import (
	"context"
	"go.uber.org/zap"
	"net"
	"sync"
)

type SubsystemManager struct {
	Config *Config

	isConfigured bool
	once         sync.Once

	subsystems []Subsystem
	managers   []*ServiceManager
	localSD    *localServiceDiscoverer
	globalSD   *globalServiceDiscoverer
}

func (m *SubsystemManager) init() {
	m.localSD = &localServiceDiscoverer{}
	m.globalSD = &globalServiceDiscoverer{
		Timeout: m.Config.RPCTimeout,
		LocalSD: m.localSD,
	}

	// 同一ノード上で動作しているcontrollerを登録。
	if m.Config.Controller.ListenAddr != nil {
		m.globalSD.AddControllers([]net.Addr{
			&net.TCPAddr{
				// 同一ノード上なので、宛先アドレスは"127.0.0.1"になる。
				IP:   net.IPv4(127, 0, 0, 1),
				Port: m.Config.Controller.ListenAddr.(*net.TCPAddr).Port,
			},
		})
	}
	// controllersの初期ノードを登録。
	if len(m.Config.Controller.Servers) > 0 {
		m.globalSD.AddControllers(m.Config.Controller.Servers)
	}
}
func (m *SubsystemManager) Add(subsystem Subsystem) {
	m.once.Do(m.init)
	if m.isConfigured {
		zap.S().Panic("SubsystemManager",
			"error", "Add() method was called after Setup()")
		panic("Add() method was called after Setup()")
	}

	mng := &ServiceManager{
		Config:   m.Config,
		LocalSD:  m.localSD,
		GlobalSD: m.globalSD,
	}
	m.subsystems = append(m.subsystems, subsystem)
	m.managers = append(m.managers, mng)
}
func (m *SubsystemManager) Setup(ctx context.Context) (errors []error) {
	m.once.Do(m.init)
	if m.isConfigured {
		zap.S().Panic("SubsystemManager",
			"error", "Setup() method was called two times")
		panic("Setup() method was called two times")
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	var lock sync.Mutex
	handleErrors := func(errs []error) {
		if len(errs) > 0 {
			lock.Lock()
			errors = append(errors, errs...)
			lock.Unlock()
		}
	}

	for i := range m.subsystems {
		s := m.subsystems[i]
		mng := m.managers[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			handleErrors(s.Setup(ctx, mng))
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

	for i := range m.subsystems {
		s := m.subsystems[i]
		mng := m.managers[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			handleErrors(s.Serve(ctx, mng))
		}()
	}
	return
}

// ServiceManagerは、同一プロセス同一サブシステム内で動作しているサービスを管理する。
// サブシステムごとに1つの ServiceManager を用意する。
type ServiceManager struct {
	Config   *Config
	LocalSD  *localServiceDiscoverer
	GlobalSD *globalServiceDiscoverer

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
		config := m.serverConfig(ctx, sock, srv)

		// サービスを登録
		m.LocalSD.Add(sock.Addr(), srv.SubsystemType(), srv.ServiceType())

		// Created eventを出す。
		zap.S().Debugw("SM.Serve", "service", srv, "status", "created")
		if err := srv.Created(config); err != nil {
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
		config := m.serverConfig(ctx, sock, srv)

		wg.Add(1)
		go func() {
			defer wg.Done()

			var innerWg sync.WaitGroup
			defer innerWg.Wait()

			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				zap.S().Debugw("SM.Serve", "service", srv, "status", "running")
				handleError(srv.Running(config))
			}()

			zap.S().Debugw("SM.Serve", "service", srv, "status", "serve")
			if handleError(srv.Serve(config)) {
				return
			}
			// Running()が終了するまで待つ。
			innerWg.Wait()

			// TODO: 外部からの終了要求を受け付ける前に、prestop()を実行する。
			zap.S().Debugw("SM.Serve", "service", srv, "status", "prestop")
			if handleError(srv.Prestop(config)) {
				return
			}

			zap.S().Debugw("SM.Serve", "service", srv, "status", "stopped")
			if handleError(srv.Stopped(config)) {
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
	return m.GlobalSD.chooseController()
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

	for i := range m.services {
		var sock net.Listener
		isController := m.services[i].SubsystemType() == SubsystemType_ControllerSubsystemType &&
			m.services[i].ServiceType() == ServiceType_ControllerServiceType

		if isController {
			// controllerは、configで指定されたportをlistenする
			addr := m.Config.Controller.ListenAddr
			sock, err = net.Listen(addr.Network(), addr.String())
		} else {
			// controller以外は、ランダムなportをlistenする
			sock, err = m.socket()
		}

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
func (m *ServiceManager) serverConfig(ctx context.Context, sock net.Listener, srv Service) *ServerConfig {
	return &ServerConfig{
		ServerInfo: *NewServerInfo(sock.Addr()),
		Ctx:        ctx,
		Listener:   sock,
		Discoverer: m.GlobalSD,
	}
}
