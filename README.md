# core

`core` 是个人使用的 Go 基础类库，模块路径为 `github.com/go-sdk/core`。项目用于统一常见基础能力和默认行为，包括配置加载、错误处理、生命周期、全局日志、环境变量、构建版本、HTTP 客户端、JSON 与 YAML 编解码、零拷贝类型转换、序列生成和测试辅助。

## 环境要求

- Go 1.27 或更高版本

## 安装

```bash
go get github.com/go-sdk/core
```

## 包概览

| 包           | 用途                                   |
|--------------|----------------------------------------|
| `cmdx`       | 创建 cobra 根命令并包装命令入口        |
| `codec/json` | 基于 encoding/json/v2 的 JSON 编解码   |
| `codec/yaml` | 基于 go.yaml.in/yaml/v3 的 YAML 编解码 |
| `config`     | 加载文件和环境变量配置                 |
| `conv`       | string 与 []byte 零拷贝互转            |
| `errx`       | 创建、包装和判断错误                   |
| `lifex`      | 管理信号、初始化和解构的进程生命周期   |
| `logx`       | 配置并使用进程级全局日志               |
| `osx`        | 读取系统、路径、环境变量和构建信息     |
| `restx`      | 创建带统一默认配置的 resty 客户端      |
| `seq`        | 生成 Snowflake ID 和 UUID v7           |
| `testx`      | 提供常用测试断言和输出辅助             |

## 使用示例

### 配置

`config` 支持 YAML 和 JSON 文件，并使用 `.` 访问嵌套配置：

```yaml
database:
  type: postgres
  port: 5432
dsn: "postgres://${database.type}:${database.port}/app"
```

```go
cfg := config.New(
	config.WithFile("config.yaml"),
	config.WithFileWatch(true),
)
if err := cfg.Load(); err != nil {
	return err
}

databaseType, ok := cfg.Get[string]("database.type")
databasePort := cfg.MustGet[int]("database.port")
databaseHost := cfg.MustGet("database.host", "localhost")
raw := cfg.Raw()
```

`MustGet` 在配置存在时返回转换后的值；配置不存在且传入默认值时返回第一个默认值，未传默认值时触发 panic。配置存在但转换失败时返回目标类型零值，不回退到默认值。

`WithFile` 的第二个参数可以在文件没有标准扩展名时显式指定格式，例如 `config.WithFile("config.data", "json")`。文件监听默认关闭；启用后由 `lifex` 在进程解构时关闭，文件重载失败会保留最后一次有效配置。

环境变量只读取 `APP__` 前缀，移除前缀后将 key 转为小写，并使用 `__` 表示层级。例如 `APP__DATABASE__TYPE=mysql` 覆盖 `database.type`。环境变量优先于文件配置，`${database.type}` 使用与 `Get` 相同的路径；缺失引用或循环引用会使加载失败。

`Raw` 返回嵌套数据的深拷贝。`DecodeTo` 使用 `json` tag，并允许将环境变量字符串弱类型转换到目标字段；字符串会按 Go duration 格式解析到 `time.Duration`，按 RFC3339 格式解析到 `time.Time`：

```go
type AppConfig struct {
	Database struct {
		Type string `json:"type"`
		Port int    `json:"port"`
	} `json:"database"`
}

var appConfig AppConfig
if err := cfg.DecodeTo(&appConfig); err != nil {
	return err
}
```

包级默认实例在初始化时解析配置文件，任何失败都会直接 panic。`CONFIG_PATH` 一经设置即直接采用该路径且不回退，空值、文件不存在或不可读均视为失败；未设置 `CONFIG_PATH` 时，按测试模块根目录 `config.yaml`、可执行文件同名的 `.yaml`、`.yml`、`.json` 顺序读取第一个存在的文件，全部不存在时仅加载环境变量。默认实例可通过 `SetDefault` 替换：

```go
databaseType, ok := config.Get[string]("database.type")
databasePort := config.MustGet[int]("database.port")
databaseHost := config.MustGet("database.host", "localhost")
hasDatabase := config.Exists("database")
raw := config.Raw()

config.SetDefault(cfg)

var appConfig AppConfig
if err := config.DecodeTo(&appConfig); err != nil {
	return err
}
```

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

