// Copyright (c) the go-ruby-dry-struct/dry-struct authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drystruct

import (
	"errors"
	"testing"

	drytypes "github.com/go-ruby-dry-types/dry-types"
)

func TestReadersAndAccess(t *testing.T) {
	u := cfgType().MustNew(omap("host", "h", "port", "80"))
	if got := u.Fetch("host"); got != "h" {
		t.Errorf("Fetch host = %v", got)
	}
	if got := u.Fetch("port"); got != int64(80) {
		t.Errorf("Fetch port = %v", got)
	}
	if v, ok := u.Get("host"); !ok || v != "h" {
		t.Errorf("Get host = %v %v", v, ok)
	}
	// absent optional reads nil and is not present.
	c := cfgType().MustNew(omap("host", "h"))
	if v, ok := c.Get("port"); ok || v != nil {
		t.Errorf("absent optional Get = %v %v", v, ok)
	}
	if v := c.Fetch("port"); v != nil {
		t.Errorf("absent optional Fetch = %v", v)
	}
	if u.Type().Name != "Cfg" {
		t.Errorf("Type().Name = %q", u.Type().Name)
	}
}

func TestAttributesVsToH(t *testing.T) {
	wrap := New("Wrap").AttributeType("u", cfgType())
	w := wrap.MustNew(omap("u", omap("host", "n", "port", "1")))
	// attributes keeps the nested struct instance.
	attrs := w.Attributes()
	inner, ok := attrs.Get(drytypes.Symbol("u"))
	if !ok {
		t.Fatal("missing u in attributes")
	}
	if _, isStruct := inner.(*Struct); !isStruct {
		t.Errorf("attributes should keep *Struct, got %T", inner)
	}
	// to_h deep-converts.
	if th := mapInspect(w.ToH()); th != `{u: {host: "n", port: 1}}` {
		t.Errorf("to_h = %q", th)
	}
	// ToHash is an alias.
	if mapInspect(w.ToHash()) != mapInspect(w.ToH()) {
		t.Error("ToHash != ToH")
	}
}

func TestWith(t *testing.T) {
	u := cfgType().MustNew(omap("host", "h", "port", "1"))
	u2, err := u.With(omap("port", "5"))
	if err != nil {
		t.Fatal(err)
	}
	if u2.Inspect() != `#<Cfg host="h" port=5>` {
		t.Errorf("with inspect = %q", u2.Inspect())
	}
	// original unchanged (immutable).
	if u.Inspect() != `#<Cfg host="h" port=1>` {
		t.Errorf("original mutated = %q", u.Inspect())
	}
	// nil changes returns an equal copy.
	u3, err := u.With(nil)
	if err != nil || !u.Eql(u3) {
		t.Errorf("with nil = %v %v", u3, err)
	}
	// a bad change surfaces the error.
	if _, err := u.With(omap("port", "abc")); err == nil {
		t.Error("expected error from bad With change")
	}
}

func TestEquality(t *testing.T) {
	// Equality is by class (same *StructType) and attributes, so instances must
	// share one type — as they do when built from a single Dry::Struct subclass.
	ct := cfgType()
	a := ct.MustNew(omap("host", "h", "port", "1"))
	b := ct.MustNew(omap("host", "h", "port", "1"))
	c := ct.MustNew(omap("host", "h", "port", "2"))
	if !a.Eql(b) {
		t.Error("a should eql b")
	}
	if a.Eql(c) {
		t.Error("a should not eql c")
	}
	if a.Eql(nil) {
		t.Error("a should not eql nil")
	}
	// different types never equal.
	other := New("Other").Attribute("host", drytypes.StrictString()).AttributeOpt("port", drytypes.CoercibleInteger())
	o := other.MustNew(omap("host", "h", "port", "1"))
	if a.Eql(o) {
		t.Error("cross-type structs should not be equal")
	}
}

