// Copyright (c) 2026 Aiven, Helsinki, Finland. https://aiven.io/

package validators

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestIPv4(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		for _, value := range []types.String{
			types.StringValue("10.20.30.40"),
			types.StringValue("172.16.0.1"),
			types.StringValue("192.168.1.2"),
			types.StringValue("8.8.8.8"),
			types.StringValue("0.0.0.0"),
			types.StringValue("255.255.255.255"),
			types.StringNull(),
			types.StringUnknown(),
		} {
			t.Run(value.String(), func(t *testing.T) {
				req := validator.StringRequest{Path: path.Root("address"), ConfigValue: value}
				var rsp validator.StringResponse
				IPv4().ValidateString(t.Context(), req, &rsp)
				require.Empty(t, rsp.Diagnostics)
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for _, value := range []types.String{
			types.StringValue(""),
			types.StringValue("abc"),
			types.StringValue("256.0.0.1"),
			types.StringValue("10.0.0.01"),
			types.StringValue("10.0.0"),
			types.StringValue("10.0.0.2/24"),
			types.StringValue("::1"),
			types.StringValue("2001:db8::1"),
			types.StringValue("::ffff:10.0.0.2"),
			types.StringValue("::ffff:0:1"),
			types.StringValue(" 10.0.0.2"),
		} {
			t.Run(value.String(), func(t *testing.T) {
				req := validator.StringRequest{Path: path.Root("address"), ConfigValue: value}
				var rsp validator.StringResponse
				IPv4().ValidateString(t.Context(), req, &rsp)
				require.Len(t, rsp.Diagnostics.Errors(), 1)
				require.Equal(t, fmt.Sprintf("%q must be a valid IPv4 address", value.ValueString()), rsp.Diagnostics.Errors()[0].Detail())
			})
		}
	})
}
