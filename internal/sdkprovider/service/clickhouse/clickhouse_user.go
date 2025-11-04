package clickhouse

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/clickhouse"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenClickhouseUserSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: schemautil.GetServiceUserValidateFunc(),
		Description:  userconfig.Desc("The name of the ClickHouse user.").ForceNew().Build(),
	},
	"password": {
		Type:        schema.TypeString,
		Sensitive:   true,
		Computed:    true,
		Description: "The password of the ClickHouse user.",
	},
	"uuid": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "UUID of the ClickHouse user.",
	},
	"required": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Indicates if a ClickHouse user is required.",
	},
}

func ResourceClickhouseUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages a ClickHouse user.",
		CreateContext: common.WithGenClient(resourceClickhouseUserCreate),
		ReadContext:   common.WithGenClient(resourceClickhouseUserRead),
		DeleteContext: common.WithGenClient(resourceClickhouseUserDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenClickhouseUserSchema,
	}
}

func resourceClickhouseUserCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)

	u, err := client.ServiceClickHouseUserCreate(ctx, projectName, serviceName, &clickhouse.ServiceClickHouseUserCreateIn{
		Name: username,
	})
	if err != nil {
		return fmt.Errorf("cannot create ClickHouse user: %w", err)
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, u.Uuid))

	if err = d.Set("password", u.Password); err != nil {
		return err
	}

	return resourceClickhouseUserRead(ctx, d, client)
}

func resourceClickhouseUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, uuid, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	var user *clickhouse.UserOut
	ul, err := client.ServiceClickHouseUserList(ctx, projectName, serviceName)
	if err != nil {
		return err
	}

	for _, u := range ul {
		if u.Uuid == uuid {
			user = &u
			break
		}
	}

	if user == nil {
		return schemautil.ResourceReadHandleNotFound(fmt.Errorf("user %q not found", d.Id()), d)
	}

	if err = d.Set("project", projectName); err != nil {
		return err
	}
	if err = d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err = d.Set("username", user.Name); err != nil {
		return err
	}
	if err = d.Set("uuid", user.Uuid); err != nil {
		return err
	}
	if err = d.Set("required", user.Required); err != nil {
		return err
	}

	return nil
}

func resourceClickhouseUserDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, uuid, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	err = client.ServiceClickHouseUserDelete(ctx, projectName, serviceName, uuid)
	if common.IsCritical(err) {
		return err
	}

	return nil
}
