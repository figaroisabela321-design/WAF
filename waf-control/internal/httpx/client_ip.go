package httpx

import (
	"net"
	"net/http"
	"strings"
)

// ClientIPFromRequest returns the client IP using trusted-proxy chain rules.
//
// If RemoteAddr is not in trustedProxies, forwarding headers are ignored and
// RemoteAddr is used.
//
// If RemoteAddr is trusted, X-Forwarded-For is treated as a proxy hop chain
// (optionally with RemoteAddr as the rightmost hop): walk right-to-left,
// strip hops that are in trustedProxies, and return the first non-trusted hop
// as the client IP. Do NOT take the leftmost IP blindly (spoofable).
//
// If XFF yields no usable client IP, X-Real-IP is used when present and valid.
// Illegal / unparseable header values fall back to RemoteAddr.
func ClientIPFromRequest(r *http.Request, trustedProxies []*net.IPNet) string {
	remoteIP := remoteIPFromAddr(r.RemoteAddr)
	if remoteIP != "" && isTrustedIP(remoteIP, trustedProxies) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			if ip := clientIPFromForwardedChain(xff, remoteIP, trustedProxies); ip != "" {
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

// clientIPFromForwardedChain walks X-Forwarded-For from right to left,
// treating RemoteAddr as an additional rightmost hop, strips trusted proxy
// hops, and returns the first non-trusted IP.
func clientIPFromForwardedChain(xff, remoteIP string, trusted []*net.IPNet) string {
	parts := strings.Split(xff, ",")
	hops := make([]string, 0, len(parts)+1)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		ip := net.ParseIP(p)
		if ip == nil {
			continue
		}
		hops = append(hops, ip.String())
	}
	if remoteIP != "" {
		if ip := net.ParseIP(remoteIP); ip != nil {
			hops = append(hops, ip.String())
		}
	}
	for i := len(hops) - 1; i >= 0; i-- {
		if !isTrustedIP(hops[i], trusted) {
			return hops[i]
		}
	}
	return ""
}
