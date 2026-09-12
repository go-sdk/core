package logx

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"

	"github.com/go-sdk/core/testx"
)

func TestGlobalKV(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := zerolog.New(buffer).Hook(globalKvHook{})

	t.Cleanup(ClearGlobalKV)
	SetGlobalKV("app", "core")
	SetGlobalKV("trace", "abc")
	logger.Info().Msg("全局键值")
	output := buffer.String()
	testx.Contains(t, output, `"app":"core"`)
	testx.Contains(t, output, `"trace":"abc"`)

	SetGlobalKV("app", "override")
	DeleteGlobalKV("trace")
	buffer.Reset()
	logger.Info().Msg("更新键值")
	output = buffer.String()
	testx.Contains(t, output, `"app":"override"`)
	testx.NotContains(t, output, `"trace"`)

	ClearGlobalKV()
	buffer.Reset()
	logger.Info().Msg("清空键值")
	testx.NotContains(t, buffer.String(), `"app"`)
	testx.NotContains(t, buffer.String(), `"override"`)

	buffer.Reset()
	logger = zerolog.New(buffer)
	logger.Info().Msg("无钩子")
	testx.NotContains(t, buffer.String(), `"app"`)
}

// reentrantKv 在序列化期间再次修改全局键值，验证 Hook 不持有写入锁。
type reentrantKv struct{}

func (reentrantKv) MarshalZerologObject(event *zerolog.Event) {
	SetGlobalKV("reentrant", "updated")
	event.Str("reentrant", "true")
}

func TestGlobalKVReentrant(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := zerolog.New(buffer).Hook(globalKvHook{})

	t.Cleanup(ClearGlobalKV)
	SetGlobalKV("value", reentrantKv{})
	logger.Info().Msg("重入序列化")
	output := buffer.String()
	testx.Contains(t, output, `"value":{"reentrant":"true"}`)

	buffer.Reset()
	logger.Info().Msg("重入后生效")
	testx.Contains(t, buffer.String(), `"reentrant":"updated"`)
}

func TestGlobalKVConflictingKey(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := zerolog.New(buffer).Hook(globalKvHook{})

	t.Cleanup(ClearGlobalKV)
	SetGlobalKV("service", "global")
	logger.Info().Str("service", "local").Msg("同名字段")
	output := buffer.String()
	testx.Contains(t, output, `"service":"local"`)
	testx.Contains(t, output, `"service":"global"`)
}

func TestGlobalKVAttachedByNew(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger := New("").Output(buffer)

	t.Cleanup(ClearGlobalKV)
	SetGlobalKV("app", "core")
	logger.Info().Msg("带全局键值")
	testx.Contains(t, buffer.String(), `"app":"core"`)
}

func TestGlobalKVConcurrent(t *testing.T) {
	t.Cleanup(ClearGlobalKV)
	buffer := &bytes.Buffer{}
	logger := zerolog.New(zerolog.SyncWriter(buffer)).Hook(globalKvHook{})

	done := make(chan struct{})
	for i := 0; i < 4; i++ {
		i := i
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 100; j++ {
				SetGlobalKV("key", i)
				DeleteGlobalKV("key")
				ClearGlobalKV()
				logger.Info().Msg("并发日志")
			}
		}()
	}
	for i := 0; i < 4; i++ {
		<-done
	}
	testx.Contains(t, buffer.String(), "并发日志")
}
