package main

import (
	"testing"

	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

// TestWalkPluginSchemaResource asserts that nested attributes and blocks are discovered.
// When they are not, their changes silently never reach the changelog.
func TestWalkPluginSchemaResource(t *testing.T) {
	s := resourceschema.Schema{
		Attributes: map[string]resourceschema.Attribute{
			"topic_name": resourceschema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"owner": resourceschema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]resourceschema.Attribute{
					"email": resourceschema.StringAttribute{Optional: true, Sensitive: true},
				},
			},
			"tags": resourceschema.SetNestedAttribute{
				Optional: true,
				NestedObject: resourceschema.NestedAttributeObject{
					Attributes: map[string]resourceschema.Attribute{
						"key": resourceschema.StringAttribute{Optional: true},
					},
				},
			},
		},
		Blocks: map[string]resourceschema.Block{
			"config": resourceschema.ListNestedBlock{
				NestedObject: resourceschema.NestedBlockObject{
					Attributes: map[string]resourceschema.Attribute{
						"retention_ms": resourceschema.Int64Attribute{Optional: true},
					},
					Blocks: map[string]resourceschema.Block{
						"limits": resourceschema.SetNestedBlock{
							NestedObject: resourceschema.NestedBlockObject{
								Attributes: map[string]resourceschema.Attribute{
									"max_bytes": resourceschema.Int64Attribute{Optional: true},
								},
							},
						},
					},
				},
			},
		},
	}

	items := walkItems(t, s, ResourceRootKind)

	assert.ElementsMatch(t, []string{
		"aiven_foo.topic_name",
		"aiven_foo.owner",
		"aiven_foo.owner.email",
		"aiven_foo.tags",
		"aiven_foo.tags.key",
		"aiven_foo.config",
		"aiven_foo.config.retention_ms",
		"aiven_foo.config.limits",
		"aiven_foo.config.limits.max_bytes",
	}, lo.Keys(items))

	assert.True(t, items["aiven_foo.topic_name"].ForceNew)
	assert.False(t, items["aiven_foo.tags.key"].ForceNew)
	assert.True(t, items["aiven_foo.owner.email"].Sensitive)
	assert.Equal(t, schema.TypeList, items["aiven_foo.config"].Type)
	assert.Equal(t, schema.TypeSet, items["aiven_foo.config.limits"].Type)
	assert.Equal(t, schema.TypeInt, items["aiven_foo.config.limits.max_bytes"].Type)
}

func TestWalkPluginSchemaDataSource(t *testing.T) {
	s := datasourceschema.Schema{
		Attributes: map[string]datasourceschema.Attribute{
			"plans": datasourceschema.MapNestedAttribute{
				Computed: true,
				NestedObject: datasourceschema.NestedAttributeObject{
					Attributes: map[string]datasourceschema.Attribute{
						"node_count": datasourceschema.Int64Attribute{Computed: true},
					},
				},
			},
		},
		Blocks: map[string]datasourceschema.Block{
			"service_plans": datasourceschema.SetNestedBlock{
				NestedObject: datasourceschema.NestedBlockObject{
					Attributes: map[string]datasourceschema.Attribute{
						"managed_cluster_plan": datasourceschema.BoolAttribute{Computed: true},
					},
				},
			},
		},
	}

	items := walkItems(t, s, DataSourceRootKind)

	assert.ElementsMatch(t, []string{
		"aiven_foo.plans",
		"aiven_foo.plans.node_count",
		"aiven_foo.service_plans",
		"aiven_foo.service_plans.managed_cluster_plan",
	}, lo.Keys(items))

	assert.Equal(t, schema.TypeMap, items["aiven_foo.plans"].Type)
	assert.Equal(t, schema.TypeBool, items["aiven_foo.service_plans.managed_cluster_plan"].Type)
}

func walkItems(t *testing.T, s any, kind RootKind) map[string]*Item {
	t.Helper()

	root := &Item{
		Name: "aiven_foo",
		Root: "aiven_foo",
		Path: "aiven_foo",
		Kind: kind,
		Type: schema.TypeList,
	}

	items := make(map[string]*Item)
	for _, item := range walkPluginSchema(s, root) {
		items[item.Path] = item
	}

	return items
}
