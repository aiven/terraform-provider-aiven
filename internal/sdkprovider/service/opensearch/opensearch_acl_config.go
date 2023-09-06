// Package opensearch implements the Aiven OpenSearch service.
package opensearch

import (
	"context"

	"github.com/aiven/aiven-go-client"
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
		Description: userconfig.Desc("Enable OpenSearch ACLs. When disabled authenticated service users have unrestricted access.").DefaultValue(true).Build(),
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
		Description:   "The OpenSearch ACL Config resource allows the creation and management of Aiven OpenSearch ACLs.",
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

func resourceOpenSearchACLConfigRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.ElasticsearchACLs.Get(project, serviceName)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("project", project); err != nil {
		return diag.Errorf("error setting ACLs `project` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.Errorf("error setting ACLs `service_name` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("extended_acl", r.ElasticSearchACLConfig.ExtendedAcl); err != nil {
		return diag.Errorf("error setting ACLs `extended_acl` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("enabled", r.ElasticSearchACLConfig.Enabled); err != nil {
		return diag.Errorf("error setting ACLs `enable` for resource %s: %s", d.Id(), err)
	}
	return nil
}

func resourceOpenSearchACLConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	modifier := resourceElasticsearchACLModifierToggleConfigFields(d.Get("enabled").(bool), d.Get("extended_acl").(bool))
	err := resourceOpenSearchACLModifyRemoteConfig(project, serviceName, client, modifier)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName))

	return resourceOpenSearchACLConfigRead(ctx, d, m)
}

func resourceOpenSearchACLConfigDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	modifier := resourceElasticsearchACLModifierToggleConfigFields(false, false)
	err := resourceOpenSearchACLModifyRemoteConfig(project, serviceName, client, modifier)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
