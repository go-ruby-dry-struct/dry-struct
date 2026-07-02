// Copyright (c) the go-ruby-dry-struct/dry-struct authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drystruct

import (
	drytypes "github.com/go-ruby-dry-types/dry-types"
)

// StructType is the analogue of a `Dry::Struct` subclass: a named, ordered list
// of [Attribute]s plus its key-transform and strictness config. It is the
// factory for [*Struct] instances (via [StructType.New] / [StructType.Call]).
//
// A StructType implements [drytypes.Type], so it can itself be used as the type
// of an attribute on another StructType — that is how nested structs and
// `Types::Array.of(SomeStruct)` compose. Applying it to a hash coerces that hash
// into a [*Struct]; applying it to an existing [*Struct] of the same type passes
// it through unchanged (mirroring the gem).
type StructType struct {
	// Name is the struct's class name, used in error messages and inspect
	// (`[Name.new] …`, `#<Name …>`).
	Name string
	attrs []Attribute
	// index maps attribute name to its position in attrs.
	index map[drytypes.Symbol]int
	// transform is the incoming-key normalization (transform_keys).
	transform KeyTransform
	// strict rejects unexpected keys (schema schema.strict).
	strict bool
	// value marks a Dry::Struct::Value (comparable-by-value; the instances are
	// still immutable, which every dry-struct instance already is).
	value bool
}

// New builds an empty [*StructType] with the given class name. Register
// attributes with [StructType.Attribute] / [StructType.AttributeOpt] (they
// return the receiver, so calls chain).
func New(name string) *StructType {
	return &StructType{Name: name, index: map[drytypes.Symbol]int{}}
}

// Attribute declares a required attribute (`attribute :name, type`) and returns
// the receiver for chaining. Re-declaring a name replaces it in place (keeping
// its position), matching a subclass overriding an inherited attribute.
func (s *StructType) Attribute(name drytypes.Symbol, t drytypes.Type) *StructType {
	return s.add(Attribute{Name: name, Type: t, Optional: false})
}

// AttributeOpt declares an optional attribute (`attribute? :name, type`): the
// key may be absent, in which case the attribute is omitted from [Struct.ToH]
// and reads back as nil. Returns the receiver for chaining.
func (s *StructType) AttributeOpt(name drytypes.Symbol, t drytypes.Type) *StructType {
	return s.add(Attribute{Name: name, Type: t, Optional: true})
}

func (s *StructType) add(a Attribute) *StructType {
	if i, ok := s.index[a.Name]; ok {
		s.attrs[i] = a
		return s
	}
	s.index[a.Name] = len(s.attrs)
	s.attrs = append(s.attrs, a)
	return s
}

// TransformKeys sets the incoming-key normalization (dry-struct's
// `transform_keys`). Returns the receiver for chaining.
func (s *StructType) TransformKeys(t KeyTransform) *StructType {
	s.transform = t
	return s
}

// Strict marks the struct's schema strict (`schema schema.strict`): construction
// rejects unexpected keys. Returns the receiver for chaining.
func (s *StructType) Strict() *StructType {
	s.strict = true
	return s
}

// AsValue marks the struct a `Dry::Struct::Value` (comparable-by-value; the
// gem also freezes it — every instance here is already immutable). Returns the
// receiver for chaining.
func (s *StructType) AsValue() *StructType {
	s.value = true
	return s
}

// IsValue reports whether the struct was declared a `Dry::Struct::Value`.
func (s *StructType) IsValue() bool { return s.value }

// Attributes returns the declared attributes in declaration order. The slice
// must not be mutated.
func (s *StructType) Attributes() []Attribute { return s.attrs }

// Inherit returns a new [*StructType] named name that begins with a copy of the
// parent's attributes and config (dry-struct subclassing). Further
// [StructType.Attribute] calls append to (or override) the inherited set.
func (s *StructType) Inherit(name string) *StructType {
	child := New(name)
	child.transform = s.transform
	child.strict = s.strict
	child.value = s.value
	for _, a := range s.attrs {
		child.add(a)
	}
	return child
}
