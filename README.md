<p align="center"><img src="https://raw.githubusercontent.com/go-ruby-dry-struct/brand/main/social/go-ruby-dry-struct-dry-struct.png" alt="go-ruby-dry-struct/dry-struct" width="720"></p>

# dry-struct — go-ruby-dry-struct

[![Docs](https://img.shields.io/badge/docs-mkdocs--material-DC2626)](https://go-ruby-dry-struct.github.io/docs/)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.26.4%2B-00ADD8)](https://go.dev/dl/)
[![Coverage](https://img.shields.io/badge/coverage-100%25-1a7f37)](#tests--coverage)

**A pure-Go (no cgo) reimplementation of Ruby's
[dry-struct](https://dry-rb.org/gems/dry-struct/) gem** — typed, immutable value
objects whose attributes are coerced and validated by
[go-ruby-dry-types](https://github.com/go-ruby-dry-types/dry-types). Declare a
struct's attributes, construct it from a hash, and every attribute is coerced and
checked through its dry-type; the first failure raises a `Dry::Struct::Error`
whose message is **byte-identical** to the gem's (`[User.new] :age is missing in
Hash input`). No Ruby runtime required.

It is the `Dry::Struct` layer for
[go-embedded-ruby](https://github.com/go-embedded-ruby/ruby), but is a
**standalone, reusable** module — a sibling of
[go-ruby-dry-types](https://github.com/go-ruby-dry-types/dry-types) (the type
system it builds on).

> **What it is — and isn't.** Coercing, validating, comparing, and inspecting a
> typed value object over dry-types attributes is fully deterministic and needs
> **no interpreter**, so it lives here as pure Go. Registering the resulting
> Ruby class, evaluating a `default { ... }` block, or running an arbitrary
> constructor is the host's job; this library hands back a small, explicit value
> model (`*StructType`, `*Struct`, `drytypes.Symbol`, …) the host maps to and
> from its own objects.

## Features

Faithful port of `Dry::Struct`, validated against the `dry-struct` gem on every
supported platform:

- **Typed attributes** — `attribute :name, Types::String`,
  `attribute :age, Types::Coercible::Integer`; every dry-types type composes
  (strict, coercible, params, enum, constrained, optional, default, …).
- **Optional attributes** — `attribute? :port, …`: the key may be absent; it
  reads back `nil` and is omitted from `to_h`.
- **Nested structs and arrays of structs** — a `*StructType` is itself a
  `drytypes.Type`, so `attribute :address, AddressStruct` and
  `Types::Array.of(MemberStruct)` compose; `to_h` deep-converts.
- **Defaults** — `Types::String.default("anon")` fills in a missing attribute.
- **Construction** — `New`/`Call`/`MustNew`; passing an instance of the same
  type through unchanged; strict schemas (`schema schema.strict`) reject
  unexpected keys.
- **Immutable value semantics** — readers, `ToH`/`ToHash` (deep), `Attributes`,
  `With` (immutable copy with overrides), value equality (`Eql`), and a
  byte-faithful `Inspect` (`#<User name="Alice" age=30>`).
- **Struct config** — `TransformKeys` (symbolize/stringify), `Strict`, `AsValue`
  (`Dry::Struct::Value`), and `Inherit` (subclassing that adds/overrides
  attributes).

## Usage

```go
import (
	drystruct "github.com/go-ruby-dry-struct/dry-struct"
	drytypes "github.com/go-ruby-dry-types/dry-types"
)

user := drystruct.New("User").
	Attribute("name", drytypes.StrictString()).
	Attribute("age", drytypes.CoercibleInteger())

u, err := user.New(map[drytypes.Symbol]any{"name": "Alice", "age": "30"})
// u.Inspect() == `#<User name="Alice" age=30>`
// u.Fetch("age") == int64(30)
```

## Tests & coverage

The suite is **100% ruby-free by default**: deterministic golden vectors keep
coverage at 100% with no interpreter. A differential *oracle* layer additionally
compares construction results, `to_h`, error messages, `inspect`, and equality
against the real `dry-struct` gem when `ruby` is present (skipped otherwise, and
gated to `RUBY_VERSION >= "4.0"`). CI runs the 100%-coverage gate + a `-race`
lane (host cgo) + the six 64-bit arches (CGO=0) across Linux, macOS and Windows.

## License

BSD-3-Clause — see [LICENSE](LICENSE). Copyright the go-ruby-dry-struct/dry-struct
authors.

## WebAssembly

Being pure Go (CGO=0), this library also compiles to **WebAssembly** — both
`GOOS=js GOARCH=wasm` (browser / Node.js) and `GOOS=wasip1 GOARCH=wasm` (WASI).
CI builds both targets on every push, alongside the six 64-bit native/qemu arches.

```sh
GOOS=js     GOARCH=wasm go build ./...   # browser / Node
GOOS=wasip1 GOARCH=wasm go build ./...   # WASI (wasmtime, wasmer, wasmedge, …)
```
