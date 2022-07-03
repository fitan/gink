package http

import (
	"github.com/go-kit/kit/sd"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"net"
	"net/http"
	"runtime"
	"time"
)

func WithClientHost(host string) ClientOption {
	return func(client *Client) {
		client.client.SetBaseURL(host)
	}
}

func WithClientTimeout(timeout time.Duration) ClientOption {
	return func(client *Client) {
		client.client.SetTimeout(timeout)
	}
}

func WithClientRetry(retryCount int, retryWaitTime, retryMaxWaitTime time.Duration) ClientOption {
	return func(client *Client) {
		client.client.SetRetryCount(retryCount).SetRetryWaitTime(retryWaitTime).SetRetryMaxWaitTime(retryMaxWaitTime)
	}
}

func WithClientKitLb(instance sd.Instancer, seed int64) ClientOption {
	return func(client *Client) {
		client.client.OnBeforeRequest(BeforeKitLb(instance, seed))
	}
}

func WithClientTrace() ClientOption {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}

	t := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}
	return func(client *Client) {
		client.client.SetTransport(otelhttp.NewTransport(t))
	}
}
