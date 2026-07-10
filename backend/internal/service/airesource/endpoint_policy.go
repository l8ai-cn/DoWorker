package airesource

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	domain "github.com/anthropics/agentsmesh/backend/internal/domain/airesource"
)

type IPResolver interface {
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
}

type EndpointPolicy struct {
	allowHTTP bool
	resolver  IPResolver
}

const (
	providerRequestTimeout        = 15 * time.Second
	providerDialTimeout           = 5 * time.Second
	providerTLSHandshakeTimeout   = 5 * time.Second
	providerResponseHeaderTimeout = 10 * time.Second
)

func NewEndpointPolicy(allowHTTP bool, resolver IPResolver) *EndpointPolicy {
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	return &EndpointPolicy{allowHTTP: allowHTTP, resolver: resolver}
}

func (policy *EndpointPolicy) Validate(ctx context.Context, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Hostname() == "" || parsed.User != nil {
		return ErrInvalidEndpoint
	}
	if parsed.Scheme != "https" && !(policy.allowHTTP && parsed.Scheme == "http") {
		return ErrInvalidEndpoint
	}
	_, err = policy.resolveSafeIPs(ctx, parsed.Hostname())
	if err != nil {
		return err
	}
	return nil
}

func (policy *EndpointPolicy) resolveSafeIPs(ctx context.Context, hostname string) ([]net.IP, error) {
	host := strings.TrimSuffix(strings.ToLower(hostname), ".")
	if host == "localhost" || host == "metadata.google.internal" || host == "metadata" {
		return nil, ErrInvalidEndpoint
	}
	if parsed := net.ParseIP(host); parsed != nil {
		if !safePublicIP(parsed) {
			return nil, ErrInvalidEndpoint
		}
		return []net.IP{parsed}, nil
	}
	addresses, err := policy.resolver.LookupIPAddr(ctx, host)
	if err != nil || len(addresses) == 0 {
		return nil, fmt.Errorf("%w: hostname resolution failed", ErrInvalidEndpoint)
	}
	resolved := make([]net.IP, 0, len(addresses))
	for _, address := range addresses {
		if !safePublicIP(address.IP) {
			return nil, ErrInvalidEndpoint
		}
		resolved = append(resolved, address.IP)
	}
	return resolved, nil
}

func safePublicIP(ip net.IP) bool {
	return ip != nil && ip.IsGlobalUnicast() && !ip.IsPrivate() && !ip.IsLoopback() &&
		!ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() && !ip.IsUnspecified() && !ip.IsMulticast() &&
		!sharedAddressRange.Contains(ip)
}

var sharedAddressRange = &net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

func NewSafeHTTPClient(policy *EndpointPolicy, transport *http.Transport) *http.Client {
	if policy == nil {
		panic("AI resource endpoint policy is required")
	}
	if transport == nil {
		transport = http.DefaultTransport.(*http.Transport).Clone()
	} else {
		transport = transport.Clone()
	}
	transport.Proxy = nil
	transport.DialTLSContext = nil
	transport.DialTLS = nil
	transport.TLSHandshakeTimeout = providerTLSHandshakeTimeout
	transport.ResponseHeaderTimeout = providerResponseHeaderTimeout
	dialer := &net.Dialer{Timeout: providerDialTimeout, KeepAlive: 30 * time.Second}
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, ErrInvalidEndpoint
		}
		ips, err := policy.resolveSafeIPs(ctx, host)
		if err != nil {
			return nil, err
		}
		var lastErr error
		for _, ip := range ips {
			connection, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if dialErr == nil {
				return connection, nil
			}
			lastErr = dialErr
		}
		return nil, lastErr
	}
	return &http.Client{Transport: transport, Timeout: providerRequestTimeout, CheckRedirect: func(*http.Request, []*http.Request) error { return ErrInvalidEndpoint }}
}

func (s *Service) validatedBaseURL(ctx context.Context, provider domain.ProviderDefinition, requested string) (string, error) {
	requested = strings.TrimSpace(requested)
	if provider.SupportsCustomEndpoint {
		if requested == "" {
			return "", ErrInvalidEndpoint
		}
	} else {
		if requested != "" && strings.TrimRight(requested, "/") != strings.TrimRight(provider.DefaultBaseURL, "/") {
			return "", ErrInvalidEndpoint
		}
		requested = provider.DefaultBaseURL
	}
	if err := s.endpoints.Validate(ctx, requested); err != nil {
		return "", fmt.Errorf("%w: endpoint policy rejected URL", ErrInvalidEndpoint)
	}
	return strings.TrimRight(requested, "/"), nil
}
