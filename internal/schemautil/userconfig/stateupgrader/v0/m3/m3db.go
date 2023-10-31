package m3

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/typeupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/dist"
)

func aivenM3DBSchema() map[string]*schema.Schema {
	schemaM3 := schemautil.ServiceCommonSchema()
	schemaM3[schemautil.ServiceTypeM3] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "M3 specific server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	schemaM3[schemautil.ServiceTypeM3+"_user_config"] = dist.ServiceTypeM3db()

	return schemaM3
}

func ResourceM3DBResource() *schema.Resource {
	return &schema.Resource{
		Description:   "The M3 DB resource allows the creation and management of Aiven M3 services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeM3),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeM3),
			schemautil.CustomizeDiffDisallowMultipleManyToOneKeys,
			customdiff.IfValueChange("tag",
				schemautil.TagsShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckUniqueTag,
			),
			customdiff.IfValueChange("disk_space",
				schemautil.DiskSpaceShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckDiskSpace,
			),
			customdiff.IfValueChange("additional_disk_space",
				schemautil.DiskSpaceShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckDiskSpace,
			),
			customdiff.IfValueChange("service_integrations",
				schemautil.ServiceIntegrationShouldNotBeEmpty,
				schemautil.CustomizeDiffServiceIntegrationAfterCreation,
			),
			customdiff.Sequence(
				schemautil.CustomizeDiffCheckPlanAndStaticIpsCannotBeModifiedTogether,
				schemautil.CustomizeDiffCheckStaticIPDisassociation,
			),
		),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: aivenM3DBSchema(),
	}
}

func ResourceM3DBStateUpgrade(
	_ context.Context,
	rawState map[string]any,
	_ any,
) (map[string]any, error) {
	userConfigSlice, ok := rawState["m3db_user_config"].([]any)
	if !ok {
		return rawState, nil
	}

	if len(userConfigSlice) == 0 {
		return rawState, nil
	}

	userConfig, ok := userConfigSlice[0].(map[string]any)
	if !ok {
		return rawState, nil
	}

	err := typeupgrader.Map(userConfig, map[string]string{
		"m3coordinator_enable_graphite_carbon_ingest": "bool",
		"static_ips": "bool",
	})
	if err != nil {
		return rawState, err
	}

	limitsSlice, ok := userConfig["limits"].([]any)
	if ok && len(limitsSlice) > 0 {
		limits, ok := limitsSlice[0].(map[string]any)
		if ok {
			err := typeupgrader.Map(limits, map[string]string{
				"max_recently_queried_series_blocks":          "int",
				"max_recently_queried_series_disk_bytes_read": "int",
				"query_docs":               "int",
				"query_require_exhaustive": "bool",
				"query_series":             "int",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	namespacesSlice, ok := userConfig["namespaces"].([]any)
	if ok && len(namespacesSlice) > 0 {
		for _, v := range namespacesSlice {
			namespace, ok := v.(map[string]any)
			if !ok {
				continue
			}

			optionsSlice, ok := namespace["options"].([]any)
			if ok && len(optionsSlice) > 0 {
				options, ok := optionsSlice[0].(map[string]any)
				if ok {
					err := typeupgrader.Map(options, map[string]string{
						"snapshot_enabled":    "bool",
						"writes_to_commitlog": "bool",
					})
					if err != nil {
						return rawState, err
					}
				}
			}
		}
	}

	privateAccessSlice, ok := userConfig["private_access"].([]any)
	if ok && len(privateAccessSlice) > 0 {
		privateAccess, ok := privateAccessSlice[0].(map[string]any)
		if ok {
			err = typeupgrader.Map(privateAccess, map[string]string{
				"m3coordinator": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	publicAccessSlice, ok := userConfig["public_access"].([]any)
	if ok && len(publicAccessSlice) > 0 {
		publicAccess, ok := publicAccessSlice[0].(map[string]any)
		if ok {
			err := typeupgrader.Map(publicAccess, map[string]string{
				"m3coordinator": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	rulesSlice, ok := userConfig["rules"].([]any)
	if ok && len(rulesSlice) > 0 {
		rules, ok := rulesSlice[0].(map[string]any)
		if ok {
			mappingSlice, ok := rules["mapping"].([]any)
			if ok && len(mappingSlice) > 0 {
				for _, v := range mappingSlice {
					mapping, ok := v.(map[string]any)
					if !ok {
						continue
					}

					err := typeupgrader.Map(mapping, map[string]string{
						"drop": "bool",
					})
					if err != nil {
						return rawState, err
					}
				}
			}
		}
	}

	return rawState, nil
}
