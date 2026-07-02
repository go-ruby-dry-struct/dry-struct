// Copyright (c) the go-ruby-dry-struct/dry-struct authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drystruct

import (
	"fmt"
	"math/big"
	"strings"

	drytypes "github.com/go-ruby-dry-types/dry-types"
)

// inspect renders a Ruby value the way Object#inspect does, for the shapes that
// appear as dry-struct attribute values (scalars, symbols, arrays, hashes and
// nested [*Struct]s). It matches the value corpus dry-struct's own `#inspect`
// prints (a nested struct renders as its own `#<Name …>`).
func inspect(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case bool:
		if x {
			return "true"
		}
		return "false"
	case string:
		return rubyStringInspect(x)
	case drytypes.Symbol:
		return ":" + string(x)
	case int:
		return fmt.Sprintf("%d", x)
	case int32:
		return fmt.Sprintf("%d", x)
	case int64:
		return fmt.Sprintf("%d", x)
	case *big.Int:
		return x.String()
	case float64:
		return formatFloat(x)
	case []any:
		parts := make([]string, len(x))
		for i, e := range x {
			parts[i] = inspect(e)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case *drytypes.Map:
		return mapInspect(x)
	case map[string]any, map[drytypes.Symbol]any, map[any]any:
		if m, ok := asMap(v); ok {
			return mapInspect(m)
		}
	case *Struct:
		return x.Inspect()
	}
	return fmt.Sprintf("%v", v)
}

// valueClass is the Ruby class name of v, used in a schema error's
// `<v> (<Class>) has invalid type …` prefix (matching dry-types' valueClass).
func valueClass(v any) string {
	switch v.(type) {
	case nil:
		return "NilClass"
	case bool:
		if v.(bool) {
			return "TrueClass"
		}
		return "FalseClass"
	case string:
		return "String"
	case drytypes.Symbol:
		return "Symbol"
	case int, int32, int64, *big.Int:
		return "Integer"
	case float64:
		return "Float"
	case []any:
		return "Array"
	case *drytypes.Map, map[string]any, map[drytypes.Symbol]any, map[any]any:
		return "Hash"
	case *Struct:
		return v.(*Struct).typ.Name
	}
	return "Object"
}

// mapInspect renders an ordered map the way Ruby 4.0's Hash#inspect does: the
// `sym: v` shorthand for symbol keys, `k => v` otherwise.
func mapInspect(m *drytypes.Map) string {
	parts := make([]string, 0, m.Len())
	for _, p := range m.Pairs() {
		if sym, ok := p.Key.(drytypes.Symbol); ok {
			parts = append(parts, string(sym)+": "+inspect(p.Val))
		} else {
			parts = append(parts, inspect(p.Key)+" => "+inspect(p.Val))
		}
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// rubyStringInspect renders a Go string the way Ruby's String#inspect does for
// the characters that appear in the struct value corpus.
func rubyStringInspect(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// formatFloat renders a float64 the way Ruby's Float#to_s does (always at least
// one fractional digit).
func formatFloat(f float64) string {
	s := fmt.Sprintf("%g", f)
	if !strings.ContainsAny(s, ".eEnN") {
		s += ".0"
	}
	return s
}
