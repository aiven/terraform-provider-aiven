// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func opensearchSchema() map[string]*schema.Schema {
	s := serviceCommonSchema()
	s[ServiceTypeOpensearch] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Opensearch server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"opensearch_dashboards_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "URI for Opensearch dashboard frontend",
					Sensitive:   true,
				},
			},
		},
	}
	s[ServiceTypeOpensearch+"_user_config"] = generateServiceUserConfiguration(ServiceTypeOpensearch)

	return s
}

func resourceOpensearch() *schema.Resource {
	return &schema.Resource{
		Description:   "The Opensearch resource allows the creation and management of Aiven Opensearch services.",
		CreateContext: resourceServiceCreateWrapper(ServiceTypeOpensearch),
		ReadContext:   resourceServiceRead,
		UpdateContext: resourceServiceUpdate,
		DeleteContext: resourceServiceDelete,
		CustomizeDiff: resourceServiceCustomizeDiffWrapper(ServiceTypeOpensearch),
		Importer: &schema.ResourceImporter{
			StateContext: resourceElasticsearchState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: opensearchSchema(),
	}
}

func resourceElasticsearchState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*aiven.Client)

	if len(strings.Split(d.Id(), "/")) != 2 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>", d.Id())
	}

	projectName, serviceName := splitResourceID2(d.Id())
	service, err := client.Services.Get(projectName, serviceName)
	if err != nil {
		return nil, err
	}

	// Hybrid Opensearch service an Aiven service type Elasticsearch but has
	// an opensearch_version user configuration option that indicates that this
	// is a hybrid opensearch service
	if _, ok := service.UserConfig["opensearch_version"]; ok && service.Type == ServiceTypeElasticsearch {
		if err := d.Set("service_type", ServiceTypeOpensearch); err != nil {
			return nil, err
		}
	}

	return resourceServiceState(ctx, d, m)
}
