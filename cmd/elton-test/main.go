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

	manager := SubsystemManager{}
	manager.Add(&ControllerSubsystem{})
	manager.Add(&ExampleSubsystem{})

	return 0
}
func main() {
	os.Exit(Main())
}
