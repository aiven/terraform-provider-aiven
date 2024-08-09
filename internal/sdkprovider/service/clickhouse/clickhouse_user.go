package clickhouse

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
		CreateContext: resourceClickhouseUserCreate,
		ReadContext:   resourceClickhouseUserRead,
		DeleteContext: resourceClickhouseUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenClickhouseUserSchema,
	}
}

func resourceClickhouseUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	u, err := client.ClickhouseUser.Create(
		ctx,
		projectName,
		serviceName,
		username,
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, u.User.UUID))

	if err := d.Set("password", u.User.Password); err != nil {
		return diag.FromErr(err)
	}

	return resourceClickhouseUserRead(ctx, d, m)
}

func resourceClickhouseUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, uuid, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	user, err := client.ClickhouseUser.Get(ctx, projectName, serviceName, uuid)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("project", projectName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("username", user.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("uuid", user.UUID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("required", user.Required); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceClickhouseUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, uuid, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ClickhouseUser.Delete(ctx, projectName, serviceName, uuid)
	if common.IsCritical(err) {
		return diag.FromErr(err)
	}

	return nil
}
