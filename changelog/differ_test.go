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
		kind     ResourceType
		old, new *Item
	}{
		{
			name:   "change enums",
			expect: "Change resource `foo` field `bar`: enum ~~`bar`, `baz`~~ -> `foo`, `baz`",
			kind:   ResourceKind,
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
			expect: "Add resource `foo` field `bar`: Foo",
			kind:   ResourceKind,
			new: &Item{
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Description: "Foo",
			},
		},
		{
			name:   "remove resource field",
			expect: "Remove resource `foo` field `bar`: Foo",
			kind:   ResourceKind,
			old: &Item{
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Description: "Foo",
			},
		},
		{
			name:   "remove beta from the field",
			expect: "Change resource `foo` field `bar`: no longer beta",
			kind:   ResourceKind,
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
			expect: "Add resource `foo` _(beta)_: does stuff, PROVIDER_AIVEN_ENABLE_BETA",
			kind:   ResourceKind,
			new: &Item{
				Type:        schema.TypeString,
				Path:        "foo",
				Description: "does stuff, PROVIDER_AIVEN_ENABLE_BETA",
			},
		},
		{
			name:   "change type",
			expect: "Change resource `foo` field `bar`: type ~~`list`~~ -> `set`",
			kind:   ResourceKind,
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
