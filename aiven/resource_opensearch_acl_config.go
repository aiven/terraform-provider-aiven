package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenOpensearchACLConfigSchema = map[string]*schema.Schema{
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
	"enabled": {
		Type:        schema.TypeBool,
		Description: "Enable Opensearch ACLs. When disabled authenticated service users have unrestricted access",
		Optional:    true,
		Default:     true,
	},
	"extended_acl": {
		Type:        schema.TypeBool,
		Description: "Index rules can be applied in a limited fashion to the _mget, _msearch and _bulk APIs (and only those) by enabling the ExtendedAcl option for the service. When it is enabled, users can use these APIs as long as all operations only target indexes they have been granted access to",
		Optional:    true,
		Default:     true,
	},
}

func resourceOpensearchACLConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceElasticsearchACLConfigUpdate,
		ReadContext:   resourceElasticsearchACLConfigRead,
		UpdateContext: resourceElasticsearchACLConfigUpdate,
		DeleteContext: resourceElasticsearchACLConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceElasticsearchACLConfigState,
		},

		Schema: aivenOpensearchACLConfigSchema,
	}
}
