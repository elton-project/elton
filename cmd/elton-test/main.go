package main

import (
	"context"
	"fmt"
	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/events/p2p"
	"gitlab.t-lab.cs.teu.ac.jp/kaimag/Elton/grpc/proto2"
	"go.uber.org/zap"
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
	fmt.Println(err)
	fmt.Println(conf)
	_ = conf

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_ = ctx

	manager := proto2.SubsystemManager{
		Config: &proto2.Config{
			Controller: &proto2.ControllerConfig{
				ListenAddr: conf.ControllerListenAddr.Addr,
				Servers:    conf.Controllers.Addrs,
			},
		},
	}
	manager.Add(&p2p.ControllerSubsystem{})
	manager.Add(&ExampleSubsystem{})

	if errs := manager.Setup(ctx); len(errs) > 0 {
		panic(errs)
	}

	if errs := manager.Serve(ctx); len(errs) > 0 {
		panic(errs)
	}

	return 0
}
func main() {
	os.Exit(Main())
}
