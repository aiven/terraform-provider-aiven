package aiven

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenElasticsearchACLConfigSchema = map[string]*schema.Schema{
	"project":      commonSchemaProjectReference,
	"service_name": commonSchemaServiceNameReference,
	"enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: complex("Enable Elasticsearch ACLs. When disabled authenticated service users have unrestricted access").defaultValue(true).build(),
	},
	"extended_acl": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: complex("Index rules can be applied in a limited fashion to the _mget, _msearch and _bulk APIs (and only those) by enabling the ExtendedAcl option for the service. When it is enabled, users can use these APIs as long as all operations only target indexes they have been granted access to").defaultValue(10).build(),
	},
}

func resourceElasticsearchACLConfig() *schema.Resource {
	return &schema.Resource{
		Description:   "The Elasticsearch ACL Config resource allows the configuration of ACL management on an Aiven Elasticsearch service.",
		CreateContext: resourceElasticsearchACLConfigUpdate,
		ReadContext:   resourceElasticsearchACLConfigRead,
		UpdateContext: resourceElasticsearchACLConfigUpdate,
		DeleteContext: resourceElasticsearchACLConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceElasticsearchACLConfigState,
		},

		Schema: aivenElasticsearchACLConfigSchema,
	}
}

func resourceElasticsearchACLConfigRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName := splitResourceID2(d.Id())
	r, err := client.ElasticsearchACLs.Get(project, serviceName)
	if err != nil {
		return diag.FromErr(resourceReadHandleNotFound(err, d))
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

func resourceElasticsearchACLConfigState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	di := resourceElasticsearchACLConfigRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get elasticsearch acl: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}

func resourceElasticsearchACLConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	modifier := resourceElasticsearchACLModifierToggleConfigFields(d.Get("enabled").(bool), d.Get("extended_acl").(bool))
	err := resourceElasticsearchACLModifyRemoteConfig(project, serviceName, client, modifier)
	if err != nil {
		return diag.FromErr(resourceReadHandleNotFound(err, d))
	}

	d.SetId(buildResourceID(project, serviceName))

	return resourceElasticsearchACLConfigRead(ctx, d, m)
}

func resourceElasticsearchACLConfigDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	modifier := resourceElasticsearchACLModifierToggleConfigFields(false, false)
	err := resourceElasticsearchACLModifyRemoteConfig(project, serviceName, client, modifier)
	if err != nil {
		return diag.FromErr(resourceReadHandleNotFound(err, d))
	}

	return nil
}
