package aiven

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var aivenElasticsearchACLSchema = map[string]*schema.Schema{
	"project":      commonSchemaProjectReference,
	"service_name": commonSchemaServiceNameReference,
	"enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: complex("Enable Elasticsearch ACLs. When disabled authenticated service users have unrestricted access.").defaultValue(true).build(),
	},
	"extended_acl": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: complex("Index rules can be applied in a limited fashion to the _mget, _msearch and _bulk APIs (and only those) by enabling the ExtendedAcl option for the service. When it is enabled, users can use these APIs as long as all operations only target indexes they have been granted access to.").defaultValue(true).build(),
	},
	"acl": {
		Type:        schema.TypeSet,
		Description: "List of Elasticsearch ACLs",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"username": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 40),
					Description:  complex("Username for the ACL entry.").maxLen(40).build(),
				},
				"rule": {
					Type:        schema.TypeSet,
					Description: "Elasticsearch rules.",
					Required:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"index": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 249),
								Description:  complex("Elasticsearch index pattern.").maxLen(249).build(),
							},

							"permission": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice([]string{"deny", "admin", "read", "readwrite", "write"}, false),
								Description:  complex("Elasticsearch permission.").possibleValues("deny", "admin", "read", "readwrite", "write").build(),
							},
						},
					},
				},
			},
		},
	},
}

func resourceElasticsearchACL() *schema.Resource {
	return &schema.Resource{
		Description: `
**This resource is deprecated, please use ` + "`aiven_elasticsearch_acl_config`" + `and ` + "`aiven_elasticsearch_acl_rule`" + `**

The Elasticsearch ACL resource allows the creation and management of ACLs for an Aiven Elasticsearch service.
`,
		DeprecationMessage: "This resource is deprecated, please use `aiven_elasticsearch_acl_config` and `aiven_elasticsearch_acl_rule`",
		CreateContext:      resourceElasticsearchACLUpdate,
		ReadContext:        resourceElasticsearchACLRead,
		UpdateContext:      resourceElasticsearchACLUpdate,
		DeleteContext:      resourceElasticsearchACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceElasticsearchACLState,
		},

		Schema: aivenElasticsearchACLSchema,
	}
}

func resourceElasticsearchACLRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName := splitResourceID2(d.Id())
	r, err := client.ElasticsearchACLs.Get(project, serviceName)
	if err != nil {
		return diag.FromErr(resourceReadHandleNotFound(err, d))
	}

	if err := d.Set("project", project); err != nil {
		return diag.Errorf("error setting Elasticsearch ACLs `project` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.Errorf("error setting Elasticsearch ACLs `service_name` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("extended_acl", r.ElasticSearchACLConfig.ExtendedAcl); err != nil {
		return diag.Errorf("error setting Elasticsearch ACLs `extended_acl` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("enabled", r.ElasticSearchACLConfig.Enabled); err != nil {
		return diag.Errorf("error setting Elasticsearch ACLs `enable` for resource %s: %s", d.Id(), err)
	}

	if err := d.Set("acl", resourceElasticsearchFlattenACLResponse(r)); err != nil {
		return diag.Errorf("error setting Elasticsearch ACLs `acls` array for resource %s: %s", d.Id(), err)
	}

	return nil
}

func resourceElasticsearchFlattenACLResponse(r *aiven.ElasticSearchACLResponse) []map[string]interface{} {
	acls := make([]map[string]interface{}, 0)

	for _, aclS := range r.ElasticSearchACLConfig.ACLs {
		var rules []map[string]interface{}

		for _, ruleS := range aclS.Rules {
			rule := map[string]interface{}{
				"index":      ruleS.Index,
				"permission": ruleS.Permission,
			}

			rules = append(rules, rule)
		}

		acl := map[string]interface{}{
			"username": aclS.Username,
			"rule":     rules,
		}
		acls = append(acls, acl)
	}

	return acls
}

func resourceElasticsearchACLState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	di := resourceElasticsearchACLRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get elasticsearch acl: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}

func resourceElasticsearchACLUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	resourceElasticsearchACLModifierMutex.Lock()
	defer resourceElasticsearchACLModifierMutex.Unlock()

	var config aiven.ElasticSearchACLConfig
	for _, aclD := range d.Get("acl").(*schema.Set).List() {
		aclM := aclD.(map[string]interface{})
		acl := aiven.ElasticSearchACL{Username: aclM["username"].(string)}

		for _, ruleD := range aclM["rule"].(*schema.Set).List() {
			ruleM := ruleD.(map[string]interface{})
			rule := aiven.ElasticsearchACLRule{Permission: ruleM["permission"].(string), Index: ruleM["index"].(string)}
			acl.Rules = append(acl.Rules, rule)
		}

		config.Add(acl)
	}

	config.Enabled = d.Get("enabled").(bool)
	config.ExtendedAcl = d.Get("extended_acl").(bool)

	_, err := client.ElasticsearchACLs.Update(
		project,
		serviceName,
		aiven.ElasticsearchACLRequest{ElasticSearchACLConfig: config})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(buildResourceID(project, serviceName))

	return resourceElasticsearchACLRead(ctx, d, m)
}

func resourceElasticsearchACLDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	resourceElasticsearchACLModifierMutex.Lock()
	defer resourceElasticsearchACLModifierMutex.Unlock()

	_, err := client.ElasticsearchACLs.Update(
		project,
		serviceName,
		aiven.ElasticsearchACLRequest{
			ElasticSearchACLConfig: aiven.ElasticSearchACLConfig{
				ACLs:        []aiven.ElasticSearchACL{},
				Enabled:     false,
				ExtendedAcl: false}})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
