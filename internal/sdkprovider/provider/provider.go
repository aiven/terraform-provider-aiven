package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/account"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/alloydbomni"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/cassandra"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/clickhouse"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/connectionpool"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/dragonfly"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/flink"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/grafana"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/influxdb"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/kafka"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/kafkaschema"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/kafkatopic"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/m3db"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/mysql"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/opensearch"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/organization"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/pg"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/project"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/redis"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/servicecomponent"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/serviceintegration"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/staticip"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/thanos"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/valkey"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/vpc"
)

// Provider returns terraform.ResourceProvider.
func Provider(version string) (*schema.Provider, error) {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				// Description should match the one in internal/plugin/provider.go.
				Description: "Aiven authentication token. Can also be set with the AIVEN_TOKEN environment variable.",
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"aiven_connection_pool":   connectionpool.DatasourceConnectionPool(),
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

			// alloydbomni
			"aiven_alloydbomni":          alloydbomni.DatasourceAlloyDBOmni(),
			"aiven_alloydbomni_user":     alloydbomni.DatasourceAlloyDBOmniUser(),
			"aiven_alloydbomni_database": alloydbomni.DatasourceAlloyDBOmniDatabase(),

			// cassandra
			"aiven_cassandra":      cassandra.DatasourceCassandra(),
			"aiven_cassandra_user": cassandra.DatasourceCassandraUser(),

			// account
			"aiven_account":                account.DatasourceAccount(),
			"aiven_account_team":           account.DatasourceAccountTeam(),
			"aiven_account_team_project":   account.DatasourceAccountTeamProject(),
			"aiven_account_team_member":    account.DatasourceAccountTeamMember(),
			"aiven_account_authentication": account.DatasourceAccountAuthentication(),

			// organization
			"aiven_organizational_unit":           organization.DatasourceOrganizationalUnit(),
			"aiven_organization_user":             organization.DatasourceOrganizationUser(),
			"aiven_organization_user_list":        organization.DatasourceOrganizationUserList(),
			"aiven_organization_user_group":       organization.DatasourceOrganizationUserGroup(),
			"aiven_organization_application_user": organization.DatasourceOrganizationApplicationUser(),

			// project
			"aiven_project":              project.DatasourceProject(),
			"aiven_project_user":         project.DatasourceProjectUser(),
			"aiven_billing_group":        project.DatasourceBillingGroup(),
			"aiven_organization_project": project.DatasourceOrganizationProject(),

			// vpc
			"aiven_aws_privatelink":                vpc.DatasourceAWSPrivatelink(),
			"aiven_aws_vpc_peering_connection":     vpc.DatasourceAWSVPCPeeringConnection(),
			"aiven_azure_privatelink":              vpc.DatasourceAzurePrivatelink(),
			"aiven_azure_vpc_peering_connection":   vpc.DatasourceAzureVPCPeeringConnection(),
			"aiven_gcp_privatelink":                vpc.DatasourceGCPPrivatelink(),
			"aiven_gcp_vpc_peering_connection":     vpc.DatasourceGCPVPCPeeringConnection(),
			"aiven_project_vpc":                    vpc.DatasourceProjectVPC(),
			"aiven_transit_gateway_vpc_attachment": vpc.DatasourceTransitGatewayVPCAttachment(),

			// service integrations
			"aiven_service_integration":          serviceintegration.DatasourceServiceIntegration(),
			"aiven_service_integration_endpoint": serviceintegration.DatasourceServiceIntegrationEndpoint(),

			// m3db
			"aiven_m3db":         m3db.DatasourceM3DB(),
			"aiven_m3db_user":    m3db.DatasourceM3DBUser(),
			"aiven_m3aggregator": m3db.DatasourceM3Aggregator(),

			// flink
			"aiven_flink":                     flink.DatasourceFlink(),
			"aiven_flink_application":         flink.DatasourceFlinkApplication(),
			"aiven_flink_application_version": flink.DatasourceFlinkApplicationVersion(),

			// opensearch
			"aiven_opensearch":                        opensearch.DatasourceOpenSearch(),
			"aiven_opensearch_user":                   opensearch.DatasourceOpenSearchUser(),
			"aiven_opensearch_acl_config":             opensearch.DatasourceOpenSearchACLConfig(),
			"aiven_opensearch_acl_rule":               opensearch.DatasourceOpenSearchACLRule(),
			"aiven_opensearch_security_plugin_config": opensearch.DatasourceOpenSearchSecurityPluginConfig(),

			// kafka
			"aiven_kafka":                        kafka.DatasourceKafka(),
			"aiven_kafka_user":                   kafka.DatasourceKafkaUser(),
			"aiven_kafka_acl":                    kafka.DatasourceKafkaACL(),
			"aiven_kafka_schema_registry_acl":    kafkaschema.DatasourceKafkaSchemaRegistryACL(),
			"aiven_kafka_topic":                  kafkatopic.DatasourceKafkaTopic(),
			"aiven_kafka_schema":                 kafkaschema.DatasourceKafkaSchema(),
			"aiven_kafka_schema_configuration":   kafkaschema.DatasourceKafkaSchemaConfiguration(),
			"aiven_kafka_connector":              kafka.DatasourceKafkaConnector(),
			"aiven_mirrormaker_replication_flow": kafka.DatasourceMirrorMakerReplicationFlowTopic(),
			"aiven_kafka_connect":                kafka.DatasourceKafkaConnect(),
			"aiven_kafka_mirrormaker":            kafka.DatasourceKafkaMirrormaker(),

			// clickhouse
			"aiven_clickhouse":          clickhouse.DatasourceClickhouse(),
			"aiven_clickhouse_database": clickhouse.DatasourceClickhouseDatabase(),
			"aiven_clickhouse_user":     clickhouse.DatasourceClickhouseUser(),

			// dragonfly
			"aiven_dragonfly": dragonfly.DatasourceDragonfly(),

			// thanos
			"aiven_thanos": thanos.DatasourceThanos(),

			// valkey
			"aiven_valkey":      valkey.DatasourceValkey(),
			"aiven_valkey_user": valkey.DatasourceValkeyUser(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"aiven_connection_pool": connectionpool.ResourceConnectionPool(),
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

			// alloydbomni
			"aiven_alloydbomni":          alloydbomni.ResourceAlloyDBOmni(),
			"aiven_alloydbomni_user":     alloydbomni.ResourceAlloyDBOmniUser(),
			"aiven_alloydbomni_database": alloydbomni.ResourceAlloyDBOmniDatabase(),

			// cassandra
			"aiven_cassandra":      cassandra.ResourceCassandra(),
			"aiven_cassandra_user": cassandra.ResourceCassandraUser(),

			// account
			"aiven_account":                account.ResourceAccount(),
			"aiven_account_team":           account.ResourceAccountTeam(),
			"aiven_account_team_project":   account.ResourceAccountTeamProject(),
			"aiven_account_team_member":    account.ResourceAccountTeamMember(),
			"aiven_account_authentication": account.ResourceAccountAuthentication(),

			// organization
			"aiven_organizational_unit":                 organization.ResourceOrganizationalUnit(),
			"aiven_organization_user":                   organization.ResourceOrganizationUser(),
			"aiven_organization_user_group":             organization.ResourceOrganizationUserGroup(),
			"aiven_organization_application_user":       organization.ResourceOrganizationApplicationUser(),
			"aiven_organization_application_user_token": organization.ResourceOrganizationApplicationUserToken(),
			"aiven_organization_permission":             organization.ResourceOrganizationalPermission(),

			// project
			"aiven_project":              project.ResourceProject(),
			"aiven_project_user":         project.ResourceProjectUser(),
			"aiven_billing_group":        project.ResourceBillingGroup(),
			"aiven_organization_project": project.ResourceOrganizationProject(),

			// vpc
			"aiven_aws_privatelink":                       vpc.ResourceAWSPrivatelink(),
			"aiven_aws_vpc_peering_connection":            vpc.ResourceAWSVPCPeeringConnection(),
			"aiven_azure_privatelink":                     vpc.ResourceAzurePrivatelink(),
			"aiven_azure_privatelink_connection_approval": vpc.ResourceAzurePrivatelinkConnectionApproval(),
			"aiven_azure_vpc_peering_connection":          vpc.ResourceAzureVPCPeeringConnection(),
			"aiven_gcp_privatelink":                       vpc.ResourceGCPPrivatelink(),
			"aiven_gcp_privatelink_connection_approval":   vpc.ResourceGCPPrivatelinkConnectionApproval(),
			"aiven_gcp_vpc_peering_connection":            vpc.ResourceGCPVPCPeeringConnection(),
			"aiven_project_vpc":                           vpc.ResourceProjectVPC(),
			"aiven_transit_gateway_vpc_attachment":        vpc.ResourceTransitGatewayVPCAttachment(),

			// service integrations
			"aiven_service_integration":          serviceintegration.ResourceServiceIntegration(),
			"aiven_service_integration_endpoint": serviceintegration.ResourceServiceIntegrationEndpoint(),

			// m3db
			"aiven_m3db":         m3db.ResourceM3DB(),
			"aiven_m3db_user":    m3db.ResourceM3DBUser(),
			"aiven_m3aggregator": m3db.ResourceM3Aggregator(),

			// flink
			"aiven_flink":                            flink.ResourceFlink(),
			"aiven_flink_application":                flink.ResourceFlinkApplication(),
			"aiven_flink_application_version":        flink.ResourceFlinkApplicationVersion(),
			"aiven_flink_application_deployment":     flink.ResourceFlinkApplicationDeployment(),
			"aiven_flink_jar_application":            flink.ResourceFlinkJarApplication(),
			"aiven_flink_jar_application_version":    flink.ResourceFlinkJarApplicationVersion(),
			"aiven_flink_jar_application_deployment": flink.ResourceFlinkJarApplicationDeployment(),

			// opensearch
			"aiven_opensearch":                        opensearch.ResourceOpenSearch(),
			"aiven_opensearch_user":                   opensearch.ResourceOpenSearchUser(),
			"aiven_opensearch_acl_config":             opensearch.ResourceOpenSearchACLConfig(),
			"aiven_opensearch_acl_rule":               opensearch.ResourceOpenSearchACLRule(),
			"aiven_opensearch_security_plugin_config": opensearch.ResourceOpenSearchSecurityPluginConfig(),

			// kafka
			"aiven_kafka":                        kafka.ResourceKafka(),
			"aiven_kafka_user":                   kafka.ResourceKafkaUser(),
			"aiven_kafka_acl":                    kafka.ResourceKafkaACL(),
			"aiven_kafka_native_acl":             kafka.ResourceKafkaNativeACL(),
			"aiven_kafka_schema_registry_acl":    kafkaschema.ResourceKafkaSchemaRegistryACL(),
			"aiven_kafka_topic":                  kafkatopic.ResourceKafkaTopic(),
			"aiven_kafka_schema":                 kafkaschema.ResourceKafkaSchema(),
			"aiven_kafka_schema_configuration":   kafkaschema.ResourceKafkaSchemaConfiguration(),
			"aiven_kafka_connector":              kafka.ResourceKafkaConnector(),
			"aiven_mirrormaker_replication_flow": kafka.ResourceMirrorMakerReplicationFlow(),
			"aiven_kafka_connect":                kafka.ResourceKafkaConnect(),
			"aiven_kafka_mirrormaker":            kafka.ResourceKafkaMirrormaker(),
			"aiven_kafka_quota":                  kafka.ResourceKafkaQuota(),

			// clickhouse
			"aiven_clickhouse":          clickhouse.ResourceClickhouse(),
			"aiven_clickhouse_database": clickhouse.ResourceClickhouseDatabase(),
			"aiven_clickhouse_user":     clickhouse.ResourceClickhouseUser(),
			"aiven_clickhouse_role":     clickhouse.ResourceClickhouseRole(),
			"aiven_clickhouse_grant":    clickhouse.ResourceClickhouseGrant(),

			// dragonfly
			"aiven_dragonfly": dragonfly.ResourceDragonfly(),

			// thanos
			"aiven_thanos": thanos.ResourceThanos(),

			// valkey
			"aiven_valkey":      valkey.ResourceValkey(),
			"aiven_valkey_user": valkey.ResourceValkeyUser(),
		},
	}

	// Adds "beta" warning to the description
	betaResources := []string{
		"aiven_alloydbomni",
		"aiven_alloydbomni_user",
		"aiven_alloydbomni_database",
		"aiven_flink_jar_application",
		"aiven_flink_jar_application_version",
		"aiven_flink_jar_application_deployment",
		"aiven_organization_project",
	}

	betaDataSources := []string{
		"aiven_alloydbomni",
		"aiven_alloydbomni_user",
		"aiven_alloydbomni_database",
		"aiven_organization_user_list",
		"aiven_organization_project",
	}

	missing := append(
		addBeta(p.ResourcesMap, betaResources...),
		addBeta(p.DataSourcesMap, betaDataSources...)...,
	)

	// Deprecates datasources along with their resources
	for k, d := range p.DataSourcesMap {
		if d.DeprecationMessage == "" {
			r, ok := p.ResourcesMap[k]
			if ok && r.DeprecationMessage != "" {
				d.DeprecationMessage = r.DeprecationMessage
			}
		}
	}

	// Adds deprecation callouts to the description, so they are visible in the Terraform registry
	for _, m := range []map[string]*schema.Resource{p.ResourcesMap, p.DataSourcesMap} {
		for _, r := range m {
			if r.DeprecationMessage != "" {
				r.Description += "\n\n" + formatDeprecation(r.DeprecationMessage)
			}
		}
	}

	// Marks sensitive fields recursively
	err := validateSensitive(p.ResourcesMap, false)
	if err != nil {
		return nil, fmt.Errorf("resource map error: %w", err)
	}
	err = validateSensitive(p.DataSourcesMap, false)
	if err != nil {
		return nil, fmt.Errorf("datasource map error: %w", err)
	}

	if missing != nil {
		return nil, fmt.Errorf("not all beta resources/datasources are found: %s", strings.Join(missing, ", "))
	}

	p.ConfigureContextFunc = func(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		token := d.Get("api_token").(string)
		if token == "" {
			token = os.Getenv("AIVEN_TOKEN")
		}

		opts := []common.ClientOpt{
			common.TokenOpt(token),
			common.TFVersionOpt(p.TerraformVersion),
			common.BuildVersionOpt(version),
		}

		client, err := common.NewAivenClient(opts...)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		// fixme: temporary solution, uses a singleton
		err = common.CachedGenAivenClient(opts...)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return client, nil
	}

	return p, nil
}

