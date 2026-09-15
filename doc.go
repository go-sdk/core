// Package core 是个人使用的 Go 基础类库，统一常见基础能力和默认行为，
// 包括配置加载、错误处理、进程生命周期、全局日志、环境变量、HTTP 客户端、
// JSON 与 YAML 编解码、序列生成和测试辅助。
//
// 根包本身不提供 API，各能力位于以下子包：
//
//	cmdx        创建 cobra 根命令并包装命令入口
//	config      加载 YAML/JSON 文件和 APP__ 前缀环境变量配置
//	conv        string 与 []byte 零拷贝互转
//	errx        创建、包装和判断错误
//	lifex       管理信号、初始化和解构的进程生命周期
//	logx        进程级全局日志，支持滚动文件和全局键值
//	osx         读取系统、路径、环境变量和构建版本信息
//	restx       创建带统一默认配置的 resty HTTP 客户端
//	seq         生成 Snowflake ID 和 UUID v7
//	testx       提供常用测试断言和输出辅助
//	codec/json  基于 encoding/json/v2 的 JSON 编解码
//	codec/yaml  基于 go.yaml.in/yaml/v3 的 YAML 编解码
//
// 安装：
//
//	go get github.com/go-sdk/core
//
// 各子包的详细用法和配置项说明见项目 README。
package core
