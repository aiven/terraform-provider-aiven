package changelog

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var reReleasedHeader = regexp.MustCompile(`^## \[\d`)

var newlines = strings.NewReplacer("\r\n", "\n", "\r", "\n")

func normalizeNL(content string) string {
	return newlines.Replace(content)
}

// SplitFile locates the draft section and returns text before the draft body,
// the draft body, text from the first released section onward, and the 1-based
// file line number of the first line in the body slice (including blank lines).
func SplitFile(content string) (before, body, after string, bodyStartLine int, err error) {
	content = normalizeNL(content)
	lines := strings.Split(content, "\n")
	draftIdx := -1
	releasedIdx := len(lines)

	for i, line := range lines {
		if line == DraftHeader && draftIdx < 0 {
			draftIdx = i
			continue
		}
		if draftIdx >= 0 && i > draftIdx && reReleasedHeader.MatchString(line) {
			releasedIdx = i
			break
		}
	}
	if draftIdx < 0 {
		return "", "", "", 0, fmt.Errorf("draft header %q not found", DraftHeader)
	}

	before = strings.Join(lines[:draftIdx+1], "\n")
	bodyStartLine = draftIdx + 2
	if draftIdx+1 < releasedIdx {
		body = strings.Join(lines[draftIdx+1:releasedIdx], "\n")
		// Leading blank lines stay so bodyStartLine keeps matching the file.
		body = strings.TrimRight(body, "\n")
	}
	if releasedIdx < len(lines) {
		after = strings.Join(lines[releasedIdx:], "\n")
	}
	return before, body, after, bodyStartLine, nil
}

// ParseDraft parses the draft body into sections and bullets.
// lineBase is the 1-based file line number of body's first line.
func ParseDraft(body string, lineBase int) (*Draft, error) {
	d := &Draft{}
	var sec *Section

	// lastBullet is the bullet that nested and wrapped lines attach to.
	lastBullet := func() *Bullet {
		if sec == nil || len(sec.Bullets) == 0 {
			return nil
		}
		return &sec.Bullets[len(sec.Bullets)-1]
	}

	for i, raw := range strings.Split(normalizeNL(body), "\n") {
		line := lineBase + i

		switch {
		case raw == "":
		case strings.HasPrefix(raw, "### "):
			name := strings.TrimSpace(strings.TrimPrefix(raw, "### "))
			if !slices.Contains(SectionOrder, name) {
				return nil, fmt.Errorf(
					"line %d: unknown section %q (want one of: %s)",
					line,
					name,
					strings.Join(SectionOrder, ", "),
				)
			}
			// A repeated header (e.g. left by a merge conflict) continues the first section.
			sec = d.section(name)
		case strings.HasPrefix(raw, topBullet):
			if sec == nil {
				return nil, fmt.Errorf(
					"line %d: bullet before any section (allowed sections: %s)",
					line,
					strings.Join(SectionOrder, ", "),
				)
			}
			sec.Bullets = append(sec.Bullets, Bullet{Text: NormalizeSpace(strings.TrimPrefix(raw, topBullet))})
		case strings.HasPrefix(raw, childBullet):
			parent := lastBullet()
			if parent == nil {
				return nil, fmt.Errorf("line %d: child bullet without parent", line)
			}
			parent.Children = append(parent.Children, NormalizeSpace(strings.TrimPrefix(raw, childBullet)))
		case strings.HasPrefix(raw, " ") || strings.HasPrefix(raw, "\t"):
			parent := lastBullet()
			if parent == nil {
				return nil, fmt.Errorf("line %d: unexpected continuation", line)
			}
			cont := NormalizeSpace(raw)
			if strings.HasPrefix(cont, "- ") || strings.HasPrefix(cont, "* ") || strings.HasPrefix(cont, "+ ") {
				return nil, fmt.Errorf("line %d: nested bullets must start with %q", line, childBullet)
			}
			if n := len(parent.Children); n > 0 {
				parent.Children[n-1] = strings.TrimSpace(parent.Children[n-1] + " " + cont)
			} else {
				parent.Text = strings.TrimSpace(parent.Text + " " + cont)
			}
		default:
			return nil, fmt.Errorf("line %d: unexpected line %q", line, raw)
		}
	}

	return d, nil
}
