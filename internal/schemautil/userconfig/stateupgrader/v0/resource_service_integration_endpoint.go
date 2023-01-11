package v0

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/typeupgrader"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader/v0/dist"
)

var aivenServiceIntegrationEndpointSchema = map[string]*schema.Schema{
	"project": {
		Description: "Project the service integration endpoint belongs to",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"endpoint_name": {
		ForceNew:    true,
		Description: "Name of the service integration endpoint",
		Required:    true,
		Type:        schema.TypeString,
	},
	"endpoint_type": {
		Description: "Type of the service integration endpoint",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"endpoint_config": {
		Description: "Integration endpoint specific backend configuration",
		Computed:    true,
		Type:        schema.TypeMap,
		Elem:        &schema.Schema{Type: schema.TypeString},
	},
	"datadog_user_config":                         dist.IntegrationEndpointTypeDatadog(),
	"prometheus_user_config":                      dist.IntegrationEndpointTypePrometheus(),
	"rsyslog_user_config":                         dist.IntegrationEndpointTypeRsyslog(),
	"external_elasticsearch_logs_user_config":     dist.IntegrationEndpointTypeExternalElasticsearchLogs(),
	"external_opensearch_logs_user_config":        dist.IntegrationEndpointTypeExternalOpensearchLogs(),
	"external_aws_cloudwatch_logs_user_config":    dist.IntegrationEndpointTypeExternalAwsCloudwatchLogs(),
	"external_google_cloud_logging_user_config":   dist.IntegrationEndpointTypeExternalGoogleCloudLogging(),
	"external_kafka_user_config":                  dist.IntegrationEndpointTypeExternalKafka(),
	"jolokia_user_config":                         dist.IntegrationEndpointTypeJolokia(),
	"signalfx_user_config":                        dist.IntegrationEndpointTypeSignalfx(),
	"external_schema_registry_user_config":        dist.IntegrationEndpointTypeExternalSchemaRegistry(),
	"external_aws_cloudwatch_metrics_user_config": dist.IntegrationEndpointTypeExternalAwsCloudwatchMetrics(),
}

func ResourceServiceIntegrationEndpoint() *schema.Resource {
	return &schema.Resource{
		Description: "The Service Integration Endpoint resource allows the creation and management of Aiven Service Integration Endpoints.",
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: aivenServiceIntegrationEndpointSchema,
	}
}

func ResourceServiceIntegrationEndpointStateUpgrade(
	_ context.Context,
	rawState map[string]interface{},
	_ interface{},
) (map[string]interface{}, error) {
	err := serviceIntegrationEndpointDatadogStateUpgrade(rawState)
	if err != nil {
		return rawState, err
	}

	err = rsyslogStateUpgrade(rawState)
	if err != nil {
		return rawState, err
	}

	err = externalElasticsearchLogsStateUpgrade(rawState)
	if err != nil {
		return rawState, err
	}

	err = externalOpensearchLogsStateUpgrade(rawState)
	if err != nil {
		return rawState, err
	}

	return rawState, nil
}

func serviceIntegrationEndpointDatadogStateUpgrade(rawState map[string]interface{}) error {
	userConfigSlice, ok := rawState["datadog_user_config"].([]interface{})
	if !ok {
		return nil
	}

	if len(userConfigSlice) == 0 {
		return nil
	}

	userConfig, ok := userConfigSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	err := typeupgrader.Map(userConfig, map[string]string{
		"disable_consumer_stats":         "bool",
		"kafka_consumer_check_instances": "int",
		"kafka_consumer_stats_timeout":   "int",
		"max_partition_contexts":         "int",
	})
	if err != nil {
		return err
	}

	return nil
}

func rsyslogStateUpgrade(rawState map[string]interface{}) error {
	userConfigSlice, ok := rawState["rsyslog_user_config"].([]interface{})
	if !ok {
		return nil
	}

	if len(userConfigSlice) == 0 {
		return nil
	}

	userConfig, ok := userConfigSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	err := typeupgrader.Map(userConfig, map[string]string{
		"port": "int",
		"tls":  "bool",
	})
	if err != nil {
		return err
	}

	return nil
}

func externalElasticsearchLogsStateUpgrade(rawState map[string]interface{}) error {
	userConfigSlice, ok := rawState["external_elasticsearch_logs_user_config"].([]interface{})
	if !ok {
		return nil
	}

	if len(userConfigSlice) == 0 {
		return nil
	}

	userConfig, ok := userConfigSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	err := typeupgrader.Map(userConfig, map[string]string{
		"index_days_max": "int",
		"timeout":        "float",
	})
	if err != nil {
		return err
	}

	return nil
}

func externalOpensearchLogsStateUpgrade(rawState map[string]interface{}) error {
	userConfigSlice, ok := rawState["external_opensearch_logs_user_config"].([]interface{})
	if !ok {
		return nil
	}

	if len(userConfigSlice) == 0 {
		return nil
	}

	userConfig, ok := userConfigSlice[0].(map[string]interface{})
	if !ok {
		return nil
	}

	err := typeupgrader.Map(userConfig, map[string]string{
		"index_days_max": "int",
		"timeout":        "float",
	})
	if err != nil {
		return err
	}

	return nil
}
