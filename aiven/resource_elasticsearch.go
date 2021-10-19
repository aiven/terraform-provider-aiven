package aiven

import (
	"strings"
	"time"

	"github.com/aiven/terraform-provider-aiven/aiven/templates"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func elasticsearchSchema() map[string]*schema.Schema {
	s := serviceCommonSchema()
	s[ServiceTypeElasticsearch] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Elasticsearch server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"kibana_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "URI for Kibana frontend",
					Sensitive:   true,
				},
			},
		},
	}
	s[ServiceTypeElasticsearch+"_user_config"] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Optional:    true,
		Description: "Elasticsearch user configurable settings",
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			if strings.Contains(k, "index_patterns") {
				return false
			}
			return emptyObjectDiffSuppressFunc(k, old, new, d)
		},
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeElasticsearch].(map[string]interface{})),
		},
	}

	return s
}

func resourceElasticsearch() *schema.Resource {
	return &schema.Resource{
		Description:   "The Elasticsearch resource allows the creation and management of Aiven Elasticsearch services.",
		CreateContext: resourceServiceCreateWrapper(ServiceTypeElasticsearch),
		ReadContext:   resourceServiceRead,
		UpdateContext: resourceServiceUpdate,
		DeleteContext: resourceServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema:             elasticsearchSchema(),
		DeprecationMessage: "Elasticsearch service is deprecated, please use aiven_opensearch",
	}
}
