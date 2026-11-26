package changelog

import "slices"

// DraftHeader is the unreleased section marker left in CHANGELOG.md between releases.
const DraftHeader = "## [MAJOR.MINOR.PATCH] - YYYY-MM-DD"

const (
	SectionActionRequired     = "Action required"
	SectionNew                = "New"
	SectionImproved           = "Improved"
	SectionFixed              = "Fixed"
	SectionUserConfigurations = "User configurations"
	childBullet               = "  * "
	topBullet                 = "- "
)

// SectionOrder is the fixed render order; empty sections are omitted.
var SectionOrder = []string{
	SectionActionRequired,
	SectionNew,
	SectionImproved,
	SectionFixed,
	SectionUserConfigurations,
}

// Bullet is a top-level draft bullet with optional nested bullets.
type Bullet struct {
	Text     string
	Children []string
}

// Section holds bullets for one ### section.
type Section struct {
	Name    string
	Bullets []Bullet
}

// Draft is the parsed unreleased changelog body (without the ## header).
// Bullet and child text is single-spaced: ParseDraft and the generator normalize it on entry.
type Draft struct {
	Sections []Section
}

// section returns the named section, appending it when missing. The pointer
// is valid until the next append to d.Sections.
func (d *Draft) section(name string) *Section {
	i := slices.IndexFunc(d.Sections, func(s Section) bool { return s.Name == name })
	if i < 0 {
		d.Sections = append(d.Sections, Section{Name: name})
		i = len(d.Sections) - 1
	}
	return &d.Sections[i]
}

// GeneratedEntry is one generated changelog line before merge.
type GeneratedEntry struct {
	Section string
	Text    string // bullet text, or the nested line when Group is set
	Group   string // set for user-config children: parent bullet text
}
