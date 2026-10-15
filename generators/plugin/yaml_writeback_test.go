package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestEditYAML_RoundTripPreservesFileLevelComments(t *testing.T) {
	t.Parallel()

	// Round-trip must preserve top-of-file, head and inline comments.
	src := `# yaml-language-server: $schema=.schema.yml
resource:
  description: Example
operations:
  - id: Foo # inline on sequence item
    type: create
remove:
  # head comment on first list item
  - alpha
  - beta
`
	out, changed, err := editYAML([]byte(src), func(root *yaml.Node) error {
		// Add a new remove entry so we trigger a write.
		appendUniqueScalars(findOrCreateSeq(root, "remove"), map[string]string{"gamma": "third greek letter"})
		return nil
	})
	require.NoError(t, err)
	require.True(t, changed)
	got := string(out)

	require.Contains(t, got, "# yaml-language-server: $schema=.schema.yml")
	require.Contains(t, got, "# head comment on first list item")
	require.Contains(t, got, "inline on sequence item")
	assertEntryWithComment(t, got, "gamma", "third greek letter")
}

func TestEditYAML_IsIdempotent_WhenMutateMakesNoChange(t *testing.T) {
	t.Parallel()

	src := `resource:
  description: Example
operations:
  - id: Foo
    type: create
`
	out, changed, err := editYAML([]byte(src), func(root *yaml.Node) error {
		return nil // noop
	})
	require.NoError(t, err)
	require.False(t, changed, "noop mutate must not report a change")
	require.Equal(t, src, string(out))
}

func TestEditYAML_PropagatesMutateError(t *testing.T) {
	t.Parallel()

	_, _, err := editYAML([]byte("resource:\n  description: x\n"), func(root *yaml.Node) error {
		return errBoom
	})
	require.ErrorIs(t, err, errBoom)
}

func TestWriteBackYAMLFile_SkipsWriteWhenUnchanged(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "demo.yml")
	src := []byte("resource:\n  description: Example\n")
	require.NoError(t, os.WriteFile(path, src, 0o644))

	require.NoError(t, writeBackYAMLFile(path, func(root *yaml.Node) error { return nil }))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, string(src), string(got))
}

func TestWriteBackYAMLFile_WritesOnChange(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "demo.yml")
	src := []byte(`resource:
  description: Example
`)
	require.NoError(t, os.WriteFile(path, src, 0o644))

	err := writeBackYAMLFile(path, func(root *yaml.Node) error {
		appendUniqueScalars(findOrCreateSeq(root, "remove"), map[string]string{"foo": "bar"})
		return nil
	})
	require.NoError(t, err)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	gotStr := string(got)
	require.Contains(t, gotStr, "remove:")
	require.Contains(t, gotStr, "- foo")
	assertEntryWithComment(t, gotStr, "foo", "bar")
}

func TestFindOrCreateSeq_ReturnsExistingWhenPresent(t *testing.T) {
	t.Parallel()

	src := `remove:
  - alpha
`
	assertCountInSeq(t, src, "remove", 1)
}

func TestFindOrCreateSeq_CreatesAtEndWhenAbsent(t *testing.T) {
	t.Parallel()

	src := `resource:
  description: Example
`
	out, changed, err := editYAML([]byte(src), func(root *yaml.Node) error {
		seq := findOrCreateSeq(root, "remove")
		require.Equal(t, yaml.SequenceNode, seq.Kind)
		appendUniqueScalars(seq, map[string]string{"new_entry": ""})
		return nil
	})
	require.NoError(t, err)
	require.True(t, changed)
	got := string(out)
	// The key was created and placed after the existing `resource:` key.
	resourceIdx := strings.Index(got, "resource:")
	removeIdx := strings.Index(got, "remove:")
	require.Greater(t, resourceIdx, -1)
	require.Greater(t, removeIdx, resourceIdx)
	require.Contains(t, got, "- new_entry")
}

func TestFindOrCreateSeq_UpgradesNullScalarToSeq(t *testing.T) {
	t.Parallel()

	// `key:` with no value parses as a null scalar; must become a seq
	// in place, keeping the key node's attached comments.
	src := `# header
resource:
  description: Example
remove:
`
	out, changed, err := editYAML([]byte(src), func(root *yaml.Node) error {
		appendUniqueScalars(findOrCreateSeq(root, "remove"), map[string]string{"only": "sole entry"})
		return nil
	})
	require.NoError(t, err)
	require.True(t, changed)
	got := string(out)
	require.Contains(t, got, "# header")
	require.Contains(t, got, "- only")
	assertEntryWithComment(t, got, "only", "sole entry")
}

