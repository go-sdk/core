package restx

import (
	"net"
	"net/http"
	"net/http/cookiejar"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/net/publicsuffix"
)

// Request 是 resty.Request 的类型别名。
type Request = resty.Request

// New 创建带有统一默认配置的 resty 客户端。
// 客户端支持环境代理、HTTP/2、连接池和 Cookie Jar，请求总超时为五分钟。
func New() *resty.Client {
	resolver := &net.Resolver{
		PreferGo: true,
	}
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Resolver:  resolver,
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
	return resty.NewWithClient(client).SetLogger(&logger{})
}
