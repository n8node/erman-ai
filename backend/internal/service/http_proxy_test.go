package service

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransportForHTTPProxyDisablesHTTP2(t *testing.T) {
	parsed, err := url.Parse("http://user:pass@127.0.0.1:3128")
	require.NoError(t, err)

	transport := transportForHTTPProxy(parsed)
	require.NotNil(t, transport)
	assert.False(t, transport.ForceAttemptHTTP2)
	assert.NotNil(t, transport.TLSNextProto)
	assert.Empty(t, transport.TLSNextProto)
}

func TestHTTPClientForProxyDisablesHTTP2OnClonedTransport(t *testing.T) {
	base := &http.Client{
		Transport: &http.Transport{},
	}
	client, err := httpClientForProxy(base, "http://user:pass@127.0.0.1:3128")
	require.NoError(t, err)

	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	assert.False(t, transport.ForceAttemptHTTP2)
	assert.Empty(t, transport.TLSNextProto)
}
