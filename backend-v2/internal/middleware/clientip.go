package middleware

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// ClientIPResolver returns a function that resolves the client IP exactly
// like Express's `app.set('trust proxy', hops)` (proxy-addr): the candidate
// list is [socket address, X-Forwarded-For entries right-to-left], and the
// first `hops` entries are trusted, so the client is the entry at index
// `hops` (clamped to the last, i.e. left-most, entry). hops <= 0 ignores
// X-Forwarded-For entirely.
//
// Gin's own SetTrustedProxies is all-or-nothing per proxy CIDR; with its
// default (trust everything) c.ClientIP() returns the LEFT-most
// X-Forwarded-For entry, which is whatever the client chose to send -- a
// trivial rate-limit bypass. main.go disables Gin's proxy trust and keys
// the rate limiters off this resolver instead.
func ClientIPResolver(hops int) func(*gin.Context) string {
	return func(c *gin.Context) string {
		return resolveClientIP(c.Request.RemoteAddr, c.GetHeader("X-Forwarded-For"), hops)
	}
}

func resolveClientIP(remoteAddr, xff string, hops int) string {
	socket := remoteAddr
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		socket = host
	}
	if hops <= 0 || xff == "" {
		return socket
	}
	parts := strings.Split(xff, ",")
	// addrs[0] is the socket peer, then XFF from nearest (right) to farthest.
	addrs := make([]string, 0, len(parts)+1)
	addrs = append(addrs, socket)
	for i := len(parts) - 1; i >= 0; i-- {
		if p := strings.TrimSpace(parts[i]); p != "" {
			addrs = append(addrs, p)
		}
	}
	return addrs[min(hops, len(addrs)-1)]
}
