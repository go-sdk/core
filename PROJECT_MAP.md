# 项目地图

## 项目定位

`github.com/go-sdk/core` 是个人 Go 基础类库，集中提供配置加载、错误处理、生命周期、日志、环境与版本信息、HTTP 客户端、JSON 与 YAML 编解码、零拷贝类型转换、序列生成和测试辅助能力。

## 目录结构

```text
core/
├── .github/workflows/golang.yml     持续集成与 Tag Release
├── .editorconfig                    编辑器格式规范
├── cmdx/                            cobra 根命令创建和入口包装
├── codec/json/                     基于 encoding/json/v2 的泛型 JSON 编解码
├── codec/yaml/                     基于 go.yaml.in/yaml/v3 的泛型 YAML 编解码
├── config/                         YAML/JSON、环境变量与文件监听配置
├── conv/                           string 与 []byte 零拷贝互转
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

### `config`

- 使用 `.` 分隔嵌套路径，分别保存扁平数据和嵌套数据；`Get`、`MustGet` 和 `Exists` 查询扁平数据，`Raw` 返回嵌套数据的深拷贝。
- `WithFile` 指定 YAML 或 JSON 文件，并可通过第二个可选参数显式指定 `yaml` 或 `json`；`WithFileWatch` 默认关闭，启用后由 `lifex` 统一关闭文件监听器。
- `Load` 先读取文件，再将 `APP__` 开头的环境变量转换为小写路径并覆盖文件值，最后解析 `${key}` 引用；引用使用与 `Get` 相同的路径，并检测缺失引用和循环引用。
- `DecodeTo` 使用 `json` tag 和弱类型转换将嵌套数据解码到目标值。
- 包初始化时解析默认配置文件，任何失败直接 panic：`CONFIG_PATH` 一经设置即直接采用该路径且不回退，空值、文件不存在或不可读均视为失败；未设置时按测试模块根目录 `config.yaml`、可执行文件同名的 `.yaml`、`.yml`、`.json` 顺序选择第一个存在的文件，全部不存在时仅加载环境变量。
- `SetDefault` 替换包级 `Get` 和 `MustGet` 使用的默认实例。

### `codec/json`

- 基于 `encoding/json/v2`，`Marshal` 和 `Unmarshal` 通过泛型参数支持 `string` 与 `[]byte` 两种载体。
- 载体转换经 `conv` 零拷贝完成，`[]byte` 路径直接复用原切片。
- `Must` 前缀版本在出错时经 `osx.Panic` 直接 panic。

### `codec/yaml`

- 基于 `go.yaml.in/yaml/v3`，方法与 `codec/json` 对应，同样通过泛型参数支持 `string` 与 `[]byte` 两种载体；yaml/v3 没有编解码选项，因此不接受额外参数。
- 载体转换经 `conv` 零拷贝完成，`[]byte` 路径直接复用原切片。
- `Must` 前缀版本在出错时经 `osx.Panic` 直接 panic。
- 上游对 `chan`、`func` 等无法表示的类型直接 panic 而非返回错误。

### `conv`

- 提供 `StringToBytes` 和 `BytesToString`，基于 unsafe 实现 string 与 []byte 的零拷贝互转。
- 空输入返回对应零值（nil 或空串）；由 string 转出的 []byte 底层只读，不可修改。

### `errx`

- 基于 `github.com/rotisserie/eris`，暴露错误创建、格式化、包装、解包、根因和类型判断的快捷入口。
- `Join`、`Is`、`As` 和 `AsType` 使用标准库 `errors` 实现，可以匹配和提取 `Join` 合并的多错误链；`Unwrap` 和 `Cause` 只沿单链展开，遇到 `Join` 合并的错误时停在该节点。
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
- `SetGlobalKV`、`DeleteGlobalKV` 和 `ClearGlobalKV` 维护进程级全局键值，仅用于 service、version 等进程级标识，并通过 Hook 附加到之后所有日志事件；键值对所有 `New` 创建的 Logger 生效。
- 全局键值只在写入时加锁，日志输出通过不可变快照无锁读取，序列化阶段不持有任何锁；全局键与事件字段同名会产生重复 JSON key，调用方应保证不冲突。
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
restx ───> logx ───> osx
config -> codec/json、codec/yaml、lifex、logx、osx
codec/json -> conv、osx
codec/yaml -> conv、osx
cmdx  ───> cobra、errx、osx
lifex ───> logx
errx  ───> eris
seq   ───> sonyflake、google/uuid、osx
testx ───> testify/require、kr/pretty
```

各业务包的测试可以依赖 `testx`，生产包不依赖 `testx`。

## 自动化流程

- 推送到 `master` 分支时整理并校验依赖文件，然后运行代码检查和全部测试。
- 推送 `v*` Tag 时先运行相同检查，全部通过后创建 GitHub Release。
- Release 使用对应的远端 Tag，并由 GitHub 自动生成标题和发布说明，不额外上传制品。

## 测试结构

- 单元测试与被测代码使用相同包名，覆盖公共行为、既定配置和关键边界。
- `cmdx` 测试验证根命令默认配置以及 `WrapRunE` 对哨兵错误和真实错误的处理。
- `config` 测试验证 YAML/JSON 文件、显式文件类型、环境变量覆盖、嵌套 Raw、json tag 解码、默认实例、变量替换及循环检测和文件监听重载。
- `codec/json` 测试验证 `string` 与 `[]byte` 两种载体的序列化、反序列化和 `Must` 版本的错误路径。
- `codec/yaml` 测试验证 `string` 与 `[]byte` 两种载体的序列化、反序列化和 `Must` 版本的错误路径。
- `conv` 测试验证空输入零值语义和常规互转结果。
- `errx` 测试验证创建、包装、根因、类型匹配、多错误合并和 `Nil` 哨兵行为。
- `lifex` 测试验证初始化顺序与失败中断、解构逆序与错误隔离、`Shutdown` 幂等与并发 `Wait`，并通过子进程重入验证信号触发退出和第二次信号强制退出，不访问外部资源。
- `logx` 测试验证全局日志包装、两阶段初始化、文件刷新、标准日志接管以及全局键值的附加、覆盖、删除、清空、同名冲突行为、序列化重入和并发安全。
- `osx` 测试验证调试模式、环境变量优先级、主机与路径信息、扩展名替换、构建版本序列化以及 Panic/Panicf 的堆栈输出与 panic 透传。
- `restx` 使用本地 `httptest` 服务验证客户端配置、调试模式、请求和 Cookie Jar，不访问外部接口。
- `seq` 同时验证格式、默认机器编号、唯一性和进程内并发安全。
- `testx` 使用内部测试替身覆盖成功与失败分支，避免依赖自身断言验证核心流程。
