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
			name: "adds limited availability warning",
			desc: Desc("Service setting").LimitedAvailability(),
			want: "Service setting. " + `

**This resource is in the limited availability stage and may change without notice.** ` + LimitedAvailabilityMessage,
		},
		{
			name: "uses data source in availability warning",
			desc: Desc("Service setting").MarkAsDataSource().LimitedAvailability(),
			want: "Service setting. " + `

**This data source is in the limited availability stage and may change without notice.** ` + LimitedAvailabilityMessage,
		},
		{
			name: "adds beta and limited availability warnings",
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
