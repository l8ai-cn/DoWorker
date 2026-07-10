package airesource

import (
	"context"
	"net"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sequenceResolver struct {
	mu        sync.Mutex
	responses [][]net.IPAddr
}

func (r *sequenceResolver) LookupIPAddr(context.Context, string) ([]net.IPAddr, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	response := r.responses[0]
	if len(r.responses) > 1 {
		r.responses = r.responses[1:]
	}
	return response, nil
}

type staticResolver struct {
	addresses map[string][]net.IPAddr
	err       error
}

func (r staticResolver) LookupIPAddr(_ context.Context, host string) ([]net.IPAddr, error) {
	return r.addresses[host], r.err
}

func TestEndpointPolicyRejectsUnsafeAddressesAndMalformedURLs(t *testing.T) {
	policy := NewEndpointPolicy(false, staticResolver{addresses: map[string][]net.IPAddr{"public.example": {{IP: net.ParseIP("203.0.113.10")}}, "mixed.example": {{IP: net.ParseIP("203.0.113.10")}, {IP: net.ParseIP("10.0.0.1")}}}})
	tests := []string{
		"http://public.example/v1", "https://user:pass@public.example/v1", "https://localhost/v1",
		"https://127.0.0.1/v1", "https://10.0.0.1/v1", "https://169.254.169.254/latest/meta-data",
		"https://100.100.100.200/latest/meta-data",
		"https://[::1]/v1", "https://mixed.example/v1", "https://metadata.google.internal/v1",
	}
	for _, rawURL := range tests {
		t.Run(rawURL, func(t *testing.T) {
			assert.ErrorIs(t, policy.Validate(context.Background(), rawURL), ErrInvalidEndpoint)
		})
	}
	require.NoError(t, policy.Validate(context.Background(), "https://public.example/v1"))
}

func TestEndpointPolicyAllowsExplicitHTTPForPrivateDeploymentOnlyWhenAddressIsPublic(t *testing.T) {
	policy := NewEndpointPolicy(true, staticResolver{addresses: map[string][]net.IPAddr{"public.example": {{IP: net.ParseIP("203.0.113.10")}}}})
	require.NoError(t, policy.Validate(context.Background(), "http://public.example/v1"))
	assert.ErrorIs(t, policy.Validate(context.Background(), "http://127.0.0.1/v1"), ErrInvalidEndpoint)
}

func TestEndpointPolicyRejectsDNSLookupFailuresAndEmptyResults(t *testing.T) {
	assert.ErrorIs(t, NewEndpointPolicy(false, staticResolver{err: errInjected}).Validate(context.Background(), "https://provider.example"), ErrInvalidEndpoint)
	assert.ErrorIs(t, NewEndpointPolicy(false, staticResolver{addresses: map[string][]net.IPAddr{}}).Validate(context.Background(), "https://provider.example"), ErrInvalidEndpoint)
}

func TestSafeHTTPClientRechecksDNSAtDialAndRejectsRebinding(t *testing.T) {
	resolver := &sequenceResolver{responses: [][]net.IPAddr{{{IP: net.ParseIP("203.0.113.10")}}, {{IP: net.ParseIP("127.0.0.1")}}}}
	policy := NewEndpointPolicy(false, resolver)
	require.NoError(t, policy.Validate(context.Background(), "https://rebind.example"))
	client := NewSafeHTTPClient(policy, nil)
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://rebind.example", nil)
	require.NoError(t, err)
	_, err = client.Do(request)
	assert.ErrorIs(t, err, ErrInvalidEndpoint)
}

func TestSafeHTTPClientDisablesEnvironmentProxy(t *testing.T) {
	client := NewSafeHTTPClient(NewEndpointPolicy(false, staticResolver{}), nil)
	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	assert.Nil(t, transport.Proxy)
	assert.Positive(t, client.Timeout)
	assert.Positive(t, transport.TLSHandshakeTimeout)
	assert.Positive(t, transport.ResponseHeaderTimeout)
}

func TestSafeHTTPClientClearsAlternateTLSDialers(t *testing.T) {
	called := false
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialTLSContext = func(context.Context, string, string) (net.Conn, error) { called = true; return nil, errInjected }
	transport.DialTLS = func(string, string) (net.Conn, error) { called = true; return nil, errInjected }
	client := NewSafeHTTPClient(NewEndpointPolicy(false, staticResolver{addresses: map[string][]net.IPAddr{"safe.example": {{IP: net.ParseIP("127.0.0.1")}}}}), transport)
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://safe.example", nil)
	require.NoError(t, err)
	_, err = client.Do(request)
	assert.ErrorIs(t, err, ErrInvalidEndpoint)
	assert.False(t, called)
}
