package opensearch

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var aivenOpensearchACLRuleSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: schemautil.GetACLUserValidateFunc(),
		Description:  schemautil.Complex("The username for the ACL entry").MaxLen(40).Referenced().ForceNew().Build(),
	},
	"index": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 249),
		Description:  schemautil.Complex("The index pattern for this ACL entry.").MaxLen(249).ForceNew().Build(),
	},
	"permission": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice([]string{"deny", "admin", "read", "readwrite", "write"}, false),
		Description:  schemautil.Complex("The permissions for this ACL entry").PossibleValues("deny", "admin", "read", "readwrite", "write").Build(),
	},
}

func ResourceOpensearchACLRule() *schema.Resource {
	return &schema.Resource{
		Description:   "The Opensearch ACL Rule resource models a single ACL Rule for an Aiven Opensearch service.",
		CreateContext: resourceOpensearchACLRuleUpdate,
		ReadContext:   resourceOpensearchACLRuleRead,
		UpdateContext: resourceOpensearchACLRuleUpdate,
		DeleteContext: resourceOpensearchACLRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: aivenOpensearchACLRuleSchema,
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

func resourceOpensearchACLRuleRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, username, index, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.ElasticsearchACLs.Get(project, serviceName)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}
	permission, found := resourceElasticsearchACLRuleGetPermissionFromACLResponse(r.ElasticSearchACLConfig, username, index)
	if !found {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
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

func resourceOpensearchACLRuleMkAivenACL(username, index, permission string) aiven.ElasticSearchACL {
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

func resourceOpensearchACLRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	index := d.Get("index").(string)
	permission := d.Get("permission").(string)

	modifier := resourceElasticsearchACLModifierUpdateACLRule(username, index, permission)
	err := resourceOpensearchACLModifyRemoteConfig(project, serviceName, client, modifier)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, username, index))

	return resourceOpensearchACLRuleRead(ctx, d, m)
}

func resourceOpensearchACLRuleDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	index := d.Get("index").(string)
	permission := d.Get("permission").(string)

	modifier := resourceElasticsearchACLModifierDeleteACLRule(username, index, permission)
	err := resourceOpensearchACLModifyRemoteConfig(project, serviceName, client, modifier)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}
	return nil
}