### JSON 编解码

```go
s, err := json.Marshal[string](data)
bs, err := json.Marshal[[]byte](data)

var out Data
err = json.Unmarshal(s, &out)
value := json.MustUnmarshal[Data](bs)
```

`Marshal` 和 `Unmarshal` 通过泛型参数支持 `string` 与 `[]byte` 两种载体，内部经 `conv` 零拷贝互转；`Must` 前缀版本在出错时直接 panic。

### YAML 编解码

```go
s, err := yaml.Marshal[string](data)
bs, err := yaml.Marshal[[]byte](data)

var out Data
err = yaml.Unmarshal(s, &out)
value := yaml.MustUnmarshal[Data](bs)
```

`codec/yaml` 的方法与 `codec/json` 一一对应，同样通过泛型参数支持 `string` 与 `[]byte` 两种载体并复用 `conv` 零拷贝互转；yaml/v3 没有编解码选项，因此不接受额外参数。上游对 `chan`、`func` 等无法表示的类型会直接 panic 而非返回错误。

### 零拷贝转换

```go
bs := conv.StringToBytes("starudream")
s := conv.BytesToString(bs)
```

`StringToBytes` 和 `BytesToString` 基于 unsafe 实现零拷贝互转，空输入返回对应零值（nil 或空串）；由 string 转出的 []byte 底层只读，不可修改。

### 错误处理

```go
cause := errx.New("操作失败")
err := errx.Wrap(cause, "保存数据")

if errx.Is(err, cause) {
	// 处理目标错误
}

joined := errx.Join(errA, errB)
if errx.Is(joined, errA) {
	// 命中合并中的任一错误
}

target, ok := errx.AsType[*MyError](joined)
```

`New`、`Newf`、`Wrap`、`Wrapf`、`Unwrap` 和 `Cause` 基于 eris；`Join`、`Is`、`As` 和 `AsType` 使用标准库 `errors`，其中 `Is` 和 `As` 可以遍历 `Join` 合并的多错误链，`AsType` 通过类型参数提取指定类型的错误。`Unwrap` 和 `Cause` 只沿单链展开，遇到 `Join` 合并的错误时停在该节点。`As` 沿用标准库契约，目标必须是合法的非空指针。

### 生命周期

```go
initialize := func() error {
	// 建立资源连接
	return nil
}

deinitialize := func() {
	// 释放资源
}

lifex.OnInit(initialize)
lifex.OnDeinit(deinitialize)

if err := lifex.Init(); err != nil {
	logx.Error().Err(err).Msg("初始化失败")
	return
}

if err := lifex.Wait(); err != nil {
	logx.Error().Err(err).Msg("进程异常退出")
}
```

`OnInit` 和 `OnDeinit` 均接受 `func()`、`func() error`、`func(context.Context)` 与 `func(context.Context) error`，普通函数、方法值和函数字面量都可以直接注册；带上下文参数时传入 `context.Background()`。`Init` 按注册顺序执行初始化函数，任一失败立即停止后续执行并返回该错误。`Wait` 阻塞等待 SIGINT/SIGTERM 信号或 `Shutdown` 触发退出，然后按注册逆序执行解构函数：信号触发视为正常退出返回 nil，主动退出返回 `Shutdown` 携带的原因；`Shutdown` 幂等，多次调用只触发一次退出。解构函数自身的错误经默认 `slog` 记录，不影响其余解构执行和 `Wait` 的返回值；解构期间再次收到退出信号时跳过剩余解构强制退出，避免解构卡死导致进程无法终止。

### 全局日志

`logx` 在包初始化时创建输出到标准输出的默认日志。应用读取配置后可以再次初始化，将日志同时写入滚动文件：

```go
logx.Init("logs/app.log")

logx.Info().Str("service", "example").Msg("服务已启动")
```

重新调用 `Init` 会先刷新并关闭此前创建的文件 Writer。`logx` 会向 `lifex` 注册关闭函数，使用 `lifex.Wait` 时在进程解构阶段自动刷新并关闭文件 Writer；未使用 `lifex` 的程序仍应在退出前调用 `logx.Close()`。`Close` 后全局日志保留控制台输出，供生命周期完成日志和后续诊断使用。日志资源按进程统一管理，不使用独立 Logger 生命周期。控制台输出经 go-colorable 包装，在 Windows 终端下也能正常显示颜色。

