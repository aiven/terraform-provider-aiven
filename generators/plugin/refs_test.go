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

func TestParseRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in      string
		want    ref
		wantErr bool
	}{
		{in: "aiven_kafka.service_name", want: ref{typeName: "aiven_kafka", attr: "service_name"}},
		{in: "data.aiven_organization.id", want: ref{data: true, typeName: "aiven_organization", attr: "id"}},
		{in: "aiven_kafka", wantErr: true},
		{in: "data.aiven_kafka", wantErr: true},
		{in: "resource.aiven_kafka.id", wantErr: true},
		{in: "data.kafka.id", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got, err := parseRef(tt.in)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveDataRefs(t *testing.T) {
	t.Parallel()

	org := func(datasource *SchemaMeta) *Definition {
		return &Definition{fileName: "aiven_org.yml", typeName: "aiven_org", Datasource: datasource}
	}

	a := refTestDef("aiven_a", map[string]string{"org_id": "data.aiven_org.id"})
	target := org(&SchemaMeta{})
	got, err := resolveRefs([]*Definition{a, target})
	require.NoError(t, err)
	assert.Equal(t, []*Definition{target, a}, got)
	assert.Equal(t, map[string]*Definition{"aiven_org": target}, a.refTargets)

	for name, tc := range map[string]struct {
		target *Definition
		errMsg string
	}{
		"not generated": {nil, "points to a data source that isn't generated"},
		"no data source": {
			&Definition{fileName: "aiven_org.yml", typeName: "aiven_org", Resource: &SchemaMeta{}},
			"points to a definition without a data source",
		},
		"deprecated": {org(&SchemaMeta{DeprecationMessage: "Use aiven_b"}), "points to a deprecated data source"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			defs := []*Definition{refTestDef("aiven_a", map[string]string{"org_id": "data.aiven_org.id"})}
			if tc.target != nil {
				defs = append(defs, tc.target)
			}
			_, err := resolveRefs(defs)
			require.ErrorContains(t, err, tc.errMsg)
		})
	}
}

func TestExampleDocDataRef(t *testing.T) {
	t.Parallel()

	org := &Definition{
		fileName:   "aiven_org.yml",
		typeName:   "aiven_org",
		Datasource: &SchemaMeta{},
		Operations: []*Operation{{
			Type:                 OperationRead,
			DatasourceLookup:     true,
			ResultListLookupKeys: map[string]string{"AccountName": "name"},
		}},
	}
	org.datasourceRoot = refTestRoot(
		&Item{Name: "id", Type: SchemaTypeString, Optional: true, Computed: true},
		&Item{Name: "name", Type: SchemaTypeString, Optional: true, Computed: true, Example: "Acme"},
	)

	group := refTestDef("aiven_group", map[string]string{"organization_id": "data.aiven_org.id"})
	group.refTargets = map[string]*Definition{"aiven_org": org}
	group.resourceRoot = refTestRoot(
		&Item{Name: "organization_id", Type: SchemaTypeString, Required: true, Ref: "data.aiven_org.id", Example: "org123"},
		&Item{Name: "name", Type: SchemaTypeString, Required: true, Example: "admins"},
		&Item{Name: "id", Type: SchemaTypeString, Computed: true},
	)
	var err error
	group.refExample, err = newRefExample(group, group.resourceRoot, 0)
	require.NoError(t, err)

	member := refTestDef("aiven_member", nil)
	member.refTargets = map[string]*Definition{"aiven_org": org, "aiven_group": group}
	root := refTestRoot(
		&Item{Name: "organization_id", Type: SchemaTypeString, Required: true, Ref: "data.aiven_org.id", Example: "org123"},
		&Item{Name: "group_id", Type: SchemaTypeString, Required: true, Ref: "aiven_group.id", Example: "ug123"},
	)

	got, err := exampleDoc(member, root)
	require.NoError(t, err)
	assert.Equal(t, `resource "aiven_member" "example" {
  group_id        = aiven_group.example.id    // e.g. "ug123"
  organization_id = data.aiven_org.example.id // e.g. "org123"
}

resource "aiven_group" "example" {
  organization_id = data.aiven_org.example.id // e.g. "org123"
  name            = "admins"
}

data "aiven_org" "example" {
  name = "Acme"
}
`, string(got))

	t.Run("required fields without a lookup", func(t *testing.T) {
		t.Parallel()
		ds := &Definition{fileName: "aiven_ds.yml", typeName: "aiven_ds", Datasource: &SchemaMeta{}}
		ds.datasourceRoot = refTestRoot(
			&Item{Name: "project", Type: SchemaTypeString, IDAttribute: true, Example: "my-project"},
			&Item{Name: "id", Type: SchemaTypeString, Computed: true},
		)
		def := refTestDef("aiven_x", nil)
		def.refTargets = map[string]*Definition{"aiven_ds": ds}

		e, err := dataRefTarget(def, "aiven_ds", "id")
		require.NoError(t, err)
		assert.Equal(t, "data \"aiven_ds\" \"example\" {\n  project = \"my-project\"\n}\n", string(e.hcl))

		_, err = dataRefTarget(def, "aiven_ds", "no_such_attribute")
		require.ErrorContains(t, err, `referenced data source "aiven_ds" has no attribute "no_such_attribute"`)
	})

	t.Run("required fields and lookup keys", func(t *testing.T) {
		t.Parallel()
		ds := &Definition{
			fileName:       "aiven_app.yml",
			typeName:       "aiven_app",
			Datasource:     &SchemaMeta{},
			IdentityFields: []string{"project", "service_name", "application_id"},
			Operations: []*Operation{{
				Type:                 OperationRead,
				DatasourceLookup:     true,
				ResultListLookupKeys: map[string]string{"Name": "name"},
			}},
		}
		ds.datasourceRoot = refTestRoot(
			&Item{Name: "project", Type: SchemaTypeString, IDAttribute: true, IDAttributePosition: 0, Example: "my-project"},
			&Item{Name: "service_name", Type: SchemaTypeString, IDAttribute: true, IDAttributePosition: 1, Example: "my-flink"},
			&Item{Name: "application_id", Type: SchemaTypeString, IDAttribute: true, IDAttributePosition: 2},
			&Item{Name: "name", Type: SchemaTypeString, Optional: true, Computed: true, Example: "my-app"},
		)
		def := refTestDef("aiven_x", nil)
		def.refTargets = map[string]*Definition{"aiven_app": ds}

		e, err := dataRefTarget(def, "aiven_app", "application_id")
		require.NoError(t, err)
		assert.Equal(t, `data "aiven_app" "example" {
  project      = "my-project"
  service_name = "my-flink"
  name         = "my-app"
}
`, string(hclwrite.Format(e.hcl)))
	})
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
  project      = "my-project"
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
	_, err := refTarget(def, "aiven_project", "no_such_attribute", 0)
	require.ErrorContains(t, err, `referenced resource "aiven_project" has no attribute "no_such_attribute"`)

	_, err = refTarget(def, "aiven_no_such_resource", "id", 0)
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
	org.refExample, err = newRefExample(org, org.resourceRoot, 0)
	require.NoError(t, err)

	group := refTestDef("aiven_group", map[string]string{"organization_id": "aiven_org.id"})
	group.refTargets = map[string]*Definition{"aiven_org": org}
	group.resourceRoot = refTestRoot(
		&Item{Name: "organization_id", Type: SchemaTypeString, Required: true, Ref: "aiven_org.id", Example: "org123"},
		&Item{Name: "name", Type: SchemaTypeString, Required: true, Example: "admins"},
	)
	group.refExample, err = newRefExample(group, group.resourceRoot, 0)
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

	member.Resource.RefDepth = 2
	got, err = exampleDoc(member, refTestRoot(
		&Item{Name: "group_id", Type: SchemaTypeString, Required: true, Ref: "aiven_group.id", Example: "ug123"},
	))
	require.NoError(t, err)
	assert.Equal(t, `resource "aiven_member" "example" {
  group_id = aiven_group.example.id // e.g. "ug123"
}

resource "aiven_group" "example" {
  organization_id = "org123"
  name            = "admins"
}
`, string(got))
	member.Resource.RefDepth = 0

	cached := group.refExample
	group.Resource.RefDepth = 1
	group.refExample, err = newRefExample(group, group.resourceRoot, 1)
	require.NoError(t, err)
	got, err = exampleDoc(member, refTestRoot(
		&Item{Name: "group_id", Type: SchemaTypeString, Required: true, Ref: "aiven_group.id", Example: "ug123"},
	))
	require.NoError(t, err)
	assert.Contains(t, string(got), `organization_id = "org123"`)
	assert.NotContains(t, string(got), `resource "aiven_org"`)
	group.Resource.RefDepth, group.refExample = 0, cached

	member.refTargets["aiven_group"] = refTestDef("aiven_group", nil)
	member.refTargets["aiven_group"].resourceRoot = refTestRoot()
	_, err = exampleDoc(member, root)
	require.ErrorContains(t, err, `has no attribute "id"`)
}

func TestRefSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		run  func(t *testing.T) (*Definition, *Item, *Definition, *Item)
	}{
		{
			name: "no ref",
			run: func(t *testing.T) (*Definition, *Item, *Definition, *Item) {
				def := refTestDef("aiven_a", nil)
				item := &Item{Name: "foo", Type: SchemaTypeString, Example: "x"}
				return def, item, def, item
			},
		},
		{
			name: "target example",
			run: func(t *testing.T) (*Definition, *Item, *Definition, *Item) {
				id := &Item{Name: "id", Type: SchemaTypeString, Example: "org123"}
				target := refTestDef("aiven_org", nil)
				target.resourceRoot = refTestRoot(
					&Item{Name: "name", Type: SchemaTypeString, Required: true},
					id,
				)
				def := refTestDef("aiven_member", nil)
				def.refTargets = map[string]*Definition{"aiven_org": target}
				item := &Item{Name: "organization_id", Type: SchemaTypeString, Ref: "aiven_org.id"}
				return def, item, target, id
			},
		},
		{
			name: "chain resolves to example on leaf",
			run: func(t *testing.T) (*Definition, *Item, *Definition, *Item) {
				cID := &Item{Name: "id", Type: SchemaTypeString, Example: "c-id"}
				c := refTestDef("aiven_c", nil)
				c.resourceRoot = refTestRoot(cID)

				bID := &Item{Name: "id", Type: SchemaTypeString, Ref: "aiven_c.id"}
				b := refTestDef("aiven_b", nil)
				b.resourceRoot = refTestRoot(bID)
				b.refTargets = map[string]*Definition{"aiven_c": c}

				def := refTestDef("aiven_a", nil)
				def.refTargets = map[string]*Definition{"aiven_b": b}
				item := &Item{Name: "b_id", Type: SchemaTypeString, Ref: "aiven_b.id"}
				return def, item, c, cID
			},
		},
		{
			name: "chain stops at example on intermediate",
			run: func(t *testing.T) (*Definition, *Item, *Definition, *Item) {
				cID := &Item{Name: "id", Type: SchemaTypeString}
				c := refTestDef("aiven_c", nil)
				c.resourceRoot = refTestRoot(cID)

				bID := &Item{Name: "id", Type: SchemaTypeString, Ref: "aiven_c.id", Example: "b-id"}
				b := refTestDef("aiven_b", nil)
				b.resourceRoot = refTestRoot(bID)
				b.refTargets = map[string]*Definition{"aiven_c": c}

				def := refTestDef("aiven_a", nil)
				def.refTargets = map[string]*Definition{"aiven_b": b}
				item := &Item{Name: "b_id", Type: SchemaTypeString, Ref: "aiven_b.id"}
				return def, item, b, bID
			},
		},
		{
			name: "target without example or default",
			run: func(t *testing.T) (*Definition, *Item, *Definition, *Item) {
				id := &Item{Name: "id", Type: SchemaTypeString}
				target := refTestDef("aiven_org", nil)
				target.resourceRoot = refTestRoot(id)
				def := refTestDef("aiven_member", nil)
				def.refTargets = map[string]*Definition{"aiven_org": target}
				item := &Item{Name: "organization_id", Type: SchemaTypeString, Ref: "aiven_org.id"}
				return def, item, def, item
			},
		},
		{
			name: "target default only",
			run: func(t *testing.T) (*Definition, *Item, *Definition, *Item) {
				id := &Item{Name: "id", Type: SchemaTypeString, Default: "default-id"}
				target := refTestDef("aiven_org", nil)
				target.resourceRoot = refTestRoot(id)
				def := refTestDef("aiven_member", nil)
				def.refTargets = map[string]*Definition{"aiven_org": target}
				item := &Item{Name: "organization_id", Type: SchemaTypeString, Ref: "aiven_org.id"}
				return def, item, target, id
			},
		},
		{
			name: "type mismatch",
			run: func(t *testing.T) (*Definition, *Item, *Definition, *Item) {
				id := &Item{Name: "id", Type: SchemaTypeInteger, Example: 42}
				target := refTestDef("aiven_org", nil)
				target.resourceRoot = refTestRoot(id)
				def := refTestDef("aiven_member", nil)
				def.refTargets = map[string]*Definition{"aiven_org": target}
				item := &Item{Name: "organization_id", Type: SchemaTypeString, Ref: "aiven_org.id"}
				return def, item, def, item
			},
		},
		{
			name: "target not generated",
			run: func(t *testing.T) (*Definition, *Item, *Definition, *Item) {
				target := refTestDef("aiven_org", nil)
				def := refTestDef("aiven_member", nil)
				def.refTargets = map[string]*Definition{"aiven_org": target}
				item := &Item{Name: "organization_id", Type: SchemaTypeString, Ref: "aiven_org.id"}
				return def, item, def, item
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			def, item, wantDef, wantItem := tt.run(t)
			gotDef, gotItem := refSource(def, item)
			assert.Same(t, wantDef, gotDef)
			assert.Same(t, wantItem, gotItem)
		})
	}
}
