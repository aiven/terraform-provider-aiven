package opensearch

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/aiven/go-client-codegen/handler/opensearch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenOpenSearchACLRuleSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: schemautil.GetACLUserValidateFunc(),
		Description:  userconfig.Desc("The username for the ACL entry").MaxLen(40).Referenced().ForceNew().Build(),
	},
	"index": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 249),
		Description:  userconfig.Desc("The index pattern for this ACL entry.").MaxLen(249).ForceNew().Build(),
	},
	"permission": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice(opensearch.PermissionTypeChoices(), false),
		Description:  userconfig.Desc("The permissions for this ACL entry").PossibleValuesString(opensearch.PermissionTypeChoices()...).Build(),
	},
}

func ResourceOpenSearchACLRule() *schema.Resource {
	return &schema.Resource{
		Description:   "The OpenSearch ACL Rule resource models a single ACL Rule for an Aiven OpenSearch service.",
		CreateContext: resourceOpenSearchACLRuleUpdate,
		ReadContext:   resourceOpenSearchACLRuleRead,
		UpdateContext: resourceOpenSearchACLRuleUpdate,
		DeleteContext: resourceOpenSearchACLRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenOpenSearchACLRuleSchema,
	}
}

func resourceOpenSearchACLRuleGetPermissionFromACLResponse(cfg aiven.OpenSearchACLConfig, username, index string) (string, bool) {
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

func resourceOpenSearchACLRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, username, index, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.OpenSearchACLs.Get(ctx, project, serviceName)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}
	permission, found := resourceOpenSearchACLRuleGetPermissionFromACLResponse(r.OpenSearchACLConfig, username, index)
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

func resourceOpenSearchACLRuleMkAivenACL(username, index, permission string) aiven.OpenSearchACL {
	return aiven.OpenSearchACL{
		Username: username,
		Rules: []aiven.OpenSearchACLRule{
			{
				Index:      index,
				Permission: permission,
			},
		},
	}
}

func resourceOpenSearchACLRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	index := d.Get("index").(string)
	permission := d.Get("permission").(string)

	modifier := resourceOpenSearchACLModifierUpdateACLRule(ctx, username, index, permission)
	err := resourceOpenSearchACLModifyRemoteConfig(ctx, project, serviceName, client, modifier)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, username, index))

	return resourceOpenSearchACLRuleRead(ctx, d, m)
}

func resourceOpenSearchACLRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	index := d.Get("index").(string)
	permission := d.Get("permission").(string)

	modifier := resourceOpenSearchACLModifierDeleteACLRule(ctx, username, index, permission)
	err := resourceOpenSearchACLModifyRemoteConfig(ctx, project, serviceName, client, modifier)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
