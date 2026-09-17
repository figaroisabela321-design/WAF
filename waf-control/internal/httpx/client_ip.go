package httpx

import (
	"net"
	"net/http"
	"strings"
)

// ClientIPFromRequest returns the client IP using trusted-proxy rules.
// If RemoteAddr is in trustedProxies, parse X-Forwarded-For (leftmost valid IP)
// or X-Real-IP; otherwise ignore forwarding headers and use RemoteAddr.
// Illegal / unparseable header values fall back to RemoteAddr.
func ClientIPFromRequest(r *http.Request, trustedProxies []*net.IPNet) string {
	remoteIP := remoteIPFromAddr(r.RemoteAddr)
	if remoteIP != "" && isTrustedIP(remoteIP, trustedProxies) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if ip := leftmostValidIP(xff); ip != "" {
				return ip
			}
		}
		if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
			if ip := net.ParseIP(xri); ip != nil {
				return ip.String()
			}
		}
	}
	if remoteIP != "" {
		return remoteIP
	}
	return r.RemoteAddr
}

// ClientIP is the legacy helper without trusted-proxy filtering.
// Prefer ClientIPFromRequest with configured TRUSTED_PROXIES.
func ClientIP(r *http.Request) string {
	return ClientIPFromRequest(r, nil)
}

func remoteIPFromAddr(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// May already be a bare IP.
		if ip := net.ParseIP(addr); ip != nil {
			return ip.String()
		}
		return ""
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String()
	}
	return ""
}

func isTrustedIP(ipStr string, trusted []*net.IPNet) bool {
	if len(trusted) == 0 {
		return false
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	for _, n := range trusted {
		if n != nil && n.Contains(ip) {
			return true
		}
	}
	return false
}

func leftmostValidIP(xff string) string {
	parts := strings.Split(xff, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if ip := net.ParseIP(p); ip != nil {
			return ip.String()
		}
	}
	return ""
}