`SetGlobalKV` 维护进程级全局键值，之后所有日志事件都会附带该键值，适合注入 `service`、`version`、`instance`、`environment` 等进程级标识：

```go
logx.SetGlobalKV("service", "example")
logx.SetGlobalKV("version", "v1.0.0")

logx.Info().Msg("服务已启动")
// {"level":"info","service":"example","version":"v1.0.0","message":"服务已启动"}

logx.DeleteGlobalKV("version") // 删除指定键
logx.ClearGlobalKV()           // 清空全部键
```

全局键值只在写入时加锁，日志输出通过不可变快照无锁读取，可并发调用；对所有由 `logx.New` 创建的 Logger 生效，不影响调用方自行创建的 zerolog Logger。全局键与事件字段同名时会产生重复 JSON key，调用方应保证全局键不与日志字段冲突。

trace、request、user 等请求级字段会随请求并发变化，不应使用全局键值，应通过上下文 Logger 传递：

```go
logger := logx.With().Str("trace", "abc123").Logger()
ctx := logger.WithContext(context.Background())

logx.Ctx(ctx).Info().Msg("处理请求")
```

默认配置加载前会先安装 zerolog 控制台 Handler，因此 `config` 使用的 `slog`、zerolog 和标准库 `log` 从启动阶段起采用相同的时间、级别和字段格式。`logx` 随后读取默认 `config` 实例，默认日志参数可以写入配置文件，也可以通过 `APP__` 环境变量覆盖：

| 配置键               | 环境变量                    | 含义             | 默认值   |
|----------------------|-----------------------------|------------------|----------|
| `log.file.path`      | `APP__LOG__FILE__PATH`      | 默认日志文件路径 | 空       |
| `log.no_color`       | `APP__LOG__NO_COLOR`        | 禁用控制台颜色   | 自动检测 |
| `log.file.size`      | `APP__LOG__FILE__SIZE`      | 单文件大小（MB） | `30`     |
| `log.file.age`       | `APP__LOG__FILE__AGE`       | 旧文件保留天数   | `90`     |
| `log.file.backups`   | `APP__LOG__FILE__BACKUPS`   | 旧文件保留数量   | `10`     |
| `log.file.localtime` | `APP__LOG__FILE__LOCALTIME` | 使用本地时间轮转 | `true`   |
| `log.file.compress`  | `APP__LOG__FILE__COMPRESS`  | 压缩旧文件       | `true`   |

配置项存在但为空或无法转换时返回目标类型零值，不使用默认值。例如 `APP__LOG__FILE__SIZE` 为空时单文件大小为零，由 lumberjack 回退到其自身的 100MB 默认值。

`log.no_color` 未配置时根据标准输出是否为支持颜色的终端自动决定；`TERM=dumb` 或标准输出被重定向时默认禁用颜色。显式配置始终优先。

### 环境变量

```go
port := osx.GetEnv(8080, "APP_PORT", "PORT")
debug := osx.GetEnv(false, "APP_DEBUG", "DEBUG")
```

`GetEnv` 按名称顺序读取第一个已设置的环境变量。变量即使为空或无法转换也会立即采用，不再尝试后续名称；转换失败时返回目标类型的零值。只有全部变量都未设置时才返回默认值。

### 异常退出

```go
osx.Panic("unexpected state")
osx.Panicf("unexpected state: %s", reason)
```

`Panic` 先将给定值和从调用方开始的调用堆栈以易读格式打印到标准错误输出，再调用内置 `panic` 保持原有行为；`Panicf` 按 format 格式化参数后行为一致。

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

Snowflake 生成器读取 `sonyflake.start_year` 和 `sonyflake.machine_id`，默认值分别为 `2020` 和 `56565`；环境变量写法为 `APP__SONYFLAKE__START_YEAR` 和 `APP__SONYFLAKE__MACHINE_ID`（机器编号取值范围为 `[0, 65535]`，起始年份须早于当前时间）。机器编号的唯一性范围为同一时间只运行一个生成器实例；非法取值会在包初始化时 panic。UUID 使用 v7 格式，并在包初始化时启用随机池。

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
