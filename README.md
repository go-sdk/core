# core

`core` 是个人使用的 Go 基础类库，模块路径为 `github.com/go-sdk/core`。项目用于统一常见基础能力和默认行为，包括错误处理、全局日志、环境变量、构建版本、HTTP 客户端、序列生成和测试辅助。

## 环境要求

- Go 1.26 或更高版本

## 安装

```bash
go get github.com/go-sdk/core
```

## 包概览

| 包      | 用途                               |
|---------|------------------------------------|
| `errx`  | 创建、包装和判断错误               |
| `logx`  | 配置并使用进程级全局日志           |
| `osx`   | 读取系统、路径、环境变量和构建信息 |
| `restx` | 创建带统一默认配置的 resty 客户端  |
| `seq`   | 生成 Snowflake ID 和 UUID v7       |
| `testx` | 提供常用测试断言和输出辅助         |

## 使用示例

### 错误处理

```go
cause := errx.New("操作失败")
err := errx.Wrap(cause, "保存数据")

if errx.Is(err, cause) {
	// 处理目标错误
}
```

### 全局日志

`logx` 在包初始化时创建输出到标准输出的默认日志。应用读取配置后可以再次初始化，将日志同时写入滚动文件：

```go
logx.Init("logs/app.log")
defer logx.Close()

logx.Info().Str("service", "example").Msg("服务已启动")
```

重新调用 `Init` 会先刷新并关闭此前创建的文件 Writer。日志资源按进程统一管理，不使用独立 Logger 生命周期。

### 环境变量

```go
port := osx.GetEnv(8080, "APP_PORT", "PORT")
debug := osx.GetEnv(false, "APP_DEBUG", "DEBUG")
```

`GetEnv` 按名称顺序读取第一个已设置的环境变量。变量即使为空或无法转换也会立即采用，不再尝试后续名称；转换失败时返回目标类型的零值。只有全部变量都未设置时才返回默认值。

### 系统与路径信息

```go
hostname := osx.Hostname()
workDir := osx.WorkDir()
exeFull := osx.ExeFull()
exeDir := osx.ExeDir()
exeName := osx.ExeName()
exeExt := osx.ExeExt()
```

主机名和工作目录在调用时读取；可执行文件完整路径在包初始化时读取一次并缓存。

### HTTP 客户端

```go
client := restx.New()
response, err := client.R().Get("https://example.com")
```

客户端默认支持环境代理、HTTP/2、连接池和 Cookie Jar，请求总超时为五分钟。

### 序列生成

```go
id := seq.NextID()
uuid := seq.UUID()
shortUUID := seq.UUIDShort()
```

Snowflake 生成器使用固定机器编号 `56565`，适用于同一时间只运行一个生成器实例的场景。UUID 使用 v7 格式，并在包初始化时启用随机池。

### 测试辅助

```go
func TestExample(t *testing.T) {
	value, err := loadValue()
	testx.P(t, err, value)
	testx.NotEmpty(t, value)
}
```

`P` 将首个参数视为可选错误：参数为 nil 时忽略，为非空错误时终止当前测试，其余参数使用易读格式写入测试日志。

## 开发约定

修改代码前先阅读 `AGENTS.md`、`PROJECT_MAP.md` 和本文件。发现文档与实际代码不一致时，应在同一次修改中自动更新对应文档。

所有新增文档和代码注释使用简体中文。`logx/log.go` 直接来源于 [zerolog 上游实现](https://github.com/rs/zerolog/blob/master/log/log.go)，保持原样，不在本仓库中修改。
