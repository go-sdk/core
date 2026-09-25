package osx

import (
	"encoding/json"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-sdk/core/conv"
)

const (
	versionEmpty = "v0.0.0"
	versionDevel = "(devel)"
)

// Version 描述当前程序的版本、版本控制和 Go 构建信息。
type Version struct {
	Type     string
	Version  string
	Revision string
	Time     string
	Dirty    bool

	GoVersion string
	Compiler  string
	Platform  string
}

// String 以缩进后的 JSON 格式返回版本信息。
func (v Version) String() string {
	bs, _ := json.MarshalIndent(v, "", "  ")
	return conv.BytesToString(bs)
}

// iVersion 是构建时通过 -ldflags "-X .../osx.iVersion=v1.2.3" 注入的版本号，
// 未注入时回退到构建信息中的主模块版本。
var (
	iVersion string

	version     Version
	versionOnce sync.Once
)

// GetVersion 返回程序构建信息，结果只计算一次并缓存。
func GetVersion() Version {
	versionOnce.Do(func() {
		bi, ok := debug.ReadBuildInfo()
		if !ok || bi == nil {
			return
		}

		if iVersion != "" {
			version.Version = iVersion
		} else if bi.Main.Version != versionDevel {
			version.Version = bi.Main.Version
		} else {
			version.Version = versionEmpty
		}

		stm := map[string]string{}
		for _, s := range bi.Settings {
			stm[s.Key] = s.Value
		}

		version.Type = stm["vcs"]
		version.Revision = stm["vcs.revision"]
		version.Dirty = stm["vcs.modified"] == "true"

		t, err := time.Parse(time.RFC3339Nano, stm["vcs.time"])
		if err == nil {
			version.Time = t.Local().Format(time.RFC3339Nano)
		}

		version.GoVersion = runtime.Version()
		version.Compiler = runtime.Compiler
		version.Platform = runtime.GOOS + "/" + runtime.GOARCH
	})
	return version
}
