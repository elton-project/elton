package proto2

import (
	"go.uber.org/zap"
	"sync"
)

var IsControllerServerRunning bool

var system = struct {
	sync.Mutex
	Status SystemStatus
}{}

type systemStateTransition struct {
	From SystemStatus
	To   SystemStatus
}

// See service.proto file.
var allowedSystemStateTransitions = map[systemStateTransition]struct{}{
	{SystemStatus_SS_UNAVAILABLE, SystemStatus_SS_INIT_PHASE_0}:        {},
	{SystemStatus_SS_INIT_PHASE_0, SystemStatus_SS_INIT_PHASE_1}:       {},
	{SystemStatus_SS_INIT_PHASE_1, SystemStatus_SS_INIT_PHASE_2}:       {},
	{SystemStatus_SS_INIT_PHASE_2, SystemStatus_SS_OPERATING_RW}:       {},
	{SystemStatus_SS_OPERATING_RW, SystemStatus_SS_OPERATING_RW_TO_RO}: {},
	{SystemStatus_SS_OPERATING_RW_TO_RO, SystemStatus_SS_OPERATING_RO}: {},
	{SystemStatus_SS_OPERATING_RO, SystemStatus_SS_OPERATING_RW}:       {},
}

// SystemStatusを更新する。
// 戻り地は、以前のSystemStatus。
func SetSystemStatus(status SystemStatus) SystemStatus {
	system.Lock()
	old := system.Status
	system.Status = status
	system.Unlock()

	// 正しい順番でステートが遷移しているか検証。
	transition := systemStateTransition{old, status}
	if _, ok := allowedSystemStateTransitions[transition]; !ok {
		zap.S().Fatalw("invalid state transition", "from", transition.From.String(), "to", transition.To.String())
	}

	zap.S().Infow("changed SystemStatus", "from", transition.From.String(), "to", transition.To.String())
	return old
}

// 現在のSystemStatusを取得する。
// Controllerが動作しているプロセス以外は、アクセスしてはいけない。
func GetSystemStatus() SystemStatus {
	if IsControllerServerRunning {
		zap.S().Fatalw("non-controller server must not access SystemStatus")
	}

	system.Lock()
	s := system.Status
	system.Unlock()
	return s
}
