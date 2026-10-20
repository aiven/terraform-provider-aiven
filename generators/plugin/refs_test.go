package main

import (
	"slices"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func refTestRoot(props ...*Item) *Item {
	root := &Item{Name: "root", Type: SchemaTypeObject, Properties: map[string]*Item{}}
	for _, p := range props {
		p.JSONName = p.Name
		p.Parent = root
		root.Properties[p.Name] = p
	}
	return root
}

func refTestDef(typeName string, refs map[string]string) *Definition {
	def := &Definition{fileName: typeName + ".yml", typeName: typeName, Resource: &SchemaMeta{}}
	if len(refs) > 0 {
		def.Schema = make(map[string]*Item)
		for k, v := range refs {
			def.Schema[k] = &Item{Ref: v}
		}
	}
	return def
}

func TestResolveRefs(t *testing.T) {
	t.Parallel()

	names := func(defs []*Definition) []string {
		out := make([]string, 0, len(defs))
		for _, d := range defs {
			out = append(out, d.typeName)
		}
		return out
	}

	t.Run("targets first", func(t *testing.T) {
		t.Parallel()
		a := refTestDef("aiven_a", map[string]string{"b_id": "aiven_b.id", "sdk": "aiven_sdk.name"})
		b := refTestDef("aiven_b", map[string]string{"c_id": "aiven_c.id"})
		c := refTestDef("aiven_c", nil)
		got, err := resolveRefs([]*Definition{a, b, c})
		require.NoError(t, err)
		assert.Equal(t, []string{"aiven_c", "aiven_b", "aiven_a"}, names(got))
		assert.Equal(t, map[string]*Definition{"aiven_b": b}, a.refTargets)
	})

	for name, defs := range map[string][]*Definition{
		"cycle": {
			refTestDef("aiven_a", map[string]string{"b": "aiven_b.id"}),
			refTestDef("aiven_b", map[string]string{"a": "aiven_a.id"}),
		},
		"self":    {refTestDef("aiven_a", map[string]string{"a": "aiven_a.id"})},
		"invalid": {refTestDef("aiven_a", map[string]string{"b": "aiven_b"})},
		"no resource": {
			refTestDef("aiven_a", map[string]string{"b": "aiven_b.id"}),
			{fileName: "aiven_b.yml", typeName: "aiven_b", Datasource: &SchemaMeta{}},
		},
		"deprecated": {
			refTestDef("aiven_a", map[string]string{"b": "aiven_b.id"}),
			{fileName: "aiven_b.yml", typeName: "aiven_b", Resource: &SchemaMeta{DeprecationMessage: "Use aiven_c"}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := resolveRefs(defs)
			require.Error(t, err)
		})
	}
}

func TestRequiredBlock(t *testing.T) {
	t.Parallel()

	src := []byte(`resource "aiven_kafka" "example_kafka" {
  project      = data.aiven_project.example_project.project
  plan         = "business-4" // Comment
  service_name = "example-kafka"
  static_ips = toset([
    aiven_static_ip.ips[0].static_ip_address_id,
  ])

  /* COMPUTED FIELDS
  id = "foo"
  */

  kafka_user_config {
    kafka_rest = true
  }

  required_block {
    foo = true
  }
}
`)
	f, diags := hclsyntax.ParseConfig(src, "", hcl.InitialPos)
	require.False(t, diags.HasErrors())

	required := func(name string) bool {
		return slices.Contains([]string{"project", "plan", "service_name", "required_block"}, name)
	}
	assert.Equal(t, `resource "aiven_kafka" "example_kafka" {
  project      = data.aiven_project.example_project.project
  plan         = "business-4"
  service_name = "example-kafka"
  required_block {
    foo = true
  }
}
`, string(hclwrite.Format(requiredBlock(src, f.Body.(*hclsyntax.Body).Blocks[0], required))))
}

func TestRefTargetSDK(t *testing.T) {
	t.Parallel()

	def := refTestDef("aiven_a", nil)
	_, err := refTarget(def, "aiven_kafka", "no_such_attribute")
	require.ErrorContains(t, err, `referenced resource "aiven_kafka" has no attribute "no_such_attribute"`)

	_, err = refTarget(def, "aiven_no_such_resource", "id")
	require.ErrorContains(t, err, `referenced resource "aiven_no_such_resource" not found`)
}

func TestExampleDoc(t *testing.T) {
	t.Parallel()

	org := refTestDef("aiven_org", nil)
	org.resourceRoot = refTestRoot(
		&Item{Name: "name", Type: SchemaTypeString, Required: true, Example: "Acme"},
		&Item{Name: "description", Type: SchemaTypeString, Optional: true, Example: "skipped"},
		&Item{Name: "id", Type: SchemaTypeString, Computed: true},
	)
	var err error
	org.refExample, err = newRefExample(org, org.resourceRoot)
	require.NoError(t, err)

	group := refTestDef("aiven_group", map[string]string{"organization_id": "aiven_org.id"})
	group.refTargets = map[string]*Definition{"aiven_org": org}
	group.resourceRoot = refTestRoot(
		&Item{Name: "organization_id", Type: SchemaTypeString, Required: true, Ref: "aiven_org.id", Example: "org123"},
		&Item{Name: "name", Type: SchemaTypeString, Required: true, Example: "admins"},
	)
	group.refExample, err = newRefExample(group, group.resourceRoot)
	require.NoError(t, err)

	member := refTestDef("aiven_member", nil)
	member.refTargets = map[string]*Definition{"aiven_org": org, "aiven_group": group}
	root := refTestRoot(
		&Item{Name: "organization_id", Type: SchemaTypeString, Required: true, Ref: "aiven_org.id", Example: "org123"},
		&Item{Name: "group_id", Type: SchemaTypeString, Required: true, Ref: "aiven_group.id", Example: "ug123"},
	)
	group.resourceRoot.Properties["id"] = &Item{Name: "id", Type: SchemaTypeString, Computed: true}

	got, err := exampleDoc(member, root)
	require.NoError(t, err)
	assert.Equal(t, `resource "aiven_member" "example" {
  group_id        = aiven_group.example.id // e.g. "ug123"
  organization_id = aiven_org.example.id   // e.g. "org123"
}

resource "aiven_group" "example" {
  organization_id = aiven_org.example.id // e.g. "org123"
  name            = "admins"
}

resource "aiven_org" "example" {
  name = "Acme"
}
`, string(got))

	plain, err := exampleRoot(member, resourceType, root)
	require.NoError(t, err)
	assert.NotContains(t, string(plain), "aiven_org.example")

	member.refTargets["aiven_group"] = refTestDef("aiven_group", nil)
	member.refTargets["aiven_group"].resourceRoot = refTestRoot()
	_, err = exampleDoc(member, root)
	require.ErrorContains(t, err, `has no attribute "id"`)
}
