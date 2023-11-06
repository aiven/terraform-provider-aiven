package pg

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/typeupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/dist"
)

func aivenPGSchema() map[string]*schema.Schema {
	schemaPG := schemautil.ServiceCommonSchema()
	schemaPG[schemautil.ServiceTypePG] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "PostgreSQL specific server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"replica_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "PostgreSQL replica URI for services with a replica",
					Sensitive:   true,
				},
				"uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "PostgreSQL master connection URI",
					Optional:    true,
					Sensitive:   true,
				},
				"dbname": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Primary PostgreSQL database name",
				},
				"host": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "PostgreSQL master node host IP or name",
				},
				"password": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "PostgreSQL admin user password",
					Sensitive:   true,
				},
				"port": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "PostgreSQL port",
				},
				"sslmode": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "PostgreSQL sslmode setting (currently always \"require\")",
				},
				"user": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "PostgreSQL admin user name",
				},
				"max_connections": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "Connection limit",
				},
			},
		},
	}
	schemaPG[schemautil.ServiceTypePG+"_user_config"] = dist.ServiceTypePg()

	return schemaPG
}

func ResourcePG() *schema.Resource {
	return &schema.Resource{
		Description:   "The PG resource allows the creation and management of Aiven PostgreSQL services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypePG),
		ReadContext:   schemautil.ResourceServiceRead,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypePG),
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
			Create:  schema.DefaultTimeout(20 * time.Minute),
			Update:  schema.DefaultTimeout(20 * time.Minute),
			Delete:  schema.DefaultTimeout(20 * time.Minute),
			Default: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: aivenPGSchema(),
	}
}

func ResourcePGStateUpgrade(
	_ context.Context,
	rawState map[string]interface{},
	_ interface{},
) (map[string]interface{}, error) {
	userConfigSlice, ok := rawState["pg_user_config"].([]interface{})
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
		"backup_hour":               "int",
		"backup_minute":             "int",
		"enable_ipv6":               "bool",
		"pg_read_replica":           "bool",
		"pg_stat_monitor_enable":    "bool",
		"shared_buffers_percentage": "float",
		"static_ips":                "bool",
		"work_mem":                  "int",
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

	pgSlice, ok := userConfig["pg"].([]interface{})
	if ok && len(pgSlice) > 0 {
		pg, ok := pgSlice[0].(map[string]interface{})
		if ok {
			err := typeupgrader.Map(pg, map[string]string{
				"autovacuum_analyze_scale_factor":     "float",
				"autovacuum_analyze_threshold":        "int",
				"autovacuum_freeze_max_age":           "int",
				"autovacuum_max_workers":              "int",
				"autovacuum_naptime":                  "int",
				"autovacuum_vacuum_cost_delay":        "int",
				"autovacuum_vacuum_cost_limit":        "int",
				"autovacuum_vacuum_scale_factor":      "float",
				"autovacuum_vacuum_threshold":         "int",
				"bgwriter_delay":                      "int",
				"bgwriter_flush_after":                "int",
				"bgwriter_lru_maxpages":               "int",
				"bgwriter_lru_multiplier":             "float",
				"deadlock_timeout":                    "int",
				"idle_in_transaction_session_timeout": "int",
				"jit":                                 "bool",
				"log_autovacuum_min_duration":         "int",
				"log_min_duration_statement":          "int",
				"log_temp_files":                      "int",
				"max_files_per_process":               "int",
				"max_locks_per_transaction":           "int",
				"max_logical_replication_workers":     "int",
				"max_parallel_workers":                "int",
				"max_parallel_workers_per_gather":     "int",
				"max_pred_locks_per_transaction":      "int",
				"max_prepared_transactions":           "int",
				"max_replication_slots":               "int",
				"max_slot_wal_keep_size":              "int",
				"max_stack_depth":                     "int",
				"max_standby_archive_delay":           "int",
				"max_standby_streaming_delay":         "int",
				"max_wal_senders":                     "int",
				"max_worker_processes":                "int",
				"pg_partman_bgw__dot__interval":       "int",
				"temp_file_limit":                     "int",
				"track_activity_query_size":           "int",
				"wal_sender_timeout":                  "int",
				"wal_writer_delay":                    "int",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	pgBouncerSlice, ok := userConfig["pgbouncer"].([]interface{})
	if ok && len(pgBouncerSlice) > 0 {
		pgBouncer, ok := pgBouncerSlice[0].(map[string]interface{})
		if ok {
			err := typeupgrader.Map(pgBouncer, map[string]string{
				"autodb_idle_timeout":       "int",
				"autodb_max_db_connections": "int",
				"autodb_pool_size":          "int",
				"min_pool_size":             "int",
				"server_idle_timeout":       "int",
				"server_lifetime":           "int",
				"server_reset_query_always": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	pgLookoutSlice, ok := userConfig["pglookout"].([]interface{})
	if ok && len(pgLookoutSlice) > 0 {
		pgLookout, ok := pgLookoutSlice[0].(map[string]interface{})
		if ok {
			err := typeupgrader.Map(pgLookout, map[string]string{
				"max_failover_replication_time_lag": "int",
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
				"pg":         "bool",
				"pgbouncer":  "bool",
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
				"pg":         "bool",
				"pgbouncer":  "bool",
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
				"pg":         "bool",
				"pgbouncer":  "bool",
				"prometheus": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	timescaleDBSlice, ok := userConfig["timescaledb"].([]interface{})
	if ok && len(timescaleDBSlice) > 0 {
		timescaleDB, ok := timescaleDBSlice[0].(map[string]interface{})
		if ok {
			err := typeupgrader.Map(timescaleDB, map[string]string{
				"max_background_workers": "int",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	return rawState, nil
}
