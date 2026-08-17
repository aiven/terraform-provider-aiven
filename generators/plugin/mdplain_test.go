package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPlainTextCallouts(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expect   string
	}{
		{
			name:     "keeps a description without callouts",
			markdown: "Creates and manages a service.",
			expect:   "Creates and manages a service.",
		},
		{
			name:     "drops a trailing callout sigil",
			markdown: "Creates and manages a service.\n\n~> This resource is in the limited availability stage.",
			expect:   "Creates and manages a service. This resource is in the limited availability stage.",
		},
		{
			name:     "drops a leading callout sigil",
			markdown: "!> **Teams have been replaced by groups**\n\nCreates and manages a team.",
			expect:   "Teams have been replaced by groups. Creates and manages a team.",
		},
		{
			name:     "ends the sentence a callout box used to carry",
			markdown: "Creates and manages a service.\n\n~> **This resource is deprecated**\nUse another one instead.",
			expect:   "Creates and manages a service. This resource is deprecated. Use another one instead.",
		},
		{
			// A callout has a title and a body, both of which the metadata keeps as prose.
			name:     "joins the title and the body of a callout",
			markdown: "Creates and manages an Aiven account.\n\n~> **This resource is deprecated**\nThis resource will be removed in v5.0.0. Use `aiven_organization` instead.",
			expect:   "Creates and manages an Aiven account. This resource is deprecated. This resource will be removed in v5.0.0. Use aiven_organization instead.",
		},
		{
			name:     "keeps the paragraphs that surround a callout",
			markdown: "Creates and manages a service.\n\n-> A hint.\n\nThe service must be powered on.",
			expect:   "Creates and manages a service. A hint. The service must be powered on.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expect, plainText(tt.markdown))
		})
	}
}
