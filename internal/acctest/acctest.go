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

	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/errmsg"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/server"
)

const (
	// Environment variables used in tests
	envToken            = "AIVEN_TOKEN"
	envProjectName      = "AIVEN_PROJECT_NAME"
	envOrganizationName = "AIVEN_ORGANIZATION_NAME"
	envBetaFeatures     = "PROVIDER_AIVEN_ENABLE_BETA"
	envUserID           = "AIVEN_ORGANIZATION_USER_ID"
	envAccountID        = "AIVEN_ORGANIZATION_ACCOUNT_ID"
	envAccountName      = "AIVEN_ACCOUNT_NAME"
)

var TestProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"aiven": func() (tfprotov6.ProviderServer, error) {
		return server.NewMuxServer(context.Background(), "test")
	},
}

// getEnvVar returns environment variable value or empty string if not set
func getEnvVar(name string) string {
	val, ok := os.LookupEnv(name)
	if !ok {
		return ""
	}
	return val
}

// RequireEnvVars checks that all specified environment variables are set.
// It skips the test if any variable is missing.
// Returns a map of environment variable names to their values.
func RequireEnvVars(t *testing.T, vars ...string) map[string]string {
	t.Helper()

	result := make(map[string]string, len(vars))
	for _, v := range vars {
		val, ok := os.LookupEnv(v)
		if !ok {
			t.Skipf("environment variable %s not set", v)
		}
		result[v] = val
	}
	return result
}

// SkipIfNotBeta skips the test if beta features are not enabled
func SkipIfNotBeta(t *testing.T) {
	t.Helper()

	if _, ok := os.LookupEnv(envBetaFeatures); !ok {
		t.Skip("This test requires beta features to be enabled. Set PROVIDER_AIVEN_ENABLE_BETA environment variable.")
	}
}

// Token returns the Aiven API token
func Token() string {
	return getEnvVar(envToken)
}

// ProjectName returns the Aiven project name
func ProjectName() string {
	return getEnvVar(envProjectName)
}

// OrganizationName returns the Aiven organization name
func OrganizationName() string {
	return getEnvVar(envOrganizationName)
}

// AccountName returns the Aiven account name
func AccountName() string {
	return getEnvVar(envAccountName)
}

// TestAccPreCheck validates the necessary test API keys exist in the testing environment
func TestAccPreCheck(t *testing.T) {
	t.Helper()

	if err := os.Setenv("TF_ACC", "1"); err != nil {
		t.Fatal("Error setting TF_ACC: ", err)
	}

	// These are required for all tests
	vars := RequireEnvVars(t, envToken, envProjectName)
	token := vars[envToken]
	projectName := vars[envProjectName]

	if err := os.Setenv(envToken, token); err != nil {
		t.Fatal("Error setting AIVEN_TOKEN: ", err)
	}

	if err := os.Setenv(envProjectName, projectName); err != nil {
		t.Fatal("Error setting AIVEN_PROJECT_NAME: ", err)
	}
}

// TestAccCheckAivenServiceResourceDestroy verifies that the given service is destroyed.
func TestAccCheckAivenServiceResourceDestroy(s *terraform.State) error {
	c := GetTestAivenClient()

	ctx := context.Background()

	// loop through the resources in state, verifying each service is destroyed
	for n, rs := range s.RootModule().Resources {
		// ignore datasource
		if strings.HasPrefix(n, "data.") {
			continue
		}

		r := func() []string {
			return []string{
				"aiven_alloydbomni",
				"aiven_influxdb",
				"aiven_grafana",
				"aiven_mysql",
				"aiven_redis",
				"aiven_valkey",
				"aiven_pg",
				"aiven_cassandra",
				"aiven_m3db",
				"aiven_flink",
				"aiven_opensearch",
				"aiven_kafka",
				"aiven_kafka_connector",
				"aiven_kafka_connect",
				"aiven_clickhouse",
				"aiven_dragonfly",
				"aiven_m3aggregator",
				"aiven_thanos",
				"aiven_valkey",
			}
		}
		if sort.SearchStrings(r(), rs.Type) > 0 {
			continue
		}

		projectName, serviceName, err := schemautil.SplitResourceID2(rs.Primary.ID)
		if err != nil {
			return err
		}

		p, err := c.Services.Get(ctx, projectName, serviceName)
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

// ResourceFromState returns a resource state from the given Terraform state.
func ResourceFromState(state *terraform.State, name string) (*terraform.ResourceState, error) {
	rs, ok := state.RootModule().Resources[name]
	if !ok {
		return nil, fmt.Errorf(errmsg.ResourceNotFound, name)
	}

	return rs, nil
}

var (
	testAivenClient     *aiven.Client
	testAivenClientOnce sync.Once
)

// GetTestAivenClient returns a new Aiven client that can be used for acceptance tests.
func GetTestAivenClient() *aiven.Client {
	testAivenClientOnce.Do(func() {
		client, err := common.NewAivenClient()
		if err != nil {
			log.Panicf("test client error: %s", err)
		}
		testAivenClient = client
	})
	return testAivenClient
}

func GetTestGenAivenClient() (avngen.Client, error) {
	client, err := common.NewAivenGenClient()
	if err != nil {
		return nil, fmt.Errorf("test generated client error: %w", err)
	}
	return client, nil
}

const (
	// DefaultResourceNamePrefix is the default prefix used for resource names in acceptance tests.
	DefaultResourceNamePrefix = "test-acc"

	// DefaultRandomSuffixLength is the default length of the random suffix used in acceptance tests.
	DefaultRandomSuffixLength = 6
)

func RandStr() string {
	return acctest.RandStringFromCharSet(DefaultRandomSuffixLength, acctest.CharSetAlphaNum)
}

func RandName(name string) string {
	return fmt.Sprintf("%s-%s-%s", DefaultResourceNamePrefix, name, RandStr())
}