func TestFindOrCreateMap_CreatesEmptyWhenAbsent(t *testing.T) {
	t.Parallel()

	src := `resource:
  description: Example
`
	out, _, err := editYAML([]byte(src), func(root *yaml.Node) error {
		m := findOrCreateMap(root, "rename")
		require.Equal(t, yaml.MappingNode, m.Kind)
		m.Content = append(m.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "old"},
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "new"},
		)
		return nil
	})
	require.NoError(t, err)
	require.Contains(t, string(out), "rename:")
	require.Contains(t, string(out), "old: new")
}

func TestAppendUniqueScalars_SkipsExisting(t *testing.T) {
	t.Parallel()

	src := `remove:
  - alpha
  - beta
`
	out, _, err := editYAML([]byte(src), func(root *yaml.Node) error {
		added := appendUniqueScalars(findOrCreateSeq(root, "remove"), map[string]string{
			"alpha": "already here",
			"gamma": "new one",
		})
		require.Equal(t, 1, added)
		return nil
	})
	require.NoError(t, err)
	got := string(out)
	// alpha stays as a plain entry (no new comment attached), gamma is new.
	for _, line := range strings.Split(got, "\n") {
		if strings.Contains(line, "- alpha") {
			require.NotContains(t, line, "already here", "existing entry must not get a comment")
		}
	}
	assertEntryWithComment(t, got, "gamma", "new one")
}

func TestAppendUniqueScalars_EmptyDescriptionYieldsPlainEntry(t *testing.T) {
	t.Parallel()

	src := `remove:
  - keep
`
	out, _, err := editYAML([]byte(src), func(root *yaml.Node) error {
		appendUniqueScalars(findOrCreateSeq(root, "remove"), map[string]string{"nodesc": ""})
		return nil
	})
	require.NoError(t, err)
	got := string(out)
	for _, line := range strings.Split(got, "\n") {
		if strings.Contains(line, "nodesc") {
			require.NotContains(t, line, "#", "no trailing comment expected: %q", line)
		}
	}
}

func TestAppendUniqueScalars_PreservesExistingInlineComments(t *testing.T) {
	t.Parallel()

	src := `remove:
  - alpha # legacy name
  - beta
`
	out, _, err := editYAML([]byte(src), func(root *yaml.Node) error {
		appendUniqueScalars(findOrCreateSeq(root, "remove"), map[string]string{"gamma": "Third greek letter."})
		return nil
	})
	require.NoError(t, err)
	got := string(out)
	require.Contains(t, got, "legacy name")
	assertEntryWithComment(t, got, "gamma", "Third greek letter.")
}

func TestDescribeForComment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"   \n\t", ""},
		{"single line", "single line"},
		{"first line\nsecond line", "first line"},
		{"\n\n  trimmed  \n\n", "trimmed"},
		{strings.Repeat("x", 130), strings.Repeat("x", 120) + "..."},
		// Multi-byte UTF-8: byte truncation at 120 would split 你.
		{strings.Repeat("a", 118) + "你zz", strings.Repeat("a", 118) + "你z..."},
	}
	for _, tt := range tests {
		got := describeForComment(tt.in)
		require.Equal(t, tt.want, got, "input=%q", tt.in)
		require.True(t, utf8.ValidString(got), "output must be valid UTF-8: %q", got)
	}
}

var errBoom = &boomErr{}

type boomErr struct{}

func (*boomErr) Error() string { return "boom" }

func assertCountInSeq(t *testing.T, src, key string, want int) {
	t.Helper()
	var doc yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(src), &doc))
	seq := findOrCreateSeq(doc.Content[0], key)
	require.Equal(t, yaml.SequenceNode, seq.Kind)
	require.Len(t, seq.Content, want)
}

// assertEntryWithComment finds a line containing both the entry value
// and the description substring. yaml.v3 picks the exact spacing.
func assertEntryWithComment(t *testing.T, yaml, entry, desc string) {
	t.Helper()
	for _, line := range strings.Split(yaml, "\n") {
		if !strings.Contains(line, "- "+entry) {
			continue
		}
		if strings.Contains(line, desc) {
			return
		}
	}
	t.Fatalf("no line found with entry %q and description %q in:\n%s", entry, desc, yaml)
}
