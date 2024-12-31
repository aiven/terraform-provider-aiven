package account

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceAccountRead),
		Description: "The Account data source provides information about the existing Aiven Account.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenAccountSchema, "name"),
	}
}

func datasourceAccountRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	name := d.Get("name").(string)

	resp, err := client.AccountList(ctx)
	if err != nil {
		return err
	}

	for _, ac := range resp {
		if ac.AccountName == name {
			d.SetId(ac.AccountId)

			return resourceAccountRead(ctx, d, client)
		}
	}

	return fmt.Errorf("account %q not found", name)
}
