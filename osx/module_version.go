package osx

import "runtime/debug"

// GetModuleVersion 返回指定模块在当前程序构建信息中的版本。
// 模块使用 replace 时返回替换模块的版本；无法取得版本时返回 v0.0.0。
func GetModuleVersion(name string) string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return versionEmpty
	}
	return getModuleVersion(bi, name)
}

func getModuleVersion(bi *debug.BuildInfo, name string) string {
	if bi == nil {
		return versionEmpty
	}
	if bi.Main.Path == name {
		return normalizeModuleVersion(&bi.Main)
	}
	for _, dep := range bi.Deps {
		if dep != nil && dep.Path == name {
			return normalizeModuleVersion(dep)
		}
	}
	return versionEmpty
}

func normalizeModuleVersion(module *debug.Module) string {
	if module.Replace != nil {
		module = module.Replace
	}
	if module.Version == "" || module.Version == versionDevel {
		return versionEmpty
	}
	return module.Version
}
