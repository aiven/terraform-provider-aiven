package main

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

func entryText(t *testing.T, old, new *Item) string {
	t.Helper()
	got, err := diffItems(old, new)
	require.NoError(t, err)
	require.NotNil(t, got)
	return generatedEntry(got).Text
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		expect   string
		old, new *Item
	}{
		{
			name:   "change enums",
			expect: "`foo` resource field `bar` (enum): add `foo`; remove `bar`",
			old: &Item{
				Kind:        ResourceRootKind,
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Root:        "foo",
				Description: userconfig.Desc("Foo.").PossibleValuesString("bar", "baz").Build(),
			},
			new: &Item{
				Kind:        ResourceRootKind,
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Root:        "foo",
				Description: userconfig.Desc("Foo.").PossibleValuesString("foo", "baz").Build(),
			},
		},
		{
			name:   "add resource field",
			expect: "`foo` resource field `bar`: Foo.",
			new: &Item{
				Kind:        ResourceRootKind,
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Root:        "foo",
				Description: "Foo",
			},
		},
		{
			name:   "remove resource field",
			expect: "`foo` resource field `bar`: removed",
			old: &Item{
				Kind:        ResourceRootKind,
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Root:        "foo",
				Description: "Foo",
			},
		},
		{
			name:   "remove beta from the field",
			expect: "`foo` resource field `bar`: no longer beta",
			old: &Item{
				Kind:        ResourceRootKind,
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Root:        "foo",
				Description: "PROVIDER_AIVEN_ENABLE_BETA",
			},
			new: &Item{
				Kind:        ResourceRootKind,
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Root:        "foo",
				Description: "Foo",
			},
		},
		{
			name:   "add beta resource",
			expect: "`foo` resource _(beta)_: Does stuff, `PROVIDER_AIVEN_ENABLE_BETA`.",
			new: &Item{
				Kind:        ResourceRootKind,
				Type:        schema.TypeString,
				Path:        "foo",
				Root:        "foo",
				Description: "does stuff, PROVIDER_AIVEN_ENABLE_BETA",
			},
		},
		{
			name:   "deprecation added",
			expect: "`foo` resource field `bar`: deprecated: use other",
			old: &Item{
				Kind: ResourceRootKind,
				Path: "foo.bar",
				Root: "foo",
			},
			new: &Item{
				Kind:        ResourceRootKind,
				Path:        "foo.bar",
				Root:        "foo",
				Deprecated:  "use other",
				Description: "Foo",
			},
		},
		{
			name:   "change type",
			expect: "`foo` resource field `bar`: type was `list`, now `set`",
			old: &Item{
				Kind: ResourceRootKind,
				Type: schema.TypeList,
				Path: "foo.bar",
				Root: "foo",
			},
			new: &Item{
				Kind: ResourceRootKind,
				Type: schema.TypeSet,
				Path: "foo.bar",
				Root: "foo",
			},
		},
		{
			name:   "user config add",
			expect: "Added `auth_azuread.org_mapping`: Map Azure AD groups.",
			new: &Item{
				Kind:        ResourceRootKind,
				Path:        "aiven_grafana.grafana_user_config.auth_azuread.org_mapping",
				Root:        "aiven_grafana",
				Description: "Map Azure AD groups",
			},
		},
		{
			name:   "user config remove",
			expect: "Removed `grafana_version`",
			old: &Item{
				Kind: ResourceRootKind,
				Path: "aiven_grafana.grafana_user_config.grafana_version",
				Root: "aiven_grafana",
			},
		},
		{
			name:   "limited availability",
			expect: "`foo` resource field `bar`: now limited availability",
			old: &Item{
				Kind:        ResourceRootKind,
				Path:        "foo.bar",
				Root:        "foo",
				Description: "Foo",
			},
			new: &Item{
				Kind:        ResourceRootKind,
				Path:        "foo.bar",
				Root:        "foo",
				Description: userconfig.LimitedAvailabilityMessage,
			},
		},
		{
			name:   "data source only suffix",
			expect: "`foo` data source field `bar`: Foo.",
			new: &Item{
				Kind:        DataSourceRootKind,
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Root:        "foo",
				Description: "Foo",
			},
		},
	}

	for _, opt := range tests {
		t.Run(opt.name, func(t *testing.T) {
			assert.Equal(t, opt.expect, entryText(t, opt.old, opt.new))
		})
	}
}

func TestSerializeDiff(t *testing.T) {
	list := []*Diff{
		{
			Action: AddDiffAction,
			Item: &Item{
				Kind: ResourceRootKind, Root: "aiven_opensearch",
				Path: "aiven_opensearch.opensearch_user_config.azure_migration.include_aliases",
			},
		},
		{
			Action: ChangeDiffAction, Description: "no longer deprecated",
			Item: &Item{
				Kind: DataSourceRootKind, Root: "aiven_pg",
				Path: "aiven_pg.pg_user_config.additional_backup_regions",
			},
		},
		{
			Action: ChangeDiffAction, Description: "no longer deprecated",
			Item: &Item{
				Kind: ResourceRootKind, Root: "aiven_pg",
				Path: "aiven_pg.pg_user_config.additional_backup_regions",
			},
		},
		{
			Action: AddDiffAction,
			Item: &Item{
				Kind: ResourceRootKind, Root: "aiven_opensearch",
				Path: "aiven_opensearch.opensearch_user_config.s3_migration.include_aliases",
			},
		},
		{
			Action: AddDiffAction,
			Item: &Item{
				Kind: DataSourceRootKind, Root: "aiven_opensearch",
				Path: "aiven_opensearch.opensearch_user_config.s3_migration.include_aliases",
			},
		},
	}

	actual := serializeDiff(list)
	texts := make([]string, 0, len(actual))
	for _, e := range actual {
		texts = append(texts, e.Text)
	}

	want := []string{
		"Added `azure_migration.include_aliases`",
		"Added `s3_migration.include_aliases`",
		"Changed `additional_backup_regions`: no longer deprecated",
	}
	assert.Empty(t, cmp.Diff(want, texts))
}

