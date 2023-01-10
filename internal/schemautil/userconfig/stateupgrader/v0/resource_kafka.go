package v0

import (
	"context"
	"strconv"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/typeupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/dist"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func aivenKafkaSchema() map[string]*schema.Schema {
	aivenKafkaSchema := schemautil.ServiceCommonSchema()
	aivenKafkaSchema["karapace"] = &schema.Schema{
		Type:             schema.TypeBool,
		Optional:         true,
		Description:      "Switch the service to use Karapace for schema registry and REST proxy",
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
	}
	aivenKafkaSchema["default_acl"] = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     true,
		Description: "Create default wildcard Kafka ACL",
	}
	aivenKafkaSchema[schemautil.ServiceTypeKafka] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Kafka server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"access_cert": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Kafka client certificate",
					Optional:    true,
					Sensitive:   true,
				},
				"access_key": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Kafka client certificate key",
					Optional:    true,
					Sensitive:   true,
				},
				"connect_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Kafka Connect URI, if any",
					Optional:    true,
					Sensitive:   true,
				},
				"rest_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Kafka REST URI, if any",
					Optional:    true,
					Sensitive:   true,
				},
				"schema_registry_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "The Schema Registry URI, if any",
					Optional:    true,
					Sensitive:   true,
				},
			},
		},
	}
	aivenKafkaSchema[schemautil.ServiceTypeKafka+"_user_config"] = dist.ServiceTypeKafka()

	return aivenKafkaSchema
}

func ResourceKafkaResourceV0() *schema.Resource {
	return &schema.Resource{
		Description:   "The Kafka resource allows the creation and management of Aiven Kafka services.",
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: aivenKafkaSchema(),
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeKafka),
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
				schemautil.CustomizeDiffCheckStaticIpDisassociation,
			),

			// if a kafka_version is >= 3.0 then this schema field is not applicable
			customdiff.ComputedIf("karapace", func(ctx context.Context, d *schema.ResourceDiff, m interface{}) bool {
				project := d.Get("project").(string)
				serviceName := d.Get("service_name").(string)
				client := m.(*aiven.Client)

				kafka, err := client.Services.Get(project, serviceName)
				if err != nil {
					return false
				}

				if v, ok := kafka.UserConfig["kafka_version"]; ok {
					if version, err := strconv.ParseFloat(v.(string), 64); err == nil {
						if version >= 3 {
							return true
						}
					}
				}

				return false
			}),
		),
	}
}

