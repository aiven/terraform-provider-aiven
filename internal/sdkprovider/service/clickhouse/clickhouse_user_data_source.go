package clickhouse

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceClickhouseUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceClickhouseUserRead),
		Description: "Gets information about a ClickHouse user.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(aivenClickhouseUserSchema,
			"project", "service_name", "username"),
	}
}

func datasourceClickhouseUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	userName := d.Get("username").(string)

	list, err := client.ServiceClickHouseUserList(ctx, projectName, serviceName)
	if err != nil {
		return err
	}

	for _, u := range list {
		if u.Name == userName {
			d.SetId(schemautil.BuildResourceID(projectName, serviceName, u.Uuid))
			return resourceClickhouseUserRead(ctx, d, client)
		}
	}

	return fmt.Errorf("clickhouse user %s/%s/%s not found",
		projectName, serviceName, userName)
}