func TestCmpList(t *testing.T) {
	cases := []struct {
		was, have []string
		expect    string
	}{
		{was: []string{"a", "b"}, have: []string{"a", "b"}, expect: ""},
		{was: []string{"a", "b"}, have: []string{"a", "b", "d"}, expect: "add `d`"},
		{was: []string{"a", "b"}, have: []string{"a"}, expect: "remove `b`"},
		{was: []string{"a", "b", "f"}, have: []string{"a", "b", "d"}, expect: "add `d`; remove `f`"},
		{was: []string{"a", "b"}, have: []string{"a", "b", "d", "f"}, expect: "add `d` and `f`"},
		{was: []string{"a", "b", "c"}, have: []string{"a"}, expect: "remove `b` and `c`"},
	}

	for _, opt := range cases {
		t.Run(opt.expect, func(t *testing.T) {
			assert.Equal(t, opt.expect, cmpList(opt.was, opt.have))
		})
	}
}

func TestChangeDetailsDeterministic(t *testing.T) {
	old := &Item{
		Kind:        ResourceRootKind,
		Type:        schema.TypeString,
		Path:        "foo.bar",
		Root:        "foo",
		Description: userconfig.Desc("").PossibleValuesString("a").Build(),
		Sensitive:   false,
		MaxItems:    0,
	}
	new := &Item{
		Kind:        ResourceRootKind,
		Type:        schema.TypeInt,
		Path:        "foo.bar",
		Root:        "foo",
		Description: userconfig.Desc("").PossibleValuesString("b").Build(),
		Sensitive:   true,
		MaxItems:    100,
	}
	got, err := diffItems(old, new)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t,
		"add `b`; remove `a`; max items was `0`, now `100`; sensitive was `false`, now `true`; type was `string`, now `int`",
		got.Description,
	)
}

func TestAddEnumDetail(t *testing.T) {
	orgListDesc := userconfig.Desc("An enumeration.").PossibleValuesString("active", "deactivated", "deleted").Build()
	sdkKafka := fmt.Sprintf("%s%s.", userconfig.EnumPrefix, "`3.9`, `4.2`, `4.3`, and newer") +
		" Kafka MirrorMaker major version."
	eventSched := fmt.Sprintf("%s%s.", userconfig.EnumPrefix, "`ON`, `OFF`") +
		" Whether the event scheduler runs events."

	tests := []struct {
		name       string
		expect     string
		new        *Item
		enumValues []string
	}{
		{
			name: "flat possible values with lead-in",
			expect: "`aiven_organization_user_list` data source field `users.state` (enum): " +
				"An enumeration. The possible values are `active`, `deactivated` and `deleted`.",
			enumValues: []string{"active", "deactivated", "deleted"},
			new: &Item{
				Kind:        DataSourceRootKind,
				Type:        schema.TypeString,
				Path:        "aiven_organization_user_list.users.state",
				Root:        "aiven_organization_user_list",
				Description: orgListDesc,
			},
		},
		{
			name:       "flat possible values only",
			expect:     "`foo` resource field `bar` (enum): The possible values are `a`, `b` and `c`.",
			enumValues: []string{"a", "b", "c"},
			new: &Item{
				Kind:        ResourceRootKind,
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Root:        "foo",
				Description: userconfig.Desc("").PossibleValuesString("a", "b", "c").Build(),
			},
		},
		{
			name: "sdk enum sentence",
			expect: "`aiven_kafka` resource field `version` (enum): " +
				"Kafka MirrorMaker major version. The possible values are `3.9`, `4.2` and `4.3`.",
			enumValues: []string{"3.9", "4.2", "4.3"},
			new: &Item{
				Kind:        ResourceRootKind,
				Type:        schema.TypeString,
				Path:        "aiven_kafka.version",
				Root:        "aiven_kafka",
				Description: sdkKafka,
			},
		},
		{
			name: "user config child enum",
			expect: "Added `event_scheduler` (enum): " +
				"Whether the event scheduler runs events. The possible values are `OFF` and `ON`.",
			enumValues: []string{"ON", "OFF"},
			new: &Item{
				Kind:        ResourceRootKind,
				Path:        "aiven_mysql.mysql_user_config.event_scheduler",
				Root:        "aiven_mysql",
				Description: eventSched,
			},
		},
		{
			name:   "non enum add unchanged",
			expect: "`foo` resource field `bar`: Plain description.",
			new: &Item{
				Kind:        ResourceRootKind,
				Type:        schema.TypeString,
				Path:        "foo.bar",
				Root:        "foo",
				Description: "Plain description",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := entryText(t, nil, tt.new)
			assert.Equal(t, tt.expect, got)
			if len(tt.enumValues) > 0 {
				assert.Contains(t, got, "The possible value")
				for _, v := range tt.enumValues {
					assert.Contains(t, got, "`"+v+"`", "missing enum value %q in %q", v, got)
				}
			}
		})
	}
}
