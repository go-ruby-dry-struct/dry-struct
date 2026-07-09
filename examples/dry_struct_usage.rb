# frozen_string_literal: true
#
# Usage of Dry::Struct — typed, immutable value objects whose attributes are
# coerced and validated through their dry-types. Runs under go-embedded-ruby
# (rbgo); see examples/README.md.

require "dry/types"
require "dry/struct"

# A nested struct is itself a type, so it composes as an attribute.
class Address < Dry::Struct
  attribute :city, Dry::Types["strict.string"]
end

class User < Dry::Struct
  attribute  :name,  Dry::Types["strict.string"]     # strict: must already be a String
  attribute  :age,   Dry::Types["coercible.integer"] # coercible: "30" -> 30
  attribute  :addr,  Address                          # nested struct attribute
  attribute? :email, Dry::Types["strict.string"]     # optional: key may be absent
end

# Construct from a hash; each attribute is coerced through its dry-type.
u = User.new(name: "Ada", age: "30", addr: { city: "Paris" })
puts u.name                 # => Ada
puts u.age                  # => 30   (coerced from the String "30")
puts u.addr.city            # => Paris
p u.email                   # => nil  (optional attribute was omitted)

# Immutable value semantics: to_h deep-converts, `with` returns a copy.
p u.to_h                    # => {name: "Ada", age: 30, addr: {city: "Paris"}}
p u.with(age: 31).age       # => 31
p u.age                     # => 30   (the original is unchanged)
puts u.inspect              # => #<User name="Ada" age=30 addr=#<Address city="Paris"> email=nil>

# Value equality: two structs with equal attributes are equal.
p(u == User.new(name: "Ada", age: 30, addr: { city: "Paris" })) # => true

# A coercion/validation failure raises Dry::Struct::Error.
begin
  User.new(name: "Ada", addr: { city: "Paris" }) # :age is missing
rescue Dry::Struct::Error => e
  puts e.message            # => [User.new] :age is missing in Hash input
end
