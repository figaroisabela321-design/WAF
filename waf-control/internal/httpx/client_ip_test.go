package httpx

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustCIDR(t *testing.T, s string) *net.IPNet {
	t.Helper()
	_, n, err := net.ParseCIDR(s)
	require.NoError(t, err)
	return n
}

func TestClientIPSpoofedXFFUntrusted(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.10:12345"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	ip := ClientIPFromRequest(req, []*net.IPNet{mustCIDR(t, "10.0.0.0/8")})
	assert.Equal(t, "203.0.113.10", ip)
}

func TestClientIPTrustedXFF(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.5:9999"
	req.Header.Set("X-Forwarded-For", "198.51.100.7, 10.0.0.5")
	ip := ClientIPFromRequest(req, []*net.IPNet{mustCIDR(t, "10.0.0.0/8")})
	assert.Equal(t, "198.51.100.7", ip)
}

func TestClientIPTrustedXRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:80"
	req.Header.Set("X-Real-IP", "198.51.100.9")
	ip := ClientIPFromRequest(req, []*net.IPNet{mustCIDR(t, "127.0.0.1/32")})
	assert.Equal(t, "198.51.100.9", ip)
}

func TestClientIPRemoteAddrFallback(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:443"
	ip := ClientIPFromRequest(req, nil)
	assert.Equal(t, "192.0.2.1", ip)
}

func TestClientIPIllegalHeaderFallback(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.1.1.1:1"
	req.Header.Set("X-Forwarded-For", "not-an-ip, also-bad")
	req.Header.Set("X-Real-IP", "also-not-valid")
	ip := ClientIPFromRequest(req, []*net.IPNet{mustCIDR(t, "10.0.0.0/8")})
	assert.Equal(t, "10.1.1.1", ip)
}
