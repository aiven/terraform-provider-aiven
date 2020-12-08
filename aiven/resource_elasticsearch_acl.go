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
	"project": {
		Type:        schema.TypeString,
		Description: "Project to link the Elasticsearch ACLs to",
		Required:    true,
		ForceNew:    true,
	},
	"service_name": {
		Type:        schema.TypeString,
		Description: "Service to link the Elasticsearch ACLs to",
		Required:    true,
		ForceNew:    true,
	},
	"enabled": {
		Type:        schema.TypeBool,
		Description: "Enable Elasticsearch ACLs. When disabled authenticated service users have unrestricted access",
		Optional:    true,
		Default:     true,
	},
	"extended_acl": {
		Type:        schema.TypeBool,
		Description: "Index rules can be applied in a limited fashion to the _mget, _msearch and _bulk APIs (and only those) by enabling the ExtendedAcl option for the service. When it is enabled, users can use these APIs as long as all operations only target indexes they have been granted access to",
		Optional:    true,
		Default:     true,
	},
	"acl": {
		Type:        schema.TypeSet,
		Description: "List of Elasticsearch ACLs",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"username": {
					Type:         schema.TypeString,
					Description:  "Username for the ACL entry",
					Required:     true,
					ValidateFunc: validation.StringLenBetween(1, 40),
				},
				"rule": {
					Type:        schema.TypeSet,
					Description: "Elasticsearch rules",
					Required:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"index": {
								Type:         schema.TypeString,
								Description:  "Elasticsearch index pattern",
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 249),
							},

							"permission": {
								Type:         schema.TypeString,
								Description:  "Elasticsearch permission",
								Required:     true,
								ValidateFunc: validation.StringInSlice([]string{"deny", "admin", "read", "readwrite", "write"}, false),
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
		CreateContext: resourceElasticsearchACLUpdate,
		ReadContext:   resourceElasticsearchACLRead,
		UpdateContext: resourceElasticsearchACLUpdate,
		DeleteContext: resourceElasticsearchACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceElasticsearchACLState,
		},

		Schema: aivenElasticsearchACLSchema,
	}
}

func flattenElasticsearchACL(r *aiven.ElasticSearchACLResponse) []map[string]interface{} {
	var acls []map[string]interface{}

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

func resourceElasticsearchACLRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName := splitResourceID2(d.Id())
	r, err := client.ElasticsearchACLs.Get(project, serviceName)
	if err != nil {
		return diag.FromErr(err)
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

	acls := flattenElasticsearchACL(r)

	if err := d.Set("acl", acls); err != nil {
		return diag.Errorf("error setting Elasticsearch ACLs `acls` array for resource %s: %s", d.Id(), err)
	}

	return nil
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
