// Code generated by user config generator. DO NOT EDIT.

package integration_endpoint

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func GetUserConfig(kind string) *schema.Schema {
	switch kind {
	case "datadog":
		return datadogUserConfig()
	case "external_aws_cloudwatch_logs":
		return externalAwsCloudwatchLogsUserConfig()
	case "external_aws_cloudwatch_metrics":
		return externalAwsCloudwatchMetricsUserConfig()
	case "external_elasticsearch_logs":
		return externalElasticsearchLogsUserConfig()
	case "external_google_cloud_bigquery":
		return externalGoogleCloudBigqueryUserConfig()
	case "external_google_cloud_logging":
		return externalGoogleCloudLoggingUserConfig()
	case "external_kafka":
		return externalKafkaUserConfig()
	case "external_opensearch_logs":
		return externalOpensearchLogsUserConfig()
	case "external_postgresql":
		return externalPostgresqlUserConfig()
	case "external_schema_registry":
		return externalSchemaRegistryUserConfig()
	case "jolokia":
		return jolokiaUserConfig()
	case "prometheus":
		return prometheusUserConfig()
	case "rsyslog":
		return rsyslogUserConfig()
	default:
		panic("unknown user config type: " + kind)
	}
}