func ResourceKafkaStateUpgradeV0(
	_ context.Context,
	rawState map[string]interface{},
	_ interface{},
) (map[string]interface{}, error) {
	userConfigSlice, ok := rawState["kafka_user_config"].([]interface{})
	if !ok {
		return rawState, nil
	}

	userConfig, ok := userConfigSlice[0].(map[string]interface{})
	if !ok {
		return rawState, nil
	}

	err := typeupgrader.Map(userConfig, map[string]string{
		"kafka_connect":   "bool",
		"kafka_rest":      "bool",
		"schema_registry": "bool",
		"static_ips":      "bool",
	})
	if err != nil {
		return rawState, err
	}

	kafkaSlice, ok := userConfig["kafka"].([]interface{})
	if ok && len(kafkaSlice) > 0 {
		kafka, ok := kafkaSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(kafka, map[string]string{
				"auto_create_topics_enable":                                  "bool",
				"connections_max_idle_ms":                                    "int",
				"default_replication_factor":                                 "int",
				"group_initial_rebalance_delay_ms":                           "int",
				"group_max_session_timeout_ms":                               "int",
				"group_min_session_timeout_ms":                               "int",
				"log_cleaner_delete_retention_ms":                            "int",
				"log_cleaner_max_compaction_lag_ms":                          "int",
				"log_cleaner_min_cleanable_ratio":                            "float",
				"log_cleaner_min_compaction_lag_ms":                          "int",
				"log_flush_interval_messages":                                "int",
				"log_flush_interval_ms":                                      "int",
				"log_index_interval_bytes":                                   "int",
				"log_index_size_max_bytes":                                   "int",
				"log_message_downconversion_enable":                          "bool",
				"log_message_timestamp_difference_max_ms":                    "int",
				"log_preallocate":                                            "bool",
				"log_retention_bytes":                                        "int",
				"log_retention_hours":                                        "int",
				"log_retention_ms":                                           "int",
				"log_roll_jitter_ms":                                         "int",
				"log_roll_ms":                                                "int",
				"log_segment_bytes":                                          "int",
				"log_segment_delete_delay_ms":                                "int",
				"max_connections_per_ip":                                     "int",
				"max_incremental_fetch_session_cache_slots":                  "int",
				"message_max_bytes":                                          "int",
				"min_insync_replicas":                                        "int",
				"num_partitions":                                             "int",
				"offsets_retention_minutes":                                  "int",
				"producer_purgatory_purge_interval_requests":                 "int",
				"replica_fetch_max_bytes":                                    "int",
				"replica_fetch_response_max_bytes":                           "int",
				"socket_request_max_bytes":                                   "int",
				"transaction_remove_expired_transaction_cleanup_interval_ms": "int",
				"transaction_state_log_segment_bytes":                        "int",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	kafkaAuthenticationMethodsSlice, ok := userConfig["kafka_authentication_methods"].([]interface{})
	if ok && len(kafkaAuthenticationMethodsSlice) > 0 {
		kafkaAuthenticationMethods, ok := kafkaAuthenticationMethodsSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(kafkaAuthenticationMethods, map[string]string{
				"certificate": "bool",
				"sasl":        "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	kafkaConnectConfigSlice, ok := userConfig["kafka_connect_config"].([]interface{})
	if ok && len(kafkaConnectConfigSlice) > 0 {
		kafkaConnectConfig, ok := kafkaConnectConfigSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(kafkaConnectConfig, map[string]string{
				"consumer_fetch_max_bytes":           "int",
				"consumer_max_partition_fetch_bytes": "int",
				"consumer_max_poll_interval_ms":      "int",
				"consumer_max_poll_records":          "int",
				"offset_flush_interval_ms":           "int",
				"offset_flush_timeout_ms":            "int",
				"producer_max_request_size":          "int",
				"session_timeout_ms":                 "int",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	kafkaRestConfigSlice, ok := userConfig["kafka_rest_config"].([]interface{})
	if ok && len(kafkaRestConfigSlice) > 0 {
		kafkaRestConfig, ok := kafkaRestConfigSlice[0].(map[string]interface{})
		if ok {
			err = typeupgrader.Map(kafkaRestConfig, map[string]string{
				"consumer_enable_auto_commit":  "bool",
				"consumer_request_max_bytes":   "int",
				"consumer_request_timeout_ms":  "int",
				"producer_linger_ms":           "int",
				"simpleconsumer_pool_size_max": "int",
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
				"jolokia":         "bool",
				"kafka":           "bool",
				"kafka_connect":   "bool",
				"kafka_rest":      "bool",
				"prometheus":      "bool",
				"schema_registry": "bool",
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
				"kafka":           "bool",
				"kafka_connect":   "bool",
				"kafka_rest":      "bool",
				"prometheus":      "bool",
				"schema_registry": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	schemaRegistryConfigSlice, ok := userConfig["schema_registry_config"].([]interface{})
	if ok && len(schemaRegistryConfigSlice) > 0 {
		schemaRegistryConfig, ok := schemaRegistryConfigSlice[0].(map[string]interface{})
		if ok {
			err := typeupgrader.Map(schemaRegistryConfig, map[string]string{
				"leader_eligibility": "bool",
			})
			if err != nil {
				return rawState, err
			}
		}
	}

	return rawState, nil
}