func TestNestedAndArrayEquality(t *testing.T) {
	pt := personType()
	p1 := pt.MustNew(omap("name", "A", "address", omap("street", "s", "city", "c")))
	p2 := pt.MustNew(omap("name", "A", "address", omap("street", "s", "city", "c")))
	if !p1.Eql(p2) {
		t.Error("nested structs should be equal")
	}
	p3 := pt.MustNew(omap("name", "A", "address", omap("street", "s", "city", "d")))
	if p1.Eql(p3) {
		t.Error("differing nested structs should not be equal")
	}
	tt := teamType()
	tm1 := tt.MustNew(omap("members", []any{omap("name", "A", "address", omap("street", "s", "city", "c"))}))
	tm2 := tt.MustNew(omap("members", []any{omap("name", "A", "address", omap("street", "s", "city", "c"))}))
	if !valuesEqual(tm1, tm2) {
		t.Error("array-of-struct structs should be equal")
	}
	// length + element + map differences.
	if valuesEqual([]any{int64(1)}, []any{int64(1), int64(2)}) {
		t.Error("different-length arrays equal")
	}
	if valuesEqual([]any{int64(1)}, []any{int64(2)}) {
		t.Error("different-element arrays equal")
	}
	if valuesEqual(omap("a", int64(1)), omap("a", int64(2))) {
		t.Error("different maps equal")
	}
	if !valuesEqual(omap("a", int64(1)), omap("a", int64(1))) {
		t.Error("equal maps unequal")
	}
	// type-mismatched operands.
	if valuesEqual(p1, []any{}) || valuesEqual([]any{}, p1) || valuesEqual(omap(), p1) {
		t.Error("type-mismatched values equal")
	}
}

func TestCallAndBracketAndPassthrough(t *testing.T) {
	ct := cfgType()
	out, err := ct.Call(omap("host", "h"))
	if err != nil {
		t.Fatal(err)
	}
	// an absent optional renders as port=nil in inspect (omitted from to_h).
	if out.(*Struct).Inspect() != `#<Cfg host="h" port=nil>` {
		t.Errorf("Call = %v", out)
	}
	// passing an instance of the same type through returns it unchanged.
	u := ct.MustNew(omap("host", "h"))
	u2, err := ct.New(u)
	if err != nil || u2 != u {
		t.Errorf("passthrough failed: %v %v", u2, err)
	}
	// an instance of a *different* type is re-coerced (not passed through) — and
	// here fails because it's not hash-shaped input.
	if _, err := personType().New(u); err == nil {
		t.Error("expected foreign struct not to pass through")
	}
}

func TestMustNewPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNew should panic on error")
		}
	}()
	cfgType().MustNew(omap()) // missing required host
}

