package userconfig

import (
	"fmt"
	"strings"
)

// Callout is a sigil that turns the paragraph it opens into a colored box in the
// Terraform Registry, prefixed with an emphasized word the Registry adds itself.
// A callout can't span multiple paragraphs.
// https://developer.hashicorp.com/terraform/registry/providers/docs#callouts
type Callout string

const (
	// CalloutBlue renders a blue "Note" box.
	CalloutBlue Callout = "->"
	// CalloutYellow renders a yellow "Note" box.
	CalloutYellow Callout = "~>"
	// CalloutRed renders a red "Warning" box.
	CalloutRed Callout = "!>"
)

// Callouts lists every sigil the Terraform Registry renders as a callout box.
func Callouts() []Callout {
	return []Callout{CalloutBlue, CalloutYellow, CalloutRed}
}

// Wrap opens a callout box titled with the given title, its text follows on the
// next line. The Registry renders the box only when the title is bold.
func (c Callout) Wrap(title, text string) string {
	return fmt.Sprintf("%s **%s**\n%s", string(c), title, text)
}

// HasCallout reports whether the paragraph opens a callout box.
func HasCallout(paragraph string) bool {
	return TrimCallout(paragraph) != strings.TrimSpace(paragraph)
}

// TrimCallout removes the sigil that opens a callout box, keeping its text.
func TrimCallout(paragraph string) string {
	trimmed := strings.TrimSpace(paragraph)
	for _, c := range Callouts() {
		if text, ok := strings.CutPrefix(trimmed, string(c)); ok {
			return strings.TrimSpace(text)
		}
	}
	return paragraph
}
