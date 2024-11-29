package alloydbomni

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceAlloyDBOmniUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceAlloyDBOmniUserRead),
		Description: "Gets information about an Aiven for AlloyDB Omni service user.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenAlloyDBOmniUserSchema,
			"project", "service_name", "username"),
	}
}

func datasourceAlloyDBOmniUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userName := d.Get("username").(string)

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, userName))
	return resourceAlloyDBOmniUserRead(ctx, d, client)
}
