package v0

import (
	"context"
	"regexp"
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/typeupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/dist"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const serviceIntegrationEndpointRegExp = "^[a-zA-Z0-9_-]*\\/{1}[a-zA-Z0-9_-]*$"

var aivenServiceIntegrationSchema = map[string]*schema.Schema{
	"integration_id": {
		Description: "Service Integration Id at aiven",
		Computed:    true,
		Type:        schema.TypeString,
	},
	"destination_endpoint_id": {
		Description: "Destination endpoint for the integration (if any)",
		ForceNew:    true,
		Optional:    true,
		Type:        schema.TypeString,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(serviceIntegrationEndpointRegExp),
			"endpoint id should have the following format: project_name/endpoint_id"),
	},
	"destination_service_name": {
		Description: "Destination service for the integration (if any)",
		ForceNew:    true,
		Optional:    true,
		Type:        schema.TypeString,
	},
	"integration_type": {
		Description: "Type of the service integration",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"project": {
		Description: "Project the integration belongs to",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"source_endpoint_id": {
		Description: "Source endpoint for the integration (if any)",
		ForceNew:    true,
		Optional:    true,
		Type:        schema.TypeString,
		ValidateFunc: validation.StringMatch(regexp.MustCompile(serviceIntegrationEndpointRegExp),
			"endpoint id should have the following format: project_name/endpoint_id"),
	},
	"source_service_name": {
		Description: "Source service for the integration (if any)",
		ForceNew:    true,
		Optional:    true,
		Type:        schema.TypeString,
	},
	"logs_user_config":                  dist.IntegrationTypeLogs(),
	"mirrormaker_user_config":           dist.IntegrationTypeMirrormaker(),
	"kafka_mirrormaker_user_config":     dist.IntegrationTypeKafkaMirrormaker(),
	"kafka_connect_user_config":         dist.IntegrationTypeKafkaConnect(),
	"kafka_logs_user_config":            dist.IntegrationTypeKafkaLogs(),
	"metrics_user_config":               dist.IntegrationTypeMetrics(),
	"datadog_user_config":               dist.IntegrationTypeDatadog(),
	"clickhouse_kafka_user_config":      dist.IntegrationTypeClickhouseKafka(),
	"clickhouse_postgresql_user_config": dist.IntegrationTypeClickhousePostgresql(),
}

func ResourceServiceIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "The Service Integration resource allows the creation and management of Aiven Service Integrations.",
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: aivenServiceIntegrationSchema,
	}
}

func ResourceServiceIntegrationStateUpgrade(
	_ context.Context,
	rawState map[string]interface{},
	_ interface{},
) (map[string]interface{}, error) {
	err := logsStateUpgrade(rawState)
	if err != nil {
		return rawState, err
	}

	err = kafkaMirrormakerStateUpgrade(rawState)
	if err != nil {
		return rawState, err
	}

	err = metricsStateUpgrade(rawState)
	if err != nil {
		return rawState, err
	}

	err = serviceIntegrationDatadogStateUpgrade(rawState)
	if err != nil {
		return rawState, err
	}

	return rawState, nil
}

func logsStateUpgrade(rawState map[string]interface{}) error {
	userConfigSlice, ok := rawState["logs_user_config"].([]interface{})
	if !ok {
		return nil
	}

	userConfig, ok := userConfigSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	err := typeupgrader.Map(userConfig, map[string]string{
		"elasticsearch_index_days_max": "int",
	})
	if err != nil {
		return err
	}

	return nil
}

func kafkaMirrormakerStateUpgrade(rawState map[string]interface{}) error {
	userConfigSlice, ok := rawState["kafka_mirrormaker_user_config"].([]interface{})
	if !ok {
		return nil
	}

	userConfig, ok := userConfigSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	kafkaMirrormakerSlice, ok := userConfig["kafka_mirrormaker"].([]interface{})
	if ok && len(kafkaMirrormakerSlice) > 0 {
		kafkaMirrormaker, ok := kafkaMirrormakerSlice[0].(map[string]interface{})
		if !ok {
			return nil
		}

		err := typeupgrader.Map(kafkaMirrormaker, map[string]string{
			"consumer_fetch_min_bytes":  "int",
			"producer_batch_size":       "int",
			"producer_buffer_memory":    "int",
			"producer_linger_ms":        "int",
			"producer_max_request_size": "int",
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func metricsStateUpgrade(rawState map[string]interface{}) error {
	userConfigSlice, ok := rawState["metrics_user_config"].([]interface{})
	if !ok {
		return nil
	}

	userConfig, ok := userConfigSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	err := typeupgrader.Map(userConfig, map[string]string{
		"retention_days": "int",
	})
	if err != nil {
		return err
	}

	sourceMySQLSlice, ok := userConfig["source_mysql"].([]interface{})
	if ok && len(sourceMySQLSlice) > 0 {
		sourceMySQL, ok := sourceMySQLSlice[0].(map[string]interface{})
		if ok {
			telegrafSlice, ok := sourceMySQL["telegraf"].([]interface{})
			if ok && len(telegrafSlice) > 0 {
				telegraf, ok := telegrafSlice[0].(map[string]interface{})
				if ok {
					err := typeupgrader.Map(telegraf, map[string]string{
						"gather_event_waits":                       "bool",
						"gather_file_events_stats":                 "bool",
						"gather_index_io_waits":                    "bool",
						"gather_info_schema_auto_inc":              "bool",
						"gather_innodb_metrics":                    "bool",
						"gather_perf_events_statements":            "bool",
						"gather_process_list":                      "bool",
						"gather_slave_status":                      "bool",
						"gather_table_io_waits":                    "bool",
						"gather_table_lock_waits":                  "bool",
						"gather_table_schema":                      "bool",
						"perf_events_statements_digest_text_limit": "int",
						"perf_events_statements_limit":             "int",
						"perf_events_statements_time_limit":        "int",
					})
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func serviceIntegrationDatadogStateUpgrade(rawState map[string]interface{}) error {
	userConfigSlice, ok := rawState["datadog_user_config"].([]interface{})
	if !ok {
		return nil
	}

	userConfig, ok := userConfigSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	err := typeupgrader.Map(userConfig, map[string]string{
		"datadog_dbm_enabled": "bool",
		"max_jmx_metrics":     "int",
	})
	if err != nil {
		return err
	}

	return nil
}