// addBeta adds resources as beta or removes them
func addBeta(m map[string]*schema.Resource, keys ...string) (missing []string) {
	isBeta := util.IsBeta()
	for _, k := range keys {
		v, ok := m[k]
		if !ok {
			missing = append(missing, k)
			continue
		}

		if isBeta {
			v.Description = userconfig.Desc(v.Description).AvailabilityType(userconfig.Beta).Build()
		} else {
			delete(m, k)
		}
	}
	return missing
}

var errSensitiveField = fmt.Errorf("must mark `Sensitive: true`")

// validateSensitive All attributes of sensitive blocks must be sensitive due to an issue in Terraform
// https://github.com/hashicorp/terraform-plugin-sdk/issues/201
func validateSensitive(m map[string]*schema.Resource, sensitive bool) error {
	for k, v := range m {
		err := validateSensitiveResource(v, sensitive)
		if err != nil {
			return fmt.Errorf("%w: %s: %w", errSensitiveField, k, err)
		}
	}
	return nil
}

func validateSensitiveResource(r *schema.Resource, sensitive bool) error {
	for k, parent := range r.Schema {
		if sensitive && !parent.Sensitive {
			return fmt.Errorf("schema %s", k)
		}

		switch child := parent.Elem.(type) {
		case *schema.Resource:
			err := validateSensitiveResource(child, parent.Sensitive)
			if err != nil {
				return err
			}

		case *schema.Schema:
			if parent.Sensitive && !child.Sensitive {
				return fmt.Errorf("element %s", k)
			}
		}
	}
	return nil
}

var (
	// reCallouts https://developer.hashicorp.com/terraform/registry/providers/docs#callouts
	reCallouts = regexp.MustCompile(`^([~>!-]>)`)
	// reAivenResourceName finds `aiven_*` resources
	reAivenResourceName = regexp.MustCompile(`(aiven_[a-z_0-9]+)`)
)

// formatDeprecation Adds a deprecation callout to the description
func formatDeprecation(s string) string {
	msg := strings.TrimSpace(reAivenResourceName.ReplaceAllString(s, "`$1`"))
	if reCallouts.MatchString(msg) {
		// Doesn't turn the deprecation into a callout if it already is one
		return msg
	}
	return "~> **This resource is deprecated**\n" + msg
}
