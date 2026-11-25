// Copyright (c) 2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/EXCCoin/exccd/blockchain/v4"
	"github.com/EXCCoin/exccd/internal/mining"
	"github.com/EXCCoin/exccd/wire"
)

func TestMain(m *testing.M) {
	flag.Parse()

	// Trim -test.* flags from the command line arguments list to allow
	// go-flags tests to succeed.
	os.Args = append([]string{os.Args[0]}, flag.Args()...)

	os.Exit(m.Run())
}

// In order to test command line arguments and environment variables, append
// the flags to the os.Args variable like so:
//   os.Args = append(os.Args, "--altdnsnames=\"hostname1,hostname2\"")
//
// For environment variables, use the following to set the variable before the
// func that loads the configuration is called:
//   os.Setenv("EXCCD_ALT_DNSNAMES", "hostname1,hostname2")
//
// These args and env variables will then get parsed during configuration load.

// TestLoadConfig ensures that basic configuration loading succeeds.
func TestLoadConfig(t *testing.T) {
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	_, _, err := loadConfig(appName)
	if err != nil {
		t.Fatalf("Failed to load exccd config: %s", err)
	}
}

// TestDefaultAltDNSNames ensures that there are no additional hostnames added
// by default during the configuration load phase.
func TestDefaultAltDNSNames(t *testing.T) {
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	cfg, _, err := loadConfig(appName)
	if err != nil {
		t.Fatalf("Failed to load exccd config: %s", err)
	}
	if len(cfg.AltDNSNames) != 0 {
		t.Fatalf("Invalid default value for altdnsnames: %s", cfg.AltDNSNames)
	}
}

// TestAltDNSNamesWithEnv ensures the DCRD_ALT_DNSNAMES environment variable is
// parsed into a slice of additional hostnames as intended.
func TestAltDNSNamesWithEnv(t *testing.T) {
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	os.Setenv("EXCCD_ALT_DNSNAMES", "hostname1,hostname2")
	cfg, _, err := loadConfig(appName)
	if err != nil {
		t.Fatalf("Failed to load exccd config: %s", err)
	}
	hostnames := strings.Join(cfg.AltDNSNames, ",")
	if hostnames != "hostname1,hostname2" {
		t.Fatalf("altDNSNames should be %s but was %s", "hostname1,hostname2",
			hostnames)
	}
}

// TestAltDNSNamesWithArg ensures the altdnsnames configuration option parses
// additional hostnames into a slice of hostnames as intended.
func TestAltDNSNamesWithArg(t *testing.T) {
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	old := os.Args
	os.Args = append(os.Args, "--altdnsnames=\"hostname1,hostname2\"")
	cfg, _, err := loadConfig(appName)
	if err != nil {
		t.Fatalf("Failed to load exccd config: %s", err)
	}
	hostnames := strings.Join(cfg.AltDNSNames, ",")
	if hostnames != "hostname1,hostname2" {
		t.Fatalf("altDNSNames should be %s but was %s", "hostname1,hostname2",
			hostnames)
	}
	os.Args = old
}

func TestWhitelists(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{
		oldArgs[0],
		"--whitelist=192.0.2.42/24",
		"--whitelist=2001:db8::1",
		"--whitelist=::ffff:198.51.100.42",
	}

	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	parsedCfg, _, err := loadConfig(appName)
	if err != nil {
		t.Fatalf("Failed to load whitelists: %v", err)
	}
	want := []netip.Prefix{
		netip.MustParsePrefix("192.0.2.0/24"),
		netip.MustParsePrefix("2001:db8::1/128"),
		netip.MustParsePrefix("198.51.100.42/32"),
	}
	if len(parsedCfg.whitelists) != len(want) {
		t.Fatalf("got %d whitelists, want %d", len(parsedCfg.whitelists),
			len(want))
	}
	for i := range want {
		if parsedCfg.whitelists[i] != want[i] {
			t.Fatalf("whitelist %d: got %v, want %v", i,
				parsedCfg.whitelists[i], want[i])
		}
	}

	oldCfg := cfg
	cfg = parsedCfg
	t.Cleanup(func() { cfg = oldCfg })
	tests := []struct {
		addr string
		want bool
	}{
		{"192.0.2.99:1234", true},
		{"192.0.3.1:1234", false},
		{"[2001:db8::1]:1234", true},
		{"[2001:db8::2]:1234", false},
		{"198.51.100.42:1234", true},
	}
	for _, test := range tests {
		addr, err := net.ResolveTCPAddr("tcp", test.addr)
		if err != nil {
			t.Fatal(err)
		}
		if got := isWhitelisted(addr); got != test.want {
			t.Errorf("isWhitelisted(%q) = %v, want %v", test.addr, got,
				test.want)
		}
	}
}

func TestRemovedUpnpFlag(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{oldArgs[0], "--upnp"}

	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	parsedCfg, _, err := loadConfig(appName)
	if err != nil {
		t.Fatalf("Failed to load deprecated upnp flag: %v", err)
	}
	if !parsedCfg.Upnp {
		t.Fatal("deprecated upnp flag was not retained for startup warning")
	}
}

func TestRejectedSpends(t *testing.T) {
	hash := strings.Repeat("ab", 32)
	values := []string{hash + ":0:0", hash + ":0:1", hash + ":1:0", hash + ":1:0"}
	got, err := parseRejectedSpends(values)
	if err != nil || len(got) != 3 {
		t.Fatalf("got %v, %v", got, err)
	}
	for _, value := range []string{hash + ":1", hash[2:] + ":1:0", hash + ":4294967296:0", hash + ":0:2", strings.Repeat("zz", 32) + ":0:0"} {
		if _, err := parseRejectedSpends([]string{value}); err == nil {
			t.Errorf("accepted invalid outpoint %q", value)
		}
	}
}

func TestMempoolLimits(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	for _, test := range []struct {
		args  []string
		valid bool
	}{
		{[]string{"--maxmempoolbytes=2097152", "--maxmempooltxs=30", "--limitfreerelay=45", "--norelaypriority"}, true},
		{[]string{"--maxmempoolbytes=0"}, false},
		{[]string{"--maxmempooltxs=-1"}, false},
		{[]string{"--maxmempooltxs=1"}, false},
	} {
		os.Args = append([]string{oldArgs[0], "--appdata=" + t.TempDir()}, test.args...)
		conf, _, err := loadConfig("exccd")
		if test.valid {
			if err != nil || conf.MaxMempoolBytes != 2097152 || conf.MaxMempoolTxs != 30 {
				t.Fatalf("load %v: %v", test.args, err)
			}
		} else if err == nil || !strings.Contains(err.Error(), "maxmempool") {
			t.Fatalf("expected limit error for %v, got %v", test.args, err)
		}
	}
}

func TestMiningStakeExclusions(t *testing.T) {
	target := wire.OutPoint{Tree: wire.TxTreeStake}
	exclusions := mining.RejectedSpends{target: {}, {Tree: wire.TxTreeRegular}: {}, {Tree: wire.TxTreeStake, Index: 1}: {}}
	for _, utxo := range []*blockchain.UtxoEntry{nil, {}} {
		lookups := 0
		err := validateMiningSpends(exclusions, func(op wire.OutPoint) (*blockchain.UtxoEntry, error) {
			lookups++
			if op != target {
				t.Fatalf("unexpected lookup of %v", op)
			}
			return utxo, nil
		})
		if lookups != 1 || (err != nil) != (utxo == nil) {
			t.Fatalf("entry %v, lookups %d, error %v", utxo, lookups, err)
		}
	}
}
