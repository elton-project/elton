package proto2

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type typeKey struct {
	Sys SubsystemType
	Srv ServiceType
}

// 同一プロセス内で動作している他のサービスを探す。
type localServiceDiscoverer struct {
	services map[typeKey]net.Addr
	lock     sync.RWMutex
}

func (d *localServiceDiscoverer) Add(addr net.Addr, subsystemType SubsystemType, serviceType ServiceType) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.services == nil {
		d.services = map[typeKey]net.Addr{}
	}

	d.services[typeKey{
		Sys: subsystemType,
		Srv: serviceType,
	}] = addr
}
func (d *localServiceDiscoverer) Get(subsystemType SubsystemType, serviceType ServiceType) net.Addr {
	d.lock.RLock()
	defer d.lock.RUnlock()

	if d.services == nil {
		return nil
	}

	return d.services[typeKey{
		Sys: subsystemType,
		Srv: serviceType,
	}]
}

// システム全体から目的のサービスを探し出す。
// この機能を利用するには、Controller serviceが動作している必要あり。
type globalServiceDiscoverer struct {
	Timeout time.Duration
	LocalSD *localServiceDiscoverer

	controllers []net.Addr
	lock        sync.RWMutex
	// TODO: controllersのアドレスを自動更新する。
}

func (d *globalServiceDiscoverer) Get(ctx context.Context, subsystemType SubsystemType) (addr net.Addr, err error) {
	return d.GetWithServiceType(ctx, subsystemType, ServiceType_ListenerServiceType)
}
func (d *globalServiceDiscoverer) GetWithServiceType(parentCtx context.Context, subsystemType SubsystemType, serviceType ServiceType) (addr net.Addr, err error) {
	addr = d.LocalSD.Get(subsystemType, serviceType)
	if addr != nil {
		// fast path
		return
	}

	// slow path
	err = WithGrpcConn(d.chooseController(), func(conn *grpc.ClientConn) error {
		ctx, _ := context.WithTimeout(parentCtx, d.Timeout)
		c := NewControllerServiceClient(conn)
		info, err := c.GetServerInfo(ctx, &ServerQuery{
			SubsystemType: subsystemType,
			ServiceType:   serviceType,
		})
		if err != nil {
			return err
		}
		addr, err = d.parseTCPAddr(info.Address)
		if err != nil {
			zap.S().Error("globalServiceDiscoverer",
				"error", err.Error(),
				"value", info.Address)
			return err
		}
		return nil
	})
	return
}
func (d *globalServiceDiscoverer) UpdateControllers(ctx context.Context) error {
	newControllers := []net.Addr{}
	// TODO: コントローラのアドレス一覧を取得してくる。

	// 排他ロックがかかる時間を最小限にするため、updateのたびにスライスを差し替える手法を採用。
	d.lock.Lock()
	d.controllers = newControllers
	d.lock.Unlock()
	return nil
}
func (d *globalServiceDiscoverer) parseTCPAddr(tcpaddr string) (net.Addr, error) {
	idx := strings.LastIndex(tcpaddr, ":")
	ip := net.ParseIP(tcpaddr[:idx])
	if ip == nil {
		return nil, errors.New("invalid ip")
	}
	port, err := strconv.ParseInt(tcpaddr[idx+1:], 10, 16)
	if err != nil {
		return nil, errors.New("invalid port")
	}
	return &net.TCPAddr{
		IP:   ip,
		Port: int(port),
	}, nil
}
func (d *globalServiceDiscoverer) chooseController() net.Addr {
	d.lock.RLock()
	defer d.lock.RUnlock()

	if len(d.controllers) > 0 {
		// 候補の中からランダムに選ぶ
		length := len(d.controllers)
		idx := rand.Intn(length)
		return d.controllers[idx]
	}
	panic("Not found controller server")
}
