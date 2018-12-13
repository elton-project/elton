package main

import (
	"context"
	"go.uber.org/zap"
	"os"
	"time"
)

func Main() int {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer zap.S().Sync()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_ = ctx

	controlelr := ServiceManager{}
	controlelr.Add(&ControllerService{
		L: zap.S(),
	})

	_ = ExampleSubsystem{
		//EMAddr: emAddr,
	}

	return 0
}
func main() {
	os.Exit(Main())
}
