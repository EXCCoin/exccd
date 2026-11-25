// Copyright (c) 2026 The ExchangeCoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcserver

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestWebsocketOrigin(t *testing.T) {
	for _, test := range []struct {
		name, host string
		origins    []string
		tls, want  bool
	}{
		{"native", "localhost:9109", nil, true, true},
		{"HTTP", "localhost:9109", []string{"http://LOCALHOST:9109"}, false, true},
		{"HTTPS", "localhost:9109", []string{"https://localhost:9109"}, true, true},
		{"default port", "example.org:443", []string{"https://example.org"}, true, true},
		{"default port on origin", "example.org", []string{"http://example.org:80"}, false, true},
		{"IPv6", "[::1]:9109", []string{"https://[::1]:9109"}, true, true},
		{"wrong host", "localhost:9109", []string{"https://example.org:9109"}, true, false},
		{"wrong scheme", "localhost:9109", []string{"http://localhost:9109"}, true, false},
		{"wrong port", "localhost:9109", []string{"https://localhost:443"}, true, false},
		{"null", "localhost:9109", []string{"null"}, true, false},
		{"file", "localhost", []string{"file://localhost"}, false, false},
		{"empty", "example.org", []string{""}, true, false},
		{"path", "example.org", []string{"https://example.org/"}, true, false},
		{"bare query", "example.org", []string{"https://example.org?"}, true, false},
		{"fragment", "example.org", []string{"https://example.org#"}, true, false},
		{"userinfo", "example.org", []string{"https://user@example.org"}, true, false},
		{"multiple", "example.org", []string{"https://example.org", "https://example.org"}, true, false},
		{"forwarded TLS", "example.org", []string{"https://example.org"}, false, false},
		{"invalid port", "example.org:80000", []string{"https://example.org:80000"}, true, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := &http.Request{Host: test.host, Header: make(http.Header)}
			for _, origin := range test.origins {
				r.Header.Add("Origin", origin)
			}
			r.Header.Set("X-Forwarded-Proto", "https")
			if test.tls {
				r.TLS = &tls.ConnectionState{}
			}
			if got := checkWebsocketOrigin(r); got != test.want {
				t.Errorf("got %t, want %t", got, test.want)
			}
		})
	}
}
