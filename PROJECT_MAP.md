# 项目地图

## 项目定位

`github.com/go-sdk/core` 是个人 Go 基础类库，集中提供错误处理、日志、环境与版本信息、HTTP 客户端、序列生成和测试辅助能力。

## 目录结构

```text
core/
├── .github/workflows/golang.yml     持续集成与 Tag Release
├── .editorconfig                    编辑器格式规范
├── errx/                            错误创建、包装、解包和判断
├── logx/                            进程级全局日志及 zerolog 配置
├── osx/                             调试状态、环境变量和构建版本信息
├── restx/                           预配置的 resty HTTP 客户端
├── seq/                             Snowflake ID 和 UUID v7 生成
├── testx/                           测试断言和输出辅助
├── AGENTS.md                        仓库协作与修改规范
├── PROJECT_MAP.md                   项目结构与调用关系
├── README.md                        使用说明与公共行为
├── Makefile                         版本注入、代码检查和命令帮助
└── go.mod                           Go 模块和依赖定义
```

## 包说明

### `errx`

- 基于 `github.com/rotisserie/eris`。
- 暴露错误创建、格式化、包装、解包、根因和类型判断的快捷入口。
- 提供需要非空错误占位时使用的 `Nil` 哨兵值。

### `logx`

- 基于 `github.com/rs/zerolog`。
- 包初始化时建立默认全局日志，并同步接管 zerolog、`log/slog` 和标准库 `log`。
- 应用读取配置后可以再次调用 `Init`，将日志同时写入标准输出和滚动文件。
- 文件日志使用异步 Writer，所有文件 Writer 由进程统一通过 `Init` 和 `Close` 管理。
- `logx/log.go` 来源于 zerolog 上游，不在本仓库中修改。

### `osx`

- `IsDebug` 根据测试参数或 `DEBUG` 环境变量判断调试模式。
- `GetEnv` 按名称优先级读取第一个已设置的环境变量；空值或转换失败也不回退，仅在全部未设置时使用默认值。
- `Hostname` 返回当前主机名，`WorkDir` 和 `Exe*` 返回工作目录及可执行文件路径信息。
- `WithExeExt` 返回替换扩展名后的可执行文件路径，扩展名缺少前导点时自动补齐。
- `GetVersion` 读取并缓存 Go 构建信息和版本控制信息，版本号可由构建参数注入。

### `restx`

- 基于 `github.com/go-resty/resty/v2`。
- 创建统一配置的 HTTP 客户端，支持环境代理、HTTP/2、连接池和 Cookie Jar。
- 调试模式下启用 resty 请求调试日志，内部日志由 `logx` 输出。

### `seq`

- 使用 Sonyflake 生成十进制字符串 ID，机器编号固定为 `56565`。
- 固定机器编号要求同一时间只运行一个生成器实例。
- 使用 `google/uuid` 生成 UUID v7，并在包初始化时启用随机池。

### `testx`

- 集中暴露常用的 `testify/require` 断言。
- `P` 用于检查首个错误参数，并以易读格式记录其余参数。

## 主要依赖关系

```text
restx ──> logx ──> osx
errx  ──> eris
seq   ──> sonyflake、google/uuid
testx ──> testify/require、kr/pretty
```

各业务包的测试可以依赖 `testx`，生产包不依赖 `testx`。

## 自动化流程

- 推送到 `master` 分支时整理并校验依赖文件，然后运行代码检查和全部测试。
- 推送 `v*` Tag 时先运行相同检查，全部通过后创建 GitHub Release。
- Release 使用对应的远端 Tag，并由 GitHub 自动生成标题和发布说明，不额外上传制品。

## 测试结构

- 单元测试与被测代码使用相同包名，覆盖公共行为、既定配置和关键边界。
- `logx` 测试验证全局日志包装、两阶段初始化、文件刷新以及标准日志接管。
- `osx` 测试验证调试模式、环境变量优先级、主机与路径信息、扩展名替换以及构建版本序列化。
- `restx` 使用本地 `httptest` 服务验证客户端配置、调试模式、请求和 Cookie Jar，不访问外部接口。
- `seq` 同时验证格式、固定机器编号、唯一性和进程内并发安全。
- `testx` 使用内部测试替身覆盖成功与失败分支，避免依赖自身断言验证核心流程。
