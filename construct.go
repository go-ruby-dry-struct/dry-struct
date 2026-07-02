// Copyright (c) the go-ruby-dry-struct/dry-struct authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drystruct

import (
	drytypes "github.com/go-ruby-dry-types/dry-types"
)

// New constructs a [*Struct] from an attribute hash, coercing and validating
// every attribute through its [AttrType]. It is `Struct.new(attrs)`:
//
//   - Each required attribute must be present (else `:key is missing in Hash
//     input`); an optional attribute may be absent.
//   - Under a strict schema, unexpected keys raise `unexpected keys […]`.
//   - Each present value is coerced through its type; the first failure raises a
//     [*Error] shaped like the gem's.
//
// As a special case matching the gem, New(existing) where existing is already a
// [*Struct] of this exact type returns it unchanged (no re-coercion).
func (s *StructType) New(attrs any) (*Struct, error) {
	if st, ok := attrs.(*Struct); ok && st.typ == s {
		return st, nil
	}
	out, err := s.build(attrs)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Call is `Struct.call(attrs)` / `Struct[attrs]` — an alias for [StructType.New].
func (s *StructType) Call(attrs any) (any, error) { return s.New(attrs) }

// MustNew is [StructType.New] but panics on error; convenient for tests and for
// statically-known-valid construction.
func (s *StructType) MustNew(attrs any) *Struct {
	out, err := s.New(attrs)
	if err != nil {
		panic(err)
	}
	return out
}

// build runs the schema coercion and assembles the ordered *Struct. It mirrors
// dry-types' Hash.schema exactly (dry-struct delegates attribute coercion to
// such a schema), so its error messages — missing key, unknown keys, and the
// per-member `<v> (<Class>) has invalid type for :key …` — are byte-identical.
func (s *StructType) build(attrs any) (*Struct, error) {
	// dry-struct coerces its input with Kernel#Hash: nil becomes an empty hash
	// (then a required attribute is reported missing), while any other non-hash
	// raises `can't convert <Class> into Hash`.
	if attrs == nil {
		attrs = drytypes.NewMap()
	}
	m, ok := asMap(attrs)
	if !ok {
		return nil, newError(s.Name, &drytypes.CoercionError{
			Message: "can't convert " + valueClass(attrs) + " into Hash",
		})
	}
	m = s.normalizeKeys(m)
	if s.strict {
		if err := s.checkUnknown(m); err != nil {
			return nil, newError(s.Name, err)
		}
	}
	values := drytypes.NewMap()
	for _, a := range s.attrs {
		val, present := lookupKey(m, a.Name)
		if !present {
			// An absent key with a default type resolves to its default (the type
			// substitutes it on the Undefined sentinel); an absent optional key is
			// dropped; any other absent required key is a missing-key error.
			if def, ok := a.defaultValue(); ok {
				values.Set(a.Name, def)
				continue
			}
			if a.Optional {
				continue
			}
			return nil, newError(s.Name, &drytypes.MissingKeyError{
				Message: ":" + string(a.Name) + " is missing in Hash input",
			})
		}
		c, err := a.Type.Coerce(val)
		if err != nil {
			return nil, newError(s.Name, wrapSchemaErr(a.Name, val, err))
		}
		values.Set(a.Name, c)
	}
	return &Struct{typ: s, values: values}, nil
}

// checkUnknown reports the strict-schema unknown-keys error for any input key not
// declared as an attribute (`unexpected keys [:a, :b] in Hash input`).
func (s *StructType) checkUnknown(m *drytypes.Map) error {
	var unknown []drytypes.Symbol
	for _, p := range m.Pairs() {
		sym, isSym := keyToSymbol(p.Key)
		if !isSym {
			continue
		}
		if _, ok := s.index[sym]; !ok {
			unknown = append(unknown, sym)
		}
	}
	if len(unknown) == 0 {
		return nil
	}
	return &drytypes.UnknownKeysError{Message: "unexpected keys " + keyList(unknown) + " in Hash input"}
}

// wrapSchemaErr turns a member failure into the gem's SchemaError shape,
// `<val> (<Class>) has invalid type for :key violates constraints (<rule>)`,
// reusing the inner constraint rule when present and appending " failed" to any
// other (coercion / nested-struct) message — exactly as dry-types' schema does.
func wrapSchemaErr(key drytypes.Symbol, val any, err error) error {
	var rule string
	if ce, ok := err.(*drytypes.ConstraintError); ok {
		rule = ce.Rule
	} else {
		rule = err.Error() + " failed"
	}
	msg := inspect(val) + " (" + valueClass(val) + ") has invalid type for :" + string(key) +
		" violates constraints (" + rule + ")"
	return &drytypes.SchemaError{Message: msg}
}

// keyToSymbol normalizes a hash key to a Symbol for attribute matching.
func keyToSymbol(k any) (drytypes.Symbol, bool) {
	switch x := k.(type) {
	case drytypes.Symbol:
		return x, true
	case string:
		return drytypes.Symbol(x), true
	}
	return "", false
}

// lookupKey finds an attribute in the input map, accepting either a Symbol or a
// String key of the same name (matching dry-types' schema key handling).
func lookupKey(m *drytypes.Map, key drytypes.Symbol) (any, bool) {
	if v, ok := m.Get(key); ok {
		return v, true
	}
	if v, ok := m.Get(string(key)); ok {
		return v, true
	}
	return nil, false
}

// keyList renders a slice of symbols the way Ruby inspects an array of symbols:
// `[:a, :b]`.
func keyList(syms []drytypes.Symbol) string {
	out := make([]any, len(syms))
	for i, s := range syms {
		out[i] = s
	}
	return inspect(out)
}

// normalizeKeys applies the transform_keys policy.
func (s *StructType) normalizeKeys(m *drytypes.Map) *drytypes.Map {
	if s.transform == KeyNone {
		return m
	}
	out := drytypes.NewMap()
	for _, p := range m.Pairs() {
		out.Set(s.transformKey(p.Key), p.Val)
	}
	return out
}

func (s *StructType) transformKey(k any) any {
	switch s.transform {
	case KeySymbolize:
		if str, ok := k.(string); ok {
			return drytypes.Symbol(str)
		}
	case KeyStringify:
		if sym, ok := k.(drytypes.Symbol); ok {
			return string(sym)
		}
	}
	return k
}
