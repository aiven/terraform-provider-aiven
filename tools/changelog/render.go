package changelog

import (
	"strings"
)

// RenderSections renders ### sections only (no ## header).
func RenderSections(d *Draft) string {
	if d == nil || len(d.Sections) == 0 {
		return ""
	}
	byName := make(map[string]*Section, len(d.Sections))
	for i := range d.Sections {
		byName[d.Sections[i].Name] = &d.Sections[i]
	}

	blocks := make([]string, 0, len(SectionOrder))
	for _, name := range SectionOrder {
		sec, ok := byName[name]
		if !ok || len(sec.Bullets) == 0 {
			continue
		}
		var lines []string
		lines = append(lines, "### "+name, "")
		for _, bul := range sec.Bullets {
			lines = append(lines, topBullet+bul.Text)
			for _, ch := range bul.Children {
				lines = append(lines, childBullet+ch)
			}
		}
		blocks = append(blocks, strings.Join(lines, "\n"))
	}
	return strings.Join(blocks, "\n\n")
}

// FormatDraftSection parses, optionally merges generated entries, and re-renders the draft.
func FormatDraftSection(fileContent string, generated ...GeneratedEntry) (string, error) {
	before, body, after, bodyStartLine, err := SplitFile(fileContent)
	if err != nil {
		return "", err
	}

	d, err := ParseDraft(body, bodyStartLine)
	if err != nil {
		return "", err
	}
	if len(generated) > 0 {
		MergeGenerated(d, generated)
	}

	rendered := RenderSections(d)

	var out strings.Builder
	out.WriteString(before)
	if rendered != "" {
		out.WriteString("\n\n")
		out.WriteString(rendered)
	}
	if after != "" {
		out.WriteString("\n\n")
		out.WriteString(after)
	}
	return out.String(), nil
}
