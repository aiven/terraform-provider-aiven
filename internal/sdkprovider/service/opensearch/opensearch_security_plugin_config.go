// Package opensearch implements the Aiven OpenSearch service.
package opensearch

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// aivenOpenSearchSecurityPluginConfigSchema holds the schema for the OpenSearch Security Plugin Config resource.
var aivenOpenSearchSecurityPluginConfigSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"admin_password": {
		Type:        schema.TypeString,
		Required:    true,
		Sensitive:   true,
		Description: "The password for the os-sec-admin user.",
	},
	"available": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the security plugin is available. This is always true for recently created services.",
	},
	"enabled": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the security plugin is enabled. This is always true for recently created services.",
	},
	"admin_enabled": {
		Type:     schema.TypeBool,
		Computed: true,
		Description: "Whether the os-sec-admin user is enabled. This indicates whether the user management with the" +
			" security plugin is enabled. This is always true when the os-sec-admin password was set at least once.",
	},
}

// ResourceOpenSearchSecurityPluginConfig defines the OpenSearch Security Plugin Config resource.
func ResourceOpenSearchSecurityPluginConfig() *schema.Resource {
	return &schema.Resource{
		Description: "The OpenSearch Security Plugin Config resource allows the creation and management of Aiven" +
			"OpenSearch Security Plugin config.",
		CreateContext: resourceOpenSearchSecurityPluginConfigCreate,
		ReadContext:   resourceOpenSearchSecurityPluginConfigRead,
		UpdateContext: resourceOpenSearchSecurityPluginConfigUpdate,
		DeleteContext: resourceOpenSearchSecurityPluginConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   aivenOpenSearchSecurityPluginConfigSchema,
	}
}

// resourceOpenSearchSecurityPluginConfigCreate applies an OpenSearch Security Plugin config to an existing OpenSearch
// service, enabling the OpenSearch Security Plugin.
func resourceOpenSearchSecurityPluginConfigCreate(
	ctx context.Context,
	d *schema.ResourceData,
	m any,
) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)

	serviceName := d.Get("service_name").(string)

	if _, err := client.OpenSearchSecurityPluginHandler.Enable(
		project,
		serviceName,
		aiven.OpenSearchSecurityPluginEnableRequest{
			AdminPassword: d.Get("admin_password").(string),
		},
	); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName))

	return resourceOpenSearchSecurityPluginConfigRead(ctx, d, m)
}

// resourceOpenSearchSecurityPluginConfigRead reads the OpenSearch Security Plugin config from an existing OpenSearch
// service.
func resourceOpenSearchSecurityPluginConfigRead(
	_ context.Context,
	d *schema.ResourceData,
	m any,
) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.OpenSearchSecurityPluginHandler.Get(project, serviceName)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("project", project); err != nil {
		return diag.Errorf("error setting `project` for resource %s: %s", d.Id(), err)
	}

	if err := d.Set("service_name", serviceName); err != nil {
		return diag.Errorf("error setting `service_name` for resource %s: %s", d.Id(), err)
	}

	if err := d.Set("available", r.SecurityPluginAvailable); err != nil {
		return diag.Errorf("error setting `available` for resource %s: %s", d.Id(), err)
	}

	if err := d.Set("enabled", r.SecurityPluginEnabled); err != nil {
		return diag.Errorf("error setting `enabled` for resource %s: %s", d.Id(), err)
	}

	if err := d.Set("admin_enabled", r.SecurityPluginAdminEnabled); err != nil {
		return diag.Errorf("error setting `admin_enabled` for resource %s: %s", d.Id(), err)
	}

	return nil
}

// resourceOpenSearchSecurityPluginConfigUpdate updates the OpenSearch Security Plugin config on an existing OpenSearch
// service.
func resourceOpenSearchSecurityPluginConfigUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m any,
) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)

	serviceName := d.Get("service_name").(string)

	oldAdminPassword, newAdminPassword := d.GetChange("admin_password")

	if _, err := client.OpenSearchSecurityPluginHandler.UpdatePassword(
		project,
		serviceName,
		aiven.OpenSearchSecurityPluginUpdatePasswordRequest{
			AdminPassword: oldAdminPassword.(string),
			NewPassword:   newAdminPassword.(string),
		},
	); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName))

	return resourceOpenSearchSecurityPluginConfigRead(ctx, d, m)
}

// resourceOpenSearchSecurityPluginConfigDelete disables the OpenSearch Security Plugin on an existing OpenSearch
// service.
//
// Currently, the Aiven API does not support disabling the OpenSearch Security Plugin, so this resource is effectively
// a no-op. Additionally, we display a warning to the user to indicate that the OpenSearch Security Plugin is not
// disabled, but the resource is still going to be deleted.
func resourceOpenSearchSecurityPluginConfigDelete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return schemautil.StringToDiagWarning("It is not possible to disable the OpenSearch Security Plugin once " +
		"it has been enabled. This resource is going to be deleted, but the OpenSearch Security Plugin will remain " +
		"enabled.")
}
