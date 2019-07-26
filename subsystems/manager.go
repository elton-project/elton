package subsystems

import (
	"context"
	"go.uber.org/zap"
	"time"
)

type ServerManager struct {
	New             func() Server
	Name            string
	AutoRestart     bool
	RestartInterval time.Duration
}

func (m *ServerManager) Serve(ctx context.Context) {
	count := 0
	for {
		cause := ""
		restartRequest := false
		needRestart := true

		err := m.serveWithoutRestart(ctx)

		select {
		case <-ctx.Done():
			// Context is cancelled.  Should not restart server.
			return
		default:
		}

		switch err.(type) {
		case nil:
			cause = "no error"
		case *RestartRequest:
			cause = err.Error()
			restartRequest = true
			continue
		default:
			cause = err.Error()
			if !m.AutoRestart {
				needRestart = false
			}
		}

		zap.S().With(
			"cause", cause,
			"receivedRestartRequest", restartRequest,
		).Errorf("%s server stopped", m.Name)

		if !needRestart {
			return
		}
		zap.S().With(
			"restartCount", count,
		).Info("restart %s server in %f seconds ...", m.Name, m.RestartInterval.Seconds())
		time.Sleep(m.RestartInterval)
		count++
	}
}

func (m *ServerManager) serveWithoutRestart(ctx context.Context) error {
	server := m.New()
	if err := server.Configure(); err != nil {
		return err
	}

	if err := server.Listen(); err != nil {
		return err
	}

	if err := server.Serve(ctx); err != nil {
		return err
	}
	return nil
}
