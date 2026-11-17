package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/provider"
)

// reRef matches a `ref` value: "<resource type>.<attribute>" or "data.<data source type>.<attribute>".
var reRef = regexp.MustCompile(`^(data\.)?(aiven_[a-z0-9_]+)\.([a-z0-9_]+)$`)

type ref struct {
	data     bool
	typeName string
	attr     string
}

func parseRef(s string) (ref, error) {
	m := reRef.FindStringSubmatch(s)
	if m == nil {
		return ref{}, fmt.Errorf(`invalid ref %q, expected "<resource type>.<attribute>" or "data.<data source type>.<attribute>"`, s)
	}
	return ref{data: m[1] != "", typeName: m[2], attr: m[3]}, nil
}

// refExample is a resource or data source example block that other examples can reference.
type refExample struct {
	data     bool
	typeName string
	label    string
	hcl      []byte
	deps     []*refExample // Transitive, dependencies first.
}

// address is the block address without the label, e.g. "data.aiven_organization".
func (e *refExample) address() string {
	if e.data {
		return "data." + e.typeName
	}
	return e.typeName
}

// newRefExample renders the compact block (required fields only) that dependents prepend to their doc example.
func newRefExample(def *Definition, root *Item, refDepth int) (*refExample, error) {
	w := &exampleWriter{def: def, entity: resourceType, withRefs: true, requiredOnly: true, refDepth: refDepth}
	b, err := w.root(root)
	if err != nil {
		return nil, err
	}
	return &refExample{
		typeName: def.typeName,
		label:    exampleLabel,
		hcl:      b,
		deps:     flattenRefExamples(w.deps),
	}, nil
}

// exampleDoc renders the resource doc example: the resource itself, then the referenced resources, dependents first.
func exampleDoc(def *Definition, root *Item) ([]byte, error) {
	w := &exampleWriter{def: def, entity: resourceType, withRefs: true, refDepth: def.Resource.RefDepth}
	b, err := w.root(root)
	if err != nil {
		return nil, err
	}

	// Flattening the reversed direct deps and rendering backwards keeps the direct deps in field order.
	direct := slices.Clone(w.deps)
	slices.Reverse(direct)
	deps := flattenRefExamples(direct)
	if len(deps) == 0 {
		return b, nil
	}

	chunks := make([][]byte, 0, len(deps)+1)
	chunks = append(chunks, bytes.TrimSpace(b))
	for _, d := range slices.Backward(deps) {
		chunks = append(chunks, bytes.TrimSpace(d.hcl))
	}
	out := append(bytes.Join(chunks, []byte("\n\n")), '\n')
	if _, diags := hclwrite.ParseConfig(out, def.typeName, hcl.InitialPos); diags.HasErrors() {
		return nil, fmt.Errorf("invalid example with refs: %w", diags)
	}
	return hclwrite.Format(out), nil
}

// refTokens renders a reference to the referenced attribute and records the target in w.deps.
func (w *exampleWriter) refTokens(item *Item) (hclwrite.Tokens, error) {
	if !item.IsScalar() {
		return nil, fmt.Errorf("ref is supported for scalar fields only, got %s %q", item.Type, item.Path())
	}

	r, err := parseRef(item.Ref)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", item.Path(), err)
	}

	var target *refExample
	if r.data {
		target, err = dataRefTarget(w.def, r.typeName, r.attr)
	} else {
		target, err = refTarget(w.def, r.typeName, r.attr, max(w.refDepth-1, 0))
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", item.Path(), err)
	}

	if !slices.ContainsFunc(w.deps, func(d *refExample) bool { return d.address() == target.address() }) {
		w.deps = append(w.deps, target)
	}

	traversal := hcl.Traversal{hcl.TraverseRoot{Name: r.typeName}}
	if r.data {
		traversal = hcl.Traversal{hcl.TraverseRoot{Name: "data"}, hcl.TraverseAttr{Name: r.typeName}}
	}
	traversal = append(traversal, hcl.TraverseAttr{Name: target.label}, hcl.TraverseAttr{Name: r.attr})
	return hclwrite.TokensForTraversal(traversal), nil
}

// refSource follows the refs of item to the last generated attribute with an example or a default,
// so a dependent shows the value of the referenced resource's example. Returns item otherwise.
func refSource(def *Definition, item *Item) (*Definition, *Item) {
	if item.Ref == "" {
		return def, item
	}
	r, err := parseRef(item.Ref)
	if err != nil {
		return def, item
	}
	t := def.refTargets[r.typeName]
	if t == nil {
		return def, item
	}
	root := t.resourceRoot
	if r.data {
		root = t.datasourceRoot
	}
	if root == nil {
		return def, item
	}
	v := root.Properties[r.attr]
	if v == nil || v.Type != item.Type {
		return def, item
	}
	t, v = refSource(t, v)
	if isEmpty(v.Example) && isEmpty(v.Default) {
		return def, item
	}
	return t, v
}

