package sweep_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/provider"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/account"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/cassandra"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/clickhouse"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/connectionpool"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/flink"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/grafana"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/influxdb"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/kafka"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/m3db"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/mysql"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/opensearch"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/organization"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/pg"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/project"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/redis"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/serviceintegration"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/staticip"
	_ "github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/vpc"
	"github.com/aiven/terraform-provider-aiven/internal/sweep"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// knownMissingSweepers returns a list of resources for which we don't have sweepers for a reason.
func knownMissingSweepers() []string {
	// These are resources for which we don't have sweepers.
	// When a correcponding serivce will be deleted, for example Kafka,
	// all the associated resources will be deleted as well, like Kafka ACLs, topics, etc.
	// Therefore, we don't need to have sweepers for them.
	return []string{
		"aiven_pg_user",
		"aiven_clickhouse_database",
		"aiven_influxdb_user",
		"aiven_influxdb_database",
		"aiven_mysql_user",
		"aiven_kafka_acl",
		"aiven_kafka_native_acl",
		"aiven_pg_database",
		"aiven_kafka_user",
		"aiven_kafka_quota",
		"aiven_redis_user",
		"aiven_valkey_user",
		"aiven_opensearch_acl_config",
		"aiven_mysql_database",
		"aiven_m3db_user",
		"aiven_kafka_topic",
		"aiven_kafka_schema",
		"aiven_azure_privatelink_connection_approval",
		"aiven_flink_application_version",
		"aiven_opensearch_acl_rule",
		"aiven_mirrormaker_replication_flow",
		"aiven_project_user",
		"aiven_clickhouse_user",
		"aiven_kafka_schema_registry_acl",
		"aiven_clickhouse_role",
		"aiven_cassandra_user",
		"aiven_flink_application_deployment",
		"aiven_gcp_privatelink_connection_approval",
		"aiven_opensearch_user",
		"aiven_kafka_schema_configuration",
		"aiven_clickhouse_grant",
		"aiven_opensearch_security_plugin_config",
		"aiven_flink_application",
		"aiven_organization_application_user_token",
		"aiven_organization_permission",
	}
}

// TestCheckSweepers checks that we have sweepers for all the resources.
func TestCheckSweepers(t *testing.T) {
	p, err := provider.Provider("test")
	require.NoError(t, err)

	resourceMap := p.ResourcesMap
	allResources := maps.Keys(resourceMap)
	allSweepers := sweep.GetTestSweepersResources()

	var missing []string
	for _, r := range allResources {
		if !slices.Contains(allSweepers, r) && !slices.Contains(knownMissingSweepers(), r) {
			missing = append(missing, r)
		}
	}

	if len(missing) > 0 {
		t.Errorf("missing sweepers for resources: %v", missing)
	}
}
