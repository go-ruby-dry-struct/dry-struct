// Copyright (c) the go-ruby-dry-struct/dry-struct authors
//
// SPDX-License-Identifier: BSD-3-Clause

package drystruct

// Error is raised when constructing a [*Struct] fails: an attribute is missing,
// an unexpected key is present under a strict schema, or an attribute's value
// fails its dry-types coercion/constraint. Its message is byte-identical to
// Dry::Struct::Error#message: the struct's failing-schema message prefixed with
// `[<Name>.new] ` (e.g. `[User.new] :age is missing in Hash input`).
type Error struct {
	// Message is the full, gem-faithful error message.
	Message string
	// Cause is the underlying dry-types error (schema / coercion / constraint /
	// missing-key / unknown-keys) that this wraps.
	Cause error
}

func (e *Error) Error() string { return e.Message }

// Unwrap exposes the underlying dry-types error for errors.Is / errors.As.
func (e *Error) Unwrap() error { return e.Cause }

// newError wraps an underlying dry-types failure into a gem-shaped [*Error] for
// the struct named name: `[<name>.new] <underlying message>`.
func newError(name string, cause error) *Error {
	return &Error{Message: "[" + name + ".new] " + cause.Error(), Cause: cause}
}
