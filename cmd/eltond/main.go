package main

import (
	"context"
	"github.com/tchap/zapext/zapsyslog"
	"gitlab.t-lab.cs.teu.ac.jp/yuuki/elton/subsystems"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"log/syslog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func handleSignals(cancel context.CancelFunc, ctx context.Context, wg *sync.WaitGroup, signals ...os.Signal) {
	defer wg.Done()
	logger := zap.S()

	stopSignal := make(chan os.Signal, 2)
	signal.Notify(stopSignal, signals...)
	defer signal.Stop(stopSignal)

	select {
	case <-stopSignal:
		logger.With("signal", stopSignal).Info("signal handled")
	case <-ctx.Done():
	}
	cancel()
}

func setupLogger() {
	writer, err := syslog.New(syslog.LOG_ERR|syslog.LOG_LOCAL0, "eltond")
	if err != nil {
		log.Fatal("failed to setup syslog: %w", err)
	}

	// Initialize Zap.
	encoder := zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig())
	core := zapsyslog.NewCore(zapcore.ErrorLevel, encoder, writer)
	logger := zap.New(core, zap.Development(), zap.AddStacktrace(zapcore.ErrorLevel))
	zap.ReplaceGlobals(logger)
}

func Main() int {
	setupLogger()

	// Start signal handler.
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer wg.Wait()
	defer cancel()
	wg.Add(1)
	go handleSignals(cancel, ctx, &wg, syscall.SIGTERM, syscall.SIGINT)

	// Load configuration.
	conf, err := loadFromEnvironment()
	if err != nil {
		zap.S().With("error", err).Error("failed to load configuration")
		return 1
	}

	// Start servers.
	swg := sync.WaitGroup{}
	for _, _role := range conf.Roles {
		role := _role
		fn := func() subsystems.Server {
			s := NewServer(role)
			if s == nil {
				zap.S().With("role", role).Error("unsupported role specified")
				cancel()
				return nil
			}
			return s
		}

		swg.Add(1)
		go func() {
			defer swg.Done()

			sm := subsystems.ServerManager{
				New:             fn,
				Name:            role,
				AutoRestart:     false,
				RestartInterval: 3 * time.Second,
			}
			sm.Serve(ctx)
		}()
	}
	swg.Wait()
	return 0
}

func main() {
	os.Exit(Main())
}
