// Copyright (c) the go-ruby-dry-struct/dry-struct authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drystruct

import (
	drytypes "github.com/go-ruby-dry-types/dry-types"
)

// New constructs a [*Struct] from an attribute hash, coercing and validating
// every attribute through its dry-types type. It is `Struct.new(attrs)`:
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

// build runs the schema coercion and assembles the ordered *Struct.
func (s *StructType) build(attrs any) (*Struct, error) {
	m, ok := asMap(attrs)
	if !ok {
		// dry-struct routes non-hash input through its Hash schema, which reports
		// the type? failure. Reuse dry-types' schema to get the exact message.
		_, err := s.schema().Call(attrs)
		return nil, newError(s.Name, err)
	}
	norm := s.normalizeKeys(m)
	coerced, err := s.schema().Call(norm)
	if err != nil {
		return nil, newError(s.Name, err)
	}
	cm := coerced.(*drytypes.Map)
	values := drytypes.NewMap()
	// Emit attributes in declaration order; optional-and-absent are skipped from
	// storage (they read back as nil and are omitted from to_h), matching the gem.
	for _, a := range s.attrs {
		if v, present := cm.Get(a.Name); present {
			values.Set(a.Name, v)
		}
	}
	return &Struct{typ: s, values: values}, nil
}

// schema builds the dry-types Hash.schema equivalent of this struct's
// attributes. dry-struct itself delegates attribute coercion to exactly such a
// schema, so this reproduces the gem's messages (missing key, unknown keys, and
// per-member `<v> (<Class>) has invalid type for :key …`) byte-for-byte.
func (s *StructType) schema() drytypes.Type {
	keys := make([]drytypes.SchemaKey, len(s.attrs))
	for i, a := range s.attrs {
		keys[i] = drytypes.SchemaKey{Key: a.Name, Type: a.Type, Optional: a.Optional}
	}
	sc := drytypes.NewSchema(keys...)
	if s.strict {
		sc = sc.Strict()
	}
	return sc
}

// normalizeKeys applies the transform_keys policy: symbolize String keys, or
// stringify Symbol keys (then treated as their symbol name for matching). The
// dry-types schema matches Symbol/String interchangeably, so the transform only
// needs to canonicalize keys the gem's transform would touch.
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
