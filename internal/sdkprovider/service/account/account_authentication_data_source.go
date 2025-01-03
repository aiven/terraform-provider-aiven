package account

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceAccountAuthentication() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceAccountAuthenticationRead),
		Description: "The Account Authentication data source provides information about the existing Aiven Account Authentication.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenAccountAuthenticationSchema,
			"account_id", "name"),
		DeprecationMessage: "This resource is deprecated",
	}
}

func datasourceAccountAuthenticationRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	name := d.Get("name").(string)
	accountID := d.Get("account_id").(string)

	resp, err := client.AccountAuthenticationMethodsList(ctx, accountID)
	if err != nil {
		return err
	}

	for _, am := range resp {
		if am.AuthenticationMethodName != nil && *am.AuthenticationMethodName == name {
			d.SetId(schemautil.BuildResourceID(am.AccountId, am.AuthenticationMethodId))

			return resourceAccountAuthenticationRead(ctx, d, client)
		}
	}

	return fmt.Errorf("account authentication method %q not found", name)
}
