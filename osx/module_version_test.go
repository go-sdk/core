package osx

import (
	"runtime/debug"
	"testing"

	"github.com/go-sdk/core/testx"
)

func TestGetModuleVersion(t *testing.T) {
	testx.Equal(t, versionEmpty, GetModuleVersion("example.com/not-exists"))
}

func TestGetModuleVersionFromBuildInfo(t *testing.T) {
	tests := []struct {
		name       string
		buildInfo  *debug.BuildInfo
		modulePath string
		want       string
	}{
		{
			name:       "构建信息为空",
			modulePath: "example.com/main",
			want:       versionEmpty,
		},
		{
			name: "读取主模块版本",
			buildInfo: &debug.BuildInfo{
				Main: debug.Module{Path: "example.com/main", Version: "v1.2.3"},
			},
			modulePath: "example.com/main",
			want:       "v1.2.3",
		},
		{
			name: "主模块开发版本返回默认值",
			buildInfo: &debug.BuildInfo{
				Main: debug.Module{Path: "example.com/main", Version: versionDevel},
			},
			modulePath: "example.com/main",
			want:       versionEmpty,
		},
		{
			name: "读取依赖模块版本",
			buildInfo: &debug.BuildInfo{
				Deps: []*debug.Module{
					{Path: "example.com/dependency", Version: "v2.3.4"},
				},
			},
			modulePath: "example.com/dependency",
			want:       "v2.3.4",
		},
		{
			name: "读取替换模块版本",
			buildInfo: &debug.BuildInfo{
				Deps: []*debug.Module{
					{
						Path:    "example.com/dependency",
						Version: "v2.3.4",
						Replace: &debug.Module{Path: "example.com/replacement", Version: "v3.4.5"},
					},
				},
			},
			modulePath: "example.com/dependency",
			want:       "v3.4.5",
		},
		{
			name: "本地替换模块返回默认值",
			buildInfo: &debug.BuildInfo{
				Deps: []*debug.Module{
					{
						Path:    "example.com/dependency",
						Version: "v2.3.4",
						Replace: &debug.Module{Path: "../dependency"},
					},
				},
			},
			modulePath: "example.com/dependency",
			want:       versionEmpty,
		},
		{
			name: "忽略空依赖并返回默认值",
			buildInfo: &debug.BuildInfo{
				Deps: []*debug.Module{nil},
			},
			modulePath: "example.com/dependency",
			want:       versionEmpty,
		},
		{
			name: "模块不存在时返回默认值",
			buildInfo: &debug.BuildInfo{
				Main: debug.Module{Path: "example.com/main", Version: "v1.2.3"},
			},
			modulePath: "example.com/dependency",
			want:       versionEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testx.Equal(t, tt.want, getModuleVersion(tt.buildInfo, tt.modulePath))
		})
	}
}
