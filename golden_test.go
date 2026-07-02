// Copyright (c) the go-ruby-dry-struct/dry-struct authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drystruct

import (
	"encoding/json"
	"os"
	"testing"
)

// goldenCase is one entry of testdata/golden.json, captured from the real
// dry-struct gem.
type goldenCase struct {
	Name       string `json:"name"`
	OK         bool   `json:"ok"`
	Inspect    string `json:"inspect"`
	ToH        string `json:"to_h"`
	Error      string `json:"error"`
	ErrorClass string `json:"error_class"`
}

// TestGolden replays every gem-captured case through the Go implementation and
// asserts inspect / to_h / error message are byte-identical. This holds the
// full suite at parity with no Ruby present.
func TestGolden(t *testing.T) {
	data, err := os.ReadFile("testdata/golden.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []goldenCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("no golden cases")
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := buildCase(c.Name)
			if c.OK {
				if err != nil {
					t.Fatalf("expected success, got error: %v", err)
				}
				if ins := got.Inspect(); ins != c.Inspect {
					t.Errorf("inspect:\n got %q\nwant %q", ins, c.Inspect)
				}
				if th := mapInspect(got.ToH()); th != c.ToH {
					t.Errorf("to_h:\n got %q\nwant %q", th, c.ToH)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error %q, got success %s", c.Error, got.Inspect())
			}
			if err.Error() != c.Error {
				t.Errorf("error:\n got %q\nwant %q", err.Error(), c.Error)
			}
			if _, ok := err.(*Error); !ok {
				t.Errorf("expected *Error (Dry::Struct::Error), got %T", err)
			}
		})
	}
}
