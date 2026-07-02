// Copyright (c) the go-ruby-dry-struct/dry-struct authors
//
// SPDX-License-Identifier: BSD-3-Clause

// Package drystruct is a pure-Go (CGO-free) MRI-faithful reimplementation of the
// Ruby dry-struct gem: typed, immutable value objects whose attributes are
// coerced and validated by [github.com/go-ruby-dry-types/dry-types] types.
//
// A [*StructType] is the analogue of a `Dry::Struct` subclass: you register
// attributes on it (each a name plus a dry-types [drytypes.Type]), then
// construct instances from an attribute hash. Construction coerces and validates
// every attribute through its type, and — on the first failure — raises a
// [*Error] whose message is byte-identical to the dry-struct gem's
// (`[Name.new] <schema error>`).
//
// # Ruby value model
//
// Attribute values use the same small Go value model the go-ruby-* ecosystem
// (and go-ruby-dry-types) uses, so a host (go-embedded-ruby / rbgo) maps its
// object graph to and from this package with no glue:
//
//	Ruby            Go
//	----            --
//	nil             nil
//	true / false    bool
//	Integer         int64, *big.Int
//	Float           float64
//	String          string
//	Symbol          drytypes.Symbol
//	Array           []any
//	Hash            *drytypes.Map (ordered)
//	Dry::Struct     *Struct
//
// A [*Struct] is one immutable instance: an ordered map of attribute name to
// coerced value, tagged with its [*StructType]. It answers the reader,
// [Struct.ToH], [Struct.With], [Struct.Eql] and [Struct.Inspect] operations the
// gem exposes.
package drystruct

import (
	drytypes "github.com/go-ruby-dry-types/dry-types"
)

// Attribute is one declared member of a [*StructType]: its symbol name, the
// [AttrType] (a dry-types type or a nested [*StructType]) that coerces-and-
// validates its value, and whether it is optional (declared with `attribute?`,
// i.e. the key may be absent).
type Attribute struct {
	// Name is the attribute's symbol name (Ruby `:name`).
	Name drytypes.Symbol
	// Type coerces and validates the attribute's value.
	Type AttrType
	// Optional reports whether the attribute was declared with `attribute?`.
	Optional bool
}

// KeyTransform names how a [*StructType] normalizes incoming hash keys before
// matching them to attributes (dry-struct's `transform_keys`).
type KeyTransform int

const (
	// KeyNone leaves keys unchanged (the default): keys are matched as given,
	// with a String key accepted for a Symbol attribute of the same name.
	KeyNone KeyTransform = iota
	// KeySymbolize maps every String key to a Symbol (`transform_keys(&:to_sym)`).
	KeySymbolize
	// KeyStringify maps every Symbol key to a String, then back to a Symbol for
	// matching (`transform_keys(&:to_s)`) — dry-struct still requires the
	// declared attribute name.
	KeyStringify
)
