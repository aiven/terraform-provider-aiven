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
			want: "Service setting. If this resource is missing (for example, after a service power off), " +
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