// refTarget returns the example of the referenced resource: the cached one for generated
// definitions, re-rendered for a refDepth below the target's own, otherwise the block from
// its hand-written example file (SDK resources).
func refTarget(def *Definition, typeName, attr string, refDepth int) (*refExample, error) {
	if t := def.refTargets[typeName]; t != nil {
		if t.resourceRoot == nil {
			return nil, fmt.Errorf("referenced resource %q is not generated yet", typeName)
		}
		if _, ok := t.resourceRoot.Properties[attr]; !ok {
			return nil, fmt.Errorf("referenced resource %q has no attribute %q", typeName, attr)
		}
		if t.refExample != nil {
			if own := t.Resource.RefDepth; refDepth == 0 || (own > 0 && own <= refDepth) {
				return t.refExample, nil
			}
			return newRefExample(t, t.resourceRoot, refDepth)
		}
	} else {
		r, err := sdkResource(typeName)
		if err != nil {
			return nil, err
		}
		// SDK resources have an implicit "id".
		if _, ok := r.SchemaMap()[attr]; !ok && attr != "id" {
			return nil, fmt.Errorf("referenced resource %q has no attribute %q", typeName, attr)
		}
	}
	return loadRefExample(def, typeName)
}

// dataRefTarget returns the data source block that looks up the referenced object.
// It sets the required fields and the lookup keys of the data source:
// the users know the name of an existing object rather than its ID.
func dataRefTarget(def *Definition, typeName, attr string) (*refExample, error) {
	t := def.refTargets[typeName]
	if t == nil || t.datasourceRoot == nil {
		return nil, fmt.Errorf("referenced data source %q is not generated yet", typeName)
	}
	root := t.datasourceRoot
	if _, ok := root.Properties[attr]; !ok {
		return nil, fmt.Errorf("referenced data source %q has no attribute %q", typeName, attr)
	}

	var keys []string
	for _, k := range sortedKeysPriority(t, datasourceType, root) {
		if root.Properties[k].IsRequired(t, datasourceType) {
			keys = append(keys, k)
		}
	}
	for _, k := range t.DatasourceLookupComposedOf() {
		if !slices.Contains(keys, k) {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("referenced data source %q has no lookup or required fields", typeName)
	}

	f := hclwrite.NewEmptyFile()
	body := f.Body().AppendNewBlock("data", []string{typeName, exampleLabel}).Body()
	for _, k := range keys {
		v, ok := root.Properties[k]
		if !ok {
			return nil, fmt.Errorf("referenced data source %q has no lookup field %q", typeName, k)
		}
		val, err := exampleItemValue(t, datasourceType, v)
		if err != nil {
			return nil, err
		}
		body.SetAttributeValue(k, val)
	}
	return &refExample{data: true, typeName: typeName, label: exampleLabel, hcl: f.Bytes()}, nil
}

var fileRefExamples = make(map[string]*refExample)

// loadRefExample returns the first `resource "<typeName>"` block from its example file, with required fields only.
func loadRefExample(def *Definition, typeName string) (*refExample, error) {
	if v, ok := fileRefExamples[typeName]; ok {
		return v, nil
	}

	required, err := refRequired(def, typeName)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(examplesDir, "resources", typeName, "resource.tf")
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("referenced resource %q has no example: %w", typeName, err)
	}

	f, diags := hclsyntax.ParseConfig(src, path, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("could not parse %s: %w", path, diags)
	}

	for _, b := range f.Body.(*hclsyntax.Body).Blocks {
		if b.Type == "resource" && len(b.Labels) == 2 && b.Labels[0] == typeName {
			v := &refExample{typeName: typeName, label: b.Labels[1], hcl: requiredBlock(src, b, required)}
			fileRefExamples[typeName] = v
			return v, nil
		}
	}
	return nil, fmt.Errorf("no resource %q block in %s", typeName, path)
}

// requiredBlock renders b with its required top-level attributes and blocks only, in source order.
func requiredBlock(src []byte, b *hclsyntax.Block, required func(name string) bool) []byte {
	type field struct {
		name   string
		isAttr bool
		rng    hcl.Range
		raw    []byte
	}

	var fields []field
	for name, a := range b.Body.Attributes {
		if required(name) {
			v := field{name: name, isAttr: true, rng: a.Expr.Range()}
			// Hand-written service examples reference an undeclared `data.aiven_project`.
			// todo: remove when services are migrated to the plugin and their examples are generated.
			if name == "project" && len(a.Expr.Variables()) > 0 {
				v.raw = []byte(`"my-project"`)
			}
			fields = append(fields, v)
		}
	}
	for _, nested := range b.Body.Blocks {
		if required(nested.Type) {
			fields = append(fields, field{name: nested.Type, rng: nested.Range()})
		}
	}
	slices.SortFunc(fields, func(a, b field) int {
		return a.rng.Start.Byte - b.rng.Start.Byte
	})

	f := hclwrite.NewEmptyFile()
	body := f.Body().AppendNewBlock(b.Type, b.Labels).Body()
	for _, v := range fields {
		if v.raw == nil {
			v.raw = v.rng.SliceBytes(src)
		}
		raw := hclwrite.Tokens{{Type: hclsyntax.TokenIdent, Bytes: v.raw}}
		if v.isAttr {
			body.SetAttributeRaw(v.name, raw)
		} else {
			body.AppendUnstructuredTokens(raw)
			body.AppendNewline()
		}
	}
	return f.Bytes()
}

