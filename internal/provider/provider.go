package provider

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/service/account"
	"github.com/aiven/terraform-provider-aiven/internal/service/cassandra"
	"github.com/aiven/terraform-provider-aiven/internal/service/clickhouse"
	"github.com/aiven/terraform-provider-aiven/internal/service/connection_pool"
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
	"github.com/aiven/terraform-provider-aiven/internal/service/service_component"
	"github.com/aiven/terraform-provider-aiven/internal/service/service_integration"
	"github.com/aiven/terraform-provider-aiven/internal/service/service_user"
	"github.com/aiven/terraform-provider-aiven/internal/service/static_ip"
	"github.com/aiven/terraform-provider-aiven/internal/service/vpc"

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
			"aiven_connection_pool":   connection_pool.DatasourceConnectionPool(),
			"aiven_database":          database.DatasourceDatabase(),
			"aiven_service_user":      service_user.DatasourceServiceUser(),
			"aiven_service_component": service_component.DatasourceServiceComponent(),

			// influxdb
			"aiven_influxdb":      influxdb.DatasourceInfluxDB(),
			"aiven_influxdb_user": influxdb.DatasourceInfluxDBUser(),

			// grafana
			"aiven_grafana": grafana.DatasourceGrafana(),

			// mysql
			"aiven_mysql":      mysql.DatasourceMySQL(),
			"aiven_mysql_user": mysql.DatasourceMySQLUser(),

			// redis
			"aiven_redis":      redis.DatasourceRedis(),
			"aiven_redis_user": redis.DatasourceRedisUser(),

			// pgs
			"aiven_pg":      pg.DatasourcePG(),
			"aiven_pg_user": pg.DatasourcePGUser(),

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
			"aiven_project_vpc":                    vpc.DatasourceProjectVPC(),
			"aiven_vpc_peering_connection":         vpc.DatasourceVPCPeeringConnection(),
			"aiven_transit_gateway_vpc_attachment": vpc.DatasourceTransitGatewayVPCAttachment(),
			"aiven_aws_privatelink":                vpc.DatasourceAWSPrivatelink(),
			"aiven_azure_privatelink":              vpc.DatasourceAzurePrivatelink(),

			// service integrations
			"aiven_service_integration":          service_integration.DatasourceServiceIntegration(),
			"aiven_service_integration_endpoint": service_integration.DatasourceServiceIntegrationEndpoint(),

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
			"aiven_connection_pool": connection_pool.ResourceConnectionPool(),
			"aiven_database":        database.ResourceDatabase(),
			"aiven_service_user":    service_user.ResourceServiceUser(),
			"aiven_static_ip":       static_ip.ResourceStaticIP(),

			// influxdb
			"aiven_influxdb":      influxdb.ResourceInfluxDB(),
			"aiven_influxdb_user": influxdb.ResourceInfluxDBUser(),

			// grafana
			"aiven_grafana": grafana.ResourceGrafana(),

			// mysql
			"aiven_mysql":      mysql.ResourceMySQL(),
			"aiven_mysql_user": mysql.ResourceMySQLUser(),

			// redis
			"aiven_redis":      redis.ResourceRedis(),
			"aiven_redis_user": redis.ResourceRedisUser(),

			// pg
			"aiven_pg":      pg.ResourcePG(),
			"aiven_pg_user": pg.ResourcePGUser(),

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
			"aiven_vpc_peering_connection":                vpc.ResourceVPCPeeringConnection(),
			"aiven_aws_privatelink":                       vpc.ResourceAWSPrivatelink(),
			"aiven_azure_privatelink":                     vpc.ResourceAzurePrivatelink(),
			"aiven_azure_privatelink_connection_approval": vpc.ResourceAzurePrivatelinkConnectionApproval(),
			"aiven_project_vpc":                           vpc.ResourceProjectVPC(),
			"aiven_transit_gateway_vpc_attachment":        vpc.ResourceTransitGatewayVPCAttachment(),

			// service integrations
			"aiven_service_integration":          service_integration.ResourceServiceIntegration(),
			"aiven_service_integration_endpoint": service_integration.ResourceServiceIntegrationEndpoint(),

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
