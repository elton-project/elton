package main

import (
	"context"
	"go.uber.org/zap"
	"os"
	"time"
)

//// EventManagerのサーバのセットアップをする。
//// サーバがlistenしているアドレスをemAddrに格納する。
//// 返した関数を実行すると、リクエストを受け付ける状況になる。
//func serveEventManager(ctx context.Context, emAddr *net.Addr) func() error {
//	m := ServiceManager{
//		Services: []Service{
//			&ManagerService{
//				L: zap.S(),
//			},
//		},
//	}
//	if err := m.Setup(ctx); err != nil {
//		zap.S().Errorw("EventManager", "error", err)
//		return func() error {
//			return err
//		}
//	}
//	*emAddr = m.Addrs()[0]
//
//	return func() error {
//		if errs := m.Serve(ctx); len(errs) > 0 {
//			zap.S().Errorw("EventManager", "errors", errs)
//			return errs[0]
//		}
//		return nil
//	}
//}
//func serveExampleSubsystem(ctx context.Context, emAddr net.Addr) func() error {
//	ex := ExampleSubsystem{
//		EMAddr: emAddr,
//	}
//	if err := ex.Setup(ctx); err != nil {
//		zap.S().Errorw("ExampleSubsystem", "error", err)
//		return func() error {
//			return err
//		}
//	}
//
//	return func() error {
//		if errs := ex.Serve(ctx); len(errs) > 0 {
//			zap.S().Errorw("ExampleSubsystem", "errors", errs)
//			return errs[0]
//		}
//		return nil
//	}
//}

func Main() int {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer zap.S().Sync()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_ = ServiceManager{
		Services: []Service{
			&ManagerService{
				L: zap.S(),
			},
		},
	}

	_ = ExampleSubsystem{
		//EMAddr: emAddr,
	}

	//var emAddr net.Addr
	//var eg errgroup.Group
	//eg.Go(serveEventManager(ctx, &emAddr))
	//eg.Go(serveExampleSubsystem(ctx, emAddr))
	//if err := eg.Wait(); err != nil {
	//	return 1
	//}
	return 0
}
func main() {
	os.Exit(Main())
}
