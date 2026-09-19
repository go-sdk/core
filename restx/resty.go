package restx

import (
	"net"
	"net/http"
	"net/http/cookiejar"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/net/publicsuffix"

	"github.com/go-sdk/core/osx"
)

// Client 是 resty.Client 的类型别名。
type Client = resty.Client

// Request 是 resty.Request 的类型别名。
type Request = resty.Request

// Response 是 resty.Response 的类型别名。
type Response = resty.Response

var (
	// ErrAutoRedirectDisabled 是 resty 的哨兵错误，禁用重定向（NoRedirectPolicy）后收到重定向响应时返回。
	ErrAutoRedirectDisabled = resty.ErrAutoRedirectDisabled
	// ErrRateLimitExceeded 是 resty 的哨兵错误，通过 SetRateLimiter 设置的限流器拒绝请求时返回。
	ErrRateLimitExceeded = resty.ErrRateLimitExceeded
	// ErrResponseBodyTooLarge 是 resty 的哨兵错误，响应体超过 SetResponseBodyLimit 设置的上限时返回。
	ErrResponseBodyTooLarge = resty.ErrResponseBodyTooLarge
)

// New 创建带有统一默认配置的 resty 客户端。
// 客户端支持环境代理、HTTP/2、连接池和 Cookie Jar，请求总超时为五分钟。
// 调试模式下启用请求调试日志，resty 内部日志统一由 logx 输出。
func New() *resty.Client {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	client := &http.Client{
		Transport: transport,
		Jar:       jar,
		Timeout:   5 * time.Minute,
	}
	return resty.NewWithClient(client).SetDebug(osx.IsDebug()).SetLogger(&logger{})
}
