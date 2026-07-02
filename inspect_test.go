// Copyright (c) the go-ruby-dry-struct/dry-struct authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drystruct

import (
	"math/big"
	"testing"

	drytypes "github.com/go-ruby-dry-types/dry-types"
)

func TestInspectScalars(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{nil, "nil"},
		{true, "true"},
		{false, "false"},
		{"a\"b\\c\nd\te\rf", `"a\"b\\c\nd\te\rf"`},
		{drytypes.Symbol("sym"), ":sym"},
		{int(3), "3"},
		{int32(4), "4"},
		{int64(5), "5"},
		{big.NewInt(6), "6"},
		{3.5, "3.5"},
		{4.0, "4.0"},
		{[]any{int64(1), "x"}, `[1, "x"]`},
		{map[string]any{"k": int64(1)}, `{"k" => 1}`}, // string-keyed hash
		{map[drytypes.Symbol]any{"k": int64(1)}, `{k: 1}`},
		{struct{ X int }{1}, "{1}"}, // unknown -> %v fallback
	}
	for _, c := range cases {
		if got := inspect(c.in); got != c.want {
			t.Errorf("inspect(%#v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestMapInspectNonSymbolKey(t *testing.T) {
	m := drytypes.NewMap()
	m.Set("strkey", int64(1))
	if got := mapInspect(m); got != `{"strkey" => 1}` {
		t.Errorf("non-symbol key inspect = %q", got)
	}
}

func TestValueClass(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{nil, "NilClass"},
		{true, "TrueClass"},
		{false, "FalseClass"},
		{"s", "String"},
		{drytypes.Symbol("x"), "Symbol"},
		{int64(1), "Integer"},
		{big.NewInt(2), "Integer"},
		{3.0, "Float"},
		{[]any{}, "Array"},
		{drytypes.NewMap(), "Hash"},
		{map[string]any{}, "Hash"},
		{struct{}{}, "Object"},
	}
	for _, c := range cases {
		if got := valueClass(c.in); got != c.want {
			t.Errorf("valueClass(%#v) = %q, want %q", c.in, got, c.want)
		}
	}
	// a nested struct reports its class name.
	st := addressType().MustNew(omap("street", "s", "city", "c"))
	if got := valueClass(st); got != "Address" {
		t.Errorf("valueClass(struct) = %q", got)
	}
}

func TestFormatFloat(t *testing.T) {
	if got := formatFloat(3.0); got != "3.0" {
		t.Errorf("formatFloat(3.0) = %q", got)
	}
	if got := formatFloat(3.25); got != "3.25" {
		t.Errorf("formatFloat(3.25) = %q", got)
	}
}

func TestStringMethod(t *testing.T) {
	st := addressType().MustNew(omap("street", "s", "city", "c"))
	if st.String() != st.Inspect() {
		t.Error("String() should equal Inspect()")
	}
}

func TestDeepToHMapBranch(t *testing.T) {
	// A struct whose attribute value is itself a *Map (not a struct) is
	// deep-copied through the map branch of deepToH.
	inner := drytypes.NewMap()
	inner.Set(drytypes.Symbol("k"), int64(1))
	s := &Struct{typ: New("X"), values: drytypes.NewMap()}
	s.values.Set(drytypes.Symbol("m"), inner)
	if got := mapInspect(s.ToH()); got != `{m: {k: 1}}` {
		t.Errorf("deepToH map branch = %q", got)
	}
}

func TestKeyToSymbolNonSymbol(t *testing.T) {
	if _, ok := keyToSymbol(int64(1)); ok {
		t.Error("int key should not convert to symbol")
	}
	if s, ok := keyToSymbol("name"); !ok || s != drytypes.Symbol("name") {
		t.Error("string key should convert to symbol")
	}
}

func TestStrictNonSymbolKeyIgnored(t *testing.T) {
	// A strict schema only complains about unexpected *symbol/string* keys; a
	// non-symbol key in the input is skipped by checkUnknown.
	s := New("S").Strict().Attribute("name", drytypes.StrictString())
	m := drytypes.NewMap()
	m.Set(drytypes.Symbol("name"), "x")
	m.Set(int64(7), "ignored") // non-symbol key
	out, err := s.New(m)
	if err != nil {
		t.Fatalf("non-symbol key should be ignored by strict check: %v", err)
	}
	if out.Inspect() != `#<S name="x">` {
		t.Errorf("inspect = %q", out.Inspect())
	}
}

func TestTransformKeyNoMatch(t *testing.T) {
	// Symbolize transform leaves a non-string key unchanged; stringify leaves a
	// non-symbol key unchanged.
	sy := New("A").TransformKeys(KeySymbolize)
	if got := sy.transformKey(int64(1)); got != int64(1) {
		t.Errorf("symbolize non-string = %v", got)
	}
	stf := New("B").TransformKeys(KeyStringify)
	if got := stf.transformKey(int64(1)); got != int64(1) {
		t.Errorf("stringify non-symbol = %v", got)
	}
}

func TestDefaultValueNonDefaultType(t *testing.T) {
	// A plain (non-default) dry-types attribute reports no default (Undefined
	// fails its coercion), and a nested struct type reports no default.
	a := Attribute{Name: "n", Type: Wrap(drytypes.StrictString())}
	if _, ok := a.defaultValue(); ok {
		t.Error("plain type should have no default")
	}
	b := Attribute{Name: "n", Type: addressType()}
	if _, ok := b.defaultValue(); ok {
		t.Error("nested struct should have no default")
	}
	c := Attribute{Name: "n", Type: Wrap(drytypes.StrictString().Default("d"))}
	if v, ok := c.defaultValue(); !ok || v != "d" {
		t.Errorf("default type = %v %v", v, ok)
	}
	// A Nominal type passes Undefined through unchanged (no error, no default):
	// it must report no default so the attribute stays required.
	d := Attribute{Name: "n", Type: Wrap(drytypes.NominalString())}
	if _, ok := d.defaultValue(); ok {
		t.Error("nominal (undefined-passthrough) type should have no default")
	}
}

func TestMapsEqualLengthMismatch(t *testing.T) {
	a := drytypes.NewMap()
	a.Set(drytypes.Symbol("x"), int64(1))
	b := drytypes.NewMap()
	b.Set(drytypes.Symbol("x"), int64(1))
	b.Set(drytypes.Symbol("y"), int64(2))
	if mapsEqual(a, b) {
		t.Error("maps of different length should not be equal")
	}
	// a key present in a but missing in b.
	c := drytypes.NewMap()
	c.Set(drytypes.Symbol("z"), int64(1))
	if mapsEqual(a, c) {
		t.Error("maps with different keys should not be equal")
	}
}
