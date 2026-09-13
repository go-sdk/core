package config

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cast"

	"github.com/go-sdk/core/osx"
)

const delimiter = "."

type Option func(*Config)

type watcher interface {
	Close() error
}

type Config struct {
	loadMu sync.Mutex
	dataMu sync.RWMutex

	flatData   map[string]any
	nestedData map[string]any

	filename  string
	fileType  string
	hasFile   bool
	fileWatch bool
	optionErr error

	watchMu       sync.Mutex
	watcher       watcher
	watchDone     chan struct{}
	lifecycleOnce sync.Once
}

// New 创建一个独立的配置实例。
func New(opts ...Option) *Config {
	c := &Config{
		flatData:   map[string]any{},
		nestedData: map[string]any{},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	return c
}

// WithFile 指定配置文件。fileType 可选值为 json 或 yaml；未指定时根据扩展名判断。
func WithFile(filename string, fileType ...string) Option {
	return func(c *Config) {
		c.filename = filename
		c.hasFile = true
		if len(fileType) > 1 {
			c.optionErr = fmt.Errorf("config: WithFile accepts at most one file type")
			return
		}
		if len(fileType) == 1 {
			c.fileType = strings.ToLower(strings.TrimSpace(fileType[0]))
		}
	}
}

// WithFileWatch 控制是否在首次成功加载后监听配置文件变化。
func WithFileWatch(enabled bool) Option {
	return func(c *Config) {
		c.fileWatch = enabled
	}
}

// Get 返回指定路径转换后的基础类型值；第二个返回值仅表示路径是否存在。
func (c *Config) Get[T cast.Basic](key string) (T, bool) {
	c.dataMu.RLock()
	value, ok := c.flatData[key]
	c.dataMu.RUnlock()
	if !ok {
		var zero T
		return zero, false
	}
	return cast.To[T](value), true
}

// MustGet 返回指定路径转换后的基础类型值，路径不存在时触发 panic。
func (c *Config) MustGet[T cast.Basic](key string) T {
	value, ok := c.Get[T](key)
	if !ok {
		osx.Panicf("config: key %q does not exist", key)
	}
	return value
}

// Exists 判断指定路径是否存在。
func (c *Config) Exists(key string) bool {
	c.dataMu.RLock()
	_, ok := c.flatData[key]
	c.dataMu.RUnlock()
	return ok
}

// Raw 返回嵌套配置数据的深拷贝。
func (c *Config) Raw() map[string]any {
	c.dataMu.RLock()
	data := cloneMap(c.nestedData)
	c.dataMu.RUnlock()
	return data
}

// DecodeTo 使用 json tag 将完整配置解码到目标值，并解析 duration 与 RFC3339 时间。
func (c *Config) DecodeTo(target any) error {
	if target == nil {
		return fmt.Errorf("config: decode target must be a non-nil pointer")
	}
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("config: decode target must be a non-nil pointer")
	}

	c.dataMu.RLock()
	data := cloneMap(c.nestedData)
	c.dataMu.RUnlock()

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           target,
		TagName:          "json",
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToTimeHookFunc(time.RFC3339),
		),
	})
	if err != nil {
		return fmt.Errorf("config: create decoder: %w", err)
	}
	if err = decoder.Decode(data); err != nil {
		return fmt.Errorf("config: decode: %w", err)
	}
	return nil
}
