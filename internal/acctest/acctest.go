package acctest

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/aiven/aiven-go-client"
<<<<<<< HEAD
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
=======
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
>>>>>>> fd0b89f6 (feat(frameworkprovider): organization resource and data source (#1283))
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
<<<<<<< HEAD
	"github.com/aiven/terraform-provider-aiven/internal/server"
=======
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/provider"
>>>>>>> fd0b89f6 (feat(frameworkprovider): organization resource and data source (#1283))
)

var (
	testAivenClient              *aiven.Client
	testAivenClientOnce          sync.Once
	TestProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"aiven": func() (tfprotov6.ProviderServer, error) {
			return server.NewMuxServer(context.Background(), "test")
		},
	}
)

func GetTestAivenClient() *aiven.Client {
	testAivenClientOnce.Do(func() {
		client, err := common.NewAivenClient()
		if err != nil {
			log.Fatal(err)
		}
		testAivenClient = client
	})
	return testAivenClient
}

const (
	// DefaultResourceNamePrefix is the default prefix used for resource names in acceptance tests.
	DefaultResourceNamePrefix = "test-acc"

	// DefaultRandomSuffixLength is the default length of the random suffix used in acceptance tests.
	DefaultRandomSuffixLength = 10
)

// TestAccPreCheck is a helper function that is called by acceptance tests prior to any test case execution.
// It is used to perform any pre-test setup, such as environment variable validation.
func TestAccPreCheck(t *testing.T) {
<<<<<<< HEAD
	if _, ok := os.LookupEnv("AIVEN_TOKEN"); !ok {
		t.Fatal("AIVEN_TOKEN environment variable must be set for acceptance tests.")
=======
	if v := os.Getenv("AIVEN_TOKEN"); v == "" {
		t.Fatal("AIVEN_TOKEN must be set for acceptance tests")
>>>>>>> fd0b89f6 (feat(frameworkprovider): organization resource and data source (#1283))
	}

	if _, ok := os.LookupEnv("AIVEN_PROJECT_NAME"); !ok {
		t.Log("AIVEN_PROJECT_NAME environment variable is not set. Some acceptance tests will be skipped.")
	}
}

func TestAccCheckAivenServiceResourceDestroy(s *terraform.State) error {
	c := GetTestAivenClient()
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
