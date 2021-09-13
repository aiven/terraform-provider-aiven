package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var aivenOpensearchACLRuleSchema = map[string]*schema.Schema{
	"project": {
		Type:        schema.TypeString,
		Description: "Project to link the Opensearch ACLs to",
		Required:    true,
		ForceNew:    true,
	},
	"service_name": {
		Type:        schema.TypeString,
		Description: "Service to link the Opensearch ACLs to",
		Required:    true,
		ForceNew:    true,
	},
	"username": {
		Type:         schema.TypeString,
		Description:  "Username for the ACL entry",
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 40),
	},
	"index": {
		Type:         schema.TypeString,
		Description:  "Elasticsearch index pattern",
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 249),
	},
	"permission": {
		Type:         schema.TypeString,
		Description:  "Elasticsearch permission",
		Required:     true,
		ValidateFunc: validation.StringInSlice([]string{"deny", "admin", "read", "readwrite", "write"}, false),
	},
}

func resourceOpensearchACLRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceElasticsearchACLRuleUpdate,
		ReadContext:   resourceElasticsearchACLRuleRead,
		UpdateContext: resourceElasticsearchACLRuleUpdate,
		DeleteContext: resourceElasticsearchACLRuleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceElasticsearchACLRuleState,
		},

		Schema: aivenOpensearchACLRuleSchema,
	}
}
