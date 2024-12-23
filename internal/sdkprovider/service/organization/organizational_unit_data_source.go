package organization

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceOrganizationalUnit() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceOrganizationalUnitRead),
		Description: "Gets information about an organizational unit.",
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenOrganizationalUnitSchema, "name"),
	}
}

func datasourceOrganizationalUnitRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	name := d.Get("name").(string)

	resp, err := client.AccountList(ctx)
	if err != nil {
		return err
	}

	for _, ac := range resp {
		if ac.AccountName == name {
			d.SetId(ac.AccountId)

			return resourceOrganizationalUnitRead(ctx, d, client)
		}
	}

	return fmt.Errorf("organizational unit %q not found", name)
}
