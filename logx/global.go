package logx

import (
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog"
)

var (
	// globalKv 保存不可变快照，写入时复制并整体替换，日志热路径无锁读取。
	globalKv   atomic.Pointer[map[string]any]
	globalKvMu sync.Mutex
)

// SetGlobalKV 设置进程级全局键值，之后所有由本包创建的日志事件都会附带该键值。
// 重复设置同一个键会覆盖旧值，可并发调用。
// 全局键值应仅用于 service、version 等进程级标识；与事件字段同名时会产生重复 JSON key，
// 调用方应保证全局键不与日志字段冲突。请求级的 trace、user 等字段应通过上下文 Logger 传递。
func SetGlobalKV(key string, value any) {
	globalKvMu.Lock()
	defer globalKvMu.Unlock()
	next := copyGlobalKv()
	next[key] = value
	globalKv.Store(&next)
}

// DeleteGlobalKV 删除指定的全局键值，之后的日志事件不再附带该键。
func DeleteGlobalKV(key string) {
	globalKvMu.Lock()
	defer globalKvMu.Unlock()
	next := copyGlobalKv()
	delete(next, key)
	globalKv.Store(&next)
}

// ClearGlobalKV 清空全部全局键值。
func ClearGlobalKV() {
	globalKvMu.Lock()
	defer globalKvMu.Unlock()
	globalKv.Store(nil)
}

func copyGlobalKv() map[string]any {
	old := globalKv.Load()
	if old == nil {
		return make(map[string]any)
	}
	next := make(map[string]any, len(*old))
	for k, v := range *old {
		next[k] = v
	}
	return next
}

// globalKvHook 将全局键值附加到每个日志事件。
type globalKvHook struct{}

// Run 读取不可变快照并在无锁状态下序列化，
// 避免值序列化过程中再次记录日志时与写入锁重入死锁。
func (globalKvHook) Run(event *zerolog.Event, _ zerolog.Level, _ string) {
	kv := globalKv.Load()
	if kv == nil {
		return
	}
	for k, v := range *kv {
		event.Any(k, v)
	}
}
