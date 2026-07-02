// Copyright (c) the go-ruby-dry-struct/dry-struct authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drystruct

import (
	drytypes "github.com/go-ruby-dry-types/dry-types"
)

// AttrType is the type of an attribute: it coerces-and-validates the attribute's
// value. Both a dry-types [drytypes.Type] (via [Wrap]) and a nested
// [*StructType] satisfy it, which is how nested structs compose as attribute
// types without dry-types needing to know about structs.
type AttrType interface {
	// Coerce coerces and validates v, returning the coerced value or an error
	// whose message matches the gem's.
	Coerce(v any) (any, error)
}

// dryType adapts a dry-types [drytypes.Type] to [AttrType].
type dryType struct{ t drytypes.Type }

func (d dryType) Coerce(v any) (any, error) { return d.t.Call(v) }

// arrayType is an [AttrType] for an array whose members are themselves an
// [AttrType] — used for `Types::Array.of(SomeStruct)` where the member is a
// nested [*StructType] (dry-types' own ArrayOf only takes a dry-types Type).
type arrayType struct{ elem AttrType }

func (a arrayType) Coerce(v any) (any, error) {
	arr, ok := v.([]any)
	if !ok {
		return nil, &drytypes.ConstraintError{
			Message: inspect(v) + " violates constraints (type?(Array, " + inspect(v) + ") failed)",
			Input:   v,
			Rule:    "type?(Array, " + inspect(v) + ") failed",
		}
	}
	out := make([]any, len(arr))
	for i, e := range arr {
		c, err := a.elem.Coerce(e)
		if err != nil {
			return nil, err
		}
		out[i] = c
	}
	return out, nil
}

// ArrayOf builds an [AttrType] for a member-typed array whose element type is
// any [AttrType] (a nested [*StructType] or a wrapped dry-types type). For a
// plain dry-types element you can also use [drytypes.ArrayOf] directly and pass
// the result to [StructType.Attribute].
func ArrayOf(elem AttrType) AttrType { return arrayType{elem: elem} }

// Wrap adapts a dry-types [drytypes.Type] into an [AttrType] so it can be used as
// an attribute type. [StructType.Attribute] accepts a [drytypes.Type] directly
// and wraps it for you; this is exposed for callers building [Attribute]s
// by hand.
func Wrap(t drytypes.Type) AttrType { return dryType{t: t} }

// Coerce lets a [*StructType] serve as an attribute type (nested struct): it
// constructs the nested [*Struct] from v, returning the gem-shaped error on
// failure so the enclosing schema can wrap it.
func (s *StructType) Coerce(v any) (any, error) { return s.New(v) }

// defaultValue resolves the attribute's default for an absent key: it feeds the
// dry-types Undefined sentinel to the type, and if the type substitutes a
// concrete value (i.e. it was declared `.default(...)`), returns that value. A
// type without a default passes Undefined through (or errors), reported as no
// default. Only dry-types attribute types can carry a default; a nested struct
// never does.
func (a Attribute) defaultValue() (any, bool) {
	dt, ok := a.Type.(dryType)
	if !ok {
		return nil, false
	}
	out, err := dt.t.Call(drytypes.Undefined)
	if err != nil {
		return nil, false
	}
	// A type without a default passes Undefined through unchanged; a defaulted
	// type substitutes a concrete value. Undefined is a comparable exported
	// sentinel, so identity distinguishes the two.
	if out == drytypes.Undefined {
		return nil, false
	}
	return out, true
}
