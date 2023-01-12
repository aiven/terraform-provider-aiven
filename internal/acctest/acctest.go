package acctest

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/aiven/terraform-provider-aiven/internal/provider"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

var (
	TestAccProvider          *schema.Provider
	TestAccProviderFactories map[string]func() (*schema.Provider, error)
)

func init() {
	TestAccProvider = provider.Provider()
	TestAccProviderFactories = map[string]func() (*schema.Provider, error){
		"aiven": func() (*schema.Provider, error) {
			return TestAccProvider, nil
		},
	}
}

func TestAccPreCheck(t *testing.T) {
	if v := os.Getenv("AIVEN_TOKEN"); v == "" {
		t.Log(v)
		t.Fatal("AIVEN_TOKEN must be set for acceptance tests")
	}

	// Provider a project name with enough credits to run acceptance
	// tests or project name with the assigned payment card.
	if v := os.Getenv("AIVEN_PROJECT_NAME"); v == "" {
		log.Print("[WARNING] AIVEN_PROJECT_NAME must be set for some acceptance tests")
	}
}

func TestAccCheckAivenServiceResourceDestroy(s *terraform.State) error {
	c := TestAccProvider.Meta().(*aiven.Client)
	// loop through the resources in state, verifying each service is destroyed
	for n, rs := range s.RootModule().Resources {
		// ignore datasource
		if strings.HasPrefix(n, "data.") {
			continue
		}

		r := func() []string {
			return []string{
				"aiven_influxdb",
				"aiven_grafana",
				"aiven_mysql",
				"aiven_redis",
				"aiven_pg",
				"aiven_cassandra",
				"aiven_m3db",
				"aiven_flink",
				"aiven_opensearch",
				"aiven_kafka",
				"aiven_kafka_connector",
				"aiven_kafka_connect",
				"aiven_clickhouse",
			}
		}
		if sort.SearchStrings(r(), rs.Type) > 0 {
			continue
		}

		projectName, serviceName, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		p, err := c.Services.Get(projectName, serviceName)
		if err != nil {
			if !aiven.IsNotFound(err) {
				return err
			}
		}

		if p != nil {
			return fmt.Errorf("common (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}
