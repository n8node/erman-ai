package service

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransportViaHTTPConnectProxyDisablesHTTP2(t *testing.T) {
	parsed, err := url.Parse("http://user:pass@127.0.0.1:3128")
	require.NoError(t, err)

	transport := transportViaHTTPConnectProxy(parsed)
	require.NotNil(t, transport)
	assert.False(t, transport.ForceAttemptHTTP2)
	assert.Nil(t, transport.Proxy)
	assert.NotNil(t, transport.DialTLSContext)
	assert.Empty(t, transport.TLSNextProto)
}

func TestConnectViaHTTPProxy(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 4096)
		n, readErr := conn.Read(buf)
		if readErr != nil {
			return
		}
		reqText := string(buf[:n])
		assert.Contains(t, reqText, "CONNECT example.com:443 HTTP/1.1")
		assert.Contains(t, reqText, "Proxy-Authorization: Basic")
		_, _ = conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	}()

	proxyURL, err := url.Parse("http://user:pass@" + ln.Addr().String())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := connectViaHTTPProxy(ctx, proxyURL, "example.com:443")
	require.NoError(t, err)
	require.NotNil(t, conn)
	_ = conn.Close()
	<-done
}

func TestHTTPClientForProxyUsesManualConnectTransport(t *testing.T) {
	client, err := httpClientForProxy(&http.Client{Timeout: 5 * time.Second}, "http://user:pass@127.0.0.1:3128")
	require.NoError(t, err)
	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok)
	assert.Nil(t, transport.Proxy)
	assert.NotNil(t, transport.DialTLSContext)
}

func TestConnectViaHTTPProxyRejectsNonHTTPScheme(t *testing.T) {
	proxyURL, err := url.Parse("socks5://127.0.0.1:1080")
	require.NoError(t, err)
	_, err = connectViaHTTPProxy(context.Background(), proxyURL, "example.com:443")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported proxy scheme")
}

func TestConnectViaHTTPProxyRejectsBadStatus(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			return
		}
		defer conn.Close()
		_, _ = io.WriteString(conn, "HTTP/1.1 407 Proxy Authentication Required\r\n\r\n")
	}()

	proxyURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", ln.Addr().(*net.TCPAddr).Port))
	require.NoError(t, err)

	_, err = connectViaHTTPProxy(context.Background(), proxyURL, "example.com:443")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "proxy connect status")
}
