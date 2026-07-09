# Ruby examples

Pure-Ruby examples of `dry-struct` — the `Dry::Struct` layer this library backs.
They run under [go-embedded-ruby](https://github.com/go-embedded-ruby/ruby) (rbgo)
via the `require "dry/struct"` binding (attribute types come from
`require "dry/types"`).

```sh
rbgo examples/dry_struct_usage.rb
```

| File | Shows |
| --- | --- |
| [`dry_struct_usage.rb`](dry_struct_usage.rb) | Declaring strict/coercible/optional/nested attributes, coercing `.new`, readers, `to_h`, `with`, value equality, `inspect`, and rescuing `Dry::Struct::Error`. |
