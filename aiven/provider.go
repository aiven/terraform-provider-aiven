// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/internal/cache"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	providerVersion = "dev"
)

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("AIVEN_TOKEN", nil),
				Description: "Aiven Authentication Token",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"aiven_connection_pool":                datasourceConnectionPool(),
			"aiven_database":                       datasourceDatabase(),
			"aiven_kafka_acl":                      datasourceKafkaACL(),
			"aiven_kafka_topic":                    datasourceKafkaTopic(),
			"aiven_kafka_connector":                datasourceKafkaConnector(),
			"aiven_kafka_schema":                   datasourceKafkaSchema(),
			"aiven_kafka_schema_configuration":     datasourceKafkaSchemaConfiguration(),
			"aiven_project":                        datasourceProject(),
			"aiven_project_user":                   datasourceProjectUser(),
			"aiven_project_vpc":                    datasourceProjectVPC(),
			"aiven_vpc_peering_connection":         datasourceVPCPeeringConnection(),
			"aiven_service_integration":            datasourceServiceIntegration(),
			"aiven_service_integration_endpoint":   datasourceServiceIntegrationEndpoint(),
			"aiven_service_user":                   datasourceServiceUser(),
			"aiven_account":                        datasourceAccount(),
			"aiven_account_team":                   datasourceAccountTeam(),
			"aiven_account_team_project":           datasourceAccountTeamProject(),
			"aiven_account_team_member":            datasourceAccountTeamMember(),
			"aiven_mirrormaker_replication_flow":   datasourceMirrorMakerReplicationFlowTopic(),
			"aiven_account_authentication":         datasourceAccountAuthentication(),
			"aiven_kafka":                          datasourceKafka(),
			"aiven_kafka_connect":                  datasourceKafkaConnect(),
			"aiven_kafka_mirrormaker":              datasourceKafkaMirrormaker(),
			"aiven_pg":                             datasourcePG(),
			"aiven_mysql":                          datasourceMySQL(),
			"aiven_cassandra":                      datasourceCassandra(),
			"aiven_elasticsearch":                  datasourceElasticsearch(),
			"aiven_elasticsearch_acl_config":       datasourceElasticsearchACLConfig(),
			"aiven_elasticsearch_acl_rule":         datasourceElasticsearchACLRule(),
			"aiven_grafana":                        datasourceGrafana(),
			"aiven_influxdb":                       datasourceInfluxDB(),
			"aiven_redis":                          datasourceRedis(),
			"aiven_transit_gateway_vpc_attachment": datasourceTransitGatewayVPCAttachment(),
			"aiven_service_component":              datasourceServiceComponent(),
			"aiven_m3db":                           datasourceM3DB(),
			"aiven_m3aggregator":                   datasourceM3Aggregator(),
			"aiven_billing_group":                  datasourceBillingGroup(),
			"aiven_aws_privatelink":                datasourceAWSPrivatelink(),
			"aiven_opensearch":                     datasourceOpensearch(),
			"aiven_opensearch_acl_config":          datasourceOpensearchACLConfig(),
			"aiven_opensearch_acl_rule":            datasourceOpensearchACLRule(),
			"aiven_flink":                          datasourceFlink(),
			"aiven_azure_privatelink":              datasourceAzurePrivatelink(),

			// clickhouse
			"aiven_clickhouse":          datasourceClickhouse(),
			"aiven_clickhouse_database": datasourceClickhouseDatabase(),

			// deprecated
			"aiven_elasticsearch_acl": datasourceElasticsearchACL(),
			"aiven_service":           datasourceService(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"aiven_connection_pool":                resourceConnectionPool(),
			"aiven_database":                       resourceDatabase(),
			"aiven_kafka_acl":                      resourceKafkaACL(),
			"aiven_kafka_topic":                    resourceKafkaTopic(),
			"aiven_kafka_connector":                resourceKafkaConnector(),
			"aiven_kafka_schema":                   resourceKafkaSchema(),
			"aiven_kafka_schema_configuration":     resourceKafkaSchemaConfiguration(),
			"aiven_project":                        resourceProject(),
			"aiven_project_user":                   resourceProjectUser(),
			"aiven_project_vpc":                    resourceProjectVPC(),
			"aiven_vpc_peering_connection":         resourceVPCPeeringConnection(),
			"aiven_service_integration":            resourceServiceIntegration(),
			"aiven_service_integration_endpoint":   resourceServiceIntegrationEndpoint(),
			"aiven_service_user":                   resourceServiceUser(),
			"aiven_account":                        resourceAccount(),
			"aiven_account_team":                   resourceAccountTeam(),
			"aiven_account_team_project":           resourceAccountTeamProject(),
			"aiven_account_team_member":            resourceAccountTeamMember(),
			"aiven_mirrormaker_replication_flow":   resourceMirrorMakerReplicationFlow(),
			"aiven_account_authentication":         resourceAccountAuthentication(),
			"aiven_kafka":                          resourceKafka(),
			"aiven_kafka_connect":                  resourceKafkaConnect(),
			"aiven_kafka_mirrormaker":              resourceKafkaMirrormaker(),
			"aiven_pg":                             resourcePG(),
			"aiven_mysql":                          resourceMySQL(),
			"aiven_cassandra":                      resourceCassandra(),
			"aiven_elasticsearch":                  resourceElasticsearch(),
			"aiven_elasticsearch_acl_config":       resourceElasticsearchACLConfig(),
			"aiven_elasticsearch_acl_rule":         resourceElasticsearchACLRule(),
			"aiven_grafana":                        resourceGrafana(),
			"aiven_influxdb":                       resourceInfluxDB(),
			"aiven_redis":                          resourceRedis(),
			"aiven_transit_gateway_vpc_attachment": resourceTransitGatewayVPCAttachment(),
			"aiven_m3db":                           resourceM3DB(),
			"aiven_m3aggregator":                   resourceM3Aggregator(),
			"aiven_billing_group":                  resourceBillingGroup(),
			"aiven_aws_privatelink":                resourceAWSPrivatelink(),
			"aiven_opensearch":                     resourceOpensearch(),
			"aiven_opensearch_acl_config":          resourceOpensearchACLConfig(),
			"aiven_opensearch_acl_rule":            resourceOpensearchACLRule(),
			"aiven_azure_privatelink":              resourceAzurePrivatelink(),
			"aiven_flink":                          resourceFlink(),
			"aiven_flink_table":                    resourceFlinkTable(),
			"aiven_flink_job":                      resourceFlinkJob(),

			// clickhouse
			"aiven_clickhouse":          resourceClickhouse(),
			"aiven_clickhouse_database": resourceClickhouseDatabase(),

			// deprecated
			"aiven_elasticsearch_acl": resourceElasticsearchACL(),
			"aiven_service":           resourceService(),
		},
	}

	p.ConfigureContextFunc = func(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		_ = cache.NewTopicCache()
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}

		client, err := aiven.NewTokenClient(
			d.Get("api_token").(string),
			fmt.Sprintf("terraform-provider-aiven/%s/%s", terraformVersion, providerVersion))
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return client, nil
	}

	return p
}

func resourceReadHandleNotFound(err error, d *schema.ResourceData) error {
	if err != nil && aiven.IsNotFound(err) {
		d.SetId("")
		return nil
	}
	return err
}
