// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"time"

	"github.com/aiven/terraform-provider-aiven/aiven/internal/service"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func elasticsearchSchema() map[string]*schema.Schema {
	s := serviceCommonSchema()
	s[ServiceTypeElasticsearch] = &schema.Schema{
		Type:        schema.TypeList,
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
	s[ServiceTypeElasticsearch+"_user_config"] = service.GenerateServiceUserConfigurationSchema(ServiceTypeElasticsearch)

	return s
}

func resourceElasticsearch() *schema.Resource {
	return &schema.Resource{
		Description:   "The Elasticsearch resource allows the creation and management of Aiven Elasticsearch services.",
		CreateContext: resourceServiceCreateWrapper(ServiceTypeElasticsearch),
		ReadContext:   resourceServiceRead,
		UpdateContext: resourceServiceUpdate,
		DeleteContext: resourceServiceDelete,
		CustomizeDiff: customdiff.All(
			customdiff.Sequence(
				service.SetServiceTypeIfEmpty(ServiceTypeElasticsearch),
				customdiff.IfValueChange("disk_space",
					service.DiskSpaceShouldNotBeEmpty,
					service.CustomizeDiffCheckDiskSpace),
			),
			customdiff.IfValueChange("service_integrations",
				service.ServiceIntegrationShouldNotBeEmpty,
				service.CustomizeDiffServiceIntegrationAfterCreation),
		),
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
