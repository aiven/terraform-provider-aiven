package userconfig

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDescriptionBuilder_Build(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		desc *DescriptionBuilder
		want string
	}{
		{
			name: "adds trailing dot",
			desc: Desc("Service setting"),
			want: "Service setting.",
		},
		{
			name: "keeps existing trailing dot",
			desc: Desc("Service setting."),
			want: "Service setting.",
		},
		{
			name: "returns empty string for empty description",
			desc: Desc(""),
			want: "",
		},
		{
			name: "empty description with possible values",
			desc: Desc("").PossibleValuesString("a", "b"),
			want: "The possible values are `a` and `b`.",
		},
		{
			name: "empty description alone",
			desc: Desc("").PossibleValuesString(),
			want: "",
		},
		{
			name: "adds format hint",
			desc: Desc("Billing group ID").FormatHint("Must be a valid UUID.").ForceNew(),
			want: "Billing group ID. Must be a valid UUID. Changing this property forces recreation of the resource.",
		},
		{
			name: "adds example",
			desc: Desc("Connection string").Example("AccountName=IDENT"),
			want: "Connection string. Example: `AccountName=IDENT`.",
		},
		{
			name: "prefers default to example, renders numbers without exponent",
			desc: Desc("Batch size").DefaultValue(1e6).Example(5.0),
			want: "Batch size. The default value is `1000000`.",
		},
		{
			name: "drops extra trailing dots",
			desc: Desc("Service setting.. "),
			want: "Service setting.",
		},
		{
			name: "keeps question mark",
			desc: Desc("Is it enabled?"),
			want: "Is it enabled?",
		},
		{
			name: "keeps a closing code fence",
			desc: Desc("Example:\n```\nfoo\n```"),
			want: "Example:\n```\nfoo\n```",
		},
		{
			name: "keeps appended sentences outside a code fence",
			desc: Desc("Example:\n\n```\ncode\n```").DefaultValue(42).ForceNew(),
			want: "Example:\n\n```\ncode\n```\n\nThe default value is `42`. Changing this property forces recreation of the resource.",
		},
		{
			name: "capitalizes a leading word",
			desc: Desc("service setting"),
			want: "Service setting.",
		},
		{
			name: "keeps a leading literal",
			desc: Desc("fnmatch pattern"),
			want: "fnmatch pattern.",
		},
		{
			name: "wraps a leading identifier in code",
			desc: Desc("user_operations records API calls"),
			want: "`user_operations` records API calls.",
		},
		{
			name: "wraps single-quoted tokens in code",
			desc: Desc("Prefix 'aiven-' is reserved. Values: 'delete', 'compact'. Symbols: '.', '_id'. 'http://'/'https://' URL. Don't touch apostrophes"),
			want: "Prefix `aiven-` is reserved. Values: `delete`, `compact`. Symbols: `.`, `_id`. `http://`/`https://` URL. Don't touch apostrophes.",
		},
		{
			name: "wraps wildcards in code",
			desc: Desc("The host or * for all hosts, foo * bar, or public-*, logs_*_foo_* names"),
			want: "The host or `*` for all hosts, foo `*` bar, or `public-*`, `logs_*_foo_*` names.",
		},
		{
			name: "keeps emphasis and code",
			desc: Desc("**Bold** and *italic*, **bold text**, a*b, (*), ends with *, already `*` and `'x'` wrapped"),
			want: "**Bold** and *italic*, **bold text**, a*b, (*), ends with *, already `*` and `'x'` wrapped.",
		},
		{
			name: "wraps placeholders in code",
			desc: Desc("Each entry is <group object ID>:<org ID>:<role>. Use ${<provider>:<key>}, or Bearer <token>"),
			want: "Each entry is `<group object ID>:<org ID>:<role>`. Use `${<provider>:<key>}`, or Bearer `<token>`.",
		},
		{
			name: "keeps code and comparisons as is",
			desc: Desc("Example: `https://example.com/<TENANT_ID>/token`. Requires a < b and c > d\n\n```\n<root>\n```"),
			want: "Example: `https://example.com/<TENANT_ID>/token`. Requires a < b and c > d\n\n```\n<root>\n```",
		},
		{
			name: "adds beta availability warning",
			desc: Desc("Service setting").Beta(),
			want: "Service setting. " + `

**This resource is in the beta stage and may change without notice.** Set
the ` + "`PROVIDER_AIVEN_ENABLE_BETA`" + ` environment variable to use the resource.`,
		},
		{
			name: "keeps limited availability warning inline",
			desc: Desc("Service setting").LimitedAvailability(),
			want: "Service setting. " + `

**This resource is in the limited availability stage and may change without notice.** ` + LimitedAvailabilityMessage,
		},
		{
			name: "keeps beta and limited availability warnings inline",
			desc: Desc("Service setting").Beta().LimitedAvailability(),
			want: "Service setting. " + fmt.Sprintf(BetaLimitedAvailabilityText, "resource"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, tt.desc.Build())
		})
	}
}

func TestDescriptionBuilder_BuildPage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		desc *DescriptionBuilder
		want string
	}{
		{
			name: "adds limited availability warning as a callout",
			desc: Desc("Service setting").LimitedAvailability(),
			want: "Service setting." + `

~> **Limited availability resource**
This resource is in the limited availability stage and may change without notice. ` + LimitedAvailabilityMessage,
		},
		{
			name: "uses data source in availability warning",
			desc: Desc("Service setting").MarkAsDataSource().LimitedAvailability(),
			want: "Service setting." + `

~> **Limited availability data source**
This data source is in the limited availability stage and may change without notice. ` + LimitedAvailabilityMessage,
		},
		{
			name: "adds beta and limited availability warnings",
			desc: Desc("Service setting").Beta().LimitedAvailability(),
			want: "Service setting.\n\n~> **Beta resource in limited availability**\n" +
				fmt.Sprintf(BetaLimitedAvailabilityText, "resource"),
		},
		{
			name: "keeps the callout last, so nothing is pulled into the box",
			desc: Desc("Service setting").LimitedAvailability().RemoveMissing(),
			want: "Service setting. If a refresh can't find this resource, " +
				"it's removed from the state and a new create plan is generated." + `

~> **Limited availability resource**
This resource is in the limited availability stage and may change without notice. ` + LimitedAvailabilityMessage,
		},
		{
			name: "moves the appended sentences out of a closing callout",
			desc: Desc("Service setting.\n\n~> **Warning**\nMind the gap.").LimitedAvailability().RemoveMissing(),
			want: "Service setting.\n\n~> **Warning**\nMind the gap." + `

If a refresh can't find this resource, ` +
				"it's removed from the state and a new create plan is generated." + `

~> **Limited availability resource**
This resource is in the limited availability stage and may change without notice. ` + LimitedAvailabilityMessage,
		},
		{
			name: "adds beta warning as a callout",
			desc: Desc("Service setting").Beta(),
			want: "Service setting.\n\n~> **Beta resource**\n" + fmt.Sprintf(BetaText, "resource"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, tt.desc.BuildPage())
		})
	}
}
