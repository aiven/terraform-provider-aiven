// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var aivenOpensearchACLRuleSchema = map[string]*schema.Schema{
	"project":      commonSchemaProjectReference,
	"service_name": commonSchemaServiceNameReference,
	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 40),
		Description:  complex("The username for the ACL entry").maxLen(40).referenced().forceNew().build(),
	},
	"index": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 249),
		Description:  complex("The index pattern for this ACL entry.").maxLen(249).forceNew().build(),
	},
	"permission": {
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice([]string{"deny", "admin", "read", "readwrite", "write"}, false),
		Description:  complex("The permissions for this ACL entry").possibleValues("deny", "admin", "read", "readwrite", "write").build(),
	},
}

func resourceOpensearchACLRule() *schema.Resource {
	return &schema.Resource{
		Description:   "The Opensearch ACL Rule resource models a single ACL Rule for an Aiven Opensearch service.",
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
