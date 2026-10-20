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

// reRef matches a `ref` value: "<resource type>.<attribute>".
var reRef = regexp.MustCompile(`^(aiven_[a-z0-9_]+)\.([a-z0-9_]+)$`)

func parseRef(ref string) (typeName, attr string, err error) {
	m := reRef.FindStringSubmatch(ref)
	if m == nil {
		return "", "", fmt.Errorf(`invalid ref %q, expected "<resource type>.<attribute>"`, ref)
	}
	return m[1], m[2], nil
}

// refExample is a resource example block that other examples can reference.
type refExample struct {
	typeName string
	label    string
	hcl      []byte
	deps     []*refExample // Transitive, dependencies first.
}

// newRefExample renders the compact block (required fields only) that dependents prepend to their doc example.
func newRefExample(def *Definition, root *Item) (*refExample, error) {
	w := &exampleWriter{def: def, entity: resourceType, withRefs: true, requiredOnly: true}
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
	w := &exampleWriter{def: def, entity: resourceType, withRefs: true}
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

	typeName, attr, err := parseRef(item.Ref)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", item.Path(), err)
	}

	target, err := refTarget(w.def, typeName, attr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", item.Path(), err)
	}

	if !slices.Contains(w.deps, target) {
		w.deps = append(w.deps, target)
	}

	return hclwrite.TokensForTraversal(hcl.Traversal{
		hcl.TraverseRoot{Name: typeName},
		hcl.TraverseAttr{Name: target.label},
		hcl.TraverseAttr{Name: attr},
	}), nil
}

// refTarget returns the example of the referenced resource: the cached one for generated
// definitions, otherwise the block from its hand-written example file (SDK resources).
func refTarget(def *Definition, typeName, attr string) (*refExample, error) {
	if t := def.refTargets[typeName]; t != nil {
		if t.resourceRoot == nil {
			return nil, fmt.Errorf("referenced resource %q is not generated yet", typeName)
		}
		if _, ok := t.resourceRoot.Properties[attr]; !ok {
			return nil, fmt.Errorf("referenced resource %q has no attribute %q", typeName, attr)
		}
		if t.refExample != nil {
			return t.refExample, nil
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
		if !seen[e.typeName] {
			seen[e.typeName] = true
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
		for _, ref := range collectRefs(d.Schema) {
			typeName, _, err := parseRef(ref)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", d.fileName, err)
			}
			if typeName == d.typeName {
				return nil, fmt.Errorf("%s: ref %q points to itself", d.fileName, ref)
			}

			t, ok := byType[typeName]
			if !ok {
				continue // SDK resource, resolved from its example file.
			}
			if t.Resource == nil {
				return nil, fmt.Errorf("%s: ref %q points to a definition without a resource", d.fileName, ref)
			}
			if t.Resource.DeprecationMessage != "" {
				return nil, fmt.Errorf("%s: ref %q points to a deprecated resource: %s", d.fileName, ref, t.Resource.DeprecationMessage)
			}
			if d.refTargets == nil {
				d.refTargets = make(map[string]*Definition)
			}
			d.refTargets[typeName] = t
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
