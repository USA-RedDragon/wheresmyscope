package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/USA-RedDragon/wheresmyscope/internal/config"
	"github.com/go-chi/chi/v5/middleware"
)

func TestClientIP(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		proxies []string
		remote  string
		xff     string
		want    string
	}{
		{name: "no proxies ignores header", remote: "203.0.113.5:1234", xff: "198.51.100.1", want: "203.0.113.5"},
		{name: "untrusted peer cannot spoof", proxies: []string{"10.0.0.0/8"}, remote: "203.0.113.5:1234", xff: "198.51.100.1", want: "203.0.113.5"},
		{name: "trusted proxy forwards client", proxies: []string{"10.0.0.0/8"}, remote: "10.1.2.3:1234", xff: "198.51.100.1", want: "198.51.100.1"},
		{name: "skips trusted hops", proxies: []string{"10.0.0.0/8"}, remote: "10.1.2.3:1234", xff: "198.51.100.1, 10.9.9.9", want: "198.51.100.1"},
		{name: "spoofed left entry ignored", proxies: []string{"10.0.0.0/8"}, remote: "10.1.2.3:1234", xff: "1.1.1.1, 198.51.100.1", want: "198.51.100.1"},
		{name: "bare ip proxy", proxies: []string{"10.1.2.3"}, remote: "10.1.2.3:1234", xff: "198.51.100.1", want: "198.51.100.1"},
		{name: "trusted proxy without header", proxies: []string{"10.0.0.0/8"}, remote: "10.1.2.3:1234", want: "10.1.2.3"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got string
			h := clientIP(&config.Config{TrustedProxies: tc.proxies})(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				got = middleware.GetClientIP(r.Context())
			}))
			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			req.RemoteAddr = tc.remote
			if tc.xff != "" {
				req.Header.Set("X-Forwarded-For", tc.xff)
			}
			h.ServeHTTP(httptest.NewRecorder(), req)
			if got != tc.want {
				t.Errorf("client IP %q, want %q", got, tc.want)
			}
		})
	}
}
