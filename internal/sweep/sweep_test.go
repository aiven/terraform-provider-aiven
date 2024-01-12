//go:build sweep

package sweep_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

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
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