// refRequired reports whether a top-level field of the referenced resource is required.
func refRequired(def *Definition, typeName string) (func(name string) bool, error) {
	if t := def.refTargets[typeName]; t != nil && t.resourceRoot != nil {
		return func(name string) bool {
			p := t.resourceRoot.Properties[name]
			return p != nil && p.IsRequired(t, resourceType)
		}, nil
	}

	// todo: remove when services are migrated to the plugin and their examples are generated.
	r, err := sdkResource(typeName)
	if err != nil {
		return nil, err
	}
	schemaMap := r.SchemaMap()
	return func(name string) bool {
		s := schemaMap[name]
		return s != nil && s.Required
	}, nil
}

var sdkResources = sync.OnceValues(func() (map[string]*schema.Resource, error) {
	p, err := provider.Provider("dev")
	if err != nil {
		return nil, fmt.Errorf("could not load the SDK provider: %w", err)
	}
	return p.ResourcesMap, nil
})

func sdkResource(typeName string) (*schema.Resource, error) {
	resources, err := sdkResources()
	if err != nil {
		return nil, err
	}
	r, ok := resources[typeName]
	if !ok {
		return nil, fmt.Errorf("referenced resource %q not found", typeName)
	}
	return r, nil
}

// flattenRefExamples returns the direct examples with their dependencies, dependencies first, without duplicates.
func flattenRefExamples(direct []*refExample) []*refExample {
	var out []*refExample
	seen := make(map[string]bool)
	add := func(e *refExample) {
		if !seen[e.address()] {
			seen[e.address()] = true
			out = append(out, e)
		}
	}
	for _, d := range direct {
		for _, dd := range d.deps {
			add(dd)
		}
		add(d)
	}
	return out
}

// resolveRefs validates `ref` values, sets Definition.refTargets,
// and returns the definitions ordered so that ref targets come before their dependents.
func resolveRefs(defs []*Definition) ([]*Definition, error) {
	byType := make(map[string]*Definition, len(defs))
	for _, d := range defs {
		byType[d.typeName] = d
	}

	for _, d := range defs {
		for _, s := range collectRefs(d.Schema) {
			r, err := parseRef(s)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", d.fileName, err)
			}
			if r.typeName == d.typeName {
				return nil, fmt.Errorf("%s: ref %q points to itself", d.fileName, s)
			}

			t, ok := byType[r.typeName]
			switch {
			case !ok && r.data:
				return nil, fmt.Errorf("%s: ref %q points to a data source that isn't generated", d.fileName, s)
			case !ok:
				continue // SDK resource, resolved from its example file.
			case r.data && t.Datasource == nil:
				return nil, fmt.Errorf("%s: ref %q points to a definition without a data source", d.fileName, s)
			case r.data && t.Datasource.DeprecationMessage != "":
				return nil, fmt.Errorf("%s: ref %q points to a deprecated data source: %s", d.fileName, s, t.Datasource.DeprecationMessage)
			case !r.data && t.Resource == nil:
				return nil, fmt.Errorf("%s: ref %q points to a definition without a resource", d.fileName, s)
			case !r.data && t.Resource.DeprecationMessage != "":
				return nil, fmt.Errorf("%s: ref %q points to a deprecated resource: %s", d.fileName, s, t.Resource.DeprecationMessage)
			}
			if d.refTargets == nil {
				d.refTargets = make(map[string]*Definition)
			}
			d.refTargets[r.typeName] = t
		}
	}

	const visiting, done = 1, 2
	state := make(map[*Definition]int, len(defs))
	ordered := make([]*Definition, 0, len(defs))
	var visit func(d *Definition) error
	visit = func(d *Definition) error {
		switch state[d] {
		case visiting:
			return fmt.Errorf("ref cycle at %q", d.typeName)
		case done:
			return nil
		}
		state[d] = visiting
		for _, k := range sortedKeys(d.refTargets) {
			if err := visit(d.refTargets[k]); err != nil {
				return err
			}
		}
		state[d] = done
		ordered = append(ordered, d)
		return nil
	}

	for _, d := range defs {
		if err := visit(d); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

// collectRefs returns the `ref` values of the schema items at any depth.
func collectRefs(schema map[string]*Item) []string {
	var refs []string
	var walk func(item *Item)
	walk = func(item *Item) {
		if item == nil {
			return
		}
		if item.Ref != "" {
			refs = append(refs, item.Ref)
		}
		for _, v := range item.Properties {
			walk(v)
		}
		walk(item.Items)
		walk(item.AdditionalProperties)
	}
	for _, v := range schema {
		walk(v)
	}
	return refs
}
