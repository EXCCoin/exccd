// Copyright (c) 2018 The ExchangeCoin team
// Copyright (c) 2014 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package exccjson_test

import (
	"testing"

	"github.com/EXCCoin/exccd/exccjson"
)

// TestErrorCodeStringer tests the stringized output for the ErrorCode type.
func TestErrorCodeStringer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   exccjson.ErrorCode
		want string
	}{
		{exccjson.ErrDuplicateMethod, "ErrDuplicateMethod"},
		{exccjson.ErrInvalidUsageFlags, "ErrInvalidUsageFlags"},
		{exccjson.ErrInvalidType, "ErrInvalidType"},
		{exccjson.ErrEmbeddedType, "ErrEmbeddedType"},
		{exccjson.ErrUnexportedField, "ErrUnexportedField"},
		{exccjson.ErrUnsupportedFieldType, "ErrUnsupportedFieldType"},
		{exccjson.ErrNonOptionalField, "ErrNonOptionalField"},
		{exccjson.ErrNonOptionalDefault, "ErrNonOptionalDefault"},
		{exccjson.ErrMismatchedDefault, "ErrMismatchedDefault"},
		{exccjson.ErrUnregisteredMethod, "ErrUnregisteredMethod"},
		{exccjson.ErrNumParams, "ErrNumParams"},
		{exccjson.ErrMissingDescription, "ErrMissingDescription"},
		{0xffff, "Unknown ErrorCode (65535)"},
	}

	// Detect additional error codes that don't have the stringer added.
	if len(tests)-1 != int(exccjson.TstNumErrorCodes) {
		t.Errorf("It appears an error code was added without adding an " +
			"associated stringer test")
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		result := test.in.String()
		if result != test.want {
			t.Errorf("String #%d\n got: %s want: %s", i, result,
				test.want)
			continue
		}
	}
}

// TestError tests the error output for the Error type.
func TestError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   exccjson.Error
		want string
	}{
		{
			exccjson.Error{Message: "some error"},
			"some error",
		},
		{
			exccjson.Error{Message: "human-readable error"},
			"human-readable error",
		},
	}

	t.Logf("Running %d tests", len(tests))
	for i, test := range tests {
		result := test.in.Error()
		if result != test.want {
			t.Errorf("Error #%d\n got: %s want: %s", i, result,
				test.want)
			continue
		}
	}
}
