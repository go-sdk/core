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
| `cmdx`  | 创建 cobra 根命令并包装命令入口    |
| `errx`  | 创建、包装和判断错误               |
| `logx`  | 配置并使用进程级全局日志           |
| `osx`   | 读取系统、路径、环境变量和构建信息 |
| `restx` | 创建带统一默认配置的 resty 客户端  |
| `seq`   | 生成 Snowflake ID 和 UUID v7       |
| `testx` | 提供常用测试断言和输出辅助         |

## 使用示例

### 命令行

```go
root := cmdx.NewRoot("app")
root.AddCommand(&cmdx.Command{
	Use: "run",
	RunE: cmdx.WrapRunE(func(cmd *cmdx.Command, args []string) error {
		return nil
	}),
})

if err := root.Execute(); err != nil {
	logx.Error().Err(err).Msg("执行失败")
}
```

`NewRoot` 隐藏 help 和 completion 子命令，`--version` 输出来自 `osx.GetVersion` 的单行版本描述。cobra 自身的错误输出被丢弃，命令错误由 `Execute` 直接返回，调用方自行处理；`WrapRunE` 将 `errx.Nil` 哨兵错误视为无错误，其余错误原样返回。

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

包初始化时若设置了 `LOGX_FILE_PATH`，默认日志会同时写入该滚动文件。文件滚动参数通过环境变量配置：

| 环境变量              | 含义             | 默认值 |
|-----------------------|------------------|--------|
| `LOGX_FILE_PATH`      | 默认日志文件路径 | 空     |
| `LOGX_FILE_SIZE`      | 单文件大小（MB） | `30`   |
| `LOGX_FILE_AGE`       | 旧文件保留天数   | `90`   |
| `LOGX_FILE_BACKUPS`   | 旧文件保留数量   | `10`   |
| `LOGX_FILE_LOCALTIME` | 使用本地时间轮转 | `true` |
| `LOGX_FILE_COMPRESS`  | 压缩旧文件       | `true` |

变量为空或无法转换时遵循 `osx.GetEnv` 语义返回零值而非默认值，例如 `LOGX_FILE_SIZE` 为空时单文件大小为零，由 lumberjack 回退到其自身的 100MB 默认值。

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
exeLog := osx.WithExeExt("log")
```

主机名和工作目录在调用时读取；可执行文件完整路径在包初始化时读取一次并缓存。`WithExeExt` 返回将可执行文件扩展名替换为指定扩展名的完整路径，常用于生成与程序同名的日志等辅助文件，扩展名缺少前导点时自动补齐。

### HTTP 客户端

```go
client := restx.New()
response, err := client.R().Get("https://example.com")
```

客户端默认支持环境代理、HTTP/2、连接池和 Cookie Jar，请求总超时为五分钟。调试模式下（判定规则见 `osx.IsDebug`）resty 会输出请求与响应调试日志，内部日志统一由 `logx` 输出。

### 序列生成

```go
id := seq.NextID()
uuid := seq.UUID()
shortUUID := seq.UUIDShort()
```

Snowflake 生成器的起始年份默认 `2020`，机器编号默认 `56565`，可分别通过 `SONYFLAKE_START_YEAR` 和 `SONYFLAKE_MACHINE_ID` 环境变量覆盖（机器编号取值范围为 `[0, 65535]`，起始年份须早于当前时间）。机器编号的唯一性范围为同一时间只运行一个生成器实例；非法取值会在包初始化时 panic。UUID 使用 v7 格式，并在包初始化时启用随机池。

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
