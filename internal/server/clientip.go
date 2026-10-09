package server

import (
	"net"
	"net/http"
	"net/netip"

	"github.com/USA-RedDragon/wheresmyscope/internal/config"
	"github.com/go-chi/chi/v5/middleware"
)

// clientIP records the client IP for the request logger. X-Forwarded-For is
// only believed when the connection comes from a configured trusted proxy;
// otherwise a client could put any address in the header.
func clientIP(cfg *config.Config) func(http.Handler) http.Handler {
	prefixes, _ := cfg.TrustedProxyPrefixes()
	strs := make([]string, len(prefixes))
	for i, p := range prefixes {
		strs[i] = p.String()
	}
	fromXFF := middleware.ClientIPFromXFF(strs...)
	return func(next http.Handler) http.Handler {
		direct := middleware.ClientIPFromRemoteAddr(next)
		proxied := fromXFF(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(prefixes) > 0 && r.Header.Get("X-Forwarded-For") != "" && fromTrusted(r.RemoteAddr, prefixes) {
				proxied.ServeHTTP(w, r)
				return
			}
			direct.ServeHTTP(w, r)
		})
	}
}

func fromTrusted(remoteAddr string, prefixes []netip.Prefix) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	addr = addr.Unmap().WithZone("")
	for _, p := range prefixes {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}
