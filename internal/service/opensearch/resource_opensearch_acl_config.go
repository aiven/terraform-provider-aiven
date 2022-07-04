package opensearch

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenOpensearchACLConfigSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: schemautil.Complex("Enable Opensearch ACLs. When disabled authenticated service users have unrestricted access.").DefaultValue(true).Build(),
	},
	"extended_acl": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: schemautil.Complex("Index rules can be applied in a limited fashion to the _mget, _msearch and _bulk APIs (and only those) by enabling the ExtendedAcl option for the service. When it is enabled, users can use these APIs as long as all operations only target indexes they have been granted access to.").DefaultValue(true).Build(),
	},
}

func ResourceOpensearchACLConfig() *schema.Resource {
	return &schema.Resource{
		Description:   "The Opensearch resource allows the creation and management of Aiven Opensearch services.",
		CreateContext: resourceOpensearchACLConfigUpdate,
		ReadContext:   resourceOpensearchACLConfigRead,
		UpdateContext: resourceOpensearchACLConfigUpdate,
		DeleteContext: resourceOpensearchACLConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: aivenOpensearchACLConfigSchema,
	}
}

func resourceOpensearchACLConfigRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName := schemautil.SplitResourceID2(d.Id())
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

func resourceOpensearchACLConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	modifier := resourceElasticsearchACLModifierToggleConfigFields(d.Get("enabled").(bool), d.Get("extended_acl").(bool))
	err := resourceOpensearchACLModifyRemoteConfig(project, serviceName, client, modifier)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName))

	return resourceOpensearchACLConfigRead(ctx, d, m)
}

func resourceOpensearchACLConfigDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	modifier := resourceElasticsearchACLModifierToggleConfigFields(false, false)
	err := resourceOpensearchACLModifyRemoteConfig(project, serviceName, client, modifier)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	return nil
}
