# 项目地图

## 项目定位

`github.com/go-sdk/core` 是个人 Go 基础类库，集中提供错误处理、生命周期、日志、环境与版本信息、HTTP 客户端、序列生成和测试辅助能力。

## 目录结构

```text
core/
├── .github/workflows/golang.yml     持续集成与 Tag Release
├── .editorconfig                    编辑器格式规范
├── cmdx/                            cobra 根命令创建和入口包装
├── errx/                            错误创建、包装、解包和判断
├── lifex/                           全局信号、初始化和解构管理
├── logx/                            进程级全局日志及 zerolog 配置
├── osx/                             调试状态、环境变量和构建版本信息
├── restx/                           预配置的 resty HTTP 客户端
├── seq/                             Snowflake ID 和 UUID v7 生成
├── testx/                           测试断言和输出辅助
├── empty.go                         根包占位文件，保证模块路径可被直接引用
├── AGENTS.md                        仓库协作与修改规范
├── PROJECT_MAP.md                   项目结构与调用关系
├── README.md                        使用说明与公共行为
├── Makefile                         版本注入、代码检查和命令帮助
└── go.mod                           Go 模块和依赖定义
```

## 包说明

### `cmdx`

- 基于 `github.com/spf13/cobra`，`Command` 是 `cobra.Command` 的类型别名。
- `NewRoot` 创建隐藏 help 和 completion 子命令的根命令，版本描述来自 `osx.GetVersion`。
- cobra 自身的错误输出被丢弃，命令错误由 `Execute` 返回，调用方自行处理。
- `WrapRunE` 将 `errx.Nil` 哨兵错误视为无错误，其余错误原样返回。

### `errx`

- 基于 `github.com/rotisserie/eris`。
- 暴露错误创建、格式化、包装、解包、根因和类型判断的快捷入口。
- 提供需要非空错误占位时使用的 `Nil` 哨兵值。

### `lifex`

- 管理进程级生命周期：`OnInit` 和 `OnDeinit` 注册初始化和解构函数，`Init` 按注册顺序执行初始化，`Wait` 阻塞等待退出后按注册逆序执行解构。
- 退出由 SIGINT/SIGTERM 信号或 `Shutdown` 触发；信号触发视为正常退出，主动退出返回 `Shutdown` 携带的原因。
- `Shutdown` 幂等，多次调用只触发一次退出；解构期间的第二个退出信号跳过剩余解构强制退出。
- 解构函数自身的错误经 `logx` 记录，不影响其余解构执行和 `Wait` 的返回值。

### `logx`

- 基于 `github.com/rs/zerolog`。
- 控制台输出经 go-colorable 包装标准输出，Windows 终端下颜色转义可正常显示。
- 包初始化时建立默认全局日志，并同步接管 zerolog、`log/slog` 和标准库 `log`；默认日志文件路径来自 `LOGX_FILE_PATH`。
- 应用读取配置后可以再次调用 `Init`，将日志同时写入标准输出和滚动文件。
- 文件滚动参数读取 `LOGX_FILE_SIZE`、`LOGX_FILE_AGE`、`LOGX_FILE_BACKUPS`、`LOGX_FILE_LOCALTIME`、`LOGX_FILE_COMPRESS` 环境变量。
- 文件日志使用异步 Writer，所有文件 Writer 由进程统一通过 `Init` 和 `Close` 管理。
- `logx/log.go` 来源于 zerolog 上游，不在本仓库中修改。

### `osx`

- `IsDebug` 根据测试参数或 `DEBUG` 环境变量判断调试模式。
- `Panic` 和 `Panicf` 将给定值或格式化消息及从调用方开始的易读调用堆栈打印到标准错误输出，再调用内置 `panic` 保持原有行为。
- `GetEnv` 按名称优先级读取第一个已设置的环境变量；空值或转换失败也不回退，仅在全部未设置时使用默认值。
- `Hostname` 返回当前主机名，`WorkDir` 和 `Exe*` 返回工作目录及可执行文件路径信息。
- `WithExeExt` 返回替换扩展名后的可执行文件路径，扩展名缺少前导点时自动补齐。
- `GetVersion` 读取并缓存 Go 构建信息和版本控制信息，版本号可由构建参数注入。

### `restx`

- 基于 `github.com/go-resty/resty/v2`。
- 创建统一配置的 HTTP 客户端，支持环境代理、HTTP/2、连接池和 Cookie Jar。
- 调试模式下启用 resty 请求调试日志，内部日志由 `logx` 输出。

### `seq`

- 使用 Sonyflake 生成十进制字符串 ID，起始年份和机器编号读取 `SONYFLAKE_START_YEAR`、`SONYFLAKE_MACHINE_ID` 环境变量（默认 2020 和 56565），初始化失败直接 panic。
- 机器编号约束为同一时间只运行一个生成器实例。
- 使用 `google/uuid` 生成 UUID v7，并在包初始化时启用随机池。

### `testx`

- 集中暴露常用的 `testify/require` 断言。
- `P` 用于检查首个错误参数，并以易读格式记录其余参数。

## 主要依赖关系

```text
restx ──> logx ──> osx
cmdx  ──> cobra、errx、osx
lifex ──> logx
errx  ──> eris
seq   ──> sonyflake、google/uuid、osx
testx ──> testify/require、kr/pretty
```

各业务包的测试可以依赖 `testx`，生产包不依赖 `testx`。

## 自动化流程

- 推送到 `master` 分支时整理并校验依赖文件，然后运行代码检查和全部测试。
- 推送 `v*` Tag 时先运行相同检查，全部通过后创建 GitHub Release。
- Release 使用对应的远端 Tag，并由 GitHub 自动生成标题和发布说明，不额外上传制品。

## 测试结构

- 单元测试与被测代码使用相同包名，覆盖公共行为、既定配置和关键边界。
- `cmdx` 测试验证根命令默认配置以及 `WrapRunE` 对哨兵错误和真实错误的处理。
- `lifex` 测试验证初始化顺序与失败中断、解构逆序与错误隔离、`Shutdown` 幂等与并发 `Wait`，并通过子进程重入验证信号触发退出和第二次信号强制退出，不访问外部资源。
- `logx` 测试验证全局日志包装、两阶段初始化、文件刷新以及标准日志接管。
- `osx` 测试验证调试模式、环境变量优先级、主机与路径信息、扩展名替换、构建版本序列化以及 Panic/Panicf 的堆栈输出与 panic 透传。
- `restx` 使用本地 `httptest` 服务验证客户端配置、调试模式、请求和 Cookie Jar，不访问外部接口。
- `seq` 同时验证格式、默认机器编号、唯一性和进程内并发安全。
- `testx` 使用内部测试替身覆盖成功与失败分支，避免依赖自身断言验证核心流程。
