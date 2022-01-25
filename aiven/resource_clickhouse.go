// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"time"

	"github.com/aiven/terraform-provider-aiven/aiven/internal/service"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func clickhouseSchema() map[string]*schema.Schema {
	s := serviceCommonSchema()
	s[ServiceTypeClickhouse] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Clickhouse server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	s[ServiceTypeClickhouse+"_user_config"] = service.GenerateServiceUserConfigurationSchema(ServiceTypeClickhouse)

	return s
}

func resourceClickhouse() *schema.Resource {
	return &schema.Resource{
		Description:   "The Clickhouse resource allows the creation and management of Aiven Clickhouse services.",
		CreateContext: resourceServiceCreateWrapper(ServiceTypeClickhouse),
		ReadContext:   resourceServiceRead,
		UpdateContext: resourceServiceUpdate,
		DeleteContext: resourceServiceDelete,
		CustomizeDiff: customdiff.All(
			customdiff.Sequence(
				service.SetServiceTypeIfEmpty(ServiceTypeClickhouse),
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

		Schema: clickhouseSchema(),
	}
}
