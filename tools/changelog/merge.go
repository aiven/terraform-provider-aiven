package changelog

import (
	"cmp"
	"slices"
)

// MergeGenerated appends generated entries into the draft (deduped).
func MergeGenerated(d *Draft, entries []GeneratedEntry) {
	if len(entries) == 0 {
		return
	}

	var uc, flat []GeneratedEntry
	for _, e := range entries {
		if e.Group != "" {
			uc = append(uc, e)
		} else {
			flat = append(flat, e)
		}
	}

	for _, e := range flat {
		sec := d.section(e.Section)
		if slices.ContainsFunc(sec.Bullets, func(b Bullet) bool { return b.Text == e.Text }) {
			continue
		}
		sec.Bullets = append(sec.Bullets, Bullet{Text: e.Text})
	}

	if len(uc) == 0 {
		return
	}

	slices.SortFunc(uc, func(a, b GeneratedEntry) int {
		return cmp.Compare(a.Group, b.Group)
	})

	sec := d.section(SectionUserConfigurations)
	for _, e := range uc {
		idx := slices.IndexFunc(sec.Bullets, func(b Bullet) bool { return b.Text == e.Group })
		if idx < 0 {
			sec.Bullets = append(sec.Bullets, Bullet{Text: e.Group})
			idx = len(sec.Bullets) - 1
		}
		if !slices.Contains(sec.Bullets[idx].Children, e.Text) {
			sec.Bullets[idx].Children = append(sec.Bullets[idx].Children, e.Text)
		}
	}

	// Plain string order is Added, Changed, Removed, then the schema path.
	for i := range sec.Bullets {
		slices.Sort(sec.Bullets[i].Children)
	}

	slices.SortFunc(sec.Bullets, func(a, b Bullet) int {
		return cmp.Compare(a.Text, b.Text)
	})
}
