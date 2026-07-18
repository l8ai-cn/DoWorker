package cache

import (
	"context"
	"net"
	"net/http"
	"net/url"
)

func newResourceHostTransport(hostAliases map[string]string) *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if len(hostAliases) == 0 {
		return transport
	}

	dialer := &net.Dialer{}
	transport.Proxy = func(req *http.Request) (*url.URL, error) {
		if _, ok := hostAliases[req.URL.Hostname()]; ok {
			return nil, nil
		}
		return http.ProxyFromEnvironment(req)
	}
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err == nil {
			if dialHost, ok := hostAliases[host]; ok {
				address = net.JoinHostPort(dialHost, port)
			}
		}
		return dialer.DialContext(ctx, network, address)
	}
	return transport
}
