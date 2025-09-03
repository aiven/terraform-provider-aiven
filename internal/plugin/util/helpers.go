package util

import (
	"os"
	"strconv"
)

// This file contains helper functions that are more generic and can be used in multiple places.
// These functions are not specific to the Aiven plugin. If you are looking for Aiven plugin specific helpers,
// please see the pluginhelpers.go file instead.

// ToPtr is a helper function that returns a pointer to the value passed in.
func ToPtr[T any](v T) *T {
	return &v
}

// NilIfZero returns a pointer to the value, or nil if the value equals its zero value
func NilIfZero[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}

	return &v
}

func EnvBool(key string, defaultValue bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	b, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return b
}

func MergeSlices[T any](slices ...[]T) []T {
	var result []T
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return result
}
