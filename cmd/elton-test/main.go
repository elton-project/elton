package main

import (
	"context"
	"fmt"
	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/events/p2p"
	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"go.uber.org/zap"
	"io"
	"os"
	"time"
)

func Main() int {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer zap.S().Sync()

	conf, err := loadFromEnvironment()
	if err != nil {
		zap.S().Fatalw("config", "error", err.Error())
		return 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	manager := proto2.SubsystemManager{
		Config: &proto2.Config{
			Controller: &proto2.ControllerConfig{
				ListenAddr: conf.ControllerListenAddr.Addr,
				Servers:    conf.Controllers.Addrs,
			},
			RPCTimeout:      10 * time.Second, // TODO: update timeout
			ShutdownTimeout: 1 * time.Minute,  // TODO: update timeout
		},
	}
	manager.Add(&p2p.ControllerSubsystem{})
	manager.Add(&ExampleSubsystem{})

	if errs := manager.Setup(ctx); len(errs) > 0 {
		fprintErrors(os.Stderr, "SubsystemManager.Setup()", errs)
		return 1
	}

	if errs := manager.Serve(ctx); len(errs) > 0 {
		fprintErrors(os.Stderr, "SubsystemManager.Serve()", errs)
		return 1
	}

	return 0
}
func main() {
	os.Exit(Main())
}
func fprintErrors(writer io.Writer, prefix string, errs []error) {
	for i, err := range errs {
		fmt.Fprintf(os.Stderr, "%s ERROR[%d]:\n%+v\n\n", prefix, i, err)
	}
}
