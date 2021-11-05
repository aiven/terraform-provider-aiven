// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenOpensearchACLConfigSchema = map[string]*schema.Schema{
	"project":      commonSchemaProjectReference,
	"service_name": commonSchemaServiceNameReference,
	"enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: complex("Enable Opensearch ACLs. When disabled authenticated service users have unrestricted access.").defaultValue(true).build(),
	},
	"extended_acl": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: complex("Index rules can be applied in a limited fashion to the _mget, _msearch and _bulk APIs (and only those) by enabling the ExtendedAcl option for the service. When it is enabled, users can use these APIs as long as all operations only target indexes they have been granted access to.").defaultValue(true).build(),
	},
}

func resourceOpensearchACLConfig() *schema.Resource {
	return &schema.Resource{
		Description:   "The Opensearch resource allows the creation and management of Aiven Opensearch services.",
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
