package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/vadicheck/gofermart/pkg/logger"
)

func NewHTTPClient(
	serviceName string,
	transport http.RoundTripper,
	timeout time.Duration,
	logger logger.LogClient,
) *http.Client {
	if transport == nil {
		transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: defaultTransportDialContext(&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}),
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			MaxConnsPerHost:       100,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}

	return &http.Client{
		Timeout: timeout,
		Transport: &DefaultTransport{
			inner:  transport,
			name:   serviceName,
			logger: logger,
		},
	}
}

type DefaultTransport struct {
	inner  http.RoundTripper
	logger logger.LogClient
	name   string
}

func (t *DefaultTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.inner.RoundTrip(req)

	return resp, err
}

func defaultTransportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}
