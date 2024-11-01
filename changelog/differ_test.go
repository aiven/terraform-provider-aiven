package main

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		expect   string
		kind     RootType
		old, new *Item
	}{
		{
			name:   "change enums",
			expect: "Change `foo` resource field `bar`: enum ~~`bar`, `baz`~~ -> `foo`, `baz`",
			kind:   ResourceRootType,
			old: &Item{
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Description: "Foo. The possible values are `bar`, `baz`.",
			},
			new: &Item{
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Description: "Foo. The possible values are `foo`, `baz`.",
			},
		},
		{
			name:   "add resource field",
			expect: "Add `foo` resource field `bar`: Foo",
			kind:   ResourceRootType,
			new: &Item{
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Description: "Foo",
			},
		},
		{
			name:   "remove resource field",
			expect: "Remove `foo` resource field `bar`: Foo",
			kind:   ResourceRootType,
			old: &Item{
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Description: "Foo",
			},
		},
		{
			name:   "remove beta from the field",
			expect: "Change `foo` resource field `bar`: no longer beta",
			kind:   ResourceRootType,
			old: &Item{
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Description: "PROVIDER_AIVEN_ENABLE_BETA",
			},
			new: &Item{
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Description: "Foo",
			},
		},
		{
			name:   "add beta resource",
			expect: "Add `foo` resource _(beta)_: does stuff, PROVIDER_AIVEN_ENABLE_BETA",
			kind:   ResourceRootType,
			new: &Item{
				Type:        schema.TypeString,
				Path:        "foo",
				Description: "does stuff, PROVIDER_AIVEN_ENABLE_BETA",
			},
		},
		{
			name:   "change type",
			expect: "Change `foo` resource field `bar`: type ~~`list`~~ -> `set`",
			kind:   ResourceRootType,
			old: &Item{
				Type: schema.TypeList,
				Path: "foo.bar",
			},
			new: &Item{
				Type: schema.TypeSet,
				Path: "foo.bar",
			},
		},
	}

	for _, opt := range tests {
		t.Run(opt.name, func(t *testing.T) {
			got, err := diffItems(opt.kind, opt.old, opt.new)
			assert.NoError(t, err)
			assert.Equal(t, opt.expect, got.String())
		})
	}
}
