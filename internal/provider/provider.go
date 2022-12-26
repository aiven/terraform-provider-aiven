package provider

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/service/account"
	"github.com/aiven/terraform-provider-aiven/internal/service/cassandra"
	"github.com/aiven/terraform-provider-aiven/internal/service/clickhouse"
	"github.com/aiven/terraform-provider-aiven/internal/service/connectionpool"
	"github.com/aiven/terraform-provider-aiven/internal/service/database"
	"github.com/aiven/terraform-provider-aiven/internal/service/flink"
	"github.com/aiven/terraform-provider-aiven/internal/service/grafana"
	"github.com/aiven/terraform-provider-aiven/internal/service/influxdb"
	"github.com/aiven/terraform-provider-aiven/internal/service/kafka"
	"github.com/aiven/terraform-provider-aiven/internal/service/m3db"
	"github.com/aiven/terraform-provider-aiven/internal/service/mysql"
	"github.com/aiven/terraform-provider-aiven/internal/service/opensearch"
	"github.com/aiven/terraform-provider-aiven/internal/service/pg"
	"github.com/aiven/terraform-provider-aiven/internal/service/project"
	"github.com/aiven/terraform-provider-aiven/internal/service/redis"
	"github.com/aiven/terraform-provider-aiven/internal/service/servicecomponent"
	"github.com/aiven/terraform-provider-aiven/internal/service/serviceintegration"
	"github.com/aiven/terraform-provider-aiven/internal/service/serviceuser"
	"github.com/aiven/terraform-provider-aiven/internal/service/staticip"
	"github.com/aiven/terraform-provider-aiven/internal/service/vpc"

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
			"aiven_connection_pool":   connectionpool.DatasourceConnectionPool(),
			"aiven_database":          database.DatasourceDatabase(),       // Deprecated
			"aiven_service_user":      serviceuser.DatasourceServiceUser(), // Deprecated
			"aiven_service_component": servicecomponent.DatasourceServiceComponent(),

			// influxdb
			"aiven_influxdb":          influxdb.DatasourceInfluxDB(),
			"aiven_influxdb_user":     influxdb.DatasourceInfluxDBUser(),
			"aiven_influxdb_database": influxdb.DatasourceInfluxDBDatabase(),

			// grafana
			"aiven_grafana": grafana.DatasourceGrafana(),

			// mysql
			"aiven_mysql":          mysql.DatasourceMySQL(),
			"aiven_mysql_user":     mysql.DatasourceMySQLUser(),
			"aiven_mysql_database": mysql.DatasourceMySQLDatabase(),

			// redis
			"aiven_redis":      redis.DatasourceRedis(),
			"aiven_redis_user": redis.DatasourceRedisUser(),

			// pg
			"aiven_pg":          pg.DatasourcePG(),
			"aiven_pg_user":     pg.DatasourcePGUser(),
			"aiven_pg_database": pg.DatasourcePGDatabase(),

			// cassandra
			"aiven_cassandra":      cassandra.DatasourceCassandra(),
			"aiven_cassandra_user": cassandra.DatasourceCassandraUser(),

			// account
			"aiven_account":                account.DatasourceAccount(),
			"aiven_account_team":           account.DatasourceAccountTeam(),
			"aiven_account_team_project":   account.DatasourceAccountTeamProject(),
			"aiven_account_team_member":    account.DatasourceAccountTeamMember(),
			"aiven_account_authentication": account.DatasourceAccountAuthentication(),

			// project
			"aiven_project":       project.DatasourceProject(),
			"aiven_project_user":  project.DatasourceProjectUser(),
			"aiven_billing_group": project.DatasourceBillingGroup(),

			// vpc
			"aiven_aws_privatelink":                vpc.DatasourceAWSPrivatelink(),
			"aiven_aws_vpc_peering_connection":     vpc.DatasourceAWSVPCPeeringConnection(),
			"aiven_azure_privatelink":              vpc.DatasourceAzurePrivatelink(),
			"aiven_azure_vpc_peering_connection":   vpc.DatasourceAzureVPCPeeringConnection(),
			"aiven_gcp_vpc_peering_connection":     vpc.DatasourceGCPVPCPeeringConnection(),
			"aiven_project_vpc":                    vpc.DatasourceProjectVPC(),
			"aiven_transit_gateway_vpc_attachment": vpc.DatasourceTransitGatewayVPCAttachment(),
			"aiven_vpc_peering_connection":         vpc.DatasourceVPCPeeringConnection(), // Deprecated

			// service integrations
			"aiven_service_integration":          serviceintegration.DatasourceServiceIntegration(),
			"aiven_service_integration_endpoint": serviceintegration.DatasourceServiceIntegrationEndpoint(),

			// m3db
			"aiven_m3db":         m3db.DatasourceM3DB(),
			"aiven_m3db_user":    m3db.DatasourceM3DBUser(),
			"aiven_m3aggregator": m3db.DatasourceM3Aggregator(),

			// flink
			"aiven_flink": flink.DatasourceFlink(),

			// opensearch
			"aiven_opensearch":            opensearch.DatasourceOpensearch(),
			"aiven_opensearch_user":       opensearch.DatasourceOpensearchUser(),
			"aiven_opensearch_acl_config": opensearch.DatasourceOpensearchACLConfig(),
			"aiven_opensearch_acl_rule":   opensearch.DatasourceOpensearchACLRule(),

			// kafka
			"aiven_kafka":                        kafka.DatasourceKafka(),
			"aiven_kafka_user":                   kafka.DatasourceKafkaUser(),
			"aiven_kafka_acl":                    kafka.DatasourceKafkaACL(),
			"aiven_kafka_schema_registry_acl":    kafka.DatasourceKafkaSchemaRegistryACL(),
			"aiven_kafka_topic":                  kafka.DatasourceKafkaTopic(),
			"aiven_kafka_schema":                 kafka.DatasourceKafkaSchema(),
			"aiven_kafka_schema_configuration":   kafka.DatasourceKafkaSchemaConfiguration(),
			"aiven_kafka_connector":              kafka.DatasourceKafkaConnector(),
			"aiven_mirrormaker_replication_flow": kafka.DatasourceMirrorMakerReplicationFlowTopic(),
			"aiven_kafka_connect":                kafka.DatasourceKafkaConnect(),
			"aiven_kafka_mirrormaker":            kafka.DatasourceKafkaMirrormaker(),

			// clickhouse
			"aiven_clickhouse":          clickhouse.DatasourceClickhouse(),
			"aiven_clickhouse_database": clickhouse.DatasourceClickhouseDatabase(),
			"aiven_clickhouse_user":     clickhouse.DatasourceClickhouseUser(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"aiven_connection_pool": connectionpool.ResourceConnectionPool(),
			"aiven_database":        database.ResourceDatabase(),       // Deprecated
			"aiven_service_user":    serviceuser.ResourceServiceUser(), // Deprecated
			"aiven_static_ip":       staticip.ResourceStaticIP(),

			// influxdb
			"aiven_influxdb":          influxdb.ResourceInfluxDB(),
			"aiven_influxdb_user":     influxdb.ResourceInfluxDBUser(),
			"aiven_influxdb_database": influxdb.ResourceInfluxDBDatabase(),

			// grafana
			"aiven_grafana": grafana.ResourceGrafana(),

			// mysql
			"aiven_mysql":          mysql.ResourceMySQL(),
			"aiven_mysql_user":     mysql.ResourceMySQLUser(),
			"aiven_mysql_database": mysql.ResourceMySQLDatabase(),

			// redis
			"aiven_redis":      redis.ResourceRedis(),
			"aiven_redis_user": redis.ResourceRedisUser(),

			// pg
			"aiven_pg":          pg.ResourcePG(),
			"aiven_pg_user":     pg.ResourcePGUser(),
			"aiven_pg_database": pg.ResourcePGDatabase(),

			// cassandra
			"aiven_cassandra":      cassandra.ResourceCassandra(),
			"aiven_cassandra_user": cassandra.ResourceCassandraUser(),

			// account
			"aiven_account":                account.ResourceAccount(),
			"aiven_account_team":           account.ResourceAccountTeam(),
			"aiven_account_team_project":   account.ResourceAccountTeamProject(),
			"aiven_account_team_member":    account.ResourceAccountTeamMember(),
			"aiven_account_authentication": account.ResourceAccountAuthentication(),

			// project
			"aiven_project":       project.ResourceProject(),
			"aiven_project_user":  project.ResourceProjectUser(),
			"aiven_billing_group": project.ResourceBillingGroup(),

			// vpc
			"aiven_aws_privatelink":                       vpc.ResourceAWSPrivatelink(),
			"aiven_azure_privatelink":                     vpc.ResourceAzurePrivatelink(),
			"aiven_azure_privatelink_connection_approval": vpc.ResourceAzurePrivatelinkConnectionApproval(),
			"aiven_aws_vpc_peering_connection":            vpc.ResourceAWSVPCPeeringConnection(),
			"aiven_azure_vpc_peering_connection":          vpc.ResourceAzureVPCPeeringConnection(),
			"aiven_gcp_vpc_peering_connection":            vpc.ResourceGCPVPCPeeringConnection(),
			"aiven_project_vpc":                           vpc.ResourceProjectVPC(),
			"aiven_transit_gateway_vpc_attachment":        vpc.ResourceTransitGatewayVPCAttachment(),
			"aiven_vpc_peering_connection":                vpc.ResourceVPCPeeringConnection(), // Deprecated

			// service integrations
			"aiven_service_integration":          serviceintegration.ResourceServiceIntegration(),
			"aiven_service_integration_endpoint": serviceintegration.ResourceServiceIntegrationEndpoint(),

			// m3db
			"aiven_m3db":         m3db.ResourceM3DB(),
			"aiven_m3db_user":    m3db.ResourceM3DBUser(),
			"aiven_m3aggregator": m3db.ResourceM3Aggregator(),

			// flink
			"aiven_flink":       flink.ResourceFlink(),
			"aiven_flink_table": flink.ResourceFlinkTable(),
			"aiven_flink_job":   flink.ResourceFlinkJob(),

			// opensearch
			"aiven_opensearch":            opensearch.ResourceOpensearch(),
			"aiven_opensearch_user":       opensearch.ResourceOpensearchUser(),
			"aiven_opensearch_acl_config": opensearch.ResourceOpensearchACLConfig(),
			"aiven_opensearch_acl_rule":   opensearch.ResourceOpensearchACLRule(),

			// kafka
			"aiven_kafka":                        kafka.ResourceKafka(),
			"aiven_kafka_user":                   kafka.ResourceKafkaUser(),
			"aiven_kafka_acl":                    kafka.ResourceKafkaACL(),
			"aiven_kafka_schema_registry_acl":    kafka.ResourceKafkaSchemaRegistryACL(),
			"aiven_kafka_topic":                  kafka.ResourceKafkaTopic(),
			"aiven_kafka_schema":                 kafka.ResourceKafkaSchema(),
			"aiven_kafka_schema_configuration":   kafka.ResourceKafkaSchemaConfiguration(),
			"aiven_kafka_connector":              kafka.ResourceKafkaConnector(),
			"aiven_mirrormaker_replication_flow": kafka.ResourceMirrorMakerReplicationFlow(),
			"aiven_kafka_connect":                kafka.ResourceKafkaConnect(),
			"aiven_kafka_mirrormaker":            kafka.ResourceKafkaMirrormaker(),

			// clickhouse
			"aiven_clickhouse":          clickhouse.ResourceClickhouse(),
			"aiven_clickhouse_database": clickhouse.ResourceClickhouseDatabase(),
			"aiven_clickhouse_user":     clickhouse.ResourceClickhouseUser(),
			"aiven_clickhouse_role":     clickhouse.ResourceClickhouseRole(),
			"aiven_clickhouse_grant":    clickhouse.ResourceClickhouseGrant(),
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
