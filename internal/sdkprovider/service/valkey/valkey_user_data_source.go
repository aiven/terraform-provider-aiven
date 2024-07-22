package valkey

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceValkeyUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceValkeyUserRead),
		Description: "The Valkey User data source provides information about the existing Aiven for Valkey user.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenValkeyUserSchema,
			"project", "service_name", "username"),
	}
}

func datasourceValkeyUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userName := d.Get("username").(string)
	d.SetId(schemautil.BuildResourceID(projectName, serviceName, userName))
	return resourceValkeyUserRead(ctx, d, client)
}
