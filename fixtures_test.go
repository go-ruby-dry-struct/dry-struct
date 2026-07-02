// Copyright (c) the go-ruby-dry-struct/dry-struct authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drystruct

import (
	drytypes "github.com/go-ruby-dry-types/dry-types"
)

// This file mirrors, in Go, the exact Dry::Struct classes the Ruby oracle and
// golden generator define (testdata/golden.json). Every case in the golden set
// is reproduced here so the deterministic suite validates parity ruby-free.

func sym(m map[string]any) map[drytypes.Symbol]any {
	out := make(map[drytypes.Symbol]any, len(m))
	for k, v := range m {
		out[drytypes.Symbol(k)] = v
	}
	return out
}

// omap builds an ordered *drytypes.Map from alternating key/value args, so a
// fixture can preserve declared key order (matching a Ruby hash literal).
func omap(kv ...any) *drytypes.Map {
	m := drytypes.NewMap()
	for i := 0; i+1 < len(kv); i += 2 {
		k := kv[i]
		if s, ok := k.(string); ok {
			k = drytypes.Symbol(s)
		}
		m.Set(k, kv[i+1])
	}
	return m
}

func addressType() *StructType {
	return New("Address").
		Attribute("street", drytypes.StrictString()).
		Attribute("city", drytypes.StrictString())
}

func personType() *StructType {
	return New("Person").
		Attribute("name", drytypes.StrictString()).
		AttributeType("address", addressType())
}

func cfgType() *StructType {
	return New("Cfg").
		Attribute("host", drytypes.StrictString()).
		AttributeOpt("port", drytypes.CoercibleInteger())
}

func srvType() *StructType {
	return New("Srv").
		Attribute("name", drytypes.StrictString().Default("anon")).
		Attribute("count", drytypes.StrictInteger().Default(int64(0)))
}

func tagsType() *StructType {
	return New("Tags").
		Attribute("names", drytypes.ArrayOf(drytypes.CoercibleString()))
}

func teamType() *StructType {
	return New("Team").
		AttributeType("members", ArrayOf(personType()))
}

func strictType() *StructType {
	return New("StrictS").Strict().
		Attribute("name", drytypes.StrictString())
}

func emptyType() *StructType { return New("Empty") }

// buildCase runs the named golden case and returns the resulting struct or error.
func buildCase(name string) (*Struct, error) {
	switch name {
	case "basic":
		return personType().New(omap("name", "A", "address", omap("street", "s", "city", "c")))
	case "nested_bad":
		return personType().New(omap("name", "A", "address", omap("street", int64(1), "city", "c")))
	case "nested_missing":
		return personType().New(omap("name", "A", "address", omap("street", "s")))
	case "cfg_noport":
		return cfgType().New(omap("host", "h"))
	case "cfg_port":
		return cfgType().New(omap("host", "h", "port", "80"))
	case "srv_default":
		return srvType().New(omap())
	case "srv_override":
		return srvType().New(omap("name", "x", "count", int64(5)))
	case "tags":
		return tagsType().New(omap("names", []any{int64(1), int64(2), "x"}))
	case "team":
		return teamType().New(omap("members", []any{omap("name", "A", "address", omap("street", "s", "city", "c"))}))
	case "strict_ok":
		return strictType().New(omap("name", "x"))
	case "strict_bad":
		return strictType().New(omap("name", "x", "extra", int64(1)))
	case "empty":
		return emptyType().New(omap())
	case "missing_key":
		return personType().New(omap("name", "A"))
	case "coerce_fail":
		return cfgType().New(omap("host", "h", "port", "abc"))
	case "wrong_type":
		return personType().New(omap("name", int64(5), "address", omap("street", "s", "city", "c")))
	case "not_hash":
		return personType().New("nothash")
	}
	return nil, nil
}
