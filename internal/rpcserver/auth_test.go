// Copyright (c) 2026 The ExchangeCoin developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpcserver

import (
	"net/http"
	"testing"
)

func TestLimitedCredentials(t *testing.T) {
	s, err := New(&Config{RPCLimitUser: "limit", RPCLimitPass: "limit"})
	if err != nil {
		t.Fatalf("unable to create RPC server: %v", err)
	}

	authed, isAdmin, err := s.checkAuth(&http.Request{}, true)
	if authed || isAdmin || err == nil {
		t.Fatalf("no-auth request = (%v, %v, %v), want (false, false, error)",
			authed, isAdmin, err)
	}

	r := &http.Request{Header: make(http.Header)}
	r.Header.Set("Authorization", "Basic bGltaXQ6bGltaXQ=") // limit:limit
	authed, isAdmin, err = s.checkAuth(r, true)
	if !authed || isAdmin || err != nil {
		t.Fatalf("limited auth = (%v, %v, %v), want (true, false, nil)",
			authed, isAdmin, err)
	}
}
