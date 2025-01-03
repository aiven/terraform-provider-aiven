package util

import (
	"math/big"

	"golang.org/x/exp/constraints"
)

// This file contains helper functions that are more generic and can be used in multiple places.
// These functions are not specific to the Aiven plugin. If you are looking for Aiven plugin specific helpers,
// please see the pluginhelpers.go file instead.

// ToPtr is a helper function that returns a pointer to the value passed in.
func ToPtr[T any](v T) *T {
	return &v
}

// First is a helper function that returns the first argument passed in out of two.
func First[T any, U any](a T, _ U) T {
	return a
}

// ToBigFloat is a helper function that converts any integer or float type to a big.Float.
func ToBigFloat[T constraints.Integer | constraints.Float](v T) *big.Float {
	return big.NewFloat(float64(v))
}
