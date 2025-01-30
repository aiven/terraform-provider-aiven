package template

import (
	"fmt"
)

// Template represents a single Terraform configuration template
type Template struct {
	Name     string
	Template string
}

// Value represents a value that can either be a literal string or a reference
type Value struct {
	Value     string
	IsLiteral bool
}

// Literal creates a new literal value
func Literal(v string) Value {
	return Value{Value: v, IsLiteral: true}
}

// Reference creates a new reference value
func Reference(v string) Value {
	return Value{Value: v, IsLiteral: false}
}

// String returns the properly formatted value based on whether it's a literal or reference
func (v Value) String() string {
	if v.IsLiteral {
		return fmt.Sprintf("%q", v.Value)
	}
	return v.Value
}

// MarshalText implements encoding.TextMarshaler
func (v Value) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
