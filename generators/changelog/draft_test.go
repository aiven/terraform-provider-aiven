package main

import (
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cl "github.com/aiven/terraform-provider-aiven/tools/changelog"
)

const testSample = `# Changelog

## [MAJOR.MINOR.PATCH] - YYYY-MM-DD

### Fixed

- Fix aiven_project: can't migrate

## [4.26.0] - 2024-09-25

- Remove aiven_valkey from beta resources

## [1.0.0] - 2018-09-27

Support all Aiven resource types.
`

func TestFormatDraftSection_mergeGenerated(t *testing.T) {
	gen := []cl.GeneratedEntry{
		{Section: cl.SectionNew, Text: "`foo` resource field `bar`: Baz."},
	}
	result, err := cl.FormatDraftSection(testSample, gen...)
	require.NoError(t, err)
	assert.Contains(t, result, "### New\n\n")
	assert.Contains(t, result, "`foo` resource field `bar`: Baz.")
	assert.Contains(t, result, "### Fixed")
	assert.Contains(t, result, "## [4.26.0] - 2024-09-25")
	assert.Contains(t, result, "Remove aiven_valkey")
}

func TestFormatDraftSection_handWrittenOrder(t *testing.T) {
	sample := `## [MAJOR.MINOR.PATCH] - YYYY-MM-DD

### New

- First hand-written
- Second hand-written
`
	gen := []cl.GeneratedEntry{{Section: cl.SectionNew, Text: "`z` resource: new."}}
	result, err := cl.FormatDraftSection(sample, gen...)
	require.NoError(t, err)
	iFirst := strings.Index(result, "First hand-written")
	iSecond := strings.Index(result, "Second hand-written")
	iGen := strings.Index(result, "`z` resource")
	assert.Less(t, iFirst, iSecond)
	assert.Less(t, iSecond, iGen)
}

func TestFormatDraftSection_userConfigMerge(t *testing.T) {
	sample := `## [MAJOR.MINOR.PATCH] - YYYY-MM-DD

### User configurations

- ` + "`aiven_grafana.grafana_user_config`" + `
  * Added ` + "`existing`" + `: Old entry
`
	gen := []cl.GeneratedEntry{
		{
			Section: cl.SectionUserConfigurations,
			Group:   "`aiven_grafana.grafana_user_config`",
			Text:    "Added `existing`: Old entry",
		},
		{
			Section: cl.SectionUserConfigurations,
			Group:   "`aiven_grafana.grafana_user_config`",
			Text:    "Changed `grafana_version` (enum): add `13`",
		},
	}
	result, err := cl.FormatDraftSection(sample, gen...)
	require.NoError(t, err)
	assert.Equal(t, 1, strings.Count(result, "`aiven_grafana.grafana_user_config`"))
	assert.Contains(t, result, "Changed `grafana_version` (enum): add `13`")
}

func TestFormatDraftSection_parseErrors(t *testing.T) {
	_, err := cl.FormatDraftSection("## [MAJOR.MINOR.PATCH] - YYYY-MM-DD\n\n- bullet before section\n")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bullet before any section")
}

func TestFormatDraftSection_releasedUntouched(t *testing.T) {
	released := "## [4.26.0] - 2024-09-25\n\n- Remove aiven_valkey from beta resources\n"
	sample := testSample
	result, err := cl.FormatDraftSection(sample)
	require.NoError(t, err)
	assert.Contains(t, result, released)
}

func TestFormatDraftSection_changelogFileIdempotent(t *testing.T) {
	b, err := os.ReadFile("../../CHANGELOG.md")
	require.NoError(t, err)

	once, err := cl.FormatDraftSection(string(b))
	require.NoError(t, err)
	twice, err := cl.FormatDraftSection(once)
	require.NoError(t, err)
	assert.Empty(t, cmp.Diff(once, twice))
}

func TestGeneratedEntry_singleSpaced(t *testing.T) {
	item := &Item{
		Path:        "aiven_x.y_user_config.opt",
		Kind:        ResourceRootKind,
		Description: "First line.\n  Second   line.",
	}
	flat := generatedEntry(&Diff{Action: ChangeDiffAction, Item: &Item{Path: "aiven_x.y", Kind: ResourceRootKind}, Description: "a\n b"})
	child := generatedEntry(&Diff{Action: AddDiffAction, Item: item})

	assert.NotContains(t, flat.Text, "\n")
	assert.NotContains(t, child.Text, "\n")
	assert.NotContains(t, child.Text, "  ")
}
