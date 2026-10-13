package main

// Generic yaml write-back toolkit. Callers compose findOrCreateSeq /
// findOrCreateMap / appendUniqueScalars (and friends) inside a mutate
// callback passed to writeBackYAMLFile or editYAML. Comments and
// formatting survive via a yaml.Node round-trip; no-op mutations skip
// the write.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// writeBackYAMLFile reads path, runs mutate on the document's root
// mapping, and writes back only when the result differs from the input.
func writeBackYAMLFile(path string, mutate func(root *yaml.Node) error) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	out, changed, err := editYAML(data, mutate)
	if err != nil {
		return fmt.Errorf("edit %s: %w", path, err)
	}
	if !changed {
		return nil
	}
	return replaceFileAtomic(path, out)
}

// replaceFileAtomic writes data via a sibling temp file + rename so an
// interrupted write cannot leave a half-populated file. Original
// permissions are preserved. The temp file is removed on error paths;
// after a successful rename the deferred Remove is a no-op.
func replaceFileAtomic(path string, data []byte) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp for %s: %w", path, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(info.Mode().Perm()); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp for %s: %w", path, err)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp for %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp for %s: %w", path, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename temp for %s: %w", path, err)
	}
	return nil
}

// editYAML is the pure bytes-in/bytes-out counterpart to
// writeBackYAMLFile. The second return value is true iff bytes changed.
func editYAML(data []byte, mutate func(root *yaml.Node) error) ([]byte, bool, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, false, fmt.Errorf("parse yaml: %w", err)
	}
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil, false, fmt.Errorf("unexpected document shape")
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, false, fmt.Errorf("root is not a mapping")
	}
	if err := mutate(root); err != nil {
		return nil, false, err
	}
	out, err := marshalYAMLDoc(&doc)
	if err != nil {
		return nil, false, fmt.Errorf("marshal yaml: %w", err)
	}
	if bytes.Equal(out, data) {
		return data, false, nil
	}
	return out, true, nil
}

// findOrCreateSeq returns (or creates at the end of mapping) the
// sequence node under key. A null scalar at key is upgraded in place so
// comments attached to the key survive.
func findOrCreateSeq(mapping *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		k := mapping.Content[i]
		if k.Value != key {
			continue
		}
		v := mapping.Content[i+1]
		if v.Kind == yaml.SequenceNode {
			return v
		}
		v.Kind = yaml.SequenceNode
		v.Tag = "!!seq"
		v.Value = ""
		v.Content = nil
		return v
	}
	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	mapping.Content = append(mapping.Content, keyNode, seq)
	return seq
}

// findOrCreateMap is findOrCreateSeq for mapping values.
func findOrCreateMap(mapping *yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		k := mapping.Content[i]
		if k.Value != key {
			continue
		}
		v := mapping.Content[i+1]
		if v.Kind == yaml.MappingNode {
			return v
		}
		v.Kind = yaml.MappingNode
		v.Tag = "!!map"
		v.Value = ""
		v.Content = nil
		return v
	}
	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
	m := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	mapping.Content = append(mapping.Content, keyNode, m)
	return m
}

// appendUniqueScalars appends new entries (in alphabetical key order,
// deduped against seq) as scalars, attaching each non-empty value as a
// trailing comment. Returns the number of entries appended.
func appendUniqueScalars(seq *yaml.Node, entries map[string]string) int {
	existing := make(map[string]bool, len(seq.Content))
	for _, n := range seq.Content {
		existing[n.Value] = true
	}
	names := make([]string, 0, len(entries))
	for name := range entries {
		if !existing[name] {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		node := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: name}
		if c := describeForComment(entries[name]); c != "" {
			node.LineComment = c
		}
		seq.Content = append(seq.Content, node)
	}
	return len(names)
}

// marshalYAMLDoc re-emits a document at 2-space indent to match the
// hand-written definition files (yaml.Marshal defaults to 4).
func marshalYAMLDoc(doc *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// describeForComment collapses desc to its first non-empty line and
// truncates at maxCommentLen. Returns "" when desc is blank.
func describeForComment(desc string) string {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return ""
	}
	for _, line := range strings.Split(desc, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		desc = line
		break
	}
	// Truncate by runes: yaml.v3 emits LineComment bytes verbatim, so
	// byte-slicing could produce invalid UTF-8.
	const maxCommentLen = 120
	if r := []rune(desc); len(r) > maxCommentLen {
		desc = strings.TrimSpace(string(r[:maxCommentLen])) + "..."
	}
	return desc
}
