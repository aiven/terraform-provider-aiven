package influxdb

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/typeupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/dist"
)

func influxDBSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchema()
	s[schemautil.ServiceTypeInfluxDB] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Sensitive:   true,
		Description: "InfluxDB server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"database_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Name of the default InfluxDB database",
				},
			},
		},
	}
	s[schemautil.ServiceTypeInfluxDB+"_user_config"] = dist.ServiceTypeInfluxdb()

	return s
}

func ResourceInfluxDB() *schema.Resource {
	return &schema.Resource{
		Description:   "The InfluxDB resource allows the creation and management of Aiven InfluxDB services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeInfluxDB),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeInfluxDB),
			schemautil.CustomizeDiffDisallowMultipleManyToOneKeys,
			customdiff.IfValueChange("tag",
				schemautil.ShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckUniqueTag,
			),
			customdiff.IfValueChange("disk_space",
				schemautil.ShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckDiskSpace,
			),
			customdiff.IfValueChange("additional_disk_space",
				schemautil.ShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckDiskSpace,
			),
			customdiff.IfValueChange("service_integrations",
				schemautil.ShouldNotBeEmpty,
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

		Schema: influxDBSchema(),
	}
}

func ResourceInfluxDBStateUpgrade(
	_ context.Context,
	rawState map[string]any,
	_ any,
) (map[string]any, error) {
	userConfigSlice, ok := rawState["influxdb_user_config"].([]any)
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
		"static_ips": "bool",
	})
	if err != nil {
		return rawState, err
	}

	influxDBSlice, ok := userConfig["influxdb"].([]any)
	if ok && len(influxDBSlice) > 0 {
		influxDB, ok := influxDBSlice[0].(map[string]any)
		if ok {
			err := typeupgrader.Map(influxDB, map[string]string{
				"log_queries_after":    "int",
				"max_connection_limit": "int",
				"max_row_limit":        "int",
				"max_select_buckets":   "int",
				"max_select_point":     "int",
				"query_timeout":        "int",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	privateAccessSlice, ok := userConfig["private_access"].([]any)
	if ok && len(privateAccessSlice) > 0 {
		privateAccess, ok := privateAccessSlice[0].(map[string]any)
		if ok {
			err = typeupgrader.Map(privateAccess, map[string]string{
				"influxdb": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	privateLinkAccessSlice, ok := userConfig["privatelink_access"].([]any)
	if ok && len(privateLinkAccessSlice) > 0 {
		privateLinkAccess, ok := privateLinkAccessSlice[0].(map[string]any)
		if ok {
			err := typeupgrader.Map(privateLinkAccess, map[string]string{
				"influxdb": "bool",
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
				"influxdb": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	return rawState, nil
}
