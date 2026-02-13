package paymentmethodlist

import (
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationbilling"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestReadView(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		methods []organizationbilling.PaymentMethodsListOut
	}{
		{
			name:    "empty list",
			methods: []organizationbilling.PaymentMethodsListOut{},
		},
		{
			name: "single payment method",
			methods: []organizationbilling.PaymentMethodsListOut{
				{
					PaymentMethodId:   "pm-1",
					PaymentMethodType: organizationbilling.PaymentMethodTypeCreditCard,
				},
			},
		},
		{
			name: "multiple payment methods",
			methods: []organizationbilling.PaymentMethodsListOut{
				{
					PaymentMethodId:   "pm-1",
					PaymentMethodType: organizationbilling.PaymentMethodTypeCreditCard,
				},
				{
					PaymentMethodId:   "pm-2",
					PaymentMethodType: organizationbilling.PaymentMethodTypeBankTransfer,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orgID := "org-id"
			client := avngen.NewMockClient(t)

			state := &tfModel{
				OrganizationID: types.StringValue(orgID),
			}

			client.EXPECT().
				PaymentMethodsList(t.Context(), orgID).
				Return(tc.methods, nil).
				Once()

			diags := readView(t.Context(), client, state)

			require.False(t, diags.HasError(), "expected no errors but got: %v", diags)
		})
	}
}
