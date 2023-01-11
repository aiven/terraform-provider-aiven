package v0

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/typeupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/dist"
)

func aivenMySQLSchema() map[string]*schema.Schema {
	schemaMySQL := schemautil.ServiceCommonSchema()
	schemaMySQL[schemautil.ServiceTypeMySQL] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "MySQL specific server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	schemaMySQL[schemautil.ServiceTypeMySQL+"_user_config"] = dist.ServiceTypeMysql()

	return schemaMySQL
}

func ResourceMySQLResource() *schema.Resource {
	return &schema.Resource{
		Description:   "The MySQL resource allows the creation and management of Aiven MySQL services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeMySQL),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeMySQL),
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

		Schema: aivenMySQLSchema(),
	}
}

func ResourceMySQLStateUpgrade(
	_ context.Context,
	rawState map[string]interface{},
	_ interface{},
) (map[string]interface{}, error) {
	userConfigSlice, ok := rawState["mysql_user_config"].([]interface{})
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
		"backup_hour":             "int",
		"backup_minute":           "int",
		"binlog_retention_period": "int",
		"static_ips":              "bool",
	})
	if err != nil {
		return rawState, err
	}

	migrationSlice, ok := userConfig["migration"].([]interface{})
	if ok && len(migrationSlice) > 0 {
		migration, ok := migrationSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(migration, map[string]string{
				"port": "int",
				"ssl":  "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	mysqlSlice, ok := userConfig["mysql"].([]interface{})
	if ok && len(mysqlSlice) > 0 {
		mysql, ok := mysqlSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(mysql, map[string]string{
				"connect_timeout":                  "int",
				"group_concat_max_len":             "int",
				"information_schema_stats_expiry":  "int",
				"innodb_change_buffer_max_size":    "int",
				"innodb_flush_neighbors":           "int",
				"innodb_ft_min_token_size":         "int",
				"innodb_lock_wait_timeout":         "int",
				"innodb_log_buffer_size":           "int",
				"innodb_online_alter_log_max_size": "int",
				"innodb_print_all_deadlocks":       "bool",
				"innodb_read_io_threads":           "int",
				"innodb_rollback_on_timeout":       "bool",
				"innodb_thread_concurrency":        "int",
				"innodb_write_io_threads":          "int",
				"interactive_timeout":              "int",
				"long_query_time":                  "float",
				"max_allowed_packet":               "int",
				"max_heap_table_size":              "int",
				"net_buffer_length":                "int",
				"net_read_timeout":                 "int",
				"net_write_timeout":                "int",
				"slow_query_log":                   "bool",
				"sort_buffer_size":                 "int",
				"sql_require_primary_key":          "bool",
				"tmp_table_size":                   "int",
				"wait_timeout":                     "int",
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
				"mysql":      "bool",
				"mysqlx":     "bool",
				"prometheus": "bool",
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
				"mysql":      "bool",
				"mysqlx":     "bool",
				"prometheus": "bool",
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
				"mysql":      "bool",
				"mysqlx":     "bool",
				"prometheus": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	return rawState, nil
}
