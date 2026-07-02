// Copyright (c) the go-ruby-dry-struct/dry-struct authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drystruct

import (
	"sort"
	"strings"

	drytypes "github.com/go-ruby-dry-types/dry-types"
)

// Struct is one immutable dry-struct instance: an ordered map from attribute
// name to coerced value, tagged with the [*StructType] that produced it. It
// answers the reader, [Struct.ToH], [Struct.With], [Struct.Eql] and
// [Struct.Inspect] operations the gem's instances expose.
type Struct struct {
	typ    *StructType
	values *drytypes.Map
}

// Type returns the [*StructType] this instance was built from.
func (s *Struct) Type() *StructType { return s.typ }

// Get returns the value of attribute name and whether it is present. An optional
// attribute that was absent at construction is not present (and reads back nil),
// matching the gem's `struct[:key]` returning nil for an unset optional.
func (s *Struct) Get(name drytypes.Symbol) (any, bool) {
	return s.values.Get(name)
}

// Fetch returns the value of attribute name, or nil if it is absent — the
// direct analogue of Ruby's attribute reader / `struct[:name]`.
func (s *Struct) Fetch(name drytypes.Symbol) any {
	v, _ := s.values.Get(name)
	return v
}

// Attributes returns the instance's attributes as an ordered [*drytypes.Map]
// (Ruby `#attributes`), with nested [*Struct] values kept as-is (not deep
// converted — that is what [Struct.ToH] does). The returned map must not be
// mutated.
func (s *Struct) Attributes() *drytypes.Map { return s.values }

// ToH deep-converts the instance to an ordered [*drytypes.Map] (Ruby `#to_h` /
// `#to_hash`): nested [*Struct] values become their own to_h, and arrays of
// structs map element-wise. Absent optional attributes are omitted.
func (s *Struct) ToH() *drytypes.Map {
	out := drytypes.NewMap()
	for _, p := range s.values.Pairs() {
		out.Set(p.Key, deepToH(p.Val))
	}
	return out
}

// deepToH recursively converts nested structs (and arrays of them) to their hash
// form, leaving scalars untouched.
func deepToH(v any) any {
	switch x := v.(type) {
	case *Struct:
		return x.ToH()
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = deepToH(e)
		}
		return out
	case *drytypes.Map:
		out := drytypes.NewMap()
		for _, p := range x.Pairs() {
			out.Set(p.Key, deepToH(p.Val))
		}
		return out
	}
	return v
}

// With returns a new [*Struct] with the given attributes overridden (Ruby's
// `struct.new(changes)`): the receiver's attributes are merged with changes and
// re-coerced through the schema, yielding a fresh immutable instance. The
// receiver is unchanged.
func (s *Struct) With(changes *drytypes.Map) (*Struct, error) {
	merged := drytypes.NewMap()
	for _, p := range s.values.Pairs() {
		merged.Set(p.Key, p.Val)
	}
	if changes != nil {
		for _, p := range changes.Pairs() {
			merged.Set(p.Key, p.Val)
		}
	}
	return s.typ.New(merged)
}

// Eql reports whether two instances are equal (Ruby `#==` / `#eql?`): same
// [*StructType] and equal attribute maps. Structs of different types are never
// equal, matching the gem.
func (s *Struct) Eql(other *Struct) bool {
	if other == nil || s.typ != other.typ {
		return false
	}
	return mapsEqual(s.values, other.values)
}

// Inspect renders the instance the way dry-struct's `#inspect` does:
// `#<Name attr=<inspect> …>` in declaration order, with an absent optional
// attribute shown as `attr=nil`. An attribute-less struct renders `#<Name>`.
func (s *Struct) Inspect() string {
	var b strings.Builder
	b.WriteString("#<")
	b.WriteString(s.typ.Name)
	for _, a := range s.typ.attrs {
		b.WriteByte(' ')
		b.WriteString(string(a.Name))
		b.WriteByte('=')
		if v, ok := s.values.Get(a.Name); ok {
			b.WriteString(inspect(v))
		} else {
			b.WriteString("nil")
		}
	}
	b.WriteByte('>')
	return b.String()
}

// String is an alias for [Struct.Inspect] so a *Struct prints faithfully via fmt.
func (s *Struct) String() string { return s.Inspect() }

// mapsEqual compares two ordered maps by key set and value equality (order
// independent, matching Ruby Hash#==).
func mapsEqual(a, b *drytypes.Map) bool {
	if a.Len() != b.Len() {
		return false
	}
	for _, p := range a.Pairs() {
		bv, ok := b.Get(p.Key)
		if !ok || !valuesEqual(p.Val, bv) {
			return false
		}
	}
	return true
}

// valuesEqual compares two Ruby values for `==`, recursing into structs, arrays
// and maps.
func valuesEqual(a, b any) bool {
	switch x := a.(type) {
	case *Struct:
		y, ok := b.(*Struct)
		return ok && x.Eql(y)
	case []any:
		y, ok := b.([]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if !valuesEqual(x[i], y[i]) {
				return false
			}
		}
		return true
	case *drytypes.Map:
		y, ok := b.(*drytypes.Map)
		return ok && mapsEqual(x, y)
	}
	return a == b
}

// asMap normalizes hash-shaped input into an ordered *drytypes.Map, accepting the
// same shapes the go-ruby-* value model uses on input. Non-hash input returns
// (nil, false) so the caller can raise the proper type error.
func asMap(v any) (*drytypes.Map, bool) {
	switch h := v.(type) {
	case *drytypes.Map:
		return h, true
	case map[string]any:
		m := drytypes.NewMap()
		keys := make([]string, 0, len(h))
		for k := range h {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			m.Set(k, h[k])
		}
		return m, true
	case map[drytypes.Symbol]any:
		m := drytypes.NewMap()
		keys := make([]string, 0, len(h))
		for k := range h {
			keys = append(keys, string(k))
		}
		sort.Strings(keys)
		for _, k := range keys {
			m.Set(drytypes.Symbol(k), h[drytypes.Symbol(k)])
		}
		return m, true
	case map[any]any:
		m := drytypes.NewMap()
		for k, val := range h {
			m.Set(k, val)
		}
		return m, true
	}
	return nil, false
}
