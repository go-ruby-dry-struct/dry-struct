// Copyright (c) the go-ruby-dry-struct/dry-struct authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drystruct

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

// oracleScript builds, in Ruby, the exact Dry::Struct classes that
// fixtures_test.go mirrors in Go, runs every golden case, and prints the
// resulting inspect / to_h / error as JSON. TestOracle compares this live gem
// output against the Go implementation, so parity is proven against the real
// gem — not just the checked-in golden vectors.
//
// The script is gated to RUBY_VERSION >= "4.0" (per the differential-oracle
// convention) and self-skips if the dry-struct gem is unavailable; the
// deterministic golden suite holds the coverage gate where ruby/the gem is
// absent (the qemu cross-arch and Windows CI lanes).
const oracleScript = `
if Gem::Version.new(RUBY_VERSION) < Gem::Version.new("4.0")
  STDERR.puts "SKIP: ruby #{RUBY_VERSION} < 4.0"; exit 42
end
begin
  require "dry-struct"
rescue LoadError
  STDERR.puts "SKIP: dry-struct gem absent"; exit 42
end
require "json"
module Types; include Dry.Types(); end

class Address < Dry::Struct
  attribute :street, Types::String
  attribute :city, Types::String
end
class Person < Dry::Struct
  attribute :name, Types::String
  attribute :address, Address
end
class Cfg < Dry::Struct
  attribute :host, Types::String
  attribute? :port, Types::Coercible::Integer
end
class Srv < Dry::Struct
  attribute :name, Types::String.default("anon".freeze)
  attribute :count, Types::Integer.default(0)
end
class Tags < Dry::Struct
  attribute :names, Types::Array.of(Types::Coercible::String)
end
class Team < Dry::Struct
  attribute :members, Types::Array.of(Person)
end
class StrictS < Dry::Struct
  schema schema.strict
  attribute :name, Types::String
end
class Empty < Dry::Struct
end

def cap(name)
  v = yield
  { "name" => name, "ok" => true, "inspect" => v.inspect, "to_h" => v.to_h.inspect }
rescue => e
  { "name" => name, "ok" => false, "error" => e.message, "error_class" => e.class.name }
end

out = []
out << cap("basic") { Person.new(name: "A", address: {street: "s", city: "c"}) }
out << cap("nested_bad") { Person.new(name: "A", address: {street: 1, city: "c"}) }
out << cap("nested_missing") { Person.new(name: "A", address: {street: "s"}) }
out << cap("cfg_noport") { Cfg.new(host: "h") }
out << cap("cfg_port") { Cfg.new(host: "h", port: "80") }
out << cap("srv_default") { Srv.new({}) }
out << cap("srv_override") { Srv.new(name: "x", count: 5) }
out << cap("tags") { Tags.new(names: [1, 2, "x"]) }
out << cap("team") { Team.new(members: [{name: "A", address: {street: "s", city: "c"}}]) }
out << cap("strict_ok") { StrictS.new(name: "x") }
out << cap("strict_bad") { StrictS.new(name: "x", extra: 1) }
out << cap("empty") { Empty.new({}) }
out << cap("missing_key") { Person.new(name: "A") }
out << cap("coerce_fail") { Cfg.new(host: "h", port: "abc") }
out << cap("wrong_type") { Person.new(name: 5, address: {street: "s", city: "c"}) }
out << cap("not_hash") { Person.new("nothash") }
$stdout.binmode
print JSON.generate(out)
`

// TestOracle runs the real dry-struct gem and asserts the Go implementation is
// byte-identical on inspect / to_h / error for every case.
func TestOracle(t *testing.T) {
	bin, err := exec.LookPath("ruby")
	if err != nil {
		t.Skip("ruby not on PATH; skipping dry-struct MRI oracle")
	}
	cmd := exec.Command(bin, "-e", oracleScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 42 {
			t.Skipf("oracle self-skip: %s", strings.TrimSpace(string(out)))
		}
		t.Fatalf("ruby error: %v\noutput:\n%s", err, out)
	}
	var cases []goldenCase
	if err := json.Unmarshal(out, &cases); err != nil {
		t.Fatalf("decode oracle output: %v\nraw:\n%s", err, out)
	}
	if len(cases) == 0 {
		t.Fatal("oracle produced no cases")
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := buildCase(c.Name)
			if c.OK {
				if err != nil {
					t.Fatalf("gem succeeded, Go errored: %v", err)
				}
				if ins := got.Inspect(); ins != c.Inspect {
					t.Errorf("inspect:\n go  %q\n gem %q", ins, c.Inspect)
				}
				if th := mapInspect(got.ToH()); th != c.ToH {
					t.Errorf("to_h:\n go  %q\n gem %q", th, c.ToH)
				}
				return
			}
			if err == nil {
				t.Fatalf("gem errored %q, Go succeeded", c.Error)
			}
			if err.Error() != c.Error {
				t.Errorf("error:\n go  %q\n gem %q", err.Error(), c.Error)
			}
		})
	}
}
