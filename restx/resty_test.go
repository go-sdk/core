package restx

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"

	"github.com/go-sdk/core/osx"
	"github.com/go-sdk/core/testx"
)

var _ resty.Logger = (*logger)(nil)

func TestNewClientConfiguration(t *testing.T) {
	client := New()
	httpClient := client.GetClient()

	testx.Equal(t, 5*time.Minute, httpClient.Timeout)
	testx.NotNil(t, httpClient.Jar)

	transport, ok := httpClient.Transport.(*http.Transport)
	testx.True(t, ok)
	testx.NotNil(t, transport.Proxy)
	testx.NotNil(t, transport.DialContext)
	testx.True(t, transport.ForceAttemptHTTP2)
	testx.Equal(t, 100, transport.MaxIdleConns)
	testx.Equal(t, runtime.GOMAXPROCS(0)+1, transport.MaxIdleConnsPerHost)
	testx.Equal(t, 90*time.Second, transport.IdleConnTimeout)
	testx.Equal(t, 10*time.Second, transport.TLSHandshakeTimeout)
	testx.Equal(t, time.Second, transport.ExpectContinueTimeout)
}

func TestNewClientDebug(t *testing.T) {
	client := New()
	testx.Equal(t, osx.IsDebug(), client.Debug)
}

func TestRequestAndCookieJar(t *testing.T) {
	type observedRequest struct {
		method string
		header string
		query  string
		cookie string
	}

	requests := make(chan observedRequest, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observed := observedRequest{
			method: r.Method,
			header: r.Header.Get("X-Test"),
			query:  r.URL.Query().Get("query"),
		}
		if cookie, err := r.Cookie("session"); err == nil {
			observed.cookie = cookie.Value
			requests <- observed
			w.WriteHeader(http.StatusNoContent)
			return
		}

		requests <- observed
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "cookie-value", Path: "/"})
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("已创建"))
	}))
	defer server.Close()

	client := New()
	t.Cleanup(client.GetClient().CloseIdleConnections)

	first, err := client.R().
		SetHeader("X-Test", "header-value").
		SetQueryParam("query", "query-value").
		Get(server.URL)
	testx.NoError(t, err)
	testx.Equal(t, http.StatusCreated, first.StatusCode())
	testx.Equal(t, "已创建", first.String())
	firstRequest := <-requests
	testx.Equal(t, http.MethodGet, firstRequest.method)
	testx.Equal(t, "header-value", firstRequest.header)
	testx.Equal(t, "query-value", firstRequest.query)
	testx.Empty(t, firstRequest.cookie)

	second, err := client.R().
		SetHeader("X-Test", "header-value").
		SetQueryParam("query", "query-value").
		Get(server.URL)
	testx.NoError(t, err)
	testx.Equal(t, http.StatusNoContent, second.StatusCode())
	secondRequest := <-requests
	testx.Equal(t, http.MethodGet, secondRequest.method)
	testx.Equal(t, "header-value", secondRequest.header)
	testx.Equal(t, "query-value", secondRequest.query)
	testx.Equal(t, "cookie-value", secondRequest.cookie)
}
