package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
			expect: "Change `foo` resource field `bar`: add `foo`, remove `bar`",
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
			name:   "change enum",
			expect: "Change `foo` resource field `bar`: add `foo`",
			kind:   ResourceRootType,
			old: &Item{
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Description: "Foo. The possible values is `bar`",
			},
			new: &Item{
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Description: "Foo. The possible values are `foo`, `bar`.",
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
			expect: "Change `foo` resource field `bar`: type ~~`list`~~ → `set`",
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

func TestSerializeDiff(t *testing.T) {
	list := []*Diff{
		{Action: AddDiffAction, RootType: ResourceRootType, Description: "foo", Item: &Item{Path: "aiven_opensearch.opensearch_user_config.azure_migration.include_aliases"}},
		{Action: ChangeDiffAction, RootType: DataSourceRootType, Description: "remove deprecation", Item: &Item{Path: "aiven_cassandra.cassandra_user_config.additional_backup_regions"}},
		{Action: ChangeDiffAction, RootType: ResourceRootType, Description: "remove deprecation", Item: &Item{Path: "aiven_cassandra.cassandra_user_config.additional_backup_regions"}},
		{Action: AddDiffAction, RootType: ResourceRootType, Description: "foo", Item: &Item{Path: "aiven_opensearch.opensearch_user_config.s3_migration.include_aliases"}},
		{Action: AddDiffAction, RootType: ResourceRootType, Description: "foo", Item: &Item{Path: "aiven_opensearch.opensearch_user_config.gcs_migration.include_aliases"}},
	}

	expect := []string{
		"Add `aiven_opensearch` resource field `opensearch_user_config.azure_migration.include_aliases`: foo",
		"Add `aiven_opensearch` resource field `opensearch_user_config.gcs_migration.include_aliases`: foo",
		"Add `aiven_opensearch` resource field `opensearch_user_config.s3_migration.include_aliases`: foo",
		"Change `aiven_cassandra` resource field `cassandra_user_config.additional_backup_regions`: remove deprecation",
		"Change `aiven_cassandra` datasource field `cassandra_user_config.additional_backup_regions`: remove deprecation",
	}

	actual := serializeDiff(list)
	assert.Empty(t, cmp.Diff(expect, actual))
}

func TestCmpList(t *testing.T) {
	cases := []struct {
		was, have []string
		expect    string
	}{
		{
			was:    []string{"a", "b", "c"},
			have:   []string{"a", "b", "c"},
			expect: "",
		},
		{
			was:    []string{"a", "b", "c"},
			have:   []string{"a", "b", "c", "d", "f"},
			expect: "add `d`, `f`",
		},
		{
			was:    []string{"a", "b", "c"},
			have:   []string{"a", "c"},
			expect: "remove `b`",
		},
		{
			was:    []string{"a", "b", "c", "f"},
			have:   []string{"a", "b", "c", "d"},
			expect: "add `d`, remove `f`",
		},
	}

	for _, opt := range cases {
		t.Run(opt.expect, func(t *testing.T) {
			assert.Equal(t, opt.expect, cmpList(opt.was, opt.have))
		})
	}
}
