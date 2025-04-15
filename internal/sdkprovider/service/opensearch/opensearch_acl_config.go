// Package opensearch implements the Aiven OpenSearch service.
package opensearch

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenOpenSearchACLConfigSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: userconfig.Desc("Enable OpenSearch ACLs. When disabled, authenticated service users have unrestricted access.").DefaultValue(true).Build(),
	},
	"extended_acl": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: userconfig.Desc("Index rules can be applied in a limited fashion to the _mget, _msearch and _bulk APIs (and only those) by enabling the ExtendedAcl option for the service. When it is enabled, users can use these APIs as long as all operations only target indexes they have been granted access to.").DefaultValue(true).Build(),
	},
}

func ResourceOpenSearchACLConfig() *schema.Resource {
	return &schema.Resource{
		Description: `Enables access control for an Aiven for OpenSearch® service.

By default, service users are granted full access rights. To limit their access, you can enable access control and [create ACLs](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/opensearch_acl_rule)
that define permissions and patterns. Alternatively, you can [enable OpenSearch Security management](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/opensearch_security_plugin_config)
to manage users and permissions with the OpenSearch Security dashboard.
`,
		CreateContext: resourceOpenSearchACLConfigUpdate,
		ReadContext:   resourceOpenSearchACLConfigRead,
		UpdateContext: resourceOpenSearchACLConfigUpdate,
		DeleteContext: resourceOpenSearchACLConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenOpenSearchACLConfigSchema,
	}
}

func resourceOpenSearchACLConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.OpenSearchACLs.Get(ctx, project, serviceName)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("project", project); err != nil {
		return diag.Errorf("error setting ACLs `project` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.Errorf("error setting ACLs `service_name` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("extended_acl", r.OpenSearchACLConfig.ExtendedAcl); err != nil {
		return diag.Errorf("error setting ACLs `extended_acl` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("enabled", r.OpenSearchACLConfig.Enabled); err != nil {
		return diag.Errorf("error setting ACLs `enable` for resource %s: %s", d.Id(), err)
	}
	return nil
}

func resourceOpenSearchACLConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	modifier := resourceOpenSearchACLModifierToggleConfigFields(d.Get("enabled").(bool), d.Get("extended_acl").(bool))
	err := resourceOpenSearchACLModifyRemoteConfig(ctx, project, serviceName, client, modifier)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName))

	return resourceOpenSearchACLConfigRead(ctx, d, m)
}

func resourceOpenSearchACLConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	modifier := resourceOpenSearchACLModifierToggleConfigFields(false, false)
	err := resourceOpenSearchACLModifyRemoteConfig(ctx, project, serviceName, client, modifier)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
