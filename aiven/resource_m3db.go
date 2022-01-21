// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"time"

	"github.com/aiven/terraform-provider-aiven/aiven/internal/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func aivenM3DBSchema() map[string]*schema.Schema {
	schemaM3 := serviceCommonSchema()
	schemaM3[ServiceTypeM3] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "M3 specific server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	schemaM3[ServiceTypeM3+"_user_config"] = service.GenerateServiceUserConfigurationSchema(ServiceTypeM3)

	return schemaM3
}
func resourceM3DB() *schema.Resource {
	return &schema.Resource{
		Description:   "The M3 DB resource allows the creation and management of Aiven M3 services.",
		CreateContext: resourceServiceCreateWrapper(ServiceTypeM3),
		ReadContext:   resourceServiceRead,
		UpdateContext: resourceServiceUpdate,
		DeleteContext: resourceServiceDelete,
		CustomizeDiff: customdiff.All(
			customdiff.Sequence(
				service.SetServiceTypeIfEmpty(ServiceTypeM3),
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

		Schema: aivenM3DBSchema(),
	}
}
