package redis

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/typeupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/dist"
)

func redisSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchema()
	s[schemautil.ServiceTypeRedis] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Redis server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	s[schemautil.ServiceTypeRedis+"_user_config"] = dist.ServiceTypeRedis()

	return s
}

func ResourceRedis() *schema.Resource {
	return &schema.Resource{
		Description:   "The Redis resource allows the creation and management of Aiven Redis services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeRedis),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeRedis),
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
				schemautil.CustomizeDiffCheckStaticIPDisassociation,
				schemautil.CustomizeDiffCheckPlanAndStaticIpsCannotBeModifiedTogether,
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

		Schema: redisSchema(),
	}
}

func ResourceRedisStateUpgrade(
	_ context.Context,
	rawState map[string]interface{},
	_ interface{},
) (map[string]interface{}, error) {
	userConfigSlice, ok := rawState["redis_user_config"].([]interface{})
	if !ok {
		return rawState, nil
	}

	if len(userConfigSlice) == 0 {
		return rawState, nil
	}

	userConfig, ok := userConfigSlice[0].(map[string]interface{})
	if !ok {
		return rawState, nil
	}

	err := typeupgrader.Map(userConfig, map[string]string{
		"redis_io_threads":                        "int",
		"redis_lfu_decay_time":                    "int",
		"redis_lfu_log_factor":                    "int",
		"redis_number_of_databases":               "int",
		"redis_pubsub_client_output_buffer_limit": "int",
		"redis_ssl":                               "bool",
		"redis_timeout":                           "int",
		"static_ips":                              "bool",
	})
	if err != nil {
		return rawState, err
	}

	migrationSlice, ok := userConfig["migration"].([]interface{})
	if ok && len(migrationSlice) > 0 {
		migration, ok := migrationSlice[0].(map[string]interface{})
		if ok {
			err := typeupgrader.Map(migration, map[string]string{
				"port": "int",
				"ssl":  "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	privateAccessSlice, ok := userConfig["private_access"].([]interface{})
	if ok && len(privateAccessSlice) > 0 {
		privateAccess, ok := privateAccessSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(privateAccess, map[string]string{
				"prometheus": "bool",
				"redis":      "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	privateLinkAccessSlice, ok := userConfig["privatelink_access"].([]interface{})
	if ok && len(privateLinkAccessSlice) > 0 {
		privateLinkAccess, ok := privateLinkAccessSlice[0].(map[string]interface{})
		if ok {
			err := typeupgrader.Map(privateLinkAccess, map[string]string{
				"prometheus": "bool",
				"redis":      "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	publicAccessSlice, ok := userConfig["public_access"].([]interface{})
	if ok && len(publicAccessSlice) > 0 {
		publicAccess, ok := publicAccessSlice[0].(map[string]interface{})
		if ok {
			err := typeupgrader.Map(publicAccess, map[string]string{
				"prometheus": "bool",
				"redis":      "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	return rawState, nil
}
