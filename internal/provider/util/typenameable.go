// Package util is the package that contains all the utility functions in the provider.
package util

// TypeNameable is an interface that defines the TypeName method.
// It is implemented by the resource and the data source structs.
type TypeNameable interface {
	// TypeName is a method that returns the type name of the resource or the data source.
	TypeName() string
}
