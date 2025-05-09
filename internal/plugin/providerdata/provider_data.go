package providerdata

import (
	"fmt"

	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
)

// ProviderData defines the interface for the Aiven provider data.
type ProviderData interface {
	// GetClient returns the handwritten Aiven client.
	GetClient() *aiven.Client

	// GetGenClient returns the generated Aiven client.
	GetGenClient() avngen.Client
}

func FromRequest(reqProviderData any) (ProviderData, diag.Diagnostics) {
	p, ok := reqProviderData.(ProviderData)
	if !ok {
		var diags diag.Diagnostics
		diags.AddError(
			errmsg.SummaryUnexpectedProviderDataType,
			fmt.Sprintf(errmsg.DetailUnexpectedProviderDataType, reqProviderData),
		)
		return nil, diags
	}
	return p, nil
}
