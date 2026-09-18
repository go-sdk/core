package lifex

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
)

var (
	mu      sync.Mutex
	inits   []func() error
	deinits []func() error

	triggerOnce sync.Once
	stopped     = make(chan struct{})
	reason      error

	deinitOnce sync.Once
	finished   = make(chan struct{})
)

type lifecycleFunc interface {
	func() | func() error | func(context.Context) | func(context.Context) error
}

// OnInit 注册初始化函数，由 Init 按注册顺序执行；带 context.Context 参数时传入 context.Background()。
func OnInit[F lifecycleFunc](fn F) {
	mu.Lock()
	defer mu.Unlock()
	inits = append(inits, wrapLifecycleFunc(fn))
}

// OnDeinit 注册解构函数，由 Wait 按注册逆序执行；带 context.Context 参数时传入 context.Background()。
func OnDeinit[F lifecycleFunc](fn F) {
	mu.Lock()
	defer mu.Unlock()
	deinits = append(deinits, wrapLifecycleFunc(fn))
}

func wrapLifecycleFunc[F lifecycleFunc](fn F) func() error {
	switch fn := any(fn).(type) {
	case func():
		return func() error {
			fn()
			return nil
		}
	case func() error:
		return fn
	case func(context.Context):
		return func() error {
			fn(context.Background())
			return nil
		}
	case func(context.Context) error:
		return func() error {
			return fn(context.Background())
		}
	default:
		panic("lifex: unsupported lifecycle function")
	}
}

// Init 按注册顺序执行全部已注册的初始化函数并清空注册表，
// 任一失败立即停止后续执行并返回该错误，不自动执行已注册的解构函数。
func Init() error {
	mu.Lock()
	fns := inits
	inits = nil
	mu.Unlock()
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

// Wait 阻塞等待 SIGINT/SIGTERM 信号或 Shutdown 触发退出，
// 然后按注册逆序执行全部解构函数；解构函数自身的错误经默认 slog 记录，不影响其余解构执行。
// 信号触发视为正常退出返回 nil，主动退出返回 Shutdown 携带的原因；
// 解构期间再次收到退出信号时跳过剩余解构强制退出，避免解构卡死导致进程无法终止。
func Wait() error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigs)

	select {
	case sig := <-sigs:
		slog.Info("shutdown signal received, running deinit functions", "signal", sig.String())
		trigger(nil)
	case <-stopped:
	}

	quit := make(chan struct{})
	defer close(quit)
	go func() {
		select {
		case sig := <-sigs:
			slog.Warn("shutdown signal received again, forcing exit", "signal", sig.String())
			os.Exit(1)
		case <-quit:
		}
	}()

	runDeinits()
	return reason
}

// Shutdown 主动触发退出，reason 作为 Wait 的返回值；重复调用无效。
func Shutdown(r error) {
	trigger(r)
}

func trigger(r error) {
	triggerOnce.Do(func() {
		reason = r
		close(stopped)
	})
}

func runDeinits() {
	deinitOnce.Do(func() {
		mu.Lock()
		fns := deinits
		deinits = nil
		mu.Unlock()
		for _, fn := range slices.Backward(fns) {
			if err := fn(); err != nil {
				slog.Error("deinit function failed", "error", err)
			}
		}
		slog.Info("deinit functions completed")
		close(finished)
	})
	<-finished
}
