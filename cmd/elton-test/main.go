package main

import (
	"context"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"os"
	"time"
)

func serveEventManager(ctx context.Context) func() error {
	return func() error {
		m := ServiceManager{
			Services: []Service{
				&ManagerService{
					L: zap.S(),
				},
			},
		}
		if err := m.Setup(ctx); err != nil {
			zap.S().Errorw("EventManager", "error", err)
			return err
		}
		if errs := m.Serve(ctx); len(errs) > 0 {
			zap.S().Errorw("EventManager", "errors", errs)
			return errs[0]
		}
		return nil
	}
}
func serveExampleSubsystem(ctx context.Context) func() error {
	return func() error {
		ex := ExampleSubsystem{}
		if err := ex.Setup(ctx); err != nil {
			zap.S().Errorw("ExampleSubsystem", "error", err)
			return err
		}
		if errs := ex.Serve(ctx); len(errs) > 0 {
			zap.S().Errorw("ExampleSubsystem", "errors", errs)
			return errs[0]
		}
		return nil
	}
}

func Main() int {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer zap.S().Sync()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var eg errgroup.Group
	eg.Go(serveEventManager(ctx))
	eg.Go(serveExampleSubsystem(ctx))
	if err := eg.Wait(); err != nil {
		return 1
	}
	return 0
}
func main() {
	os.Exit(Main())
}
