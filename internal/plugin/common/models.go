package common

import (
	"reflect"
	"strings"
)

type Index int

const (
	ProjectIndex  Index = 0
	ServiceIndex  Index = 1
	DatabaseIndex Index = 2
)

func PathIndex(s string, i Index) string {
	list := strings.Split(s, "/")
	if int(i) < len(list) {
		return list[i]
	}
	return ""
}

func PathIndexPointer(s string, i Index) *string {
	v := PathIndex(s, i)
	if v == "" {
		return nil
	}
	return &v
}

func SplitN(s string, n int) ([]string, bool) {
	parts := strings.Split(s, "/")
	return parts, len(parts) == n
}

// Copy
// Plugin framework doesn't support embedded structs.
// https://github.com/hashicorp/terraform-plugin-framework/issues/242
// We can't share the same functions for resource and datasource by default.
// To do so, we design all functions for resource and copy it to datasource.
func Copy(from, dest any) func() {
	copyModel(from, dest)
	return func() {
		copyModel(dest, from)
	}
}

func copyModel(from, dest any) {
	typeFrom := reflect.TypeOf(from).Elem()
	typeDest := reflect.TypeOf(dest).Elem()

	fields := make(map[string]int, typeDest.NumField())
	for i := 0; i < typeDest.NumField(); i++ {
		fields[typeDest.Field(i).Name] = i
	}

	elemFrom := reflect.ValueOf(from).Elem()
	elemDest := reflect.ValueOf(dest).Elem()
	for j := 0; j < typeFrom.NumField(); j++ {
		if i, ok := fields[typeFrom.Field(j).Name]; ok {
			elemDest.Field(i).Set(elemFrom.Field(j))
		}
	}
}
