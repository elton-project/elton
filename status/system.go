package status

import (
	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"go.uber.org/zap"
	"sync"
)

var IsControllerServerRunning bool

var system = struct {
	sync.Mutex
	Status proto2.SystemStatus
}{}

// SystemStatusを更新する。
// 戻り地は、以前のSystemStatus。
func SetSystemStatus(status proto2.SystemStatus) proto2.SystemStatus {
	system.Lock()
	old := system.Status
	system.Status = status
	system.Unlock()
	return old
}

// 現在のSystemStatusを取得する。
// Controllerが動作しているプロセス以外は、アクセスしてはいけない。
func GetSystemStatus() proto2.SystemStatus {
	if IsControllerServerRunning {
		zap.S().Fatalw("non-controller server must not access SystemStatus")
	}

	system.Lock()
	s := system.Status
	system.Unlock()
	return s
}
