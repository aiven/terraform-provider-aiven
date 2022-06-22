//go:build sweep
// +build sweep

package sweep_test

import (
	"testing"

	_ "github.com/aiven/terraform-provider-aiven/internal/service/account"
	_ "github.com/aiven/terraform-provider-aiven/internal/service/cassandra"
	_ "github.com/aiven/terraform-provider-aiven/internal/service/clickhouse"
	_ "github.com/aiven/terraform-provider-aiven/internal/service/flink"
	_ "github.com/aiven/terraform-provider-aiven/internal/service/grafana"
	_ "github.com/aiven/terraform-provider-aiven/internal/service/influxdb"
	_ "github.com/aiven/terraform-provider-aiven/internal/service/kafka"
	_ "github.com/aiven/terraform-provider-aiven/internal/service/m3db"
	_ "github.com/aiven/terraform-provider-aiven/internal/service/mysql"
	_ "github.com/aiven/terraform-provider-aiven/internal/service/opensearch"
	_ "github.com/aiven/terraform-provider-aiven/internal/service/pg"
	_ "github.com/aiven/terraform-provider-aiven/internal/service/project"
	_ "github.com/aiven/terraform-provider-aiven/internal/service/redis"
	_ "github.com/aiven/terraform-provider-aiven/internal/service/service_integration"
	_ "github.com/aiven/terraform-provider-aiven/internal/service/static_ip"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
