package provider

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	account2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/account"
	cassandra2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/cassandra"
	clickhouse2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/clickhouse"
	connectionpool2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/connectionpool"
	flink2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/flink"
	grafana2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/grafana"
	influxdb2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/influxdb"
	kafka2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/kafka"
	m3db2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/m3db"
	mysql2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/mysql"
	opensearch2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/opensearch"
	pg2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/pg"
	project2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/project"
	redis2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/redis"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/servicecomponent"
	serviceintegration2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/serviceintegration"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/staticip"
	vpc2 "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/vpc"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	version = "dev"
)

// Provider returns terraform.ResourceProvider.
//
//goland:noinspection GoDeprecation
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
			"aiven_connection_pool":   connectionpool2.DatasourceConnectionPool(),
			"aiven_service_component": servicecomponent.DatasourceServiceComponent(),

			// influxdb
			"aiven_influxdb":          influxdb2.DatasourceInfluxDB(),
			"aiven_influxdb_user":     influxdb2.DatasourceInfluxDBUser(),
			"aiven_influxdb_database": influxdb2.DatasourceInfluxDBDatabase(),

			// grafana
			"aiven_grafana": grafana2.DatasourceGrafana(),

			// mysql
			"aiven_mysql":          mysql2.DatasourceMySQL(),
			"aiven_mysql_user":     mysql2.DatasourceMySQLUser(),
			"aiven_mysql_database": mysql2.DatasourceMySQLDatabase(),

			// redis
			"aiven_redis":      redis2.DatasourceRedis(),
			"aiven_redis_user": redis2.DatasourceRedisUser(),

			// pg
			"aiven_pg":          pg2.DatasourcePG(),
			"aiven_pg_user":     pg2.DatasourcePGUser(),
			"aiven_pg_database": pg2.DatasourcePGDatabase(),

			// cassandra
			"aiven_cassandra":      cassandra2.DatasourceCassandra(),
			"aiven_cassandra_user": cassandra2.DatasourceCassandraUser(),

			// account
			"aiven_account":                account2.DatasourceAccount(),
			"aiven_account_team":           account2.DatasourceAccountTeam(),
			"aiven_account_team_project":   account2.DatasourceAccountTeamProject(),
			"aiven_account_team_member":    account2.DatasourceAccountTeamMember(),
			"aiven_account_authentication": account2.DatasourceAccountAuthentication(),

			// project
			"aiven_project":       project2.DatasourceProject(),
			"aiven_project_user":  project2.DatasourceProjectUser(),
			"aiven_billing_group": project2.DatasourceBillingGroup(),

			// vpc
			"aiven_aws_privatelink":                vpc2.DatasourceAWSPrivatelink(),
			"aiven_aws_vpc_peering_connection":     vpc2.DatasourceAWSVPCPeeringConnection(),
			"aiven_azure_privatelink":              vpc2.DatasourceAzurePrivatelink(),
			"aiven_azure_vpc_peering_connection":   vpc2.DatasourceAzureVPCPeeringConnection(),
			"aiven_gcp_vpc_peering_connection":     vpc2.DatasourceGCPVPCPeeringConnection(),
			"aiven_project_vpc":                    vpc2.DatasourceProjectVPC(),
			"aiven_transit_gateway_vpc_attachment": vpc2.DatasourceTransitGatewayVPCAttachment(),

			// service integrations
			"aiven_service_integration":          serviceintegration2.DatasourceServiceIntegration(),
			"aiven_service_integration_endpoint": serviceintegration2.DatasourceServiceIntegrationEndpoint(),

			// m3db
			"aiven_m3db":         m3db2.DatasourceM3DB(),
			"aiven_m3db_user":    m3db2.DatasourceM3DBUser(),
			"aiven_m3aggregator": m3db2.DatasourceM3Aggregator(),

			// flink
			"aiven_flink":                     flink2.DatasourceFlink(),
			"aiven_flink_application":         flink2.DatasourceFlinkApplication(),
			"aiven_flink_application_version": flink2.DatasourceFlinkApplicationVersion(),

			// opensearch
			"aiven_opensearch":            opensearch2.DatasourceOpensearch(),
			"aiven_opensearch_user":       opensearch2.DatasourceOpensearchUser(),
			"aiven_opensearch_acl_config": opensearch2.DatasourceOpensearchACLConfig(),
			"aiven_opensearch_acl_rule":   opensearch2.DatasourceOpensearchACLRule(),

			// kafka
			"aiven_kafka":                        kafka2.DatasourceKafka(),
			"aiven_kafka_user":                   kafka2.DatasourceKafkaUser(),
			"aiven_kafka_acl":                    kafka2.DatasourceKafkaACL(),
			"aiven_kafka_schema_registry_acl":    kafka2.DatasourceKafkaSchemaRegistryACL(),
			"aiven_kafka_topic":                  kafka2.DatasourceKafkaTopic(),
			"aiven_kafka_schema":                 kafka2.DatasourceKafkaSchema(),
			"aiven_kafka_schema_configuration":   kafka2.DatasourceKafkaSchemaConfiguration(),
			"aiven_kafka_connector":              kafka2.DatasourceKafkaConnector(),
			"aiven_mirrormaker_replication_flow": kafka2.DatasourceMirrorMakerReplicationFlowTopic(),
			"aiven_kafka_connect":                kafka2.DatasourceKafkaConnect(),
			"aiven_kafka_mirrormaker":            kafka2.DatasourceKafkaMirrormaker(),

			// clickhouse
			"aiven_clickhouse":          clickhouse2.DatasourceClickhouse(),
			"aiven_clickhouse_database": clickhouse2.DatasourceClickhouseDatabase(),
			"aiven_clickhouse_user":     clickhouse2.DatasourceClickhouseUser(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"aiven_connection_pool": connectionpool2.ResourceConnectionPool(),
			"aiven_static_ip":       staticip.ResourceStaticIP(),

			// influxdb
			"aiven_influxdb":          influxdb2.ResourceInfluxDB(),
			"aiven_influxdb_user":     influxdb2.ResourceInfluxDBUser(),
			"aiven_influxdb_database": influxdb2.ResourceInfluxDBDatabase(),

			// grafana
			"aiven_grafana": grafana2.ResourceGrafana(),

			// mysql
			"aiven_mysql":          mysql2.ResourceMySQL(),
			"aiven_mysql_user":     mysql2.ResourceMySQLUser(),
			"aiven_mysql_database": mysql2.ResourceMySQLDatabase(),

			// redis
			"aiven_redis":      redis2.ResourceRedis(),
			"aiven_redis_user": redis2.ResourceRedisUser(),

			// pg
			"aiven_pg":          pg2.ResourcePG(),
			"aiven_pg_user":     pg2.ResourcePGUser(),
			"aiven_pg_database": pg2.ResourcePGDatabase(),

			// cassandra
			"aiven_cassandra":      cassandra2.ResourceCassandra(),
			"aiven_cassandra_user": cassandra2.ResourceCassandraUser(),

			// account
			"aiven_account":                account2.ResourceAccount(),
			"aiven_account_team":           account2.ResourceAccountTeam(),
			"aiven_account_team_project":   account2.ResourceAccountTeamProject(),
			"aiven_account_team_member":    account2.ResourceAccountTeamMember(),
			"aiven_account_authentication": account2.ResourceAccountAuthentication(),

			// project
			"aiven_project":       project2.ResourceProject(),
			"aiven_project_user":  project2.ResourceProjectUser(),
			"aiven_billing_group": project2.ResourceBillingGroup(),

			// vpc
			"aiven_aws_privatelink":                       vpc2.ResourceAWSPrivatelink(),
			"aiven_azure_privatelink":                     vpc2.ResourceAzurePrivatelink(),
			"aiven_azure_privatelink_connection_approval": vpc2.ResourceAzurePrivatelinkConnectionApproval(),
			"aiven_aws_vpc_peering_connection":            vpc2.ResourceAWSVPCPeeringConnection(),
			"aiven_azure_vpc_peering_connection":          vpc2.ResourceAzureVPCPeeringConnection(),
			"aiven_gcp_vpc_peering_connection":            vpc2.ResourceGCPVPCPeeringConnection(),
			"aiven_project_vpc":                           vpc2.ResourceProjectVPC(),
			"aiven_transit_gateway_vpc_attachment":        vpc2.ResourceTransitGatewayVPCAttachment(),

			// service integrations
			"aiven_service_integration":          serviceintegration2.ResourceServiceIntegration(),
			"aiven_service_integration_endpoint": serviceintegration2.ResourceServiceIntegrationEndpoint(),

			// m3db
			"aiven_m3db":         m3db2.ResourceM3DB(),
			"aiven_m3db_user":    m3db2.ResourceM3DBUser(),
			"aiven_m3aggregator": m3db2.ResourceM3Aggregator(),

			// flink
			"aiven_flink":                     flink2.ResourceFlink(),
			"aiven_flink_application":         flink2.ResourceFlinkApplication(),
			"aiven_flink_application_version": flink2.ResourceFlinkApplicationVersion(),

			// opensearch
			"aiven_opensearch":            opensearch2.ResourceOpensearch(),
			"aiven_opensearch_user":       opensearch2.ResourceOpensearchUser(),
			"aiven_opensearch_acl_config": opensearch2.ResourceOpensearchACLConfig(),
			"aiven_opensearch_acl_rule":   opensearch2.ResourceOpensearchACLRule(),

			// kafka
			"aiven_kafka":                        kafka2.ResourceKafka(),
			"aiven_kafka_user":                   kafka2.ResourceKafkaUser(),
			"aiven_kafka_acl":                    kafka2.ResourceKafkaACL(),
			"aiven_kafka_schema_registry_acl":    kafka2.ResourceKafkaSchemaRegistryACL(),
			"aiven_kafka_topic":                  kafka2.ResourceKafkaTopic(),
			"aiven_kafka_schema":                 kafka2.ResourceKafkaSchema(),
			"aiven_kafka_schema_configuration":   kafka2.ResourceKafkaSchemaConfiguration(),
			"aiven_kafka_connector":              kafka2.ResourceKafkaConnector(),
			"aiven_mirrormaker_replication_flow": kafka2.ResourceMirrorMakerReplicationFlow(),
			"aiven_kafka_connect":                kafka2.ResourceKafkaConnect(),
			"aiven_kafka_mirrormaker":            kafka2.ResourceKafkaMirrormaker(),

			// clickhouse
			"aiven_clickhouse":          clickhouse2.ResourceClickhouse(),
			"aiven_clickhouse_database": clickhouse2.ResourceClickhouseDatabase(),
			"aiven_clickhouse_user":     clickhouse2.ResourceClickhouseUser(),
			"aiven_clickhouse_role":     clickhouse2.ResourceClickhouseRole(),
			"aiven_clickhouse_grant":    clickhouse2.ResourceClickhouseGrant(),
		},
	}

	p.ConfigureContextFunc = func(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		client, err := common.NewCustomAivenClient(d.Get("api_token").(string), p.TerraformVersion, version)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return client, nil
	}

	return p
}