func TestTransformKeysSymbolize(t *testing.T) {
	sk := New("Sk").TransformKeys(KeySymbolize).Attribute("name", drytypes.StrictString())
	out, err := sk.New(map[string]any{"name": "x"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Inspect() != `#<Sk name="x">` {
		t.Errorf("symbolize = %q", out.Inspect())
	}
}

func TestTransformKeysStringify(t *testing.T) {
	// A stringify transform maps the incoming symbol key to a string; the schema
	// then matches the declared attribute by name.
	st := New("St").TransformKeys(KeyStringify).AttributeType("n", Wrap(drytypes.StrictString()))
	// Declare the attribute name as the string form so lookup finds it.
	st = New("St2").TransformKeys(KeyStringify).Attribute("n", drytypes.StrictString())
	out, err := st.New(omap("n", "x"))
	if err != nil {
		t.Fatal(err)
	}
	if out.Inspect() != `#<St2 n="x">` {
		t.Errorf("stringify = %q", out.Inspect())
	}
}

func TestValueAndInherit(t *testing.T) {
	pt := New("Pt").AsValue().
		Attribute("x", drytypes.StrictInteger()).
		Attribute("y", drytypes.StrictInteger())
	if !pt.IsValue() {
		t.Error("IsValue should be true")
	}
	p := pt.MustNew(omap("x", int64(1), "y", int64(2)))
	if p.Inspect() != "#<Pt x=1 y=2>" {
		t.Errorf("value inspect = %q", p.Inspect())
	}

	base := New("Base").Attribute("id", drytypes.StrictInteger())
	derived := base.Inherit("Derived").Attribute("name", drytypes.StrictString())
	d := derived.MustNew(omap("id", int64(1), "name", "x"))
	if d.Inspect() != `#<Derived id=1 name="x">` {
		t.Errorf("inherit inspect = %q", d.Inspect())
	}
	if mapInspect(d.ToH()) != `{id: 1, name: "x"}` {
		t.Errorf("inherit to_h = %q", mapInspect(d.ToH()))
	}
	// inheriting carries value/strict/transform flags.
	vchild := pt.Inherit("Pt2")
	if !vchild.IsValue() {
		t.Error("inherited struct should keep Value flag")
	}
	// declaration order and count.
	if len(derived.Attributes()) != 2 {
		t.Errorf("derived attrs = %d", len(derived.Attributes()))
	}
}

func TestAttributeOverride(t *testing.T) {
	// re-declaring an attribute replaces it in place (subclass override).
	s := New("S").
		Attribute("a", drytypes.StrictInteger()).
		Attribute("b", drytypes.StrictString()).
		Attribute("a", drytypes.CoercibleInteger()) // override a's type
	if len(s.Attributes()) != 2 {
		t.Fatalf("override should keep 2 attrs, got %d", len(s.Attributes()))
	}
	out := s.MustNew(omap("a", "7", "b", "x")) // coercible now accepts string
	if out.Inspect() != `#<S a=7 b="x">` {
		t.Errorf("override inspect = %q", out.Inspect())
	}
}

func TestDefine(t *testing.T) {
	u := Define("U", func(s *StructType) {
		s.Attribute("name", drytypes.StrictString())
		s.AttributeOpt("nick", drytypes.StrictString())
	})
	out := u.MustNew(omap("name", "x"))
	if out.Inspect() != `#<U name="x" nick=nil>` {
		t.Errorf("Define inspect = %q", out.Inspect())
	}
}

func TestAttributeTypeOptAndArrayOfBad(t *testing.T) {
	// AttributeTypeOpt: optional nested struct.
	wrap := New("W").AttributeTypeOpt("addr", addressType())
	out := wrap.MustNew(omap())
	if out.Inspect() != "#<W addr=nil>" {
		t.Errorf("opt nested inspect = %q", out.Inspect())
	}
	// ArrayOf with a non-array value.
	tm := teamType()
	if _, err := tm.New(omap("members", "notarray")); err == nil {
		t.Error("expected array type error")
	}
	// ArrayOf element failure surfaces.
	if _, err := tm.New(omap("members", []any{omap("name", int64(1), "address", omap("street", "s", "city", "c"))})); err == nil {
		t.Error("expected array element error")
	}
}

func TestErrorUnwrap(t *testing.T) {
	_, err := cfgType().New(omap())
	var de *Error
	if !errors.As(err, &de) {
		t.Fatal("expected *Error")
	}
	if de.Unwrap() == nil {
		t.Error("Unwrap should expose the underlying dry-types error")
	}
	var mk *drytypes.MissingKeyError
	if !errors.As(err, &mk) {
		t.Error("Unwrap should reach MissingKeyError")
	}
}

func TestGoMapInputsAndGetShapes(t *testing.T) {
	// map[string]any input path (sorted keys).
	out, err := cfgType().New(map[string]any{"host": "h", "port": "9"})
	if err != nil || out.Inspect() != `#<Cfg host="h" port=9>` {
		t.Errorf("map[string]any input = %v %v", out, err)
	}
	// map[Symbol]any input path.
	out2, err := cfgType().New(sym(map[string]any{"host": "h"}))
	if err != nil || out2.Inspect() != `#<Cfg host="h" port=nil>` {
		t.Errorf("map[Symbol]any input = %v %v", out2, err)
	}
	// map[any]any input path.
	out3, err := cfgType().New(map[any]any{drytypes.Symbol("host"): "h"})
	if err != nil || out3.Inspect() != `#<Cfg host="h" port=nil>` {
		t.Errorf("map[any]any input = %v %v", out3, err)
	}
	// nil input -> treated as empty hash -> missing key.
	if _, err := cfgType().New(nil); err == nil {
		t.Error("nil input should error on missing required key")
	}
	// non-hash scalar -> can't convert.
	_, err = cfgType().New(int64(5))
	if err == nil || err.Error() != "[Cfg.new] can't convert Integer into Hash" {
		t.Errorf("scalar input = %v", err)
	}
}
