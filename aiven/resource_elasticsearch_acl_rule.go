package aiven

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var aivenElasticsearchACLRuleSchema = map[string]*schema.Schema{
	"project":      commonSchemaProjectReference,
	"service_name": commonSchemaServiceNameReference,
	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 40),
		Description:  complex("The username for the ACL entry.").maxLen(40).forceNew().referenced().build(),
	},
	"index": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 249),
		Description:  complex("The index pattern for the ACL entry.").maxLen(249).forceNew().build(),
	},
	"permission": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice([]string{"deny", "admin", "read", "readwrite", "write"}, false),
		Description:  complex("The permission for the ACL entry.").possibleValues("deny", "admin", "read", "readwrite", "write").build(),
	},
}

func resourceElasticsearchACLRule() *schema.Resource {
	return &schema.Resource{
		Description:   "The Elasticsearch ACL Rule resource models a single ACL Rule for an Aiven Elasticsearch service.",
		CreateContext: resourceElasticsearchACLRuleUpdate,
		ReadContext:   resourceElasticsearchACLRuleRead,
		UpdateContext: resourceElasticsearchACLRuleUpdate,
		DeleteContext: resourceElasticsearchACLRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceElasticsearchACLRuleState,
		},

		Schema: aivenElasticsearchACLRuleSchema,
	}
}

func resourceElasticsearchACLRuleGetPermissionFromACLResponse(cfg aiven.ElasticSearchACLConfig, username, index string) (string, bool) {
	for _, acl := range cfg.ACLs {
		if acl.Username != username {
			continue
		}
		for _, rule := range acl.Rules {
			if rule.Index == index {
				return rule.Permission, true
			}
		}
	}
	return "", false
}

func resourceElasticsearchACLRuleRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, username, index := splitResourceID4(d.Id())
	r, err := client.ElasticsearchACLs.Get(project, serviceName)
	if err != nil {
		return diag.FromErr(resourceReadHandleNotFound(err, d))
	}
	permission, found := resourceElasticsearchACLRuleGetPermissionFromACLResponse(r.ElasticSearchACLConfig, username, index)
	if !found {
		d.SetId("")
		return nil
	}

	if err := d.Set("project", project); err != nil {
		return diag.Errorf("error setting ACL Rules `project` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.Errorf("error setting ACL Rules `service_name` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("username", username); err != nil {
		return diag.Errorf("error setting ACLs Rules `username` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("index", index); err != nil {
		return diag.Errorf("error setting ACLs Rules `index` for resource %s: %s", d.Id(), err)
	}
	if err := d.Set("permission", permission); err != nil {
		return diag.Errorf("error setting ACLs Rules `permission` for resource %s: %s", d.Id(), err)
	}

	return nil
}

func resourceElasticsearchACLRuleState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	di := resourceElasticsearchACLRuleRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get elasticsearch acl rule: %v", di)
	}
	return []*schema.ResourceData{d}, nil
}

func resourceElasticsearchACLRuleMkAivenACL(username, index, permission string) aiven.ElasticSearchACL {
	return aiven.ElasticSearchACL{
		Username: username,
		Rules: []aiven.ElasticsearchACLRule{
			{
				Index:      index,
				Permission: permission,
			},
		},
	}
}

func resourceElasticsearchACLRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	index := d.Get("index").(string)
	permission := d.Get("permission").(string)

	modifier := resourceElasticsearchACLModifierUpdateACLRule(username, index, permission)
	err := resourceElasticsearchACLModifyRemoteConfig(project, serviceName, client, modifier)
	if err != nil {
		return diag.FromErr(resourceReadHandleNotFound(err, d))
	}

	d.SetId(buildResourceID(project, serviceName, username, index))

	return resourceElasticsearchACLRuleRead(ctx, d, m)
}

func resourceElasticsearchACLRuleDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	index := d.Get("index").(string)
	permission := d.Get("permission").(string)

	modifier := resourceElasticsearchACLModifierDeleteACLRule(username, index, permission)
	err := resourceElasticsearchACLModifyRemoteConfig(project, serviceName, client, modifier)
	if err != nil {
		return diag.FromErr(resourceReadHandleNotFound(err, d))
	}
	return nil
}
