package lifex

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/go-sdk/core/testx"
)

// reset 重置全局状态，保证测试之间互不影响。
func reset() {
	mu.Lock()
	inits = nil
	deinits = nil
	mu.Unlock()

	triggerOnce = sync.Once{}
	stopped = make(chan struct{})
	reason = nil

	deinitOnce = sync.Once{}
	finished = make(chan struct{})
}

func TestInitOrder(t *testing.T) {
	reset()

	var order []string
	OnInit(func() error { order = append(order, "a"); return nil })
	OnInit(func() error { order = append(order, "b"); return nil })

	testx.NoError(t, Init())
	testx.Equal(t, []string{"a", "b"}, order)
}

func TestInitStopOnError(t *testing.T) {
	reset()

	failure := errors.New("初始化失败")
	var order []string
	OnInit(func() error { order = append(order, "a"); return nil })
	OnInit(func() error { order = append(order, "b"); return failure })
	OnInit(func() error { order = append(order, "c"); return nil })

	testx.ErrorIs(t, Init(), failure)
	testx.Equal(t, []string{"a", "b"}, order)
}

func TestWaitDeinitReverseOrder(t *testing.T) {
	reset()

	var order []string
	OnDeinit(func() error { order = append(order, "a"); return nil })
	OnDeinit(func() error { order = append(order, "b"); return errors.New("解构失败") })
	OnDeinit(func() error { order = append(order, "c"); return nil })

	Shutdown(nil)
	testx.NoError(t, Wait())
	testx.Equal(t, []string{"c", "b", "a"}, order)
}

func TestWaitAfterDeinitReturnsSameReason(t *testing.T) {
	reset()

	OnDeinit(func() error { return nil })
	Shutdown(nil)

	testx.NoError(t, Wait())
	testx.NoError(t, Wait())
}

func TestShutdownIdempotent(t *testing.T) {
	reset()

	first := errors.New("第一次退出")
	Shutdown(first)
	Shutdown(errors.New("第二次退出"))

	testx.ErrorIs(t, Wait(), first)
}

func TestWaitConcurrent(t *testing.T) {
	reset()

	target := errors.New("主动退出")
	OnDeinit(func() error { return nil })

	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- Wait()
		}()
	}

	Shutdown(target)
	wg.Wait()
	close(errs)

	for err := range errs {
		testx.ErrorIs(t, err, target)
	}
}

// childSignalMode 子进程模式：注册解构函数后向自身发送 SIGTERM，
// Wait 返回 nil 且解构已执行时以退出码 0 结束。
func childSignalMode(t *testing.T) {
	t.Helper()

	marker := os.Getenv("LIFEX_TEST_MARKER")
	OnDeinit(func() error {
		return os.WriteFile(marker, []byte("done"), 0o600)
	})

	go func() {
		time.Sleep(200 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	if err := Wait(); err != nil {
		os.Exit(2)
	}
	os.Exit(0)
}

func TestWaitSignal(t *testing.T) {
	mode := os.Getenv("LIFEX_TEST_MODE")
	if mode == "signal" {
		childSignalMode(t)
		return
	}

	marker := filepath.Join(t.TempDir(), "marker")
	cmd := exec.Command(os.Args[0], "-test.run=TestWaitSignal", "-test.v=false")
	cmd.Env = append(os.Environ(),
		"LIFEX_TEST_MODE=signal",
		"LIFEX_TEST_MARKER="+marker,
	)
	cmd.Stdout = nil
	cmd.Stderr = nil

	testx.NoError(t, cmd.Run())

	content, err := os.ReadFile(marker)
	testx.NoError(t, err)
	testx.Equal(t, "done", string(content))
}

// childForceMode 子进程模式：解构函数耗时较长，
// 解构期间收到第二次 SIGTERM 时跳过剩余解构并强制退出。
func childForceMode(t *testing.T) {
	t.Helper()

	OnDeinit(func() error {
		time.Sleep(time.Second)
		return nil
	})

	go func() {
		time.Sleep(200 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		time.Sleep(50 * time.Millisecond)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	_ = Wait()
	os.Exit(0) // 不应到达，正常路径会先被第二次信号强制退出
}

func TestWaitForceExitOnSecondSignal(t *testing.T) {
	mode := os.Getenv("LIFEX_TEST_MODE")
	if mode == "force" {
		childForceMode(t)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestWaitForceExitOnSecondSignal", "-test.v=false")
	cmd.Env = append(os.Environ(), "LIFEX_TEST_MODE=force")
	cmd.Stdout = nil
	cmd.Stderr = nil

	var ee *exec.ExitError
	err := cmd.Run()
	testx.ErrorAs(t, err, &ee)
	testx.Equal(t, 1, ee.ExitCode())
}
