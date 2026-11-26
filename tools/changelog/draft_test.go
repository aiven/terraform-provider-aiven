package changelog

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDraft_nested(t *testing.T) {
	body := `### User configurations

- ` + "`group`" + `
  * Added ` + "`child`" + `: detail
  * Changed ` + "`other`" + `: more`

	d, err := ParseDraft(body, 1)
	require.NoError(t, err)
	require.Len(t, d.Sections, 1)
	require.Len(t, d.Sections[0].Bullets, 1)
	assert.Len(t, d.Sections[0].Bullets[0].Children, 2)
}

func TestParseDraft_joinsWrappedContinuation(t *testing.T) {
	body := `### New

- Long bullet starts here
  and continues indented
`
	d, err := ParseDraft(body, 1)
	require.NoError(t, err)
	assert.Equal(t, "Long bullet starts here and continues indented", d.Sections[0].Bullets[0].Text)
}

func TestFormatDraftSection_joinWrappedInput(t *testing.T) {
	file := DraftHeader + `

### New

- First part of bullet
  continues on the next line

### User configurations

- ` + "`group`" + `
  * Added ` + "`child`" + `: first line
    and second line
`
	got, err := FormatDraftSection(file)
	require.NoError(t, err)
	assert.Contains(t, got, "- First part of bullet continues on the next line\n")
	assert.Contains(t, got, "  * Added `child`: first line and second line")
}

func TestFormatDraftSection_longBulletOneLine(t *testing.T) {
	long := strings.Repeat("word ", 60)
	file := DraftHeader + `

### New

- ` + long + `
`
	got, err := FormatDraftSection(file)
	require.NoError(t, err)
	lines := strings.Split(got, "\n")
	var bullet string
	for _, line := range lines {
		if strings.HasPrefix(line, "- ") && !strings.HasPrefix(line, "###") {
			bullet = line
			break
		}
	}
	require.NotEmpty(t, bullet)
	assert.Less(t, len(bullet), 500)
	assert.NotContains(t, got, "\n  word")
}

func TestFormatDraftSection_idempotent(t *testing.T) {
	file := DraftHeader + `

### New

- One entry

### Fixed

- Another entry
`
	once, err := FormatDraftSection(file)
	require.NoError(t, err)
	twice, err := FormatDraftSection(once)
	require.NoError(t, err)
	assert.Empty(t, cmp.Diff(once, twice))
}

func TestFormatDraftSection_CRLF(t *testing.T) {
	const lf = "## [MAJOR.MINOR.PATCH] - YYYY-MM-DD\n\n### New\n\n- One bullet\n"
	crlf := strings.ReplaceAll(lf, "\n", "\r\n")
	gotLF, err := FormatDraftSection(lf)
	require.NoError(t, err)
	gotCRLF, err := FormatDraftSection(crlf)
	require.NoError(t, err)
	assert.Empty(t, cmp.Diff(gotLF, gotCRLF))
	assert.NotContains(t, gotCRLF, "\r")
}

func TestMergeGenerated_emptyPreservesUserConfigOrder(t *testing.T) {
	file := DraftHeader + `

### User configurations

- ` + "`zebra.service`" + `
  * Added ` + "`z_first`" + `: first
  * Changed ` + "`z_second`" + `: second
- ` + "`alpha.service`" + `
  * Added ` + "`a_only`" + `: only
`
	got, err := FormatDraftSection(file)
	require.NoError(t, err)
	alpha := strings.Index(got, "`alpha.service`")
	zebra := strings.Index(got, "`zebra.service`")
	require.Positive(t, alpha)
	require.Positive(t, zebra)
	assert.Less(t, zebra, alpha, "hand-written group order must be preserved")

	zFirst := strings.Index(got, "`z_first`")
	zSecond := strings.Index(got, "`z_second`")
	assert.Less(t, zFirst, zSecond, "hand-written child order must be preserved")
}

func TestParseDraft_fileLineError(t *testing.T) {
	file := "# title\n\n" + DraftHeader + "\n\n- orphan bullet\n"
	_, err := FormatDraftSection(file)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "line 5:")
}

func TestParseDraft_fileLineErrorAfterBlankLines(t *testing.T) {
	file := "# title\n\n" + DraftHeader + "\n\n\n- orphan bullet\n"
	_, err := FormatDraftSection(file)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "line 6:")
}

func TestParseDraft_misindentedChildIsError(t *testing.T) {
	for name, child := range map[string]string{
		"tab":        "\t* Added `x`: y",
		"one space":  " * Added `x`: y",
		"four space": "    * Added `x`: y",
		"dash":       "  - Added `x`: y",
	} {
		t.Run(name, func(t *testing.T) {
			file := DraftHeader + "\n\n### User configurations\n\n- `g`\n" + child + "\n"
			_, err := FormatDraftSection(file)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "nested bullets must start with")
		})
	}
}

func TestMergeGenerated_existingSectionAfterNewOne(t *testing.T) {
	file := DraftHeader + "\n\n### Fixed\n\n- keep me\n"
	got, err := FormatDraftSection(
		file,
		GeneratedEntry{Section: SectionNew, Text: "new added"},
		GeneratedEntry{Section: SectionImproved, Text: "new changed"},
		GeneratedEntry{Section: SectionFixed, Text: "new fixed"},
	)
	require.NoError(t, err)
	assert.Contains(t, got, "- keep me\n- new fixed")
	assert.Contains(t, got, "- new added")
	assert.Contains(t, got, "- new changed")
}

func TestParseDraft_duplicateSectionIsMerged(t *testing.T) {
	file := DraftHeader + "\n\n### Fixed\n\n- first\n\n### New\n\n- a\n\n### Fixed\n\n- second\n"
	got, err := FormatDraftSection(file)
	require.NoError(t, err)
	assert.Contains(t, got, "### Fixed\n\n- first\n- second")
	assert.Equal(t, 1, strings.Count(got, "### Fixed"))
}

func TestMergeGenerated_childrenSortAddedChangedRemoved(t *testing.T) {
	file := DraftHeader + "\n\n### User configurations\n\n- `g`\n  * Removed `z`\n"
	got, err := FormatDraftSection(
		file,
		GeneratedEntry{Section: SectionUserConfigurations, Group: "`g`", Text: "Changed `b`: x"},
		GeneratedEntry{Section: SectionUserConfigurations, Group: "`g`", Text: "Added `c`: y"},
		GeneratedEntry{Section: SectionUserConfigurations, Group: "`g`", Text: "Added `a`: y"},
	)
	require.NoError(t, err)
	assert.Contains(t, got, "- `g`\n  * Added `a`: y\n  * Added `c`: y\n  * Changed `b`: x\n  * Removed `z`")
}
